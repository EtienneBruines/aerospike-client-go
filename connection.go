// Copyright 2013-2014 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"net"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"
)

// Connection represents a connection with a timeout
type Connection struct {
	// timeout
	timeout time.Duration

	// connection object
	conn net.Conn
}

func errToTimeoutErr(err error) error {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return NewAerospikeError(TIMEOUT, err.Error())
	}
	return err
}

// NewConnection creates a connection on the network and returns the pointer
// A minimum timeout of 2 seconds will always be applied.
// If the connection is not established in the specified timeout,
// an error will be returned
func NewConnection(address string, timeout time.Duration) (*Connection, error) {
	newConn := &Connection{}

	// connection timeout should be finite
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		Logger.Error("Connection to address `" + address + "` failed to establish with error: " + err.Error())
		return nil, errToTimeoutErr(err)
	}
	newConn.conn = conn

	// set timeout at the last possible moment
	if err := newConn.SetTimeout(timeout); err != nil {
		return nil, err
	}
	return newConn, nil
}

// Writes the slice to the connection buffer.
func (ctn *Connection) Write(buf []byte) (total int, err error) {
	// make sure all bytes are written
	// Don't worry about the loop, timeout has been set elsewhere
	length := len(buf)
	var r int
	for total < length {
		if r, err = ctn.conn.Write(buf[total:]); err != nil {
			break
		}
		total += r
	}

	if err == nil {
		return total, nil
	}
	return total, errToTimeoutErr(err)
}

// Reads from connection buffer to the slice
func (ctn *Connection) Read(buf []byte, length int) (total int, err error) {
	// if all bytes are not read, retry until successful
	// Don't worry about the loop; we've already set the timeout elsewhere
	var r int
	for total < length {
		if r, err = ctn.conn.Read(buf[total:length]); err != nil {
			break
		}
		total += r
	}

	if err == nil && total == length {
		return total, nil
	} else if err != nil {
		return total, errToTimeoutErr(err)
	} else {
		return total, NewAerospikeError(SERVER_ERROR)
	}
}

// Returns true if the connection is not closed
func (ctn *Connection) IsConnected() bool {
	return ctn.conn != nil
}

// sets connection timeout
func (ctn *Connection) SetTimeout(timeout time.Duration) error {
	ctn.timeout = timeout

	// important: remove deadline when not needed; connections are pooled
	if ctn.conn != nil {
		var deadline time.Time
		if timeout > 0 {
			deadline = time.Now().Add(timeout)
		}
		if err := ctn.conn.SetDeadline(deadline); err != nil {
			return err
		}
	}

	return nil
}

// Closes the connection
func (ctn *Connection) Close() {
	if ctn != nil && ctn.conn != nil {
		if err := ctn.conn.Close(); err != nil {
			Logger.Warn(err.Error())
		}
		ctn.conn = nil
	}
}

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
	"strings"
)

// Task used to poll for UDF registration completion.
type RegisterTask struct {
	BaseTask

	packageName string
}

// Initialize task with fields needed to query server nodes.
func NewRegisterTask(cluster *Cluster, packageName string) *RegisterTask {
	return &RegisterTask{
		BaseTask:    *NewTask(cluster, false),
		packageName: packageName,
	}
}

// Query all nodes for task completion status.
func (tskr *RegisterTask) IsDone() (bool, error) {
	command := "udf-list"
	nodes := tskr.cluster.GetNodes()
	done := false

	for _, node := range nodes {
		responseMap, err := RequestNodeInfo(node, command)
		if err != nil {
			return false, err
		}

		for _, response := range responseMap {
			find := "filename=" + tskr.packageName
			index := strings.Index(response, find)

			if index < 0 {
				return false, nil
			}
			done = true
		}
	}
	return done, nil
}

func (tskr *RegisterTask) OnComplete() chan error {
	return tskr.onComplete(tskr)
}

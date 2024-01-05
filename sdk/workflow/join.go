//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package workflow

import (
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
)

type JoinTask struct {
	WorkflowTaskEx
	joinOn []string
}

func NewJoinTask(taskRefName string, joinOn ...string) *JoinTask {
	return &JoinTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          JOIN,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		joinOn: joinOn,
	}
}

func (task *JoinTask) toWorkflowTask() []model.WorkflowTask {
	workflowTasks := task.WorkflowTaskEx.toWorkflowTask()
	workflowTasks[0].JoinOn = task.joinOn
	return workflowTasks
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *JoinTask) Optional(optional bool) *JoinTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Description of the task
func (task *JoinTask) Description(description string) *JoinTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

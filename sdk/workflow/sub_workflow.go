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

type SubWorkflowTask struct {
	WorkflowTaskEx
	workflowName    string
	version         int32
	taskToDomainMap map[string]string
	workflow        *WorkflowDefEx
}

func NewSubWorkflowTask(taskRefName string, workflowName string, version int32) *SubWorkflowTask {
	return &SubWorkflowTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          SUB_WORKFLOW,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		workflowName: workflowName,
		version:      version,
	}
}

func NewSubWorkflowInlineTask(taskRefName string, workflow *WorkflowDefEx) *SubWorkflowTask {
	return &SubWorkflowTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          SUB_WORKFLOW,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		workflow: workflow,
	}
}

func (task *SubWorkflowTask) TaskToDomain(taskToDomainMap map[string]string) *SubWorkflowTask {
	task.taskToDomainMap = taskToDomainMap
	return task
}
func (task *SubWorkflowTask) toWorkflowTask() []model.WorkflowTask {
	workflowTasks := task.WorkflowTaskEx.toWorkflowTask()
	if task.workflow != nil {
		workflowTasks[0].SubWorkflowParam = &model.SubWorkflowParams{
			Name:               task.workflow.name,
			TaskToDomain:       task.taskToDomainMap,
			WorkflowDefinition: task.workflow.ToWorkflowDef(),
		}
	} else {
		workflowTasks[0].SubWorkflowParam = &model.SubWorkflowParams{
			Name:               task.workflowName,
			Version:            task.version,
			TaskToDomain:       task.taskToDomainMap,
			WorkflowDefinition: nil,
		}
	}
	return workflowTasks
}

// Description of the task
func (task *SubWorkflowTask) Description(description string) *SubWorkflowTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *SubWorkflowTask) Optional(optional bool) *SubWorkflowTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *SubWorkflowTask) Input(key string, value interface{}) *SubWorkflowTask {
	task.WorkflowTaskEx.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *SubWorkflowTask) InputMap(inputMap map[string]interface{}) *SubWorkflowTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

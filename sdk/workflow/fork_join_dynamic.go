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
	"swiftconductor.com/swift-conductor-client/sdk/model"
)

type DynamicForkTask struct {
	WorkflowTaskEx
	preForkTask *WorkflowTaskInterface
	join        JoinTask
}

const (
	forkedTasks       = "forkedTasks"
	forkedTasksInputs = "forkedTasksInputs"
)

func NewDynamicForkTask(taskRefName string, forkPrepareTask WorkflowTaskInterface) *DynamicForkTask {
	return &DynamicForkTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          FORK_JOIN_DYNAMIC,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		preForkTask: &forkPrepareTask,
	}
}

func NewDynamicForkTaskWithoutPrepareTask(taskRefName string) *DynamicForkTask {
	return &DynamicForkTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          FORK_JOIN_DYNAMIC,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		preForkTask: nil,
	}
}

func NewDynamicForkWithJoinTask(taskRefName string, forkPrepareTask WorkflowTaskInterface, join JoinTask) *DynamicForkTask {
	return &DynamicForkTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          FORK_JOIN_DYNAMIC,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		preForkTask: &forkPrepareTask,
		join:        join,
	}
}

func (task *DynamicForkTask) toWorkflowTask() []model.WorkflowTask {
	forkWorkflowTask := task.WorkflowTaskEx.toWorkflowTask()[0]
	forkWorkflowTask.DynamicForkTasksParam = forkedTasks
	forkWorkflowTask.DynamicForkTasksInputParamName = forkedTasksInputs
	if task.preForkTask != nil {
		forkWorkflowTask.InputParameters[forkedTasks] = (*task.preForkTask).OutputRef(forkedTasks)
		forkWorkflowTask.InputParameters[forkedTasksInputs] = (*task.preForkTask).OutputRef(forkedTasksInputs)
		tasks := (*task.preForkTask).toWorkflowTask()
		tasks = append(tasks, forkWorkflowTask, task.getJoinTask())
		return tasks
	}
	return []model.WorkflowTask{
		forkWorkflowTask,
		task.getJoinTask(),
	}
}

func (task *DynamicForkTask) getJoinTask() model.WorkflowTask {
	join := NewJoinTask(task.taskReferenceName + "_join")
	return (join.toWorkflowTask())[0]
}

// Input to the task
func (task *DynamicForkTask) Input(key string, value interface{}) *DynamicForkTask {
	task.WorkflowTaskEx.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *DynamicForkTask) InputMap(inputMap map[string]interface{}) *DynamicForkTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *DynamicForkTask) Optional(optional bool) *DynamicForkTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Description of the task
func (task *DynamicForkTask) Description(description string) *DynamicForkTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

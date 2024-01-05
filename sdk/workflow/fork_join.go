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

type ForkTask struct {
	WorkflowTaskEx
	forkedTasks [][]WorkflowTaskInterface
	join        *JoinTask
}

//NewForkTask creates a new fork task that executes the given tasks in parallel
/**
 * execute task specified in the forkedTasks parameter in parallel.
 *
 * <p>forkedTask is a two-dimensional list that executes the outermost list in parallel and list
 * within that is executed sequentially.
 *
 * <p>e.g. [[task1, task2],[task3, task4],[task5]] are executed as:
 *
 * <pre>
 *                    ---------------
 *                    |     fork    |
 *                    ---------------
 *                    |       |     |
 *                    |       |     |
 *                  task1  task3  task5
 *                  task2  task4    |
 *                    |      |      |
 *                 ---------------------
 *                 |       join        |
 *                 ---------------------
 * </pre>
 *
 *
 */
func NewForkTask(taskRefName string, forkedTask ...[]WorkflowTaskInterface) *ForkTask {
	return &ForkTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          FORK_JOIN,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		forkedTasks: forkedTask,
	}
}

func NewForkTaskWithJoin(taskRefName string, join *JoinTask, forkedTask ...[]WorkflowTaskInterface) *ForkTask {
	return &ForkTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          FORK_JOIN,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		forkedTasks: forkedTask,
		join:        join,
	}
}

func (task *ForkTask) toWorkflowTask() []model.WorkflowTask {
	forkWorkflowTask := task.WorkflowTaskEx.toWorkflowTask()[0]
	forkWorkflowTask.ForkTasks = make([][]model.WorkflowTask, len(task.forkedTasks))
	for i, forkedTask := range task.forkedTasks {
		forkWorkflowTask.ForkTasks[i] = make([]model.WorkflowTask, len(forkedTask))
		for j, innerForkedTask := range forkedTask {
			forkWorkflowTask.ForkTasks[i][j] = innerForkedTask.toWorkflowTask()[0]
		}
	}
	return []model.WorkflowTask{
		forkWorkflowTask,
		task.getJoinTask(),
	}
}

func (task *ForkTask) getJoinTask() model.WorkflowTask {
	join := task.join
	if join == nil {
		join = NewJoinTask(task.taskReferenceName + "_join")
	}
	return (join.toWorkflowTask())[0]
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *ForkTask) Input(key string, value interface{}) *ForkTask {
	task.WorkflowTaskEx.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *ForkTask) InputMap(inputMap map[string]interface{}) *ForkTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Optional if set to true, the task will not fail the workflow if one of the loop task fails
func (task *ForkTask) Optional(optional bool) *ForkTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Description of the task
func (task *ForkTask) Description(description string) *ForkTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

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
	"fmt"

	"github.com/swift-conductor/conductor-client-golang/sdk/model"
)

const (
	loopCondition = "loop_count"
)

// DoWhileTask Do...While task
type DoWhileTask struct {
	WorkflowTaskEx
	loopCondition string
	loopOver      []WorkflowTaskInterface
}

// NewDoWhileTask DoWhileTask Crate a new DoWhile task.
// terminationCondition is a Javascript expression that evaluates to True or False
func NewDoWhileTask(taskRefName string, terminationCondition string, tasks ...WorkflowTaskInterface) *DoWhileTask {
	return &DoWhileTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			taskType:          DO_WHILE,
			inputParameters:   map[string]interface{}{},
		},
		loopCondition: terminationCondition,
		loopOver:      tasks,
	}
}

// NewLoopTask Loop over N times when N is specified as iterations
func NewLoopTask(taskRefName string, iterations int32, tasks ...WorkflowTaskInterface) *DoWhileTask {
	return &DoWhileTask{
		WorkflowTaskEx: WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			taskType:          DO_WHILE,
			inputParameters: map[string]interface{}{
				loopCondition: iterations,
			},
		},
		loopCondition: getForLoopCondition(taskRefName, loopCondition),
		loopOver:      tasks,
	}
}

func (task *DoWhileTask) toWorkflowTask() []model.WorkflowTask {
	workflowTasks := task.WorkflowTaskEx.toWorkflowTask()
	workflowTasks[0].LoopCondition = task.loopCondition
	workflowTasks[0].LoopOver = []model.WorkflowTask{}
	for _, loopTask := range task.loopOver {
		workflowTasks[0].LoopOver = append(
			workflowTasks[0].LoopOver,
			loopTask.toWorkflowTask()...,
		)
	}
	return workflowTasks
}
func getForLoopCondition(loopValue string, taskReferenceName string) string {
	return fmt.Sprintf(
		"if ( $.%s['iteration'] < $.%s ) { true; } else { false; }",
		taskReferenceName, loopValue,
	)
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *DoWhileTask) Optional(optional bool) *DoWhileTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *DoWhileTask) Input(key string, value interface{}) *DoWhileTask {
	task.WorkflowTaskEx.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *DoWhileTask) InputMap(inputMap map[string]interface{}) *DoWhileTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Description of the task
func (task *DoWhileTask) Description(description string) *DoWhileTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

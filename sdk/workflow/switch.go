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

type SwitchTask struct {
	WorkflowTaskBuilder
	DecisionCases map[string][]WorkflowTaskInterface
	defaultCase   []WorkflowTaskInterface
	expression    string
	useJavascript bool
	evaluatorType string
}

func NewSwitchTask(taskRefName string, caseExpression string) *SwitchTask {
	return &SwitchTask{
		WorkflowTaskBuilder: WorkflowTaskBuilder{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			description:       "",
			taskType:          SWITCH,
			optional:          false,
			inputParameters:   map[string]interface{}{},
		},
		DecisionCases: make(map[string][]WorkflowTaskInterface),
		defaultCase:   make([]WorkflowTaskInterface, 0),
		expression:    caseExpression,
		useJavascript: false,
		evaluatorType: "value-param",
	}
}

func (task *SwitchTask) SwitchCase(caseName string, tasks ...WorkflowTaskInterface) *SwitchTask {
	task.DecisionCases[caseName] = tasks
	return task
}
func (task *SwitchTask) DefaultCase(tasks ...WorkflowTaskInterface) *SwitchTask {
	task.defaultCase = tasks
	return task
}

func (task *SwitchTask) toWorkflowTask() []model.WorkflowTask {
	if task.useJavascript {
		task.evaluatorType = "javascript"
	} else {
		task.evaluatorType = "value-param"
		task.WorkflowTaskBuilder.inputParameters["switchCaseValue"] = task.expression
		task.expression = "switchCaseValue"
	}
	var DecisionCases = map[string][]model.WorkflowTask{}
	for caseValue, tasks := range task.DecisionCases {
		for _, task := range tasks {
			for _, caseTask := range task.toWorkflowTask() {
				DecisionCases[caseValue] = append(DecisionCases[caseValue], caseTask)
			}
		}
	}
	var defaultCase []model.WorkflowTask
	for _, task := range task.defaultCase {
		for _, defaultTask := range task.toWorkflowTask() {
			defaultCase = append(defaultCase, defaultTask)
		}
	}
	workflowTasks := task.WorkflowTaskBuilder.toWorkflowTask()
	workflowTasks[0].DecisionCases = DecisionCases
	workflowTasks[0].DefaultCase = defaultCase
	workflowTasks[0].EvaluatorType = task.evaluatorType
	workflowTasks[0].Expression = task.expression
	return workflowTasks
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *SwitchTask) Input(key string, value interface{}) *SwitchTask {
	task.WorkflowTaskBuilder.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *SwitchTask) InputMap(inputMap map[string]interface{}) *SwitchTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Description of the task
func (task *SwitchTask) Description(description string) *SwitchTask {
	task.WorkflowTaskBuilder.Description(description)
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *SwitchTask) Optional(optional bool) *SwitchTask {
	task.WorkflowTaskBuilder.Optional(optional)
	return task
}

// UseJavascript If set to to true, the caseExpression parameter is treated as a Javascript.
// If set to false, the caseExpression follows the regular task input mapping format as described in https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html
func (task *SwitchTask) UseJavascript(use bool) *SwitchTask {
	task.useJavascript = use
	return task
}

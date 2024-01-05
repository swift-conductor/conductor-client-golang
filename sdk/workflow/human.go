//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package workflow

type HumanTask struct {
	WorkflowTaskEx
}

func NewHumanTask(taskRefName string) *HumanTask {
	return &HumanTask{
		WorkflowTaskEx{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			taskType:          HUMAN,
			inputParameters:   map[string]interface{}{},
		},
	}
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *HumanTask) Input(key string, value interface{}) *HumanTask {
	task.WorkflowTaskEx.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *HumanTask) InputMap(inputMap map[string]interface{}) *HumanTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *HumanTask) Optional(optional bool) *HumanTask {
	task.WorkflowTaskEx.Optional(optional)
	return task
}

// Description of the task
func (task *HumanTask) Description(description string) *HumanTask {
	task.WorkflowTaskEx.Description(description)
	return task
}

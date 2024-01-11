//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package workflow

type CustomTask struct {
	WorkflowTaskBuilder
}

func NewCustomTask(taskType string, taskRefName string) *CustomTask {
	return &CustomTask{
		WorkflowTaskBuilder{
			name:              taskType,
			taskReferenceName: taskRefName,
			taskType:          CUSTOM,
			inputParameters:   map[string]interface{}{},
		},
	}
}

// Input to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *CustomTask) Input(key string, value interface{}) *CustomTask {
	task.WorkflowTaskBuilder.Input(key, value)
	return task
}

// InputMap to the task.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (task *CustomTask) InputMap(inputMap map[string]interface{}) *CustomTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *CustomTask) Optional(optional bool) *CustomTask {
	task.WorkflowTaskBuilder.Optional(optional)
	return task
}

// Description of the task
func (task *CustomTask) Description(description string) *CustomTask {
	task.WorkflowTaskBuilder.Description(description)
	return task
}

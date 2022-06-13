//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package definition

type KafkaPublishTask struct {
	Task
}

type KafkaPublishTaskInput struct {
	BootStrapServers string                 `json:"bootStrapServers"`
	Key              string                 `json:"key"`
	KeySerializer    string                 `json:"keySerializer,omitempty"`
	Value            string                 `json:"value"`
	RequestTimeoutMs string                 `json:"requestTimeoutMs,omitempty"`
	MaxBlockMs       string                 `json:"maxBlockMs,omitempty"`
	Headers          map[string]interface{} `json:"headers,omitempty"`
	Topic            string                 `json:"topic"`
}

func NewKafkaPublishTask(taskRefName string, kafkaPublishTaskInput *KafkaPublishTaskInput) *KafkaPublishTask {
	return &KafkaPublishTask{
		Task: Task{
			name:              taskRefName,
			taskReferenceName: taskRefName,
			taskType:          KAFKA_PUBLISH,
			inputParameters: map[string]interface{}{
				"kafka_request": kafkaPublishTaskInput,
			},
		},
	}
}

// Input to the task.  See https://conductor.netflix.com/how-tos/Tasks/task-inputs.html for details
func (task *KafkaPublishTask) Input(key string, value interface{}) *KafkaPublishTask {
	task.Task.Input(key, value)
	return task
}

// InputMap to the task.  See https://conductor.netflix.com/how-tos/Tasks/task-inputs.html for details
func (task *KafkaPublishTask) InputMap(inputMap map[string]interface{}) *KafkaPublishTask {
	for k, v := range inputMap {
		task.inputParameters[k] = v
	}
	return task
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (task *KafkaPublishTask) Optional(optional bool) *KafkaPublishTask {
	task.Task.Optional(optional)
	return task
}

// Description of the task
func (task *KafkaPublishTask) Description(description string) *KafkaPublishTask {
	task.Task.Description(description)
	return task
}

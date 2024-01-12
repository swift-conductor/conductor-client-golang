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

type TaskType string

const (
	CUSTOM            TaskType = "CUSTOM"
	DYNAMIC           TaskType = "DYNAMIC"
	FORK_JOIN         TaskType = "FORK_JOIN"
	FORK_JOIN_DYNAMIC TaskType = "FORK_JOIN_DYNAMIC"
	SWITCH            TaskType = "SWITCH"
	JOIN              TaskType = "JOIN"
	DO_WHILE          TaskType = "DO_WHILE"
	SUB_WORKFLOW      TaskType = "SUB_WORKFLOW"
	START_WORKFLOW    TaskType = "START_WORKFLOW"
	EVENT             TaskType = "EVENT"
	WAIT              TaskType = "WAIT"
	HUMAN             TaskType = "HUMAN"
	HTTP              TaskType = "HTTP"
	INLINE            TaskType = "INLINE"
	TERMINATE         TaskType = "TERMINATE"
	KAFKA_PUBLISH     TaskType = "KAFKA_PUBLISH"
	JSON_JQ_TRANSFORM TaskType = "JSON_JQ_TRANSFORM"
	SET_VARIABLE      TaskType = "SET_VARIABLE"
)

type IWorkflowTask interface {
	toWorkflowTask() []model.WorkflowTask
	ToTaskDef() *model.TaskDef
	OutputRef(path string) string
}

type WorkflowTaskBuilder struct {
	name              string
	taskReferenceName string
	description       string
	taskType          TaskType
	optional          bool
	inputParameters   map[string]interface{}
}

func (builder *WorkflowTaskBuilder) toWorkflowTask() []model.WorkflowTask {
	return []model.WorkflowTask{
		{
			Name:              builder.name,
			TaskReferenceName: builder.taskReferenceName,
			Description:       builder.description,
			InputParameters:   builder.inputParameters,
			Optional:          builder.optional,
			Type_:             string(builder.taskType),
		},
	}
}

func (builder *WorkflowTaskBuilder) ToTaskDef() *model.TaskDef {
	return &model.TaskDef{
		Name:        builder.name,
		Description: builder.description,
	}
}

func (builder *WorkflowTaskBuilder) OutputRef(path string) string {
	if path == "" {
		return fmt.Sprintf("${%s.output}", builder.taskReferenceName)
	}
	return fmt.Sprintf("${%s.output.%s}", builder.taskReferenceName, path)
}

//Note: All the below method should be implemented by the
//Implementing interface given its a fluent interface
//If not, the return type is a Task which makes it impossible to use fluent interface
//For the tasks like Switch which has other methods too - quirks with Golang!

func (builder *WorkflowTaskBuilder) ReferenceName() string {
	return builder.taskReferenceName
}

// Input to the builder.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (builder *WorkflowTaskBuilder) Input(key string, value interface{}) *WorkflowTaskBuilder {
	builder.inputParameters[key] = value
	return builder
}

// InputMap to the builder.  See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for details
func (builder *WorkflowTaskBuilder) InputMap(inputMap map[string]interface{}) *WorkflowTaskBuilder {
	for k, v := range inputMap {
		builder.inputParameters[k] = v
	}
	return builder
}

// Description of the task
func (builder *WorkflowTaskBuilder) Description(description string) *WorkflowTaskBuilder {
	builder.description = description
	return builder
}

// Optional if set to true, the task will not fail the workflow if the task fails
func (builder *WorkflowTaskBuilder) Optional(optional bool) *WorkflowTaskBuilder {
	builder.optional = optional
	return builder
}

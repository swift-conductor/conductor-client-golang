//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package testdata

import (
	"fmt"
	"strconv"

	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
)

func NewKitchenSinkWorkflowBuilder() *workflow.WorkflowBuilder {
	task := workflow.NewCustomTask("custom_task", "custom_task_0")
	simpleWorkflow := workflow.NewWorkflowBuilder().
		Name("inline_sub").
		OwnerEmail("test@test.com").
		Add(
			workflow.NewCustomTask("custom_task", "custom_task_1"),
		)

	subWorkflowInline := workflow.NewSubWorkflowInlineTask(
		"sub_flow_inline",
		simpleWorkflow,
	)

	decide := workflow.NewSwitchTask("fact_length", "$.number < 15 ? 'LONG':'LONG'").
		Description("Fail if the fact is too short").
		Input("number", "${custom_task_0.output.key2}").
		UseJavascript(true).
		SwitchCase(
			"LONG",
			workflow.NewCustomTask("custom_task", "custom_task_2"),
			workflow.NewCustomTask("custom_task", "custom_task_3"),
		).
		SwitchCase(
			"SHORT",
			workflow.NewTerminateTask(
				"too_short",
				model.FailedWorkflow,
				"value too short",
			),
		)

	doWhile := workflow.NewLoopTask("loop_until_success", 2, decide).Optional(true)

	fork := workflow.NewForkTask(
		"fork",
		[]workflow.WorkflowTaskInterface{
			doWhile,
			subWorkflowInline,
		},
		[]workflow.WorkflowTaskInterface{
			workflow.NewCustomTask("custom_task", "custom_task_5"),
		},
	)

	join := workflow.NewJoinTask("new_join_ref", "custom_task_fork_ref2", "custom_task_fork_ref4")
	join.InputMap(map[string]interface{}{
		"param1": "value",
	})

	forkWithJoin := workflow.NewForkTaskWithJoin("fork_with_join_fork_ref", join,
		[]workflow.WorkflowTaskInterface{
			workflow.NewCustomTask("custom_task", "custom_task_fork_ref1"),
			workflow.NewCustomTask("custom_task", "custom_task_fork_ref2"),
		}, []workflow.WorkflowTaskInterface{
			workflow.NewCustomTask("custom_task", "custom_task_fork_ref3"),
			workflow.NewCustomTask("custom_task", "custom_task_fork_ref4"),
		})

	dynamicFork := workflow.NewDynamicForkTask(
		"dynamic_fork",
		workflow.NewCustomTask("dynamic_fork_prep", "dynamic_fork_prep"),
	)

	setVariable := workflow.NewSetVariableTask("set_state").
		Input("call_made", true).
		Input("number", task.OutputRef("number"))

	subWorkflow := workflow.NewSubWorkflowTask("sub_flow", "PopulationMinMax", 0)

	jqTask := workflow.NewJQTask("jq", "{ key3: (.key1.value1 + .key2.value2) }")
	jqTask.Input("key1", map[string]interface{}{
		"value1": []string{"a", "b"},
	})

	jqTask.InputMap(map[string]interface{}{
		"value2": []string{"d", "e"},
	})

	graalTask := workflow.NewInlineGraalJSTask("graaljstask", "(function () { return $.value1 + $.value2; })();")
	graalTask.Input("value1", "value-1")
	graalTask.Input("value2", 23.4)

	workflowBuilder := workflow.NewWorkflowBuilder().
		Name("sdk_kitchen_sink2").
		Version(1).
		OwnerEmail("test@test.com").
		Add(task).
		Add(jqTask).
		Add(graalTask).
		Add(setVariable).
		Add(subWorkflow).
		Add(dynamicFork).
		Add(fork).
		Add(forkWithJoin)

	return workflowBuilder
}

type WorkflowTask struct {
	Name              string `json:"name"`
	TaskReferenceName string `json:"taskReferenceName"`
	Type              string `json:"type,omitempty"`
}

func DynamicForkWorker(t *model.WorkerTask) (output interface{}, err error) {
	taskResult := model.NewTaskResultFromTask(t)

	tasks := []WorkflowTask{
		{
			Name:              "custom_task",
			TaskReferenceName: "custom_task_6",
			Type:              "CUSTOM",
		},
		{
			Name:              "custom_task",
			TaskReferenceName: "custom_task_7",
			Type:              "CUSTOM",
		},
		{
			Name:              "custom_task",
			TaskReferenceName: "custom_task_8",
			Type:              "CUSTOM",
		},
	}

	inputs := map[string]interface{}{
		"custom_task_6": map[string]interface{}{
			"key1": "value1",
			"key2": 121,
		},
		"custom_task_7": map[string]interface{}{
			"key1": "value2",
			"key2": 122,
		},
		"custom_task_8": map[string]interface{}{
			"key1": "value3",
			"key2": 123,
		},
	}

	taskResult.OutputData = map[string]interface{}{
		"forkedTasks":       tasks,
		"forkedTasksInputs": inputs,
	}

	taskResult.Status = model.CompletedTask

	err = nil
	return taskResult, err
}

func GetWorkflowBuilderWithComplexSwitchTask() *workflow.WorkflowBuilder {
	task := workflow.NewSwitchTask("complex_switch_task", "${workflow.input.value}")

	for i := 0; i < 3; i += 1 {
		var subtasks []workflow.WorkflowTaskInterface
		for j := 0; j <= i; j += 1 {
			httpTask := workflow.NewHttpTask(
				fmt.Sprintf("ComplexSwitchTaskGoSDK-%d-%d", i, j),
				&workflow.HttpInput{
					Uri: "http://swiftconductor.com",
				},
			)
			subtasks = append(subtasks, httpTask)
		}

		task.SwitchCase(strconv.Itoa(i), subtasks...)
	}

	return workflow.NewWorkflowBuilder().
		Name("ComplexSwitchWorkflowGoSDK").
		OwnerEmail("test@test.com").
		Version(1).
		Add(task)
}

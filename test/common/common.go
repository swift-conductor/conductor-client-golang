package common

import (
	"time"

	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
)

const (
	WorkflowValidationTimeout = 7 * time.Second
	WorkflowBulkQty           = 10
)

var (
	TestHttpTask = workflow.NewHttpTask(
		"TEST_GO_TASK_HTTP",
		&workflow.HttpInput{
			Uri: "https://swiftconductor.com",
		},
	)

	TestCustomTask = workflow.NewCustomTask(
		"TEST_GO_TASK_CUSTOM", "TEST_GO_TASK_CUSTOM",
	)

	TestTerminateTask = workflow.NewTerminateTask(
		"TEST_GO_TASK_TERMINATE",
		model.FailedWorkflow,
		"Task used to mark workflow as failed",
	)

	TestSwitchTask = workflow.NewSwitchTask(
		"TEST_GO_TASK_SWITCH",
		"switchCaseValue",
	).
		Input("switchCaseValue", "${workflow.input.service}").
		UseJavascript(true).
		SwitchCase(
			"REQUEST",
			TestHttpTask,
		).
		SwitchCase(
			"STOP",
			TestTerminateTask,
		)

	TestInlineTask = workflow.NewInlineTask(
		"TEST_GO_TASK_INLINE",
		"function e() { if ($.value == 1){return {\"result\": true}} else { return {\"result\": false}}} e();",
	)

	TestKafkaPublishTask = workflow.NewKafkaPublishTask(
		"TEST_GO_TASK_KAFKA_PUBLISH",
		&workflow.KafkaPublishTaskInput{
			Topic:            "userTopic",
			Value:            "Message to publish",
			BootStrapServers: "localhost:9092",
			Headers: map[string]interface{}{
				"x-Auth": "Auth-key",
			},
			Key:           "123",
			KeySerializer: "org.apache.kafka.common.serialization.IntegerSerializer",
		},
	)

	TestSqsEventTask = workflow.NewSqsEventTask(
		"TEST_GO_TASK_EVENT_SQS",
		"QUEUE",
	)

	TestConductorEventTask = workflow.NewConductorEventTask(
		"TEST_GO_TASK_EVENT_CONDUCTOR",
		"EVENT_NAME",
	)
)

// Package pollable contains mechanisms for setting up and managing a task that can poll SFN for work.
package pollable

import (
	"reflect"
	"time"

	"github.com/angel-one/sfn-poller/sfnpoller/cancellablecontext/cancellablecontextiface"
	"github.com/angel-one/sfn-poller/sfnpoller/pollable/pollableiface"
	"github.com/angel-one/sfn-poller/sfnpoller/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/go-logr/logr"
)

// Task is an action that supports polling.
type Task struct {
	handlerFn         interface{}
	activityArn       string
	workerName        string
	heartbeatInterval time.Duration
	sfnAPI            sfniface.SFNAPI
	started           chan struct{}
	done              chan struct{}
	logger            logr.Logger
}

// NewTask returns a reference to a new pollable task.
func NewTask(handlerFn interface{}, activityArn, workerName string, heartbeatInterval time.Duration, sfnAPI sfniface.SFNAPI, logger logr.Logger) *Task {
	return &Task{
		handlerFn:         handlerFn,
		activityArn:       activityArn,
		workerName:        workerName,
		heartbeatInterval: heartbeatInterval,
		sfnAPI:            sfnAPI,
		logger:            logger,
	}
}

// ResourceInfo is the interface for any resource that knows its ARN and ActivityName.
type ResourceInfo interface {
	ARN() string
	ActivityName() string
}

// NewTask2 returns a reference to a new pollable task using ResourceInfo.
func NewTask2(handlerFn interface{}, resourceInfo ResourceInfo, heartbeatInterval time.Duration, sfnAPI sfniface.SFNAPI, logger logr.Logger) *Task {
	return NewTask(handlerFn, resourceInfo.ARN(), resourceInfo.ActivityName(), heartbeatInterval, sfnAPI, logger)
}

// Start initializes polling for the task.
func (task *Task) Start(ctx cancellablecontextiface.Context) {
	task.started = make(chan struct{})
	task.done = make(chan struct{})
	go func() {
		defer close(task.started)
		defer close(task.done)
		task.started <- struct{}{}
		for {
			var ctxDone bool
			select {
			case <-ctx.Done():
				ctxDone = true
			default:
				ctxDone = false
			}

			if ctxDone == true {
				task.done <- struct{}{}
				task.logger.Info("Task execution done.")
				break
			}

			getActivityTaskOutput, err := task.sfnAPI.GetActivityTask(&sfn.GetActivityTaskInput{
				ActivityArn: aws.String(task.activityArn),
				WorkerName:  aws.String(task.workerName),
			})
			if err != nil {
				task.logger.Error(err, "Error getting activity task")
				continue
			}
			if getActivityTaskOutput.TaskToken == nil {
				continue
			}

			task.logger.Info("starting work on task token", "workerName", task.workerName, "token", *getActivityTaskOutput.TaskToken)

			handler := reflect.ValueOf(task.handlerFn)
			handlerType := reflect.TypeOf(task.handlerFn)
			eventType := handlerType.In(1)
			event := reflect.New(eventType)
			ctxValue := reflect.ValueOf(ctx)
			err = utils.Unmarshal(getActivityTaskOutput.Input, event.Interface())
			if err != nil {
				task.logger.Error(err, "An error occured while Unmarshalling Activity Input")
				continue
			}
			args := []reflect.Value{
				ctxValue,
				event.Elem(),
			}
			out, err := task.keepAlive(handler.Call, args, getActivityTaskOutput.TaskToken, task.heartbeatInterval)
			if err != nil {
				task.logger.Error(err, "An error occured while reporting a heartbeat to SFN!")
				continue
			}

			result := out[0].Interface()
			var callErr error
			if !out[1].IsNil() {
				callErr = out[1].Interface().(error)
			}
			if callErr != nil {
				task.logger.Info("sending failure notification to SFN...", "workerName", task.workerName, "token", *getActivityTaskOutput.TaskToken)
				_, err := task.sfnAPI.SendTaskFailure(&sfn.SendTaskFailureInput{
					Cause:     aws.String(callErr.Error()),
					Error:     aws.String(callErr.Error()),
					TaskToken: getActivityTaskOutput.TaskToken,
				})
				if err != nil {
					task.logger.Error(err, "An error occured while reporting failure to SFN!")
					continue
				}
			} else {
				task.logger.Info("sending success notification to SFN...", "workerName", task.workerName, "token", *getActivityTaskOutput.TaskToken)
				taskOutputJSON, err := utils.Marshal(result)
				if err != nil {
					task.logger.Error(err, "An error occured while marshalling output to JSON")
					continue
				}
				_, err = task.sfnAPI.SendTaskSuccess(&sfn.SendTaskSuccessInput{
					Output:    taskOutputJSON,
					TaskToken: getActivityTaskOutput.TaskToken,
				})
				if err != nil {
					task.logger.Error(err, "An error occured while reporting success to SFN!")
					continue
				}
			}
		}
	}()
}

// Done returns a channel that blocks until the task is done polling.
func (task *Task) Done() <-chan struct{} {
	return task.done
}

// Started returns a channel that blocks until the task has started polling.
func (task *Task) Started() <-chan struct{} {
	return task.started
}

// keepAlive calls the handler function then periodicially sends heartbeat notifications to SFN until the handler function returns.
// This method blocks until the handler returns.
func (task *Task) keepAlive(handler func([]reflect.Value) []reflect.Value, args []reflect.Value, taskToken *string, heartbeatInterval time.Duration) (result []reflect.Value, err error) {
	resultSource := make(chan []reflect.Value, 1)
	go func() {
		resultSource <- handler(args)
		close(resultSource)
	}()
	for {
		select {
		case result = <-resultSource:
			return
		case <-time.After(heartbeatInterval):
			task.logger.Info("Sending Heartbeat")
			_, err = task.sfnAPI.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
				TaskToken: taskToken,
			})
			if err != nil {
				task.logger.Error(err, "An error occured while sending heartbeat to SFN!")
				return
			}
		}
	}
}

var _ pollableiface.PollableTask = (*Task)(nil)

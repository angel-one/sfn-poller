// Package sfnpoller provides a 'generic' mechanism for to poll a StepFunction state machine for tasks.
package sfnpoller

import (
	"context"
	"log"

	"github.com/angel-one/sfn-poller/sfnpoller/cancellablecontext"
	"github.com/angel-one/sfn-poller/sfnpoller/pollable/pollableiface"
	"github.com/aws/aws-sdk-go/service/sfn"
)

type SFNAPI interface {
	GetActivityTask(*sfn.GetActivityTaskInput) (*sfn.GetActivityTaskOutput, error)
	SendTaskFailure(*sfn.SendTaskFailureInput) (*sfn.SendTaskFailureOutput, error)
	SendTaskHeartbeat(*sfn.SendTaskHeartbeatInput) (*sfn.SendTaskHeartbeatOutput, error)
	SendTaskSuccess(*sfn.SendTaskSuccessInput) (*sfn.SendTaskSuccessOutput, error)
}

// API is the sfnpoller's API.
type API struct {
	registeredTasks []pollableiface.PollableTask
	sfnAPI          SFNAPI
	done            chan struct{}
}

// New returns a reference to a new sfnpoller API.
func New() *API {
	return &API{
		registeredTasks: make([]pollableiface.PollableTask, 0),
	}
}

// RegisterTask adds the specified task to the poller's internal list of tasks to execute.
func (a *API) RegisterTask(task pollableiface.PollableTask) *API {
	a.registeredTasks = append(a.registeredTasks, task)
	return a
}

// BeginPolling initiates polling on registered tasks.
// This method blocks until all all pollers have reported that they have started.
func (a *API) BeginPolling(parentCtx context.Context) *API {
	log.Println("Starting tasks...")
	ctx := cancellablecontext.New(parentCtx)
	for _, task := range a.registeredTasks {
		task.Start(ctx)
	}
	numberOfStartedTasks := 0
	for i := 0; numberOfStartedTasks < len(a.registeredTasks); i++ {
		log.Println("Waiting for task to report that it has started...")
		<-a.registeredTasks[i].Started()
		numberOfStartedTasks++
	}
	log.Println("All tasks have started.")
	return a
}

// Stops them once no activity is available or current activity is finished.
func (a *API) Stop() {
	for _, task := range a.registeredTasks {
		task.Stop()
	}
}

// Done returns a channel that blocks until all pollers have reported that they are done polling.
func (a *API) Done() <-chan struct{} {
	a.done = make(chan struct{})
	go func() {
		numberOfDoneTasks := 0
		for i := 0; numberOfDoneTasks < len(a.registeredTasks); i++ {
			<-a.registeredTasks[i].Done()
			numberOfDoneTasks++
		}
		a.done <- struct{}{}
	}()
	return a.done
}

/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package telemetry

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// runnable struct contains the dynamic scheduling logic.
type runnable struct {
	client runtimeclient.Client
	cron   *cron.Cron
	*TelemetryOptions
	taskID cron.EntryID
	// taskFunc func()
	mu sync.Mutex
}

// newRunnable creates a new runnable instance with the initial schedule.
func NewRunnable(ctx context.Context, options *TelemetryOptions, client runtimeclient.Client) (*runnable, error) {
	r := &runnable{
		cron:             cron.New(),
		TelemetryOptions: options,
		client:           client,
	}

	// Initialize and start the task.
	if err := r.startTask(); err != nil {
		return nil, err
	}
	r.cron.Start()
	go func() {
		<-ctx.Done()
		r.cron.Stop()
	}()
	return r, nil
}

// startTask adds the task to the cron scheduler using the current schedule.
func (r *runnable) startTask() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add the task to the cron scheduler
	id, err := r.cron.AddFunc(r.TelemetryOptions.Schedule, func() {
		var args = []string{
			"--url", r.TelemetryOptions.KSCloudURL,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "telemetry", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			klog.Errorf("failed to exec command for telemetry %v. output is %s", err, output)
		}
	})
	if err != nil {
		return err
	}
	r.taskID = id
	return nil
}

// UpdateSchedule dynamically updates the task's schedule.
func (r *runnable) UpdateSchedule(newSchedule string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If the schedule hasn't changed, don't update.
	if newSchedule == r.TelemetryOptions.Schedule {
		return nil
	}

	// Remove the current task from the cron scheduler.
	r.cron.Remove(r.taskID)

	// Update the schedule and re-add the task.
	r.TelemetryOptions.Schedule = newSchedule
	return r.startTask()
}

func (r *runnable) Close() {
	r.cron.Stop()
}

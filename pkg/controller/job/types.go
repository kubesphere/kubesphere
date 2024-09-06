/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package job

import "time"

const (
	Running    = "running"
	Failed     = "failed"
	Unfinished = "unfinished"
	Completed  = "completed"
	Pause      = "pause"
)

type JobRevisions map[int]JobRevision

type JobRevision struct {
	Status         string    `json:"status"`
	Reasons        []string  `json:"reasons,omitempty"`
	Messages       []string  `json:"messages,omitempty"`
	Succeed        int32     `json:"succeed,omitempty"`
	DesirePodNum   int32     `json:"desire,omitempty"`
	Failed         int32     `json:"failed,omitempty"`
	Uid            string    `json:"uid"`
	StartTime      time.Time `json:"start-time,omitempty"`
	CompletionTime time.Time `json:"completion-time,omitempty"`
}

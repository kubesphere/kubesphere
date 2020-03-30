package v1alpha3

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJenkinsEvent_ToEvent(t *testing.T) {
	testdata := `
{
    "timestamp":1583995334434,
    "type":"jenkins.job.started",
    "args":{
        "jobState":{
            "name":"artifacts_test",
            "displayName":"artifacts_test",
            "url":"job/folder/job/artifacts_test/",
            "build":{
                "artifacts":{

                },
                "fullUrl":"http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/2/",
                "number":2,
                "queueId":1,
                "phase":"STARTED",
                "timestamp":1583995334364,
                "url":"job/folder/job/artifacts_test/2/",
                "testSummary":{
                    "total":0,
                    "failed":0,
                    "passed":0,
                    "skipped":0
                }
            },
            "previousCompletedBuild":{
                "artifacts":{
                    "result.xml":{
                        "archive":"http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/1/artifact/result.xml"
                    }
                },
                "fullUrl":"http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/1/",
                "number":1,
                "queueId":3,
                "phase":"COMPLETED",
                "timestamp":1583825730989,
                "status":"UNSTABLE",
                "url":"job/folder/job/artifacts_test/1/",
                "testSummary":{
                    "total":3,
                    "failed":1,
                    "passed":2,
                    "skipped":0,
                    "failedTests":[
                        "foo3.AFailingTest"
                    ]
                }
            }
        }
    }
}`
	var jenkinsEvent JenkinsEvent
	err := json.Unmarshal([]byte(testdata), &jenkinsEvent)
	if err != nil {
		t.Fatal(err)
	}
	expectEvent := Event{
		Timestamp: 1583995334434,
		Type:      PipelineStarted,
		Args: EventArgs{
			PipelineState: PipelineState{
				Name:        "artifacts_test",
				DisplayName: "artifacts_test",
				Url:         "job/folder/job/artifacts_test/",
				ProjectId:   "folder",
				Pipeline:    "artifacts_test",
				Build: BuildState{
					Artifacts:  map[string]Artifact{},
					FullUrl:    "http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/2/",
					Number:     2,
					QueueId:    1,
					Phase:      "STARTED",
					Timestamp:  1583995334364,
					Url:        "job/folder/job/artifacts_test/2/",
					Parameters: nil,
					TestSummary: TestState{
						Total:   0,
						Failed:  0,
						Passed:  0,
						Skipped: 0,
					},
				},
				PreviousCompletedBuild: BuildState{
					Artifacts: map[string]Artifact{
						"result.xml": {Archive: "http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/1/artifact/result.xml"},
					},
					FullUrl:    "http://127.0.0.1:8080/jenkins/job/folder/job/artifacts_test/1/",
					Number:     1,
					QueueId:    3,
					Phase:      "COMPLETED",
					Timestamp:  1583825730989,
					Status:     "UNSTABLE",
					Url:        "job/folder/job/artifacts_test/1/",
					Parameters: nil,
					TestSummary: TestState{
						Total:       3,
						Failed:      1,
						Passed:      2,
						Skipped:     0,
						FailedTests: []string{"foo3.AFailingTest"},
					},
				},
			},
		},
	}
	event := jenkinsEvent.ToEvent()

	if !reflect.DeepEqual(event, expectEvent) {
		t.Fatalf("expect \n %v \n, but got \n %v", expectEvent, event)
	}
}

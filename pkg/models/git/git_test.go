package git

import (
	"testing"
)

func TestGitReadVerifyWithBasicAuth(t *testing.T) {
	shouldSuccess := []map[string]string{
		{
			"username": "",
			"password": "",
			"remote":   "https://github.com/kubesphere/kubesphere",
		},
	}
	shouldFailed := []map[string]string{
		{
			"username": "",
			"password": "",
			"remote":   "https://github.com/kubesphere/kubesphere12222",
		},
		{
			"username": "",
			"password": "",
			"remote":   "git@github.com:kubesphere/kubesphere.git",
		},
		{
			"username": "runzexia",
			"password": "",
			"remote":   "git@github.com:kubesphere/kubesphere.git",
		},
		{
			"username": "",
			"password": "",
			"remote":   "git@fdsfs41342`@@@2414!!!!github.com:kubesphere/kubesphere.git",
		},
	}
	for _, item := range shouldSuccess {
		err := gitReadVerifyWithBasicAuth(item["username"], item["password"], item["remote"])
		if err != nil {

			t.Errorf("should could access repo [%s] with %s:%s, %v", item["username"], item["password"], item["remote"], err)
		}
	}

	for _, item := range shouldFailed {
		err := gitReadVerifyWithBasicAuth(item["username"], item["password"], item["remote"])
		if err == nil {
			t.Errorf("should could access repo [%s] with %s:%s ", item["username"], item["password"], item["remote"])
		}
	}
}

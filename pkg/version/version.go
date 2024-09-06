/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
)

var (
	gitVersion   = "v0.0.0"
	gitCommit    = "unknown"
	gitTreeState = "unknown"
	buildDate    = "unknown"
	gitMajor     = "unknown"
	gitMinor     = "unknown"
)

type Info struct {
	GitVersion   string                    `json:"gitVersion"`
	GitMajor     string                    `json:"gitMajor"`
	GitMinor     string                    `json:"gitMinor"`
	GitCommit    string                    `json:"gitCommit"`
	GitTreeState string                    `json:"gitTreeState"`
	BuildDate    string                    `json:"buildDate"`
	GoVersion    string                    `json:"goVersion"`
	Compiler     string                    `json:"compiler"`
	Platform     string                    `json:"platform"`
	Kubernetes   *apimachineryversion.Info `json:"kubernetes,omitempty"`
}

func (info Info) String() string {
	jsonString, _ := json.Marshal(info)
	return string(jsonString)
}

// Get returns the overall codebase version. It's for
// detecting what code a binary was built from.
func Get() Info {
	// These variables typically come from -ldflags settings and
	// in their absence fallback to the default settings
	return Info{
		GitVersion:   gitVersion,
		GitMajor:     gitMajor,
		GitMinor:     gitMinor,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// SetGitVersion sets the gitVersion, this is mainly for testing use.
func SetGitVersion(v string) {
	gitVersion = v
}

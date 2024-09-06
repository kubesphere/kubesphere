/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// Following code copied from https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apiserver/pkg/server/healthz/healthz.go

package healthz

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog/v2"
)

func AddToContainer(container *restful.Container, path string, checks ...HealthChecker) error {
	name := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
	container.Handle(path, handleRootHealth(name, nil, checks...))

	for _, check := range checks {
		container.Handle(fmt.Sprintf("%s/%v", path, check.Name()), adaptCheckToHandler(check))
	}

	return nil
}

func InstallHandler(container *restful.Container, checks ...HealthChecker) error {
	if len(checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	return AddToContainer(container, "/healthz", checks...)
}

func InstallLivezHandler(container *restful.Container, checks ...HealthChecker) error {
	if len(checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	return AddToContainer(container, "/livez", checks...)
}

// handleRootHealth returns an http.HandlerFunc that serves the provided checks.
func handleRootHealth(name string, firstTimeHealthy func(), checks ...HealthChecker) http.HandlerFunc {
	var notifyOnce sync.Once
	return func(w http.ResponseWriter, r *http.Request) {
		excluded := getExcludedChecks(r)
		// failedVerboseLogOutput is for output to the log.  It indicates detailed failed output information for the log.
		var failedVerboseLogOutput bytes.Buffer
		var failedChecks []string
		var individualCheckOutput bytes.Buffer
		for _, check := range checks {
			// no-op the check if we've specified we want to exclude the check
			if excluded.Has(check.Name()) {
				excluded.Delete(check.Name())
				fmt.Fprintf(&individualCheckOutput, "[+]%s excluded: ok\n", check.Name())
				continue
			}
			if err := check.Check(r); err != nil {
				// don't include the error since this endpoint is public.  If someone wants more detail
				// they should have explicit permission to the detailed checks.
				fmt.Fprintf(&individualCheckOutput, "[-]%s failed: reason withheld\n", check.Name())
				// but we do want detailed information for our log
				fmt.Fprintf(&failedVerboseLogOutput, "[-]%s failed: %v\n", check.Name(), err)
				failedChecks = append(failedChecks, check.Name())
			} else {
				fmt.Fprintf(&individualCheckOutput, "[+]%s ok\n", check.Name())
			}
		}
		if excluded.Len() > 0 {
			fmt.Fprintf(&individualCheckOutput, "warn: some health checks cannot be excluded: no matches for %s\n", formatQuoted(excluded.UnsortedList()...))
			klog.Warningf("cannot exclude some health checks, no health checks are installed matching %s",
				formatQuoted(excluded.UnsortedList()...))
		}
		// always be verbose on failure
		if len(failedChecks) > 0 {
			klog.V(2).Infof("%s check failed: %s\n%v", strings.Join(failedChecks, ","), name, failedVerboseLogOutput.String())
			httplog.SetStacktracePredicate(r.Context(), func(int) bool { return false })
			http.Error(w, fmt.Sprintf("%s%s check failed", individualCheckOutput.String(), name), http.StatusInternalServerError)
			return
		}

		// signal first time this is healthy
		if firstTimeHealthy != nil {
			notifyOnce.Do(firstTimeHealthy)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, found := r.URL.Query()["verbose"]; !found {
			fmt.Fprint(w, "ok")
			return
		}

		individualCheckOutput.WriteTo(w)
		fmt.Fprintf(w, "%s check passed\n", name)
	}
}

// adaptCheckToHandler returns an http.HandlerFunc that serves the provided checks.
func adaptCheckToHandler(checks HealthChecker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Header().Set("X-Content-Type-Options", "nosniff")

		err := checks.Check(request)
		if err != nil {
			http.Error(writer, fmt.Sprintf("internal server error: %v", err), http.StatusInternalServerError)
		} else {
			fmt.Fprint(writer, "ok")
		}
	}
}

// HealthChecker is a named healthz checker.
type HealthChecker interface {
	Name() string
	Check(req *http.Request) error
}

// getExcludedChecks extracts the health check names to be excluded from the query param
func getExcludedChecks(r *http.Request) sets.Set[string] {
	checks, found := r.URL.Query()["exclude"]
	if found {
		return sets.New(checks...)
	}
	return sets.New[string]()
}

// PingHealthz returns true automatically when checked
var PingHealthz HealthChecker = ping{}

// ping implements the simplest possible healthz checker.
type ping struct{}

func (ping) Name() string {
	return "ping"
}

// PingHealthz is a health check that returns true.
func (ping) Check(_ *http.Request) error {
	return nil
}

// formatQuoted returns a formatted string of the health check names,
// preserving the order passed in.
func formatQuoted(names ...string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ",")
}

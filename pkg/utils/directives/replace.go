/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package directives

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Following code copied from github.com/caddyserver/caddy/modules/caddyhttp/rewrite/rewrite.go

type substrReplacer struct {
	Find    string `json:"find,omitempty"`
	Replace string `json:"replace,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

func (rep substrReplacer) do(r *http.Request, repl *Replacer) {
	if rep.Find == "" {
		return
	}

	lim := rep.Limit
	if lim == 0 {
		lim = -1
	}

	find := repl.ReplaceAll(rep.Find, "")
	replace := repl.ReplaceAll(rep.Replace, "")

	mergeSlashes := !strings.Contains(rep.Find, "//")

	changePath(r, func(pathOrRawPath string) string {
		return strings.Replace(CleanPath(pathOrRawPath, mergeSlashes), find, replace, lim)
	})

	r.URL.RawQuery = strings.Replace(r.URL.RawQuery, find, replace, lim)
}

type regexReplacer struct {
	Find    string `json:"find,omitempty"`
	Replace string `json:"replace,omitempty"`
	re      *regexp.Regexp
}

func (rep regexReplacer) do(r *http.Request, repl *Replacer) {
	if rep.Find == "" || rep.re == nil {
		return
	}
	replace := repl.ReplaceAll(rep.Replace, "")
	changePath(r, func(pathOrRawPath string) string {
		return rep.re.ReplaceAllString(pathOrRawPath, replace)
	})
}

func NewReplacer() *Replacer {
	rep := &Replacer{
		static: make(map[string]any),
	}
	rep.providers = []ReplacerFunc{
		globalDefaultReplacements,
		rep.fromStatic,
	}
	return rep
}

type Replacer struct {
	providers []ReplacerFunc
	static    map[string]any
}

func (r *Replacer) Map(mapFunc ReplacerFunc) {
	r.providers = append(r.providers, mapFunc)
}

func (r *Replacer) Set(variable string, value any) {
	r.static[variable] = value
}

func (r *Replacer) Get(variable string) (any, bool) {
	for _, mapFunc := range r.providers {
		if val, ok := mapFunc(variable); ok {
			return val, true
		}
	}
	return nil, false
}

func (r *Replacer) GetString(variable string) (string, bool) {
	s, found := r.Get(variable)
	return ToString(s), found
}

func (r *Replacer) Delete(variable string) {
	delete(r.static, variable)
}

func (r *Replacer) fromStatic(key string) (any, bool) {
	val, ok := r.static[key]
	return val, ok
}

func (r *Replacer) ReplaceOrErr(input string, errOnEmpty, errOnUnknown bool) (string, error) {
	return r.replace(input, "", false, errOnEmpty, errOnUnknown, nil)
}

func (r *Replacer) ReplaceKnown(input, empty string) string {
	out, _ := r.replace(input, empty, false, false, false, nil)
	return out
}

func (r *Replacer) ReplaceAll(input, empty string) string {
	out, _ := r.replace(input, empty, true, false, false, nil)
	return out
}

func (r *Replacer) ReplaceFunc(input string, f ReplacementFunc) (string, error) {
	return r.replace(input, "", true, false, false, f)
}

func (r *Replacer) replace(input, empty string,
	treatUnknownAsEmpty, errOnEmpty, errOnUnknown bool,
	f ReplacementFunc,
) (string, error) {
	if !strings.Contains(input, string(phOpen)) {
		return input, nil
	}

	var sb strings.Builder

	// it is reasonable to assume that the output
	// will be approximately as long as the input
	sb.Grow(len(input))

	// iterate the input to find each placeholder
	var lastWriteCursor int

	// fail fast if too many placeholders are unclosed
	var unclosedCount int

scan:
	for i := 0; i < len(input); i++ {
		// check for escaped braces
		if i > 0 && input[i-1] == phEscape && (input[i] == phClose || input[i] == phOpen) {
			sb.WriteString(input[lastWriteCursor : i-1])
			lastWriteCursor = i
			continue
		}

		if input[i] != phOpen {
			continue
		}

		// our iterator is now on an unescaped open brace (start of placeholder)

		// too many unclosed placeholders in absolutely ridiculous input can be extremely slow (issue #4170)
		if unclosedCount > 100 {
			return "", fmt.Errorf("too many unclosed placeholders")
		}

		// find the end of the placeholder
		end := strings.Index(input[i:], string(phClose)) + i
		if end < i {
			unclosedCount++
			continue
		}

		// if necessary look for the first closing brace that is not escaped
		for end > 0 && end < len(input)-1 && input[end-1] == phEscape {
			nextEnd := strings.Index(input[end+1:], string(phClose))
			if nextEnd < 0 {
				unclosedCount++
				continue scan
			}
			end += nextEnd + 1
		}

		// write the substring from the last cursor to this point
		sb.WriteString(input[lastWriteCursor:i])

		// trim opening bracket
		key := input[i+1 : end]

		// try to get a value for this key, handle empty values accordingly
		val, found := r.Get(key)
		if !found {
			// placeholder is unknown (unrecognized); handle accordingly
			if errOnUnknown {
				return "", fmt.Errorf("unrecognized placeholder %s%s%s",
					string(phOpen), key, string(phClose))
			} else if !treatUnknownAsEmpty {
				// if treatUnknownAsEmpty is true, we'll handle an empty
				// val later; so only continue otherwise
				lastWriteCursor = i
				continue
			}
		}

		// apply any transformations
		if f != nil {
			var err error
			val, err = f(key, val)
			if err != nil {
				return "", err
			}
		}

		valStr := ToString(val)

		if valStr == "" {
			if errOnEmpty {
				return "", fmt.Errorf("evaluated placeholder %s%s%s is empty",
					string(phOpen), key, string(phClose))
			} else if empty != "" {
				sb.WriteString(empty)
			}
		} else {
			sb.WriteString(valStr)
		}

		i = end
		lastWriteCursor = i + 1
	}

	sb.WriteString(input[lastWriteCursor:])

	return sb.String(), nil
}

func ToString(val any) string {
	switch v := val.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	case byte:
		return string(v)
	case []byte:
		return string(v)
	case []rune:
		return string(v)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.Itoa(int(v))
	case uint:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint64:
		return strconv.Itoa(int(v))
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%+v", v)
	}
}

type ReplacerFunc func(key string) (any, bool)

func globalDefaultReplacements(key string) (any, bool) {
	const envPrefix = "env."
	if strings.HasPrefix(key, envPrefix) {
		return os.Getenv(key[len(envPrefix):]), true
	}

	switch key {
	case "system.hostname":
		name, _ := os.Hostname()
		return name, true
	case "system.slash":
		return string(filepath.Separator), true
	case "system.os":
		return runtime.GOOS, true
	case "system.wd":
		// OK if there is an error; just return empty string
		wd, _ := os.Getwd()
		return wd, true
	case "system.arch":
		return runtime.GOARCH, true
	case "time.now":
		return nowFunc(), true
	case "time.now.http":
		return nowFunc().UTC().Format(http.TimeFormat), true
	case "time.now.common_log":
		return nowFunc().Format("02/Jan/2006:15:04:05 -0700"), true
	case "time.now.year":
		return strconv.Itoa(nowFunc().Year()), true
	case "time.now.unix":
		return strconv.FormatInt(nowFunc().Unix(), 10), true
	case "time.now.unix_ms":
		return strconv.FormatInt(nowFunc().UnixNano()/int64(time.Millisecond), 10), true
	}

	return nil, false
}

// ReplacementFunc is a function that is called when a
// replacement is being performed. It receives the
// variable (i.e. placeholder name) and the value that
// will be the replacement, and returns the value that
// will actually be the replacement, or an error. Note
// that errors are sometimes ignored by replacers.
type ReplacementFunc func(variable string, val any) (any, error)

// nowFunc is a variable so tests can change it
// in order to obtain a deterministic time.
var nowFunc = time.Now

type ContextKey string

// ReplacerCtxKey is the context key for a replacer.
const ReplacerCtxKey ContextKey = "replacer"

const phOpen, phClose, phEscape = '{', '}', '\\'

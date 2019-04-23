package gojenkins

import (
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Parse jenkins ajax response in order find the current jenkins build history
func parseBuildHistory(d io.Reader) []*History {
	z := html.NewTokenizer(d)
	depth := 0
	buildRowCellDepth := -1
	builds := make([]*History, 0)
	var curBuild *History
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return builds
			}
		case html.SelfClosingTagToken:
			tn, hasAttr := z.TagName()
			// fmt.Println("START__", string(tn), hasAttr)
			if hasAttr {
				a := attr(z)
				// <img src="/static/f2881562/images/16x16/red.png" alt="Failed &gt; Console Output" tooltip="Failed &gt; Console Output" style="width: 16px; height: 16px; " class="icon-red icon-sm" />
				if string(tn) == "img" {
					if hasCSSClass(a, "icon-sm") && buildRowCellDepth > -1 {
						if alt, found := a["alt"]; found {
							curBuild.BuildStatus = strings.Fields(alt)[0]
						}
					}
				}
			}
		case html.StartTagToken:
			depth++
			tn, hasAttr := z.TagName()
			// fmt.Println("START__", string(tn), hasAttr)
			if hasAttr {
				a := attr(z)
				// <td class="build-row-cell">
				if string(tn) == "td" {
					if hasCSSClass(a, "build-row-cell") {
						buildRowCellDepth = depth
						curBuild = &History{}
						builds = append(builds, curBuild)
					}
				}
				// <a update-parent-class=".build-row" href="/job/appscode/job/43/job/build-binary/227/" class="tip model-link inside build-link display-name">#227</a>
				if string(tn) == "a" {
					if hasCSSClass(a, "build-link") && buildRowCellDepth > -1 {
						if href, found := a["href"]; found {
							parts := strings.Split(href, "/")
							if num, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
								curBuild.BuildNumber = num
							}
						}
					}
				}
				// <div time="1469024602546" class="pane build-details"> ... </div>
				if string(tn) == "div" {
					if hasCSSClass(a, "build-details") && buildRowCellDepth > -1 {
						if t, found := a["time"]; found {
							if msec, err := strconv.ParseInt(t, 10, 0); err == nil {
								curBuild.BuildTimestamp = msec / 1000
							}
						}
					}
				}
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if string(tn) == "td" && depth == buildRowCellDepth {
				buildRowCellDepth = -1
				curBuild = nil
			}
			depth--
		}
	}
}

func attr(z *html.Tokenizer) map[string]string {
	a := make(map[string]string)
	for {
		k, v, more := z.TagAttr()
		if k != nil && v != nil {
			a[string(k)] = string(v)
		}
		if !more {
			break
		}
	}
	return a
}

func hasCSSClass(a map[string]string, className string) bool {
	if classes, found := a["class"]; found {
		for _, class := range strings.Fields(classes) {
			if class == className {
				return true
			}
		}
	}
	return false
}

package util

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/kiali/kiali/log"
)

// UpdateBaseURL updates index.html base href with web root string
func UpdateBaseURL(webRootPath string) {
	log.Infof("Updating base URL in index.html with [%v]", webRootPath)
	path, _ := filepath.Abs("./console/index.html")
	b, err := ioutil.ReadFile(path)
	if isError(err) {
		return
	}

	html := string(b)

	searchStr := `<base href="/">`
	newStr := `<base href="` + webRootPath + `/">`
	newHTML := strings.Replace(html, searchStr, newStr, -1)

	err = ioutil.WriteFile(path, []byte(newHTML), 0)
	if isError(err) {
		return
	}
}

func isError(err error) bool {
	if err != nil {
		log.Errorf("File I/O error [%v]", err.Error())
	}

	return (err != nil)
}

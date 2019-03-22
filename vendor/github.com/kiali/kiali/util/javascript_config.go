package util

import (
	"io/ioutil"
	"path/filepath"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/log"
)

// ConfigToJS generates env.js file from Kiali config
func ConfigToJS() {
	log.Info("Generating env.js from config")
	path, _ := filepath.Abs("./console/env.js")

	content := "window.WEB_ROOT='" + config.Get().Server.WebRoot + "';"

	log.Debugf("The content of %v will be:\n%v", path, content)

	err := ioutil.WriteFile(path, []byte(content), 0)
	if isError(err) {
		return
	}
}

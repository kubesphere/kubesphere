package multiconfig

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

var (
	// ErrSourceNotSet states that neither the path or the reader is set on the loader
	ErrSourceNotSet = errors.New("config path or reader is not set")

	// ErrFileNotFound states that given file is not exists
	ErrFileNotFound = errors.New("config file not found")
)

// TOMLLoader satisifies the loader interface. It loads the configuration from
// the given toml file or Reader.
type TOMLLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (t *TOMLLoader) Load(s interface{}) error {
	var r io.Reader

	if t.Reader != nil {
		r = t.Reader
	} else if t.Path != "" {
		file, err := getConfig(t.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	if _, err := toml.DecodeReader(r, s); err != nil {
		return err
	}

	return nil
}

// JSONLoader satisifies the loader interface. It loads the configuration from
// the given json file or Reader.
type JSONLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s.
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (j *JSONLoader) Load(s interface{}) error {
	var r io.Reader
	if j.Reader != nil {
		r = j.Reader
	} else if j.Path != "" {
		file, err := getConfig(j.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	return json.NewDecoder(r).Decode(s)
}

// YAMLLoader satisifies the loader interface. It loads the configuration from
// the given yaml file.
type YAMLLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s.
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (y *YAMLLoader) Load(s interface{}) error {
	var r io.Reader

	if y.Reader != nil {
		r = y.Reader
	} else if y.Path != "" {
		file, err := getConfig(y.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, s)
}

func getConfig(path string) (*os.File, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := path
	if !filepath.IsAbs(path) {
		configPath = filepath.Join(pwd, path)
	}

	// check if file with combined path is exists(relative path)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return os.Open(configPath)
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}
	return f, err
}

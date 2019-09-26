package multiconfig

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
)

// Loader loads the configuration from a source. The implementer of Loader is
// responsible for setting the default values of the struct.
type Loader interface {
	// Load loads the source into the config defined by struct s
	Load(s interface{}) error
}

// DefaultLoader implements the Loader interface. It initializes the given
// pointer of struct s with configuration from the default sources. The order
// of load is TagLoader, FileLoader, EnvLoader and lastly FlagLoader. An error
// in any step stops the loading process. Each step overrides the previous
// step's config (i.e: defining a flag will override previous environment or
// file config). To customize the order use the individual load functions.
type DefaultLoader struct {
	Loader
	Validator
}

// NewWithPath returns a new instance of Loader to read from the given
// configuration file.
func NewWithPath(path string) *DefaultLoader {
	loaders := []Loader{}

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &TagLoader{})

	// Choose what while is passed
	if strings.HasSuffix(path, "toml") {
		loaders = append(loaders, &TOMLLoader{Path: path})
	}

	if strings.HasSuffix(path, "json") {
		loaders = append(loaders, &JSONLoader{Path: path})
	}

	if strings.HasSuffix(path, "yml") || strings.HasSuffix(path, "yaml") {
		loaders = append(loaders, &YAMLLoader{Path: path})
	}

	e := &EnvironmentLoader{}
	f := &FlagLoader{}

	loaders = append(loaders, e, f)
	loader := MultiLoader(loaders...)

	d := &DefaultLoader{}
	d.Loader = loader
	d.Validator = MultiValidator(&RequiredValidator{})
	return d
}

// New returns a new instance of DefaultLoader without any file loaders.
func New() *DefaultLoader {
	loader := MultiLoader(
		&TagLoader{},
		&EnvironmentLoader{},
		&FlagLoader{},
	)

	d := &DefaultLoader{}
	d.Loader = loader
	d.Validator = MultiValidator(&RequiredValidator{})
	return d
}

// MustLoadWithPath loads with the DefaultLoader settings and from the given
// Path. It exits if the config cannot be parsed.
func MustLoadWithPath(path string, conf interface{}) {
	d := NewWithPath(path)
	d.MustLoad(conf)
}

// MustLoad loads with the DefaultLoader settings. It exits if the config
// cannot be parsed.
func MustLoad(conf interface{}) {
	d := New()
	d.MustLoad(conf)
}

// MustLoad is like Load but panics if the config cannot be parsed.
func (d *DefaultLoader) MustLoad(conf interface{}) {
	if err := d.Load(conf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// we at koding, believe having sane defaults in our system, this is the
	// reason why we have default validators in DefaultLoader. But do not cause
	// nil pointer panics if one uses DefaultLoader directly.
	if d.Validator != nil {
		d.MustValidate(conf)
	}
}

// MustValidate validates the struct. It exits with status 1 if it can't
// validate.
func (d *DefaultLoader) MustValidate(conf interface{}) {
	if err := d.Validate(conf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// fieldSet sets field value from the given string value. It converts the
// string value in a sane way and is usefulf or environment variables or flags
// which are by nature in string types.
func fieldSet(field *structs.Field, v string) error {
	switch f := field.Value().(type) {
	case flag.Value:
		if v := reflect.ValueOf(field.Value()); v.IsNil() {
			typ := v.Type()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}

			if err := field.Set(reflect.New(typ).Interface()); err != nil {
				return err
			}

			f = field.Value().(flag.Value)
		}

		return f.Set(v)
	}

	// TODO: add support for other types
	switch field.Kind() {
	case reflect.Bool:
		val, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}

		if err := field.Set(val); err != nil {
			return err
		}
	case reflect.Int:
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}

		if err := field.Set(i); err != nil {
			return err
		}
	case reflect.String:
		if err := field.Set(v); err != nil {
			return err
		}
	case reflect.Slice:
		switch t := field.Value().(type) {
		case []string:
			if err := field.Set(strings.Split(v, ",")); err != nil {
				return err
			}
		case []int:
			var list []int
			for _, in := range strings.Split(v, ",") {
				i, err := strconv.Atoi(in)
				if err != nil {
					return err
				}

				list = append(list, i)
			}

			if err := field.Set(list); err != nil {
				return err
			}
		default:
			return fmt.Errorf("multiconfig: field '%s' of type slice is unsupported: %s (%T)",
				field.Name(), field.Kind(), t)
		}
	case reflect.Float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}

		if err := field.Set(f); err != nil {
			return err
		}
	case reflect.Int64:
		switch t := field.Value().(type) {
		case time.Duration:
			d, err := time.ParseDuration(v)
			if err != nil {
				return err
			}

			if err := field.Set(d); err != nil {
				return err
			}
		case int64:
			p, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return err
			}

			if err := field.Set(p); err != nil {
				return err
			}
		default:
			return fmt.Errorf("multiconfig: field '%s' of type int64 is unsupported: %s (%T)",
				field.Name(), field.Kind(), t)
		}

	default:
		return fmt.Errorf("multiconfig: field '%s' has unsupported type: %s", field.Name(), field.Kind())
	}

	return nil
}

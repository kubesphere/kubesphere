package multiconfig

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/fatih/structs"
)

// EnvironmentLoader satisifies the loader interface. It loads the
// configuration from the environment variables in the form of
// STRUCTNAME_FIELDNAME.
type EnvironmentLoader struct {
	// Prefix prepends given string to every environment variable
	// {STRUCTNAME}_FIELDNAME will be {PREFIX}_FIELDNAME
	Prefix string

	// CamelCase adds a separator for field names in camelcase form. A
	// fieldname of "AccessKey" would generate a environment name of
	// "STRUCTNAME_ACCESSKEY". If CamelCase is enabled, the environment name
	// will be generated in the form of "STRUCTNAME_ACCESS_KEY"
	CamelCase bool
}

func (e *EnvironmentLoader) getPrefix(s *structs.Struct) string {
	if e.Prefix != "" {
		return e.Prefix
	}

	return s.Name()
}

// Load loads the source into the config defined by struct s
func (e *EnvironmentLoader) Load(s interface{}) error {
	strct := structs.New(s)
	strctMap := strct.Map()
	prefix := e.getPrefix(strct)

	for key, val := range strctMap {
		field := strct.Field(key)

		if err := e.processField(prefix, field, key, val); err != nil {
			return err
		}
	}

	return nil
}

// processField gets leading name for the env variable and combines the current
// field's name and generates environment variable names recursively
func (e *EnvironmentLoader) processField(prefix string, field *structs.Field, name string, strctMap interface{}) error {
	fieldName := e.generateFieldName(prefix, name)

	switch strctMap.(type) {
	case map[string]interface{}:
		for key, val := range strctMap.(map[string]interface{}) {
			field := field.Field(key)

			if err := e.processField(fieldName, field, key, val); err != nil {
				return err
			}
		}
	default:
		v := os.Getenv(fieldName)
		if v == "" {
			return nil
		}

		if err := fieldSet(field, v); err != nil {
			return err
		}
	}

	return nil
}

// PrintEnvs prints the generated environment variables to the std out.
func (e *EnvironmentLoader) PrintEnvs(s interface{}) {
	strct := structs.New(s)
	strctMap := strct.Map()
	prefix := e.getPrefix(strct)

	keys := make([]string, 0, len(strctMap))
	for key := range strctMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		field := strct.Field(key)
		e.printField(prefix, field, key, strctMap[key])
	}
}

// printField prints the field of the config struct for the flag.Usage
func (e *EnvironmentLoader) printField(prefix string, field *structs.Field, name string, strctMap interface{}) {
	fieldName := e.generateFieldName(prefix, name)

	switch strctMap.(type) {
	case map[string]interface{}:
		smap := strctMap.(map[string]interface{})
		keys := make([]string, 0, len(smap))
		for key := range smap {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			field := field.Field(key)
			e.printField(fieldName, field, key, smap[key])
		}
	default:
		fmt.Println("  ", fieldName)
	}
}

// generateFieldName generates the field name combined with the prefix and the
// struct's field name
func (e *EnvironmentLoader) generateFieldName(prefix string, name string) string {
	fieldName := strings.ToUpper(name)
	if e.CamelCase {
		fieldName = strings.ToUpper(strings.Join(camelcase.Split(name), "_"))
	}

	return strings.ToUpper(prefix) + "_" + fieldName
}

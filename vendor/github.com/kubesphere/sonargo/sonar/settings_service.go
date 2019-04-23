// Manage settings.
package sonargo

import "net/http"

type SettingsService struct {
	client *Client
}

type SettingsListDefinitionsObject struct {
	Definitions []*Definition `json:"definitions,omitempty"`
}

type Definition struct {
	Category      string   `json:"category,omitempty"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	DeprecatedKey string   `json:"deprecatedKey,omitempty"`
	Description   string   `json:"description,omitempty"`
	Fields        []*Field `json:"fields,omitempty"`
	Key           string   `json:"key,omitempty"`
	MultiValues   bool     `json:"multiValues,omitempty"`
	Name          string   `json:"name,omitempty"`
	Options       []string `json:"options,omitempty"`
	SubCategory   string   `json:"subCategory,omitempty"`
	Type          string   `json:"type,omitempty"`
}

type Field struct {
	Description string   `json:"description,omitempty"`
	Key         string   `json:"key,omitempty"`
	Name        string   `json:"name,omitempty"`
	Options     []string `json:"options,omitempty"`
	Type        string   `json:"type,omitempty"`
}

type SettingsValuesObject struct {
	Settings []*Setting `json:"settings,omitempty"`
}
type FieldValue struct {
	Boolean string `json:"boolean,omitempty"`
	Text    string `json:"text,omitempty"`
}
type Setting struct {
	Key         string        `json:"key,omitempty"`
	Value       string        `json:"value,omitempty"`
	Inherited   bool          `json:"inherited,omitempty"`
	Values      []string      `json:"values,omitempty"`
	FieldValues []*FieldValue `json:"fieldValues,omitempty"`
}
type SettingsListDefinitionsOption struct {
	Component string `url:"component,omitempty"` // Description:"Component key",ExampleValue:"my_project"
}

// ListDefinitions List settings definitions.<br>Requires 'Browse' permission when a component is specified<br/>
func (s *SettingsService) ListDefinitions(opt *SettingsListDefinitionsOption) (v *SettingsListDefinitionsObject, resp *http.Response, err error) {
	err = s.ValidateListDefinitionsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "settings/list_definitions", opt)
	if err != nil {
		return
	}
	v = new(SettingsListDefinitionsObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type SettingsResetOption struct {
	Component string `url:"component,omitempty"` // Description:"Component key",ExampleValue:"my_project"
	Keys      string `url:"keys,omitempty"`      // Description:"Comma-separated list of keys",ExampleValue:"sonar.links.scm,sonar.debt.hoursInDay"
}

// Reset Remove a setting value.<br>The settings defined in config/sonar.properties are read-only and can't be changed.<br/>Requires one of the following permissions: <ul><li>'Administer System'</li><li>'Administer' rights on the specified component</li></ul>
func (s *SettingsService) Reset(opt *SettingsResetOption) (resp *http.Response, err error) {
	err = s.ValidateResetOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "settings/reset", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type SettingsSetOption struct {
	Component   string   `url:"component,omitempty"`   // Description:"Component key",ExampleValue:"my_project"
	FieldValues string   `url:"fieldValues,omitempty"` // Description:"Setting field values. To set several values, the parameter must be called once for each value.",ExampleValue:"fieldValues={"firstField":"first value", "secondField":"second value", "thirdField":"third value"}"
	Key         string   `url:"key,omitempty"`         // Description:"Setting key",ExampleValue:"sonar.links.scm"
	Value       string   `url:"value,omitempty"`       // Description:"Setting value. To reset a value, please use the reset web service.",ExampleValue:"git@github.com:SonarSource/sonarqube.git"
	Values      []string `url:"values,omitempty"`      // Description:"Setting multi value. To set several values, the parameter must be called once for each value.",ExampleValue:"values=firstValue&values=secondValue&values=thirdValue"
}

// Set Update a setting value.<br>Either 'value' or 'values' must be provided.<br> The settings defined in config/sonar.properties are read-only and can't be changed.<br/>Requires one of the following permissions: <ul><li>'Administer System'</li><li>'Administer' rights on the specified component</li></ul>
func (s *SettingsService) Set(opt *SettingsSetOption) (resp *http.Response, err error) {
	err = s.ValidateSetOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "settings/set", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type SettingsValuesOption struct {
	Component string `url:"component,omitempty"` // Description:"Component key",ExampleValue:"my_project"
	Keys      string `url:"keys,omitempty"`      // Description:"List of setting keys",ExampleValue:"sonar.test.inclusions,sonar.dbcleaner.cleanDirectory"
}

// Values List settings values.<br>If no value has been set for a setting, then the default value is returned.<br>The settings from conf/sonar.properties are excluded from results.<br>Requires 'Browse' or 'Execute Analysis' permission when a component is specified<br/>
func (s *SettingsService) Values(opt *SettingsValuesOption) (v *SettingsValuesObject, resp *http.Response, err error) {
	err = s.ValidateValuesOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "settings/values", opt)
	if err != nil {
		return
	}
	v = new(SettingsValuesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

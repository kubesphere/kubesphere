// Get and update some details of automatic rules, and manage custom rules.
package sonargo

import "net/http"

type RulesService struct {
	client *Client
}

type Rule struct {
	Key             string       `json:"key,omitempty"`
	Repo            string       `json:"repo,omitempty"`
	Name            string       `json:"name,omitempty"`
	CreatedAt       string       `json:"createdAt,omitempty"`
	HTMLDesc        string       `json:"htmlDesc,omitempty"`
	MdDesc          string       `json:"mdDesc,omitempty"`
	Severity        string       `json:"severity,omitempty"`
	Status          string       `json:"status,omitempty"`
	IsTemplate      bool         `json:"isTemplate,omitempty"`
	TemplateKey     string       `json:"templateKey,omitempty"`
	Tags            []string     `json:"tags,omitempty"`
	SysTags         []string     `json:"sysTags,omitempty"`
	Lang            string       `json:"lang,omitempty"`
	LangName        string       `json:"langName,omitempty"`
	DebtOverloaded  bool         `json:"debtOverloaded,omitempty"`
	RemFnOverloaded bool         `json:"remFnOverloaded,omitempty"`
	Params          []*RuleParam `json:"params,omitempty"`
	Scope           string       `json:"scope,omitempty"`
	IsExternal      bool         `json:"isExternal,omitempty"`
	Type            string       `json:"type,omitempty"`
}

type RuleParam struct {
	DefaultValue string `json:"defaultValue,omitempty"`
	Key          string `json:"key,omitempty"`
	HTMLDesc     string `json:"htmlDesc,omitempty"`
	Type         string `json:"type,omitempty"`
}
type RulesRepositoriesObject struct {
	Repositories []*Repositorie `json:"repositories,omitempty"`
}

type Repositorie struct {
	Key      string `json:"key,omitempty"`
	Language string `json:"language,omitempty"`
	Name     string `json:"name,omitempty"`
}
type RuleCreateObject struct {
	Rule *Rule `json:"rule,omitempty"`
}
type RulesSearchObject struct {
	Actives *Actives `json:"actives,omitempty"`
	Facets  []*Facet `json:"facets,omitempty"`
	P       int64    `json:"p,omitempty"`
	Ps      int64    `json:"ps,omitempty"`
	Rules   []*Rule  `json:"rules,omitempty"`
	Total   int64    `json:"total,omitempty"`
}

type RuleUpdateObject RuleCreateObject
type Actives struct {
	SquidClassCyclomaticComplexity  []*SquidClassCyclomaticComplexity `json:"squid:ClassCyclomaticComplexity,omitempty"`
	SquidMethodCyclomaticComplexity []*SquidClassCyclomaticComplexity `json:"squid:MethodCyclomaticComplexity,omitempty"`
	Squid_S1067                     []*SquidClassCyclomaticComplexity `json:"squid:S1067,omitempty"`
}

type Facet struct {
	Name     string        `json:"name,omitempty"`
	Property string        `json:"property,omitempty"`
	Values   []*FacetValue `json:"values,omitempty"`
}

type SquidParam struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type SquidClassCyclomaticComplexity struct {
	Inherit  string        `json:"inherit,omitempty"`
	Params   []*SquidParam `json:"params,omitempty"`
	QProfile string        `json:"qProfile,omitempty"`
	Severity string        `json:"severity,omitempty"`
}

type FacetValue struct {
	Count int64  `json:"count,omitempty"`
	Val   string `json:"val,omitempty"`
}

type RulesShowObject struct {
	Actives []*SquidClassCyclomaticComplexity `json:"actives,omitempty"`
	Rule    *Rule                             `json:"rule,omitempty"`
}

type RulesCreateOption struct {
	CustomKey           string `url:"custom_key,omitempty"`           // Description:"Key of the custom rule",ExampleValue:"Todo_should_not_be_used"
	ManualKey           string `url:"manual_key,omitempty"`           // Description:"Manual rules are no more supported. This parameter is ignored",ExampleValue:"Error_handling"
	MarkdownDescription string `url:"markdown_description,omitempty"` // Description:"Rule description",ExampleValue:"Description of my custom rule"
	Name                string `url:"name,omitempty"`                 // Description:"Rule name",ExampleValue:"My custom rule"
	Params              string `url:"params,omitempty"`               // Description:"Parameters as semi-colon list of <key>=<value>, for example 'params=key1=v1;key2=v2' (Only for custom rule)",ExampleValue:""
	PreventReactivation string `url:"prevent_reactivation,omitempty"` // Description:"If set to true and if the rule has been deactivated (status 'REMOVED'), a status 409 will be returned",ExampleValue:""
	Severity            string `url:"severity,omitempty"`             // Description:"Rule severity",ExampleValue:""
	Status              string `url:"status,omitempty"`               // Description:"Rule status",ExampleValue:""
	TemplateKey         string `url:"template_key,omitempty"`         // Description:"Key of the template rule in order to create a custom rule (mandatory for custom rule)",ExampleValue:"java:XPath"
	Type                string `url:"type,omitempty"`                 // Description:"Rule type",ExampleValue:""
}

// Create Create a custom rule.<br>Requires the 'Administer Quality Profiles' permission
func (s *RulesService) Create(opt *RulesCreateOption) (v *RuleCreateObject, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "rules/create", opt)
	if err != nil {
		return
	}
	v = new(RuleCreateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type RulesDeleteOption struct {
	Key string `url:"key,omitempty"` // Description:"Rule key",ExampleValue:"squid:XPath_1402065390816"
}

// Delete Delete custom rule.<br/>Requires the 'Administer Quality Profiles' permission
func (s *RulesService) Delete(opt *RulesDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "rules/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type RulesRepositoriesOption struct {
	Language string `url:"language,omitempty"` // Description:"A language key; if provided, only repositories for the given language will be returned",ExampleValue:"java"
	Q        string `url:"q,omitempty"`        // Description:"A pattern to match repository keys/names against",ExampleValue:"squid"
}

// Repositories List available rule repositories
func (s *RulesService) Repositories(opt *RulesRepositoriesOption) (v *RulesRepositoriesObject, resp *http.Response, err error) {
	err = s.ValidateRepositoriesOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "rules/repositories", opt)
	if err != nil {
		return
	}
	v = new(RulesRepositoriesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type RulesSearchOption struct {
	Activation       string `url:"activation,omitempty"`        // Description:"Filter rules that are activated or deactivated on the selected Quality profile. Ignored if the parameter 'qprofile' is not set.",ExampleValue:""
	ActiveSeverities string `url:"active_severities,omitempty"` // Description:"Comma-separated list of activation severities, i.e the severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Asc              string `url:"asc,omitempty"`               // Description:"Ascending sort",ExampleValue:""
	AvailableSince   string `url:"available_since,omitempty"`   // Description:"Filters rules added since date. Format is yyyy-MM-dd",ExampleValue:"2014-06-22"
	Facets           string `url:"facets,omitempty"`            // Description:"Comma-separated list of the facets to be computed. No facet is computed by default.",ExampleValue:"languages,repositories"
	Inheritance      string `url:"inheritance,omitempty"`       // Description:"Comma-separated list of values of inheritance for a rule within a quality profile. Used only if the parameter 'activation' is set.",ExampleValue:"INHERITED,OVERRIDES"
	IsTemplate       string `url:"is_template,omitempty"`       // Description:"Filter template rules",ExampleValue:""
	Languages        string `url:"languages,omitempty"`         // Description:"Comma-separated list of languages",ExampleValue:"java,js"
	P                string `url:"p,omitempty"`                 // Description:"1-based page number",ExampleValue:"42"
	Ps               string `url:"ps,omitempty"`                // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	Q                string `url:"q,omitempty"`                 // Description:"UTF-8 search query",ExampleValue:"xpath"
	Qprofile         string `url:"qprofile,omitempty"`          // Description:"Quality profile key to filter on. Used only if the parameter 'activation' is set.",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Repositories     string `url:"repositories,omitempty"`      // Description:"Comma-separated list of repositories",ExampleValue:"checkstyle,findbugs"
	RuleKey          string `url:"rule_key,omitempty"`          // Description:"Key of rule to search for",ExampleValue:"squid:S001"
	S                string `url:"s,omitempty"`                 // Description:"Sort field",ExampleValue:"name"
	Severities       string `url:"severities,omitempty"`        // Description:"Comma-separated list of default severities. Not the same than severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Statuses         string `url:"statuses,omitempty"`          // Description:"Comma-separated list of status codes",ExampleValue:"READY"
	Tags             string `url:"tags,omitempty"`              // Description:"Comma-separated list of tags. Returned rules match any of the tags (OR operator)",ExampleValue:"security,java8"
	TemplateKey      string `url:"template_key,omitempty"`      // Description:"Key of the template rule to filter on. Used to search for the custom rules based on this template.",ExampleValue:"java:S001"
	Types            string `url:"types,omitempty"`             // Description:"Comma-separated list of types. Returned rules match any of the tags (OR operator)",ExampleValue:"BUG"
}

// Search Search for a collection of relevant rules matching a specified query.<br/>Since 5.5, following fields in the response have been deprecated :<ul><li>"effortToFixDescription" becomes "gapDescription"</li><li>"debtRemFnCoeff" becomes "remFnGapMultiplier"</li><li>"defaultDebtRemFnCoeff" becomes "defaultRemFnGapMultiplier"</li><li>"debtRemFnOffset" becomes "remFnBaseEffort"</li><li>"defaultDebtRemFnOffset" becomes "defaultRemFnBaseEffort"</li><li>"debtOverloaded" becomes "remFnOverloaded"</li></ul>
func (s *RulesService) Search(opt *RulesSearchOption) (v *RulesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "rules/search", opt)
	if err != nil {
		return
	}
	v = new(RulesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type RulesShowOption struct {
	Actives string `url:"actives,omitempty"` // Description:"Show rule's activations for all profiles ("active rules")",ExampleValue:""
	Key     string `url:"key,omitempty"`     // Description:"Rule key",ExampleValue:"javascript:EmptyBlock"
}

// Show Get detailed information about a rule<br>Since 5.5, following fields in the response have been deprecated :<ul><li>"effortToFixDescription" becomes "gapDescription"</li><li>"debtRemFnCoeff" becomes "remFnGapMultiplier"</li><li>"defaultDebtRemFnCoeff" becomes "defaultRemFnGapMultiplier"</li><li>"debtRemFnOffset" becomes "remFnBaseEffort"</li><li>"defaultDebtRemFnOffset" becomes "defaultRemFnBaseEffort"</li><li>"debtOverloaded" becomes "remFnOverloaded"</li></ul>In 7.1, the field 'scope' has been added.
func (s *RulesService) Show(opt *RulesShowOption) (v *RulesShowObject, resp *http.Response, err error) {
	err = s.ValidateShowOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "rules/show", opt)
	if err != nil {
		return
	}
	v = new(RulesShowObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type RulesTagsOption struct {
	Ps int    `url:"ps,omitempty"` // Description:"Page size. Must be greater than 0 and less or equal than 100",ExampleValue:"20"
	Q  string `url:"q,omitempty"`  // Description:"Limit search to tags that contain the supplied string.",ExampleValue:"misra"
}

// Tags List rule tags
func (s *RulesService) Tags(opt *RulesTagsOption) (v *IssuesTagsObject, resp *http.Response, err error) {
	err = s.ValidateTagsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "rules/tags", opt)
	if err != nil {
		return
	}
	v = new(IssuesTagsObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type RulesUpdateOption struct {
	DebtRemediationFnOffset    string `url:"debt_remediation_fn_offset,omitempty"`    // Description:"",ExampleValue:""
	DebtRemediationFnType      string `url:"debt_remediation_fn_type,omitempty"`      // Description:"",ExampleValue:""
	DebtRemediationFyCoeff     string `url:"debt_remediation_fy_coeff,omitempty"`     // Description:"",ExampleValue:""
	DebtSubCharacteristic      string `url:"debt_sub_characteristic,omitempty"`       // Description:"Debt characteristics are no more supported. This parameter is ignored.",ExampleValue:""
	Key                        string `url:"key,omitempty"`                           // Description:"Key of the rule to update",ExampleValue:"javascript:NullCheck"
	MarkdownDescription        string `url:"markdown_description,omitempty"`          // Description:"Rule description (mandatory for custom rule and manual rule)",ExampleValue:"Description of my custom rule"
	MarkdownNote               string `url:"markdown_note,omitempty"`                 // Description:"Optional note in markdown format. Use empty value to remove current note. Note is not changedif the parameter is not set.",ExampleValue:"my *note*"
	Name                       string `url:"name,omitempty"`                          // Description:"Rule name (mandatory for custom rule)",ExampleValue:"My custom rule"
	Params                     string `url:"params,omitempty"`                        // Description:"Parameters as semi-colon list of <key>=<value>, for example 'params=key1=v1;key2=v2' (Only when updating a custom rule)",ExampleValue:""
	RemediationFnBaseEffort    string `url:"remediation_fn_base_effort,omitempty"`    // Description:"Base effort of the remediation function of the rule",ExampleValue:"1d"
	RemediationFnType          string `url:"remediation_fn_type,omitempty"`           // Description:"Type of the remediation function of the rule",ExampleValue:""
	RemediationFyGapMultiplier string `url:"remediation_fy_gap_multiplier,omitempty"` // Description:"Gap multiplier of the remediation function of the rule",ExampleValue:"3min"
	Severity                   string `url:"severity,omitempty"`                      // Description:"Rule severity (Only when updating a custom rule)",ExampleValue:""
	Status                     string `url:"status,omitempty"`                        // Description:"Rule status (Only when updating a custom rule)",ExampleValue:""
	Tags                       string `url:"tags,omitempty"`                          // Description:"Optional comma-separated list of tags to set. Use blank value to remove current tags. Tags are not changed if the parameter is not set.",ExampleValue:"java8,security"
}

// Update Update an existing rule.<br>Requires the 'Administer Quality Profiles' permission
func (s *RulesService) Update(opt *RulesUpdateOption) (v *RuleUpdateObject, resp *http.Response, err error) {
	err = s.ValidateUpdateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "rules/update", opt)
	if err != nil {
		return
	}
	v = new(RuleUpdateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

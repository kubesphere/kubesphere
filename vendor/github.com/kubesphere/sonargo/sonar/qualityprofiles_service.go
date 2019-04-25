// Manage quality profiles.
package sonargo

import (
	"net/http"
)

type QualityProfilesService struct {
	client *Client
}
type QualityProfile struct {
	Actions                   *Actions `json:"actions,omitempty"`
	ActiveRuleCount           int64    `json:"activeRuleCount,omitempty"`
	IsBuiltIn                 bool     `json:"isBuiltIn,omitempty"`
	IsDefault                 bool     `json:"isDefault,omitempty"`
	IsInherited               bool     `json:"isInherited,omitempty"`
	Key                       string   `json:"key,omitempty"`
	Language                  string   `json:"language,omitempty"`
	LanguageName              string   `json:"languageName,omitempty"`
	Name                      string   `json:"name,omitempty"`
	OverridingRuleCount       int64    `json:"overridingRuleCount,omitempty"`
	Organization              string   `json:"organization,omitempty"`
	Parent                    string   `json:"parent,omitempty"`
	ParentKey                 string   `json:"parentKey,omitempty"`
	ActiveDeprecatedRuleCount int      `json:"activeDeprecatedRuleCount,omitempty"`
	RulesUpdatedAt            string   `json:"rulesUpdatedAt,omitempty"`
	UserUpdatedAt             string   `json:"userUpdatedAt,omitempty"`
	LastUsed                  string   `json:"lastUsed,omitempty"`
	ProjectCount              int      `json:"projectCount,omitempty"`
}

type QualityProfilesChangelogObject struct {
	Events []*QualityProfilesEvent `json:"events,omitempty"`
	P      int64                   `json:"p,omitempty"`
	Ps     int64                   `json:"ps,omitempty"`
	Total  int64                   `json:"total,omitempty"`
}

type QualityProfilesEvent struct {
	Action      string      `json:"action,omitempty"`
	AuthorLogin string      `json:"authorLogin,omitempty"`
	AuthorName  string      `json:"authorName,omitempty"`
	Date        string      `json:"date,omitempty"`
	Params      interface{} `json:"params,omitempty"`
	RuleKey     string      `json:"ruleKey,omitempty"`
	RuleName    string      `json:"ruleName,omitempty"`
}

type QualityProfilesActiveRulesObject struct {
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Errors    []struct {
		Msg string `json:"msg"`
	} `json:"errors"`
}

type QualityProfilesDeactiveRulesObject QualityProfilesActiveRulesObject

type QualityProfilesCreateObject struct {
	Profile  *QualityProfile `json:"profile,omitempty"`
	Warnings []string        `json:"warnings,omitempty"`
}

type ExportImporter struct {
	Key       string   `json:"key,omitempty"`
	Name      string   `json:"name,omitempty"`
	Languages []string `json:"languages,omitempty"`
}
type QualityProfilesExportersObject struct {
	Exporters []*ExportImporter `json:"exporters,omitempty"`
}

type QualityProfilesImportersObject struct {
	Importers []*ExportImporter `json:"importers,omitempty"`
}

type QualityProfilesInheritanceObject struct {
	Ancestors []*QualityProfile `json:"ancestors,omitempty"`
	Children  []*QualityProfile `json:"children,omitempty"`
	Profile   QualityProfile    `json:"profile,omitempty"`
}

type QualityProfilesSearchObject struct {
	Actions  *Actions          `json:"actions,omitempty"`
	Profiles []*QualityProfile `json:"profiles,omitempty"`
}

type QualityProfilesActivateRuleOption struct {
	Key      string `url:"key,omitempty"`      // Description:"Quality Profile key. Can be obtained through <code>api/qualityprofiles/search</code>",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Params   string `url:"params,omitempty"`   // Description:"Parameters as semi-colon list of <code>key=value</code>. Ignored if parameter reset is true.",ExampleValue:"params=key1=v1;key2=v2"
	Reset    string `url:"reset,omitempty"`    // Description:"Reset severity and parameters of activated rule. Set the values defined on parent profile or from rule default values.",ExampleValue:""
	Rule     string `url:"rule,omitempty"`     // Description:"Rule key",ExampleValue:"squid:AvoidCycles"
	Severity string `url:"severity,omitempty"` // Description:"Severity. Ignored if parameter reset is true.",ExampleValue:""
}

// ActivateRule Activate a rule on a Quality Profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) ActivateRule(opt *QualityProfilesActivateRuleOption) (resp *http.Response, err error) {
	err = s.ValidateActivateRuleOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/activate_rule", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesActivateRulesOption struct {
	Activation       string `url:"activation,omitempty"`        // Description:"Filter rules that are activated or deactivated on the selected Quality profile. Ignored if the parameter 'qprofile' is not set.",ExampleValue:""
	ActiveSeverities string `url:"active_severities,omitempty"` // Description:"Comma-separated list of activation severities, i.e the severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Asc              string `url:"asc,omitempty"`               // Description:"Ascending sort",ExampleValue:""
	AvailableSince   string `url:"available_since,omitempty"`   // Description:"Filters rules added since date. Format is yyyy-MM-dd",ExampleValue:"2014-06-22"
	Inheritance      string `url:"inheritance,omitempty"`       // Description:"Comma-separated list of values of inheritance for a rule within a quality profile. Used only if the parameter 'activation' is set.",ExampleValue:"INHERITED,OVERRIDES"
	IsTemplate       string `url:"is_template,omitempty"`       // Description:"Filter template rules",ExampleValue:""
	Languages        string `url:"languages,omitempty"`         // Description:"Comma-separated list of languages",ExampleValue:"java,js"
	Q                string `url:"q,omitempty"`                 // Description:"UTF-8 search query",ExampleValue:"xpath"
	Qprofile         string `url:"qprofile,omitempty"`          // Description:"Quality profile key to filter on. Used only if the parameter 'activation' is set.",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Repositories     string `url:"repositories,omitempty"`      // Description:"Comma-separated list of repositories",ExampleValue:"checkstyle,findbugs"
	RuleKey          string `url:"rule_key,omitempty"`          // Description:"Key of rule to search for",ExampleValue:"squid:S001"
	S                string `url:"s,omitempty"`                 // Description:"Sort field",ExampleValue:"name"
	Severities       string `url:"severities,omitempty"`        // Description:"Comma-separated list of default severities. Not the same than severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Statuses         string `url:"statuses,omitempty"`          // Description:"Comma-separated list of status codes",ExampleValue:"READY"
	Tags             string `url:"tags,omitempty"`              // Description:"Comma-separated list of tags. Returned rules match any of the tags (OR operator)",ExampleValue:"security,java8"
	TargetKey        string `url:"targetKey,omitempty"`         // Description:"Quality Profile key on which the rule activation is done. To retrieve a quality profile key please see <code>api/qualityprofiles/search</code>",ExampleValue:"AU-TpxcA-iU5OvuD2FL0"
	TargetSeverity   string `url:"targetSeverity,omitempty"`    // Description:"Severity to set on the activated rules",ExampleValue:""
	TemplateKey      string `url:"template_key,omitempty"`      // Description:"Key of the template rule to filter on. Used to search for the custom rules based on this template.",ExampleValue:"java:S001"
	Types            string `url:"types,omitempty"`             // Description:"Comma-separated list of types. Returned rules match any of the tags (OR operator)",ExampleValue:"BUG"
}

// ActivateRules Bulk-activate rules on one quality profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) ActivateRules(opt *QualityProfilesActivateRulesOption) (v *QualityProfilesActiveRulesObject, resp *http.Response, err error) {
	err = s.ValidateActivateRulesOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/activate_rules", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesActiveRulesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type QualityProfilesAddProjectOption struct {
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:""
	Project        string `url:"project,omitempty"`        // Description:"Project key",ExampleValue:"my_project"
	ProjectUuid    string `url:"projectUuid,omitempty"`    // Description:"Project ID. Either this parameter or 'project' must be set.",ExampleValue:"AU-TpxcA-iU5OvuD2FL5"
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"Sonar way"
	Organization   string `json:"organization,omitempty"`
}

// AddProject Associate a project with a quality profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li>  <li>Administer right on the specified project</li></ul>
func (s *QualityProfilesService) AddProject(opt *QualityProfilesAddProjectOption) (resp *http.Response, err error) {
	err = s.ValidateAddProjectOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/add_project", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesBackupOption struct {
	ProfileKey string `url:"profileKey,omitempty"` // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// Backup Backup a quality profile in XML form. The exported profile can be restored through api/qualityprofiles/restore.
func (s *QualityProfilesService) Backup(opt *QualityProfilesBackupOption) (v *string, resp *http.Response, err error) {
	err = s.ValidateBackupOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/backup", opt)
	if err != nil {
		return
	}
	v = new(string)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	// b := new(bytes.Buffer)
	// b.ReadFrom(resp.Body)
	// *v = b.String()
	return
}

type QualityProfilesChangeParentOption struct {
	Key                  string `url:"key,omitempty"`                  // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language             string `url:"language,omitempty"`             // Description:"Quality profile language",ExampleValue:""
	ParentKey            string `url:"parentKey,omitempty"`            // Description:"New parent profile key.<br> If no profile is provided, the inheritance link with current parent profile (if any) is broken, which deactivates all rules which come from the parent and are not overridden.",ExampleValue:"AU-TpxcA-iU5OvuD2FLz"
	ParentQualityProfile string `url:"parentQualityProfile,omitempty"` // Description:"Quality profile name. If this parameter is set, 'parentKey' must not be set and 'language' must be set to disambiguate.",ExampleValue:"Sonar way"
	QualityProfile       string `url:"qualityProfile,omitempty"`       // Description:"Quality profile name",ExampleValue:"Sonar way"
}

// ChangeParent Change a quality profile's parent.<br>Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) ChangeParent(opt *QualityProfilesChangeParentOption) (resp *http.Response, err error) {
	err = s.ValidateChangeParentOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/change_parent", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesChangelogOption struct {
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:""
	P              string `url:"p,omitempty"`              // Description:"1-based page number",ExampleValue:"42"
	Ps             string `url:"ps,omitempty"`             // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"Sonar way"
	Since          string `url:"since,omitempty"`          // Description:"Start date for the changelog. <br>Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	To             string `url:"to,omitempty"`             // Description:"End date for the changelog. <br>Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
}

// Changelog Get the history of changes on a quality profile: rule activation/deactivation, change in parameters/severity. Events are ordered by date in descending order (most recent first).
func (s *QualityProfilesService) Changelog(opt *QualityProfilesChangelogOption) (v *QualityProfilesChangelogObject, resp *http.Response, err error) {
	err = s.ValidateChangelogOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/changelog", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesChangelogObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesCopyOption struct {
	FromKey string `url:"fromKey,omitempty"` // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	ToName  string `url:"toName,omitempty"`  // Description:"Name for the new quality profile.",ExampleValue:"My Sonar way"
}

// Copy Copy a quality profile.<br> Requires to be logged in and the 'Administer Quality Profiles' permission.
func (s *QualityProfilesService) Copy(opt *QualityProfilesCopyOption) (v *QualityProfile, resp *http.Response, err error) {
	err = s.ValidateCopyOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/copy", opt)
	if err != nil {
		return
	}
	v = new(QualityProfile)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesCreateOption struct {
	BackupSonarlintVsCsFake string `url:"backup_sonarlint-vs-cs-fake,omitempty"` // Description:"A configuration file for Technical importer for the MSBuild SonarQube Scanner.",ExampleValue:""
	Language                string `url:"language,omitempty"`                    // Description:"Quality profile language",ExampleValue:"js"
	Name                    string `url:"name,omitempty"`                        // Description:"Quality profile name",ExampleValue:"My Sonar way"
	Organization            string `url:"organization,omitempty"`
}

// Create Create a quality profile.<br>Requires to be logged in and the 'Administer Quality Profiles' permission.
func (s *QualityProfilesService) Create(opt *QualityProfilesCreateOption) (v *QualityProfilesCreateObject, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/create", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesCreateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesDeactivateRuleOption struct {
	Key  string `url:"key,omitempty"`  // Description:"Quality Profile key. Can be obtained through <code>api/qualityprofiles/search</code>",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Rule string `url:"rule,omitempty"` // Description:"Rule key",ExampleValue:"squid:AvoidCycles"
}

// DeactivateRule Deactivate a rule on a quality profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) DeactivateRule(opt *QualityProfilesDeactivateRuleOption) (resp *http.Response, err error) {
	err = s.ValidateDeactivateRuleOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/deactivate_rule", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesDeactivateRulesOption struct {
	Activation       string `url:"activation,omitempty"`        // Description:"Filter rules that are activated or deactivated on the selected Quality profile. Ignored if the parameter 'qprofile' is not set.",ExampleValue:""
	ActiveSeverities string `url:"active_severities,omitempty"` // Description:"Comma-separated list of activation severities, i.e the severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Asc              string `url:"asc,omitempty"`               // Description:"Ascending sort",ExampleValue:""
	AvailableSince   string `url:"available_since,omitempty"`   // Description:"Filters rules added since date. Format is yyyy-MM-dd",ExampleValue:"2014-06-22"
	Inheritance      string `url:"inheritance,omitempty"`       // Description:"Comma-separated list of values of inheritance for a rule within a quality profile. Used only if the parameter 'activation' is set.",ExampleValue:"INHERITED,OVERRIDES"
	IsTemplate       string `url:"is_template,omitempty"`       // Description:"Filter template rules",ExampleValue:""
	Languages        string `url:"languages,omitempty"`         // Description:"Comma-separated list of languages",ExampleValue:"java,js"
	Q                string `url:"q,omitempty"`                 // Description:"UTF-8 search query",ExampleValue:"xpath"
	Qprofile         string `url:"qprofile,omitempty"`          // Description:"Quality profile key to filter on. Used only if the parameter 'activation' is set.",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Repositories     string `url:"repositories,omitempty"`      // Description:"Comma-separated list of repositories",ExampleValue:"checkstyle,findbugs"
	RuleKey          string `url:"rule_key,omitempty"`          // Description:"Key of rule to search for",ExampleValue:"squid:S001"
	S                string `url:"s,omitempty"`                 // Description:"Sort field",ExampleValue:"name"
	Severities       string `url:"severities,omitempty"`        // Description:"Comma-separated list of default severities. Not the same than severity of rules in Quality profiles.",ExampleValue:"CRITICAL,BLOCKER"
	Statuses         string `url:"statuses,omitempty"`          // Description:"Comma-separated list of status codes",ExampleValue:"READY"
	Tags             string `url:"tags,omitempty"`              // Description:"Comma-separated list of tags. Returned rules match any of the tags (OR operator)",ExampleValue:"security,java8"
	TargetKey        string `url:"targetKey,omitempty"`         // Description:"Quality Profile key on which the rule deactivation is done. To retrieve a profile key please see <code>api/qualityprofiles/search</code>",ExampleValue:"AU-TpxcA-iU5OvuD2FL1"
	TemplateKey      string `url:"template_key,omitempty"`      // Description:"Key of the template rule to filter on. Used to search for the custom rules based on this template.",ExampleValue:"java:S001"
	Types            string `url:"types,omitempty"`             // Description:"Comma-separated list of types. Returned rules match any of the tags (OR operator)",ExampleValue:"BUG"
}

// DeactivateRules Bulk deactivate rules on Quality profiles.<br>Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) DeactivateRules(opt *QualityProfilesDeactivateRulesOption) (v *QualityProfilesDeactiveRulesObject, resp *http.Response, err error) {
	err = s.ValidateDeactivateRulesOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/deactivate_rules", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesDeactiveRulesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type QualityProfilesDeleteOption struct {
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:""
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"Sonar way"
}

// Delete Delete a quality profile and all its descendants. The default quality profile cannot be deleted.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) Delete(opt *QualityProfilesDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesExportOption struct {
	ExporterKey    string `url:"exporterKey,omitempty"`    // Description:"Output format. If left empty, the same format as api/qualityprofiles/backup is used. Possible values are described by api/qualityprofiles/exporters.",ExampleValue:""
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:"py"
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name to export. If left empty, the default profile for the language is exported.",ExampleValue:"My Sonar way"
}

// Export Export a quality profile.
func (s *QualityProfilesService) Export(opt *QualityProfilesExportOption) (v *string, resp *http.Response, err error) {
	err = s.ValidateExportOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/export", opt)
	if err != nil {
		return
	}
	v = new(string)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Exporters Lists available profile export formats.
func (s *QualityProfilesService) Exporters() (v *QualityProfilesExportersObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "qualityprofiles/exporters", nil)
	if err != nil {
		return
	}
	v = new(QualityProfilesExportersObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Importers List supported importers.
func (s *QualityProfilesService) Importers() (v *QualityProfilesImportersObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "qualityprofiles/importers", nil)
	if err != nil {
		return
	}
	v = new(QualityProfilesImportersObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesInheritanceOption struct {
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:""
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"Sonar way"
}

// Inheritance Show a quality profile's ancestors and children.
func (s *QualityProfilesService) Inheritance(opt *QualityProfilesInheritanceOption) (v *QualityProfilesInheritanceObject, resp *http.Response, err error) {
	err = s.ValidateInheritanceOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/inheritance", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesInheritanceObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesProjectsOption struct {
	Key      string `url:"key,omitempty"`      // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	P        int    `url:"p,omitempty"`        // Description:"1-based page number",ExampleValue:"42"
	Ps       int    `url:"ps,omitempty"`       // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	Q        string `url:"q,omitempty"`        // Description:"Limit search to projects that contain the supplied string.",ExampleValue:"sonar"
	Selected string `url:"selected,omitempty"` // Description:"Depending on the value, show only selected items (selected=selected), deselected items (selected=deselected), or all items with their selection status (selected=all).",ExampleValue:""
}

// Projects List projects with their association status regarding a quality profile
func (s *QualityProfilesService) Projects(opt *QualityProfilesProjectsOption) (v *QualitygatesSearchObject, resp *http.Response, err error) {
	err = s.ValidateProjectsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/projects", opt)
	if err != nil {
		return
	}
	v = new(QualitygatesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesRemoveProjectOption QualityProfilesAddProjectOption

// RemoveProject Remove a project's association with a quality profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li>  <li>Administer right on the specified project</li></ul>
func (s *QualityProfilesService) RemoveProject(opt *QualityProfilesRemoveProjectOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveProjectOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/remove_project", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesRenameOption struct {
	Key  string `url:"key,omitempty"`  // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Name string `url:"name,omitempty"` // Description:"New quality profile name",ExampleValue:"My Sonar way"
}

// Rename Rename a quality profile.<br> Requires one of the following permissions:<ul>  <li>'Administer Quality Profiles'</li>  <li>Edit right on the specified quality profile</li></ul>
func (s *QualityProfilesService) Rename(opt *QualityProfilesRenameOption) (resp *http.Response, err error) {
	err = s.ValidateRenameOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/rename", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesRestoreOption struct {
	Backup string `url:"backup,omitempty"` // Description:"A profile backup file in XML format, as generated by api/qualityprofiles/backup or the former api/profiles/backup.",ExampleValue:""
}

// Restore Restore a quality profile using an XML file. The restored profile name is taken from the backup file, so if a profile with the same name and language already exists, it will be overwritten.<br> Requires to be logged in and the 'Administer Quality Profiles' permission.
func (s *QualityProfilesService) Restore(opt *QualityProfilesRestoreOption) (resp *http.Response, err error) {
	err = s.ValidateRestoreOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/restore", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualityProfilesSearchOption struct {
	Defaults       string `url:"defaults,omitempty"`       // Description:"If set to true, return only the quality profiles marked as default for each language",ExampleValue:""
	Language       string `url:"language,omitempty"`       // Description:"Language key. If provided, only profiles for the given language are returned.",ExampleValue:""
	Project        string `url:"project,omitempty"`        // Description:"Project key",ExampleValue:"my_project"
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"SonarQube Way"
}

// Search Search quality profiles
func (s *QualityProfilesService) Search(opt *QualityProfilesSearchOption) (v *QualityProfilesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualityprofiles/search", opt)
	if err != nil {
		return
	}
	v = new(QualityProfilesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualityProfilesSetDefaultOption struct {
	Key            string `url:"key,omitempty"`            // Description:"Quality profile key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Language       string `url:"language,omitempty"`       // Description:"Quality profile language",ExampleValue:""
	QualityProfile string `url:"qualityProfile,omitempty"` // Description:"Quality profile name",ExampleValue:"Sonar way"
}

// SetDefault Select the default profile for a given language.<br> Requires to be logged in and the 'Administer Quality Profiles' permission.
func (s *QualityProfilesService) SetDefault(opt *QualityProfilesSetDefaultOption) (resp *http.Response, err error) {
	err = s.ValidateSetDefaultOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualityprofiles/set_default", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

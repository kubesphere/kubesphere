// Manage the plugins on the server, including installing, uninstalling, and upgrading.
package sonargo

import "net/http"

type PluginsService struct {
	client *Client
}

type Plugin struct {
	Category              string    `json:"category,omitempty"`
	Description           string    `json:"description,omitempty"`
	EditionBundled        bool      `json:"editionBundled,omitempty"`
	Filename              string    `json:"filename,omitempty"`
	Hash                  string    `json:"hash,omitempty"`
	HomepageURL           string    `json:"homepageUrl,omitempty"`
	ImplementationBuild   string    `json:"implementationBuild,omitempty"`
	IssueTrackerURL       string    `json:"issueTrackerUrl,omitempty"`
	Key                   string    `json:"key,omitempty"`
	License               string    `json:"license,omitempty"`
	Name                  string    `json:"name,omitempty"`
	OrganizationName      string    `json:"organizationName,omitempty"`
	OrganizationURL       string    `json:"organizationUrl,omitempty"`
	Release               *Release  `json:"release,omitempty"`
	SonarLintSupported    bool      `json:"sonarLintSupported,omitempty"`
	TermsAndConditionsURL string    `json:"termsAndConditionsUrl,omitempty"`
	Update                *Update   `json:"update,omitempty"`
	Updates               []*Update `json:"updates,omitempty"`
	UpdatedAt             int64     `json:"updatedAt,omitempty"`
	Version               string    `json:"version,omitempty"`
}

type PluginsAvailableObject struct {
	Plugins             []*Plugin `json:"plugins,omitempty"`
	UpdateCenterRefresh string    `json:"updateCenterRefresh,omitempty"`
}

type Release struct {
	Date         string `json:"date,omitempty"`
	Version      string `json:"version,omitempty"`
	Description  string `json:"description,omitempty"`
	ChangeLogURL string `json:"changeLogUrl,omitempty"`
}

type Require struct {
	Description string `json:"description,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
}

type Update struct {
	Release  *Release   `json:"release,omitempty"`
	Requires []*Require `json:"requires,omitempty"`
	Status   string     `json:"status,omitempty"`
}

type PluginsInstalledObject struct {
	Plugins []*Plugin `json:"plugins,omitempty"`
}

type PluginsPendingObject struct {
	Installing []*Plugin `json:"installing,omitempty"`
	Removing   []*Plugin `json:"removing,omitempty"`
	Updating   []*Plugin `json:"updating,omitempty"`
}

type PluginsUpdatesObject PluginsAvailableObject

// Available Get the list of all the plugins available for installation on the SonarQube instance, sorted by plugin name.<br/>Plugin information is retrieved from Update Center. Date and time at which Update Center was last refreshed is provided in the response.<br/>Update status values are: <ul><li>COMPATIBLE: plugin is compatible with current SonarQube instance.</li><li>INCOMPATIBLE: plugin is not compatible with current SonarQube instance.</li><li>REQUIRES_SYSTEM_UPGRADE: plugin requires SonarQube to be upgraded before being installed.</li><li>DEPS_REQUIRE_SYSTEM_UPGRADE: at least one plugin on which the plugin is dependent requires SonarQube to be upgraded.</li></ul>Require 'Administer System' permission.
func (s *PluginsService) Available() (v *PluginsAvailableObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "plugins/available", nil)
	if err != nil {
		return
	}
	v = new(PluginsAvailableObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// CancelAll Cancels any operation pending on any plugin (install, update or uninstall)<br/>Requires user to be authenticated with Administer System permissions
func (s *PluginsService) CancelAll() (resp *http.Response, err error) {
	req, err := s.client.NewRequest("POST", "plugins/cancel_all", nil)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PluginsInstallOption struct {
	Key string `url:"key,omitempty"` // Description:"The key identifying the plugin to install",ExampleValue:""
}

// Install Installs the latest version of a plugin specified by its key.<br/>Plugin information is retrieved from Update Center.<br/>Requires user to be authenticated with Administer System permissions
func (s *PluginsService) Install(opt *PluginsInstallOption) (resp *http.Response, err error) {
	err = s.ValidateInstallOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "plugins/install", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PluginsInstalledOption struct {
	F string `url:"f,omitempty"` // Description:"Comma-separated list of the additional fields to be returned in response. No additional field is returned by default. Possible values are:<ul><li>category - category as defined in the Update Center. A connection to the Update Center is needed</li></lu>",ExampleValue:""
}

// Installed Get the list of all the plugins installed on the SonarQube instance, sorted by plugin name.
func (s *PluginsService) Installed(opt *PluginsInstalledOption) (v *PluginsInstalledObject, resp *http.Response, err error) {
	err = s.ValidateInstalledOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "plugins/installed", opt)
	if err != nil {
		return
	}
	v = new(PluginsInstalledObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Pending Get the list of plugins which will either be installed or removed at the next startup of the SonarQube instance, sorted by plugin name.<br/>Require 'Administer System' permission.
func (s *PluginsService) Pending() (v *PluginsPendingObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "plugins/pending", nil)
	if err != nil {
		return
	}
	v = new(PluginsPendingObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type PluginsUninstallOption struct {
	Key string `url:"key,omitempty"` // Description:"The key identifying the plugin to uninstall",ExampleValue:""
}

// Uninstall Uninstalls the plugin specified by its key.<br/>Requires user to be authenticated with Administer System permissions.
func (s *PluginsService) Uninstall(opt *PluginsUninstallOption) (resp *http.Response, err error) {
	err = s.ValidateUninstallOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "plugins/uninstall", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PluginsUpdateOption struct {
	Key string `url:"key,omitempty"` // Description:"The key identifying the plugin to update",ExampleValue:""
}

// Update Updates a plugin specified by its key to the latest version compatible with the SonarQube instance.<br/>Plugin information is retrieved from Update Center.<br/>Requires user to be authenticated with Administer System permissions
func (s *PluginsService) Update(opt *PluginsUpdateOption) (resp *http.Response, err error) {
	err = s.ValidateUpdateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "plugins/update", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

// Updates Lists plugins installed on the SonarQube instance for which at least one newer version is available, sorted by plugin name.<br/>Each newer version is listed, ordered from the oldest to the newest, with its own update/compatibility status.<br/>Plugin information is retrieved from Update Center. Date and time at which Update Center was last refreshed is provided in the response.<br/>Update status values are: [COMPATIBLE, INCOMPATIBLE, REQUIRES_UPGRADE, DEPS_REQUIRE_UPGRADE].<br/>Require 'Administer System' permission.
func (s *PluginsService) Updates() (v *PluginsUpdatesObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "plugins/updates", nil)
	if err != nil {
		return
	}
	v = new(PluginsUpdatesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

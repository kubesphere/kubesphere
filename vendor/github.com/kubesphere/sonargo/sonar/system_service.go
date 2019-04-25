// Get system details, and perform some management actions, such as restarting, and initiating a database migration (as part of a system upgrade).
package sonargo

import "net/http"

type SystemService struct {
	client *Client
}
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelTrace LogLevel = "TRACE"
)

type Cause struct {
	Message   string `json:"message,omitempty"`
	StartedAt string `json:"startedAt,omitempty"`
	State     string `json:"state,omitempty"`
}

type SystemHealthObject struct {
	Causes []*Cause `json:"causes,omitempty"`
	Health string   `json:"health,omitempty"`
	Nodes  []*Node  `json:"nodes,omitempty"`
}

type Node struct {
	Causes    []*Cause `json:"causes,omitempty"`
	Health    string   `json:"health,omitempty"`
	Host      string   `json:"host,omitempty"`
	Name      string   `json:"name,omitempty"`
	Port      int64    `json:"port,omitempty"`
	StartedAt string   `json:"startedAt,omitempty"`
	Type      string   `json:"type,omitempty"`
}

type SystemStatusObject struct {
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Version string `json:"version,omitempty"`
}

type SystemUpgradesObject struct {
	UpdateCenterRefresh string     `json:"updateCenterRefresh,omitempty"`
	Upgrades            []*Upgrade `json:"upgrades,omitempty"`
}

type Incompatible struct {
	Category         string `json:"category,omitempty"`
	Description      string `json:"description,omitempty"`
	EditionBundled   bool   `json:"editionBundled,omitempty"`
	Key              string `json:"key,omitempty"`
	License          string `json:"license,omitempty"`
	Name             string `json:"name,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	OrganizationURL  string `json:"organizationUrl,omitempty"`
}

type UpgradePlugins struct {
	Incompatible  []*Incompatible  `json:"incompatible,omitempty"`
	RequireUpdate []*RequireUpdate `json:"requireUpdate,omitempty"`
}

type RequireUpdate struct {
	Category              string `json:"category,omitempty"`
	Description           string `json:"description,omitempty"`
	EditionBundled        bool   `json:"editionBundled,omitempty"`
	Key                   string `json:"key,omitempty"`
	License               string `json:"license,omitempty"`
	Name                  string `json:"name,omitempty"`
	OrganizationName      string `json:"organizationName,omitempty"`
	OrganizationURL       string `json:"organizationUrl,omitempty"`
	TermsAndConditionsURL string `json:"termsAndConditionsUrl,omitempty"`
	Version               string `json:"version,omitempty"`
}

type Upgrade struct {
	ChangeLogURL string          `json:"changeLogUrl,omitempty"`
	Description  string          `json:"description,omitempty"`
	DownloadURL  string          `json:"downloadUrl,omitempty"`
	Plugins      *UpgradePlugins `json:"plugins,omitempty"`
	ReleaseDate  string          `json:"releaseDate,omitempty"`
	Version      string          `json:"version,omitempty"`
}

type SystemChangeLogLevelOption struct {
	Level LogLevel `url:"level,omitempty"` // Description:"The new level. Be cautious: DEBUG, and even more TRACE, may have performance impacts.",ExampleValue:""
}

// ChangeLogLevel Temporarily changes level of logs. New level is not persistent and is lost when restarting server. Requires system administration permission.
func (s *SystemService) ChangeLogLevel(opt *SystemChangeLogLevelOption) (resp *http.Response, err error) {
	err = s.ValidateChangeLogLevelOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "system/change_log_level", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

// DbMigrationStatus Display the database migration status of SonarQube.<br/>State values are:<ul><li>NO_MIGRATION: DB is up to date with current version of SonarQube.</li><li>NOT_SUPPORTED: Migration is not supported on embedded databases.</li><li>MIGRATION_RUNNING: DB migration is under go.</li><li>MIGRATION_SUCCEEDED: DB migration has run and has been successful.</li><li>MIGRATION_FAILED: DB migration has run and failed. SonarQube must be restarted in order to retry a DB migration (optionally after DB has been restored from backup).</li><li>MIGRATION_REQUIRED: DB migration is required.</li></ul>
func (s *SystemService) DbMigrationStatus() (v *Cause, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "system/db_migration_status", nil)
	if err != nil {
		return
	}
	v = new(Cause)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Health Provide health status of SonarQube.<p>Require 'Administer System' permission or authentication with passcode</p><p>  <ul> <li>GREEN: SonarQube is fully operational</li> <li>YELLOW: SonarQube is usable, but it needs attention in order to be fully operational</li> <li>RED: SonarQube is not operational</li> </ul></p>
func (s *SystemService) Health() (v *SystemHealthObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "system/health", nil)
	if err != nil {
		return
	}
	v = new(SystemHealthObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type SystemLogsOption struct {
	Process string `url:"process,omitempty"` // Description:"Process to get logs from",ExampleValue:""
}

// Logs Get system logs in plain-text format. Requires system administration permission.
func (s *SystemService) Logs(opt *SystemLogsOption) (v *string, resp *http.Response, err error) {
	err = s.ValidateLogsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "system/logs", opt)
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

// MigrateDb Migrate the database to match the current version of SonarQube.<br/>Sending a POST request to this URL starts the DB migration. It is strongly advised to <strong>make a database backup</strong> before invoking this WS.<br/>State values are:<ul><li>NO_MIGRATION: DB is up to date with current version of SonarQube.</li><li>NOT_SUPPORTED: Migration is not supported on embedded databases.</li><li>MIGRATION_RUNNING: DB migration is under go.</li><li>MIGRATION_SUCCEEDED: DB migration has run and has been successful.</li><li>MIGRATION_FAILED: DB migration has run and failed. SonarQube must be restarted in order to retry a DB migration (optionally after DB has been restored from backup).</li><li>MIGRATION_REQUIRED: DB migration is required.</li></ul>
func (s *SystemService) MigrateDb() (v *Cause, resp *http.Response, err error) {
	req, err := s.client.NewRequest("POST", "system/migrate_db", nil)
	if err != nil {
		return
	}
	v = new(Cause)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Ping Answers "pong" as plain-text
func (s *SystemService) Ping() (v *string, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "system/ping", nil)
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

// Restart Restart server. Require 'Administer System' permission. Perform a full restart of the Web, Search and Compute Engine Servers processes.
func (s *SystemService) Restart() (resp *http.Response, err error) {
	req, err := s.client.NewRequest("POST", "system/restart", nil)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

// Status Get state information about SonarQube.<p>status: the running status <ul> <li>STARTING: SonarQube Web Server is up and serving some Web Services (eg. api/system/status) but initialization is still ongoing</li> <li>UP: SonarQube instance is up and running</li> <li>DOWN: SonarQube instance is up but not running because migration has failed (refer to WS /api/system/migrate_db for details) or some other reason (check logs).</li> <li>RESTARTING: SonarQube instance is still up but a restart has been requested (refer to WS /api/system/restart for details).</li> <li>DB_MIGRATION_NEEDED: database migration is required. DB migration can be started using WS /api/system/migrate_db.</li> <li>DB_MIGRATION_RUNNING: DB migration is running (refer to WS /api/system/migrate_db for details)</li> </ul></p>
func (s *SystemService) Status() (v *SystemStatusObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "system/status", nil)
	if err != nil {
		return
	}
	v = new(SystemStatusObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Upgrades Lists available upgrades for the SonarQube instance (if any) and for each one, lists incompatible plugins and plugins requiring upgrade.<br/>Plugin information is retrieved from Update Center. Date and time at which Update Center was last refreshed is provided in the response.
func (s *SystemService) Upgrades() (v *SystemUpgradesObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "system/upgrades", nil)
	if err != nil {
		return
	}
	v = new(SystemUpgradesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

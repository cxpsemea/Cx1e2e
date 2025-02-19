package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

const (
	MOD_ACCESS      = "AccessAssignment"
	MOD_APPLICATION = "Application"
	MOD_FLAG        = "Flag"
	MOD_GROUP       = "Group"
	MOD_IMPORT      = "Import"
	MOD_PRESET      = "Preset"
	MOD_PROJECT     = "Project"
	MOD_QUERY       = "Query"
	MOD_REPORT      = "Report"
	MOD_RESULT      = "Result"
	MOD_ROLE        = "Role"
	MOD_SCAN        = "Scan"
	MOD_USER        = "User"
	MOD_CLIENT      = "OIDCClient"
)

const (
	OP_CREATE = "Create"
	OP_READ   = "Read"
	OP_UPDATE = "Update"
	OP_DELETE = "Delete"
)

var RepoCreds *regexp.Regexp = regexp.MustCompile(`//(.*)@`)

type EnabledEngines struct {
	SAST   bool
	KICS   bool
	SCA    bool
	APISEC bool
}

func (e EnabledEngines) IsEnabled(engine string) bool {
	requestedEngine := strings.ToLower(engine)
	switch requestedEngine {
	case "sast":
		return e.SAST
	case "sca":
		return e.SCA
	case "kics":
		return e.KICS
	case "apisec":
		return e.APISEC
	}
	return false
}

type CRUDTest struct {
	Test         string     `yaml:"Test"`         // CRUD [create, read, update, delete]
	FailTest     bool       `yaml:"FailTest"`     // is it a negative test
	Flags        []string   `yaml:"FeatureFlags"` // are there specific feature flags needed for this test, with ! for negative-flag-test
	Version      string     `yaml:"Version"`      // is there a specific minimum version for this test, with a ! for "less than this version"
	TestSource   string     // filename
	ForceRun     bool       `yaml:"ForceRun"` // should this test run even if it is unsupported by the backend (unlicensed engine, disabled flag). this is to force a failed test.
	OnFailAction FailAction `yaml:"OnFail"`   // actions to take if this command fails
}

type FailAction struct {
	RetryCount uint     `yaml:"Retries"`     // how many times to retry the action, 0 for none
	RetryDelay uint     `yaml:"RetryDelay"`  // delay (in seconds) between retries
	FailSet    bool     `yaml:"FailTestSet"` // whole test set fails if this test fails (skip remaining tests)
	Commands   []string `yaml:"Commands"`    // command to run when the test fails
}

type AccessAssignmentCRUD struct {
	CRUDTest     `yaml:",inline"`
	EntityType   string   `yaml:"EntityType"`
	EntityName   string   `yaml:"EntityName"`
	ResourceType string   `yaml:"ResourceType"`
	ResourceName string   `yaml:"ResourceName"`
	Roles        []string `yaml:"Roles"`
}

func (o AccessAssignmentCRUD) String() string {
	return fmt.Sprintf("%v %v to access %v %v with roles: %v", o.EntityType, o.EntityName, o.ResourceType, o.ResourceName, strings.Join(o.Roles, ", "))
}

type ApplicationCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string            `yaml:"Name"`
	Groups      []string          `yaml:"Groups"`
	Criticality uint              `yaml:"Criticality"`
	Rules       []ApplicationRule `yaml:"Rules"`
	Tags        []Tag             `yaml:"Tags"`
	Application *Cx1ClientGo.Application
}

func (o ApplicationCRUD) String() string {
	return o.Name
}

type ApplicationRule struct {
	Type  string `yaml:"Type"`
	Value string `yaml:"Value"`
}

func (o ApplicationRule) String() string {
	return fmt.Sprintf("%v: %v", o.Type, o.Value)
}

type CxQLCRUD struct {
	CRUDTest `yaml:",inline"`
	//QueryID       uint64 `yaml:"ID"`
	QueryLanguage string    `yaml:"Language"`
	QueryGroup    string    `yaml:"Group"`
	QueryName     string    `yaml:"Name"`
	Source        string    `yaml:"Source"`
	Scope         CxQLScope `yaml:"Scope"`
	Severity      string    `yaml:"Severity"`
	IsExecutable  bool      `yaml:"IsExecutable"`
	Compile       bool      `yaml:"Compile"`
	DeleteSession bool      `yaml:"DeleteAuditSession"`
	OldAPI        bool      `yaml:"OldAPI"`
	ScopeID       string
	ScopeStr      string
	Query         *Cx1ClientGo.Query
	LastScan      *Cx1ClientGo.Scan
}

func (s CxQLScope) String() string {
	if s.Corp {
		return "Corp"
	} else {
		if s.Application != "" {
			return fmt.Sprintf("App: %v", s.Application)
		} else {
			return fmt.Sprintf("Proj: %v", s.Project)
		}
	}
}

func (o CxQLCRUD) String() string {
	//if o.QueryName != "" {
	return fmt.Sprintf("%v: %v -> %v -> %v", o.Scope.String(), o.QueryLanguage, o.QueryGroup, o.QueryName)
	/*} else {
		return fmt.Sprintf("QueryID#%d", o.QueryID)
	} // */
}

type CxQLScope struct {
	Corp        bool   `yaml:"Tenant"`
	Project     string `yaml:"Project"`
	Application string `yaml:"Application"`
	ProjectID   string `yaml:"-"`
}

type FlagCRUD struct {
	CRUDTest `yaml:",inline"`
	Name     string `yaml:"Name"`
	Parent   string `yaml:"Parent"`
}

func (o FlagCRUD) String() string {
	return fmt.Sprintf("%v set to %v", o.Name, !o.FailTest)
}

type GroupCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string `yaml:"Name"`
	Path        string `yaml:"Path"`
	Parent      string `yaml:"Parent"`
	ParentPath  string `yaml:"ParentPath"`
	ClientRoles []struct {
		Client string   `yaml:"Client"`
		Roles  []string `yaml:"Roles"`
	} `yaml:"ClientRoles"`
	Group *Cx1ClientGo.Group
}

func (o GroupCRUD) String() string {
	if o.Name != "" {
		return o.Name
	}

	return o.Path
}

type ImportCRUD struct {
	CRUDTest       `yaml:",inline"`
	Name           string `yaml:"Name"`
	ZipFile        string `yaml:"ZipFile"`
	EncryptionKey  string `yaml:"EncryptionKey"`
	ProjectMapFile string `yaml:"ProjectMapFile"`
	Parent         string `yaml:"Parent"`
	TimeoutSeconds int    `yaml:"Timeout"`
}

func (o ImportCRUD) String() string {
	if o.TimeoutSeconds == 0 {
		return o.Name
	} else {
		return fmt.Sprintf("%v (%d sec timeout)", o.Name, o.TimeoutSeconds)
	}
}

type OIDCClientCRUD struct {
	CRUDTest `yaml:",inline"`
	Name     string   `yaml:"Name"`
	Groups   []string `yaml:"Groups"`
	Roles    []string `yaml:"Roles"`
	Client   *Cx1ClientGo.OIDCClient
	User     *Cx1ClientGo.User
}

func (o OIDCClientCRUD) String() string {
	if o.Client == nil {
		return "New OIDC Client: " + o.Name
	}
	return o.Client.String()
}

type PresetCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string `yaml:"Name"`
	Description string `yaml:"Description"`
	Queries     []struct {
		QueryID       uint64 `yaml:"ID"`
		QueryLanguage string `yaml:"Language"`
		QueryGroup    string `yaml:"Group"`
		QueryName     string `yaml:"Name"`
	} `yaml:"Queries"`
	Preset *Cx1ClientGo.Preset
}

func (o PresetCRUD) String() string {
	return o.Name
}

type ProjectCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string   `yaml:"Name"`
	Groups      []string `yaml:"Groups"`
	Application string   `yaml:"Application"`
	Tags        []Tag    `yaml:"Tags"`
	Preset      string   `yaml:"Preset"`
	Project     *Cx1ClientGo.Project
}

func (o ProjectCRUD) String() string {
	return o.Name
}

type QueryCRUD struct {
	CRUDTest      `yaml:",inline"`
	QueryID       uint64 `yaml:"ID"`
	QueryLanguage string `yaml:"Language"`
	QueryGroup    string `yaml:"Group"`
	QueryName     string `yaml:"Name"`
	Query         *Cx1ClientGo.Query
}

func (o QueryCRUD) String() string {
	if o.QueryName != "" {
		return fmt.Sprintf("%v -> %v -> %v", o.QueryLanguage, o.QueryGroup, o.QueryName)
	} else {
		return fmt.Sprintf("QueryID#%d", o.QueryID)
	}
}

type ReportCRUD struct {
	CRUDTest    `yaml:",inline"`
	ProjectName string `yaml:"Project"`
	Number      uint   `yaml:"Number"`
	Status      string `yaml:"ScanStatus"`
	Branch      string `yaml:"Branch"`
	Format      string `yaml:"Format"`
	Timeout     int    `yaml:"Timeout"`
	Scan        *Cx1ClientGo.Scan
}

func (o ReportCRUD) String() string {
	filters := []string{}

	if o.Status != "" {
		filters = append(filters, fmt.Sprintf("Scan status: %v", o.Status))
	}

	if o.Branch != "" {
		filters = append(filters, fmt.Sprintf("Branch: %v", o.Branch))
	}

	if len(filters) > 0 {
		return fmt.Sprintf("Report for project %v scan #%d matching filter %v, in %v format", o.ProjectName, o.Number, strings.Join(filters, ", "), o.Format)
	}
	return fmt.Sprintf("Report for project %v scan #%d in %v format", o.ProjectName, o.Number, o.Format)
}

type ResultCRUD struct {
	CRUDTest    `yaml:",inline"`
	ProjectName string           `yaml:"Project"`
	Number      uint64           `yaml:"FindingNumber"`
	State       string           `yaml:"State"`
	Severity    string           `yaml:"Severity"`
	Comment     string           `yaml:"Comment"`
	Type        string           `yaml:"Type"`
	SASTFilter  SASTResultFilter `yaml:"SASTFilter"`
	KICSFilter  KICSResultFilter `yaml:"KICSFilter"`
	SCAFilter   SCAResultFilter  `yaml:"SCAFilter"`
	Results     *Cx1ClientGo.ScanResultSet
	Project     *Cx1ClientGo.Project
}

func (o *ResultCRUD) String() string {
	switch o.Type {
	case "SAST":
		filter := o.SASTFilter.String()
		if filter != "" {
			return fmt.Sprintf("%v: SAST finding #%d matching filter: %v", o.ProjectName, o.Number, filter)
		}
	case "SCA":
		filter := o.SCAFilter.String()
		if filter != "" {
			return fmt.Sprintf("%v: SCA finding #%d matching filter: %v", o.ProjectName, o.Number, filter)
		}
	case "KICS":
		filter := o.KICSFilter.String()
		if filter != "" {
			return fmt.Sprintf("%v: KICS finding #%d matching filter: %v", o.ProjectName, o.Number, filter)
		}
	}
	return fmt.Sprintf("%v: finding #%d", o.ProjectName, o.Number)
}

type ResultFilter struct {
	State        string `yaml:"State"`
	Severity     string `yaml:"Severity"`
	SimilarityID string `yaml:"SimilarityID"`
}

type SASTResultFilter struct {
	ResultFilter  `yaml:",inline"`
	QueryID       string `yaml:"QueryID"`
	QueryLanguage string `yaml:"Language"`
	QueryGroup    string `yaml:"Group"`
	QueryName     string `yaml:"Query"`
	ResultHash    string `yaml:"ResultHash"`
	CweID         int    `yaml:"CweID"`
}

func (o *SASTResultFilter) String() string {
	var filters []string

	if o.QueryID != "" {
		filters = append(filters, fmt.Sprintf("QueryID = %v", o.QueryID))
	}
	if o.QueryLanguage != "" {
		filters = append(filters, fmt.Sprintf("Language = %v", o.QueryLanguage))
	}
	if o.QueryGroup != "" {
		filters = append(filters, fmt.Sprintf("Group = %v", o.QueryGroup))
	}
	if o.QueryName != "" {
		filters = append(filters, fmt.Sprintf("Query = %v", o.QueryName))
	}
	if o.ResultHash != "" {
		filters = append(filters, fmt.Sprintf("ResultHash = %v", o.ResultHash))
	}
	if o.Severity != "" {
		filters = append(filters, fmt.Sprintf("Severity = %v", o.Severity))
	}
	if o.State != "" {
		filters = append(filters, fmt.Sprintf("State = %v", o.State))
	}
	if o.SimilarityID != "" {
		filters = append(filters, fmt.Sprintf("SimilarityID = %v", o.SimilarityID))
	}
	if o.CweID != 0 {
		filters = append(filters, fmt.Sprintf("CweID = %d", o.CweID))
	}

	return strings.Join(filters, ", ")
}

type KICSResultFilter struct {
	ResultFilter `yaml:",inline"`
	QueryID      string `yaml:"QueryID"`
	QueryName    string `yaml:"Name"`
	QueryGroup   string `yaml:"Group"`
}

func (o *KICSResultFilter) String() string {
	var filters []string

	if o.QueryID != "" {
		filters = append(filters, fmt.Sprintf("QueryID = %v", o.QueryID))
	}
	if o.QueryGroup != "" {
		filters = append(filters, fmt.Sprintf("Group = %v", o.QueryGroup))
	}
	if o.QueryName != "" {
		filters = append(filters, fmt.Sprintf("Query = %v", o.QueryName))
	}
	if o.Severity != "" {
		filters = append(filters, fmt.Sprintf("Severity = %v", o.Severity))
	}
	if o.State != "" {
		filters = append(filters, fmt.Sprintf("State = %v", o.State))
	}
	if o.SimilarityID != "" {
		filters = append(filters, fmt.Sprintf("SimilarityID = %v", o.SimilarityID))
	}

	return strings.Join(filters, ", ")
}

type SCAResultFilter struct {
	ResultFilter `yaml:",inline"`
	CveName      string `yaml:"CveName"`
	PackageMatch string `yaml:"PackageMatch"`
}

func (o *SCAResultFilter) String() string {
	var filters []string

	if o.Severity != "" {
		filters = append(filters, fmt.Sprintf("Severity = %v", o.Severity))
	}
	if o.State != "" {
		filters = append(filters, fmt.Sprintf("State = %v", o.State))
	}
	if o.SimilarityID != "" {
		filters = append(filters, fmt.Sprintf("SimilarityID = %v", o.SimilarityID))
	}
	if o.CveName != "" {
		filters = append(filters, fmt.Sprintf("CveName = %v", o.CveName))
	}
	if o.PackageMatch != "" {
		filters = append(filters, fmt.Sprintf("PackageMatch = %v", o.PackageMatch))
	}
	return strings.Join(filters, ", ")
}

type RoleCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string   `yaml:"Name"`
	Permissions []string `yaml:"Permissions"`
	Filter      []string `yaml:"Filter"`
	Role        *Cx1ClientGo.Role
}

func (o RoleCRUD) String() string {
	return o.Name
}

type ScanCRUD struct {
	CRUDTest      `yaml:",inline"`
	Project       string      `yaml:"Project"`
	Branch        string      `yaml:"Branch"`
	Repository    string      `yaml:"Repository"`
	Engine        string      `yaml:"Engine"`
	Incremental   bool        `yaml:"Incremental"`
	WaitForEnd    bool        `yaml:"WaitForEnd"`
	ZipFile       string      `yaml:"ZipFile"`
	Preset        string      `yaml:"Preset"`
	Status        string      `yaml:"Status"`
	Timeout       int         `yaml:"Timeout"`
	Cancel        bool        `yaml:"CancelOnTimeout"`
	Filter        *ScanFilter `yaml:"Filter"`
	Cx1ScanFilter *Cx1ClientGo.ScanFilter
	Scan          *Cx1ClientGo.Scan
}

func (o ScanCRUD) String() string {
	if o.Repository != "" {
		safeRepo := RepoCreds.ReplaceAllString(o.Repository, "//****@")
		return fmt.Sprintf("%v scan of %v: repo %v, branch %v", o.Engine, o.Project, safeRepo, o.Branch)
	} else {
		return fmt.Sprintf("%v scan of %v: zip %v, branch %v", o.Engine, o.Project, o.ZipFile, o.Branch)
	}
}

type ScanFilter struct {
	Index    int      `yaml:"Index"` // which scan are we looking for
	Statuses []string `yaml:"Statuses"`
	Branches []string `yaml:"Branches"`
}

func (f ScanFilter) String() string {
	var str string
	if len(f.Statuses) > 0 {
		str = "with status " + strings.Join(f.Statuses, " or ")
	}
	if len(f.Branches) > 0 {
		if str == "" {
			str = "for branch " + strings.Join(f.Branches, " or ")
		} else {
			str += ", for branch " + strings.Join(f.Branches, " or ")
		}
	}
	return str
}

type UserCRUD struct {
	CRUDTest  `yaml:",inline"`
	Name      string   `yaml:"Name"`
	Email     string   `yaml:"Email"`
	FirstName string   `yaml:"FirstName"`
	LastName  string   `yaml:"LastName"`
	Groups    []string `yaml:"Groups"`
	Roles     []string `yaml:"Roles"`

	User *Cx1ClientGo.User
}

func (o UserCRUD) String() string {
	if o.Email != "" {
		return fmt.Sprintf("%v (%v)", o.Name, o.Email)
	}
	return fmt.Sprintf("%v", o.Name)
}

type Tag struct {
	Key   string `yaml:"Key"`
	Value string `yaml:"Value"`
}

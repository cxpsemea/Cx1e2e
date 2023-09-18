package types

import (
	"fmt"
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
)

type CRUDTest struct {
	Test       string   `yaml:"Test"`         // CRUD [create, read, update, delete]
	FailTest   bool     `yaml:"FailTest"`     // is it a negative test
	Flags      []string `yaml:"FeatureFlags"` // are there specific feature flags needed for this test
	TestSource string   // filename
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
	ScopeID       string
	Query         *Cx1ClientGo.AuditQuery
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
	Parent      string `yaml:"Parent"`
	ClientRoles []struct {
		Client string   `yaml:"Client"`
		Roles  []string `yaml:"Roles"`
	} `yaml:"ClientRoles"`
	Group *Cx1ClientGo.Group
}

func (o GroupCRUD) String() string {
	return o.Name
}

type ImportCRUD struct {
	CRUDTest       `yaml:",inline"`
	Name           string `yaml:"Name"`
	ZipFile        string `yaml:"ZipFile"`
	EncryptionKey  string `yaml:"EncryptionKey"`
	ProjectMapFile string `yaml:"ProjectMapFile"`
	Parent         string `yaml:"Parent"`
}

func (o ImportCRUD) String() string {
	return o.Name
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
	ProjectName string       `yaml:"Project"`
	Number      uint64       `yaml:"FindingNumber"`
	State       string       `yaml:"State"`
	Severity    string       `yaml:"Severity"`
	Comment     string       `yaml:"Comment"`
	Filter      ResultFilter `yaml:"Filter"`
	Result      *Cx1ClientGo.ScanResult
	Project     *Cx1ClientGo.Project
}

func (o *ResultCRUD) String() string {
	filter := o.Filter.String()
	if filter != "" {
		return fmt.Sprintf("%v: finding #%d matching filter: %v", o.ProjectName, o.Number, filter)
	}
	return fmt.Sprintf("%v: finding #%d", o.ProjectName, o.Number)
}

type ResultFilter struct {
	QueryID       uint64 `yaml:"QueryID"`
	QueryLanguage string `yaml:"Language"`
	QueryGroup    string `yaml:"Group"`
	QueryName     string `yaml:"Query"`
	SimilarityID  int64  `yaml:"SimilarityID"`
	ResultHash    string `yaml:"ResultHash"`
	State         string `yaml:"State"`
	Severity      string `yaml:"Severity"`
}

func (o *ResultFilter) String() string {
	var filters []string
	if o.QueryID != 0 {
		filters = append(filters, fmt.Sprintf("QueryID = %d", o.QueryID))
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
	if o.SimilarityID != 0 {
		filters = append(filters, fmt.Sprintf("SimilarityID = %d", o.SimilarityID))
	}

	return strings.Join(filters, ", ")
}

type RoleCRUD struct {
	CRUDTest    `yaml:",inline"`
	Name        string   `yaml:"Name"`
	Permissions []string `yaml:"Permissions"`
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
	Filter        *ScanFilter `yaml:"Filter"`
	Cx1ScanFilter *Cx1ClientGo.ScanFilter
	Scan          *Cx1ClientGo.Scan
}

func (o ScanCRUD) String() string {
	if o.Repository != "" {
		return fmt.Sprintf("%v: repo %v, branch %v", o.Project, o.Repository, o.Branch)
	} else {
		return fmt.Sprintf("%v: zip %v, branch %v", o.Project, o.ZipFile, o.Branch)
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
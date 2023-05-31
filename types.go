package main

import (
	"fmt"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

type ApplicationCRUD struct {
	Name        string            `yaml:"Name"`
	Test        string            `yaml:"Test"`
	Groups      []string          `yaml:"Groups"`
	Criticality uint              `yaml:"Criticality"`
	Rules       []ApplicationRule `yaml:"Rules"`
	Tags        []Tag             `yaml:"Tags"`
	FailTest    bool              `yaml:"FailTest"`
	TestSource  string
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
	//QueryID       uint64 `yaml:"ID"`
	QueryLanguage string    `yaml:"Language"`
	QueryGroup    string    `yaml:"Group"`
	QueryName     string    `yaml:"Name"`
	Test          string    `yaml:"Test"`
	Source        string    `yaml:"Source"`
	Scope         CxQLScope `yaml:"Scope"`
	Severity      string    `yaml:"Severity"`
	IsExecutable  bool      `yaml:"IsExecutable"`
	FailTest      bool      `yaml:"FailTest"`
	Compile       bool      `yaml:"Compile"`
	TestSource    string
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

type GroupCRUD struct {
	Name        string `yaml:"Name"`
	Test        string `yaml:"Test"`
	Parent      string `yaml:"Parent"`
	ClientRoles []struct {
		Client string   `yaml:"Client"`
		Roles  []string `yaml:"Roles"`
	} `yaml:"ClientRoles"`
	FailTest   bool `yaml:"FailTest"`
	TestSource string
	Group      *Cx1ClientGo.Group
}

func (o GroupCRUD) String() string {
	return o.Name
}

type ProjectCRUD struct {
	Name        string   `yaml:"Name"`
	Test        string   `yaml:"Test"`
	Groups      []string `yaml:"Groups"`
	Application string   `yaml:"Application"`
	Tags        []Tag    `yaml:"Tags"`
	FailTest    bool     `yaml:"FailTest"`
	TestSource  string
	Project     *Cx1ClientGo.Project
}

func (o ProjectCRUD) String() string {
	return o.Name
}

type PresetCRUD struct {
	Name        string `yaml:"Name"`
	Description string `yaml:"Description"`
	Test        string `yaml:"Test"`
	Queries     []struct {
		QueryID       uint64 `yaml:"ID"`
		QueryLanguage string `yaml:"Language"`
		QueryGroup    string `yaml:"Group"`
		QueryName     string `yaml:"Name"`
	} `yaml:"Queries"`
	FailTest   bool `yaml:"FailTest"`
	TestSource string
	Preset     *Cx1ClientGo.Preset
}

func (o PresetCRUD) String() string {
	return o.Name
}

type QueryCRUD struct {
	QueryID       uint64 `yaml:"ID"`
	QueryLanguage string `yaml:"Language"`
	QueryGroup    string `yaml:"Group"`
	QueryName     string `yaml:"Name"`
	Test          string `yaml:"Test"`
	FailTest      bool   `yaml:"FailTest"`
	TestSource    string
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
	ProjectName string `yaml:"Project"`
	Number      uint   `yaml:"Number"`
	Status      string `yaml:"ScanStatus"`
	Branch      string `yaml:"Branch"`
	Format      string `yaml:"Format"`
	Test        string `yaml:"Test"`
	FailTest    bool   `yaml:"FailTest"`
	TestSource  string
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
	ProjectName string `yaml:"Project"`
	Number      uint64 `yaml:"FindingNumber"`
	State       string `yaml:"State"`
	Severity    string `yaml:"Severity"`
	Comment     string `yaml:"Comment"`
	Test        string `yaml:"Test"`
	FailTest    bool   `yaml:"FailTest"`
	TestSource  string
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
	Name        string   `yaml:"Name"`
	Test        string   `yaml:"Test"`
	Permissions []string `yaml:"Permissions"`
	FailTest    bool     `yaml:"FailTest"`
	TestSource  string
	Role        *Cx1ClientGo.Role
}

func (o RoleCRUD) String() string {
	return o.Name
}

type ScanCRUD struct {
	Test        string `yaml:"Test"`
	Project     string `yaml:"Project"`
	Branch      string `yaml:"Branch"`
	Repository  string `yaml:"Repository"`
	Engine      string `yaml:"Engine"`
	Incremental bool   `yaml:"Incremental"`
	WaitForEnd  bool   `yaml:"WaitForEnd"`
	ZipFile     string `yaml:"ZipFile"`
	Preset      string `yaml:"Preset"`
	FailTest    bool   `yaml:"FailTest"`
	TestSource  string
	Scan        *Cx1ClientGo.Scan
}

func (o ScanCRUD) String() string {
	if o.Repository != "" {
		return fmt.Sprintf("%v: repo %v, branch %v", o.Project, o.Repository, o.Branch)
	} else {
		return fmt.Sprintf("%v: zip %v, branch %v", o.Project, o.ZipFile, o.Branch)
	}
}

type UserCRUD struct {
	Name       string   `yaml:"Name"`
	Email      string   `yaml:"Email"`
	Test       string   `yaml:"Test"`
	FirstName  string   `yaml:"FirstName"`
	LastName   string   `yaml:"LastName"`
	Groups     []string `yaml:"Groups"`
	Roles      []string `yaml:"Roles"`
	FailTest   bool     `yaml:"FailTest"`
	TestSource string
	User       *Cx1ClientGo.User
}

func (o UserCRUD) String() string {
	return fmt.Sprintf("%v (%v)", o.Name, o.Email)
}

type Tag struct {
	Key   string `yaml:"Key"`
	Value string `yaml:"Value"`
}

type TestSet struct {
	Name         string            `yaml:"Name"`
	File         string            `yaml:"File"`
	Groups       []GroupCRUD       `yaml:"Groups"`
	Users        []UserCRUD        `yaml:"Users"`
	Applications []ApplicationCRUD `yaml:"Applications"`
	Projects     []ProjectCRUD     `yaml:"Projects"`
	Queries      []CxQLCRUD        `yaml:"Queries"`
	Presets      []PresetCRUD      `yaml:"Presets"`
	Roles        []RoleCRUD        `yaml:"Roles"`
	Scans        []ScanCRUD        `yaml:"Scans"`
	Results      []ResultCRUD      `yaml:"Results"`
	Reports      []ReportCRUD      `yaml:"Reports"`
	Wait         uint              `yaml:"Wait"`
}

type TestConfig struct {
	Cx1URL     string    `yaml:"Cx1URL"`
	IAMURL     string    `yaml:"IAMURL"`
	Tenant     string    `yaml:"Tenant"`
	ProxyURL   string    `yaml:"ProxyURL"`
	Tests      []TestSet `yaml:"Tests"`
	LogLevel   string    `yaml:"LogLevel"`
	ConfigPath string    `yaml:"-"`
}

type TestResult struct {
	FailTest   bool
	Result     int
	CRUD       string
	Module     string
	Duration   float64
	Name       string
	Id         int
	TestObject string
	Reason     string
	TestSource string
}

// test result output
type Counter struct {
	Pass uint
	Fail uint
	Skip uint
}
type CounterSet struct {
	Create Counter
	Read   Counter
	Update Counter
	Delete Counter
}

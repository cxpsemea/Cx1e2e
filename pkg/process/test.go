package process

import "github.com/cxpsemea/cx1e2e/pkg/types"

type TestSet struct {
	Name              string                       `yaml:"Name"`
	File              string                       `yaml:"File"`
	AccessAssignments []types.AccessAssignmentCRUD `yaml:"AccessAssignments"`
	Applications      []types.ApplicationCRUD      `yaml:"Applications"`
	Flags             []types.FlagCRUD             `yaml:"Flags"`
	Groups            []types.GroupCRUD            `yaml:"Groups"`
	Imports           []types.ImportCRUD           `yaml:"Imports"`
	Presets           []types.PresetCRUD           `yaml:"Presets"`
	Projects          []types.ProjectCRUD          `yaml:"Projects"`
	Queries           []types.CxQLCRUD             `yaml:"Queries"`
	Reports           []types.ReportCRUD           `yaml:"Reports"`
	Results           []types.ResultCRUD           `yaml:"Results"`
	Roles             []types.RoleCRUD             `yaml:"Roles"`
	Scans             []types.ScanCRUD             `yaml:"Scans"`
	Users             []types.UserCRUD             `yaml:"Users"`

	Wait uint `yaml:"Wait"`
}

type TestConfig struct {
	Cx1URL     string               `yaml:"Cx1URL"`
	IAMURL     string               `yaml:"IAMURL"`
	Tenant     string               `yaml:"Tenant"`
	ProxyURL   string               `yaml:"ProxyURL"`
	Tests      []TestSet            `yaml:"Tests"`
	LogLevel   string               `yaml:"LogLevel"`
	ConfigPath string               `yaml:"-"`
	AuthType   string               `yaml:"-"`
	AuthUser   string               `yaml:"-"`
	ReportType string               `yaml:"ReportType"`
	ReportName string               `yaml:"ReportName"`
	Engines    types.EnabledEngines `yaml:"-"`
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

type ReportSettings struct {
	Target    string `json:"TestTarget"`
	Auth      string `json:"Authentication"`
	Config    string `json:"TestConfig"`
	Timestamp string `json:"ExecutionTime"`
	E2ESuffix string `json:"E2ESuffix"`
}

type ReportSummary struct {
	Total Counter `json:"Total"`
	Area  struct {
		Access      CounterSet
		Application CounterSet
		Flag        CounterSet
		Group       CounterSet
		Import      CounterSet
		Preset      CounterSet
		Project     CounterSet
		Query       CounterSet
		Result      CounterSet
		Report      CounterSet
		Role        CounterSet
		Scan        CounterSet
		User        CounterSet
	} `json:"Area"`
}

type ReportTestDetails struct {
	Name       string
	Source     string
	Test       string
	Duration   float64
	ResultType int `json:"-"`
	Result     string
}

type Report struct {
	Settings ReportSettings      `json:"Settings"`
	Summary  ReportSummary       `json:"Summary"`
	Details  []ReportTestDetails `json:"Details"`
}

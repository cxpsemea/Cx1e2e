package process

import (
	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/cxpsemea/cx1e2e/pkg/types"
)

type TestSet struct {
	Name  string `yaml:"Name"`
	File  string `yaml:"File"`
	RunAs struct {
		APIKey       string `yaml:"APIKey"`
		ClientID     string `yaml:"ClientID"`
		ClientSecret string `yaml:"ClientSecret"`
		OIDCClient   string `yaml:"OIDCClient"`
	} `yaml:"RunAs"`

	AccessAssignments []types.AccessAssignmentCRUD `yaml:"AccessAssignments"`
	Analytics         []types.AnalyticsCRUD        `yaml:"Analytics"`
	Applications      []types.ApplicationCRUD      `yaml:"Applications"`
	Clients           []types.OIDCClientCRUD       `yaml:"OIDCClients"`
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
	Wait              uint                         `yaml:"Wait"`
	Thread            uint                         `yaml:"Thread"`
	ActiveThread      int                          `yaml:"-"`

	SubTests   []TestSet `yaml:"-"`
	TestSource string    `yaml:"-"`
}

type TestConfig struct {
	Cx1URL             string                  `yaml:"Cx1URL"`
	IAMURL             string                  `yaml:"IAMURL"`
	Tenant             string                  `yaml:"Tenant"`
	ProxyURL           string                  `yaml:"ProxyURL"`
	NoTLS              bool                    `yaml:"NoTLS"`
	Tests              []TestSet               `yaml:"Tests"`
	LogLevel           string                  `yaml:"LogLevel"`
	MultiThreadable    bool                    `yaml:"MultiThreadable"`
	ConfigPath         string                  `yaml:"-"`
	AuthType           string                  `yaml:"-"`
	AuthUser           string                  `yaml:"-"`
	ReportType         string                  `yaml:"ReportType"`
	ReportName         string                  `yaml:"ReportName"`
	Engines            types.EnabledEngines    `yaml:"-"`
	EnvironmentVersion Cx1ClientGo.VersionInfo `yaml:"-"`
	TestCount          int                     `yaml:"-"`
}

type TestResult struct {
	FailTest   bool
	Result     int
	CRUD       string
	Module     string
	Duration   float64
	Name       string
	Id         uint
	TestObject string
	Reason     []string
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
	Target    string                  `json:"TestTarget"`
	Auth      string                  `json:"Authentication"`
	Config    string                  `json:"TestConfig"`
	StartTime string                  `json:"StartTime"`
	EndTime   string                  `json:"EndTime"`
	Duration  string                  `json:"Duration"`
	E2ESuffix string                  `json:"E2ESuffix"`
	Threads   int                     `json:"Threads"`
	Version   Cx1ClientGo.VersionInfo `json:"TargetVersions"`
}

type ReportSummary struct {
	Total Counter `json:"Total"`
	Area  struct {
		Access      CounterSet
		Application CounterSet
		Analytics   CounterSet
		Client      CounterSet
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
	Name        string
	Source      string
	Test        string
	Duration    float64
	ResultType  int `json:"-"`
	Result      string
	ID          uint
	FailOutputs []string `json:"FailOutputs,omitempty"`
}

type Report struct {
	Settings ReportSettings      `json:"Settings"`
	Summary  ReportSummary       `json:"Summary"`
	Details  []ReportTestDetails `json:"Details"`
}

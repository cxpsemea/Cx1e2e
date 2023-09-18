package process

import (
	"github.com/cxpsemea/cx1e2e/pkg/types"
)

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

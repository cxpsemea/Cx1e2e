package main

import (
	"strings"
)

type TestSet struct {
	Name              string                 `yaml:"Name"`
	File              string                 `yaml:"File"`
	AccessAssignments []AccessAssignmentCRUD `yaml:"AccessAssignments"`
	Applications      []ApplicationCRUD      `yaml:"Applications"`
	Flags             []FlagCRUD             `yaml:"Flags"`
	Groups            []GroupCRUD            `yaml:"Groups"`
	Imports           []ImportCRUD           `yaml:"Imports"`
	Presets           []PresetCRUD           `yaml:"Presets"`
	Projects          []ProjectCRUD          `yaml:"Projects"`
	Queries           []CxQLCRUD             `yaml:"Queries"`
	Reports           []ReportCRUD           `yaml:"Reports"`
	Results           []ResultCRUD           `yaml:"Results"`
	Roles             []RoleCRUD             `yaml:"Roles"`
	Scans             []ScanCRUD             `yaml:"Scans"`
	Users             []UserCRUD             `yaml:"Users"`

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

func (c CRUDTest) IsNegative() bool {
	return c.FailTest
}

func (c CRUDTest) IsType(CRUD string) bool {
	switch CRUD {
	case OP_CREATE:
		return strings.Contains(c.Test, "C")
	case OP_READ:
		return strings.Contains(c.Test, "R")
	case OP_UPDATE:
		return strings.Contains(c.Test, "U")
	case OP_DELETE:
		return strings.Contains(c.Test, "D")
	}
	return false
}

func (c CRUDTest) GetSource() string {
	return c.TestSource
}

func (c CRUDTest) IsCreate() bool {
	return strings.Contains(c.Test, "C")
}
func (c CRUDTest) IsRead() bool {
	return strings.Contains(c.Test, "R")
}
func (c CRUDTest) IsUpdate() bool {
	return strings.Contains(c.Test, "U")
}
func (c CRUDTest) IsDelete() bool {
	return strings.Contains(c.Test, "D")
}

package process

import (
	"fmt"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/cxpsemea/cx1e2e/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	TST_FAIL = 0
	TST_PASS = 1
	TST_SKIP = 2
)

type TestRunner interface {
	Validate(testType string) error
	String() string
	IsType(testType string) bool
	IsForced() bool
	IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testType string, Engines *types.EnabledEngines) error
	IsNegative() bool
	GetSource() string
	GetModule() string
	GetFlags() []string

	RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *types.EnabledEngines) error
	RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *types.EnabledEngines) error
	RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *types.EnabledEngines) error
	RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *types.EnabledEngines) error
}

func MakeResult(test TestRunner) TestResult {
	return TestResult{
		FailTest:   test.IsNegative(),
		Result:     TST_SKIP,
		Module:     test.GetModule(),
		Duration:   0,
		Id:         -1,
		TestObject: test.String(),
		TestSource: test.GetSource(),
	}
}

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig) float32 {
	all_results := []TestResult{}

	for id := range Config.Tests {
		all_results = append(all_results, Config.Tests[id].RunTests(cx1client, logger, Config)...)
	}

	status, err := GenerateReport(&all_results, logger, Config)
	if err != nil {
		logger.Errorf("Failed to generate the report: %s", err)
	}

	return status
}

func (t *TestSet) RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig) []TestResult {
	logger.Tracef("Running test set: %v", t.Name)

	if t.Wait > 0 {
		logger.Infof("Waiting for %d seconds", t.Wait)
		time.Sleep(time.Duration(t.Wait) * time.Second)
	}

	all_results := []TestResult{}

	all_results = append(all_results, t.Run(cx1client, logger, types.OP_CREATE, Config)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_READ, Config)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_UPDATE, Config)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_DELETE, Config)...)

	return all_results
}

func (t *TestSet) Run(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Config *TestConfig) []TestResult {
	results := []TestResult{}

	for id := range t.Flags {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Flags[id]), &results, Config)
	}
	for id := range t.Imports {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Imports[id]), &results, Config)
	}
	for id := range t.Groups {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Groups[id]), &results, Config)
	}
	for id := range t.Applications {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Applications[id]), &results, Config)
	}
	for id := range t.Projects {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Projects[id]), &results, Config)
	}
	for id := range t.Roles {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Roles[id]), &results, Config)
	}
	for id := range t.Users {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Users[id]), &results, Config)
	}
	for id := range t.AccessAssignments {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.AccessAssignments[id]), &results, Config)
	}
	for id := range t.Queries {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Queries[id]), &results, Config)
	}
	for id := range t.Presets {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Presets[id]), &results, Config)
	}
	for id := range t.Scans {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Scans[id]), &results, Config)
	}
	for id := range t.Results {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Results[id]), &results, Config)
	}
	for id := range t.Reports {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Reports[id]), &results, Config)
	}

	return results
}

func RunTest(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD, testName string, test TestRunner, results *[]TestResult, Config *TestConfig) {
	if test.IsType(CRUD) {
		var result TestResult
		err := test.IsSupported(cx1client, logger, CRUD, &Config.Engines)

		if err == nil && !CheckFlags(cx1client, logger, test) {
			err = fmt.Errorf("test requires feature flag(s) %v to be enabled", strings.Join(test.GetFlags(), ","))
		}

		if err != nil && !test.IsForced() {
			result = MakeResult(test)
			result.CRUD = CRUD
			result.Name = testName
			result.Duration = 0
			result.Reason = err.Error()
			result.Result = TST_SKIP
			logger.Warnf("Test for %v %v is not supported and will be skipped. Reason: %s", CRUD, test.String(), err)
		} else {
			result = Run(cx1client, logger, CRUD, testName, test, Config)
		}

		LogResult(logger, result)
		*results = append(*results, result)
	}
}

func Run(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD, testName string, test TestRunner, Config *TestConfig) TestResult {
	//logger.Infof("Running test: %v %v", CRUD, test.String())
	LogStart(logger, test, CRUD, testName)
	result := MakeResult(test)
	result.CRUD = CRUD
	result.Name = testName

	err := test.Validate(CRUD)
	if err != nil {
		result.Result = TST_SKIP
		result.Reason = err.Error()
		return result
	}
	start := time.Now().UnixNano()

	switch CRUD {
	case types.OP_CREATE:
		err = test.RunCreate(cx1client, logger, &Config.Engines)
	case types.OP_READ:
		err = test.RunRead(cx1client, logger, &Config.Engines)
	case types.OP_UPDATE:
		err = test.RunUpdate(cx1client, logger, &Config.Engines)
	case types.OP_DELETE:
		err = test.RunDelete(cx1client, logger, &Config.Engines)
	}

	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	result.Duration = duration
	if err != nil {
		result.Reason = err.Error()
		if test.IsNegative() { // negative test with error = pass
			result.Result = TST_PASS
			return result
		} else {
			result.Result = TST_FAIL
			return result
		}
	} else {
		if test.IsNegative() { // negative test with no error = fail
			result.Result = TST_FAIL
			result.Reason = "action succeeded but should have failed"
			return result
		} else {
			result.Result = TST_PASS
			return result
		}
	}
}

func CheckFlags(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, test TestRunner) bool {
	for _, flag := range test.GetFlags() {
		val, err := cx1client.CheckFlag(flag)
		if err != nil {
			logger.Errorf("Failed to check flag %v: %s", flag, err)
			return false
		}
		if !val {
			logger.Debugf("Test requires feature flag %v but it is disabled", flag)
			return false
		}
	}

	return true
}

func LogStart(logger *logrus.Logger, test TestRunner, CRUD, testName string) {
	logger.Infof("")
	testType := "Test"
	if test.IsNegative() {
		testType = "Negative-Test"
	}

	logger.Infof("Starting %v %v %v '%v' - %v", CRUD, test.GetModule(), testType, testName, test.String())
}

func LogResult(logger *logrus.Logger, result TestResult) {
	testType := "Test"
	if result.FailTest {
		testType = "Negative-Test"
	}
	switch result.Result {
	case TST_FAIL:
		logger.Errorf("FAIL [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
		logger.Errorf("Failure reason: %v", result.Reason)
	case TST_SKIP:
		logger.Warnf("SKIP [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
		logger.Warnf("Skip reason: %v", result.Reason)
	case TST_PASS:
		logger.Infof("PASS [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
	}
}

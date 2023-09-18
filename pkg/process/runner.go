package process

import (
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
	IsSupported(testType string) bool
	IsNegative() bool
	GetSource() string
	GetModule() string
	GetFlags() []string

	RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
}

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig) float32 {
	all_results := []TestResult{}

	for id := range Config.Tests {
		all_results = append(all_results, Config.Tests[id].RunTests(cx1client, logger)...)
	}

	status, err := GenerateReport(&all_results, logger, Config)
	if err != nil {
		logger.Errorf("Failed to generate the report: %s", err)
	}

	return status
}

func (t *TestSet) RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) []TestResult {
	logger.Tracef("Running test set: %v", t.Name)

	if t.Wait > 0 {
		logger.Infof("Waiting for %d seconds", t.Wait)
		time.Sleep(time.Duration(t.Wait) * time.Second)
	}

	all_results := []TestResult{}

	all_results = append(all_results, t.Run(cx1client, logger, types.OP_CREATE)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_READ)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_UPDATE)...)
	all_results = append(all_results, t.Run(cx1client, logger, types.OP_DELETE)...)

	return all_results
}

func (t *TestSet) Run(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string) []TestResult {
	results := []TestResult{}

	for id := range t.Flags {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Flags[id]), &results)
	}
	for id := range t.Imports {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Imports[id]), &results)
	}
	for id := range t.Groups {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Groups[id]), &results)
	}
	for id := range t.Applications {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Applications[id]), &results)
	}
	for id := range t.Projects {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Projects[id]), &results)
	}
	for id := range t.Roles {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Roles[id]), &results)
	}
	for id := range t.Users {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Users[id]), &results)
	}
	for id := range t.AccessAssignments {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.AccessAssignments[id]), &results)
	}
	for id := range t.Queries {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Queries[id]), &results)
	}
	for id := range t.Presets {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Presets[id]), &results)
	}
	for id := range t.Scans {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Scans[id]), &results)
	}
	for id := range t.Results {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Results[id]), &results)
	}
	for id := range t.Reports {
		RunTest(cx1client, logger, CRUD, t.Name, &(t.Reports[id]), &results)
	}

	return results
}

func RunTest(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD, testName string, test TestRunner, results *[]TestResult) {
	if test.IsType(CRUD) {
		if !test.IsSupported(CRUD) {
			logger.Warnf("Test for %v %v is not supported and will be skipped.", CRUD, test.String())
		} else if !CheckFlags(cx1client, logger, test) {
			logger.Warnf("Test for %v %v requires features that are not enabled in this environment and will be skipped.", CRUD, test.String())
		} else {
			result := Run(cx1client, logger, CRUD, testName, test)
			LogResult(logger, result)
			*results = append(*results, result)
		}
	}
}

func Run(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD, testName string, test TestRunner) TestResult {
	//logger.Infof("Running test: %v %v", CRUD, test.String())
	LogStart(logger, test, CRUD, testName)
	err := test.Validate(CRUD)
	if err != nil {
		//LogSkip(test.FailTest, logger, OP_CREATE, MOD_GROUP, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
		return TestResult{
			test.IsNegative(), TST_SKIP, CRUD, test.GetModule(), 0, testName, -1, test.String(), err.Error(), test.GetSource(),
		}
	}
	start := time.Now().UnixNano()

	switch CRUD {
	case types.OP_CREATE:
		err = test.RunCreate(cx1client, logger)
	case types.OP_READ:
		err = test.RunRead(cx1client, logger)
	case types.OP_UPDATE:
		err = test.RunUpdate(cx1client, logger)
	case types.OP_DELETE:
		err = test.RunDelete(cx1client, logger)
	}

	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if err != nil {
		if test.IsNegative() { // negative test with error = pass
			return TestResult{
				test.IsNegative(), TST_PASS, CRUD, test.GetModule(), duration, testName, -1, test.String(), err.Error(), test.GetSource(),
			}
		} else {
			return TestResult{
				test.IsNegative(), TST_FAIL, CRUD, test.GetModule(), duration, testName, -1, test.String(), err.Error(), test.GetSource(),
			}
		}
	} else {
		if test.IsNegative() { // negative test with no error = fail
			return TestResult{
				test.IsNegative(), TST_FAIL, CRUD, test.GetModule(), duration, testName, -1, test.String(), "action succeeded but should have failed", test.GetSource(),
			}
		} else {
			return TestResult{
				test.IsNegative(), TST_PASS, CRUD, test.GetModule(), duration, testName, -1, test.String(), "", test.GetSource(),
			}
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
			logger.Warnf("Test requires feature flag %v but it is disabled", flag)
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

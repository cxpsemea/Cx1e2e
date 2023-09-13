package main

import (
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

type TestRunner interface {
	Validate(testType string) error
	String() string
	IsType(testType string) bool
	IsNegative() bool
	GetSource() string
	GetModule() string

	RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
	RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error
}

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig) []TestResult {
	all_results := []TestResult{}

	for id := range Config.Tests {
		all_results = append(all_results, Config.Tests[id].RunTests(cx1client, logger)...)
	}

	return all_results
}

func (t *TestSet) RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) []TestResult {
	logger.Tracef("Running test set: %v", t.Name)

	if t.Wait > 0 {
		logger.Infof("Waiting for %d seconds", t.Wait)
		time.Sleep(time.Duration(t.Wait) * time.Second)
	}

	all_results := []TestResult{}

	all_results = append(all_results, t.Run(cx1client, logger, OP_CREATE)...)
	all_results = append(all_results, t.Run(cx1client, logger, OP_READ)...)
	all_results = append(all_results, t.Run(cx1client, logger, OP_UPDATE)...)
	all_results = append(all_results, t.Run(cx1client, logger, OP_DELETE)...)

	/*t.TestCreate(cx1client, logger)
	t.TestRead(cx1client, logger)
	t.TestUpdate(cx1client, logger)
	t.TestDelete(cx1client, logger)*/

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
		result := Run(cx1client, logger, CRUD, testName, test)
		LogResult(logger, result)
		*results = append(*results, result)
	}
}

func Run(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD, testName string, test TestRunner) TestResult {
	logger.Infof("Running test: %v %v", CRUD, test.String())
	err := test.Validate(CRUD)
	if err != nil {
		//LogSkip(test.FailTest, logger, OP_CREATE, MOD_GROUP, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
		return TestResult{
			test.IsNegative(), TST_SKIP, CRUD, test.GetModule(), 0, testName, -1, test.String(), err.Error(), test.GetSource(),
		}
	}
	start := time.Now().UnixNano()

	switch CRUD {
	case OP_CREATE:
		err = test.RunCreate(cx1client, logger)
	case OP_READ:
		err = test.RunRead(cx1client, logger)
	case OP_UPDATE:
		err = test.RunUpdate(cx1client, logger)
	case OP_DELETE:
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

func (t *TestSet) TestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	logger.Tracef("Create tests: %v", t.Name)

	/*FlagTestsCreate(cx1client, logger, t.Name, &tests.Flags)
	ImportTestsCreate(cx1client, logger, t.Name, &tests.Imports)
	GroupTestsCreate(cx1client, logger, t.Name, &tests.Groups)
	ApplicationTestsCreate(cx1client, logger, t.Name, &tests.Applications)
	ProjectTestsCreate(cx1client, logger, t.Name, &tests.Projects)
	RoleTestsCreate(cx1client, logger, t.Name, &tests.Roles)
	UserTestsCreate(cx1client, logger, t.Name, &tests.Users)
	AccessTestsCreate(cx1client, logger, t.Name, &tests.AccessAssignments)
	QueryTestsCreate(cx1client, logger, t.Name, &tests.Queries)
	PresetTestsCreate(cx1client, logger, t.Name, &tests.Presets)
	ScanTestsCreate(cx1client, logger, t.Name, &tests.Scans)
	ResultTestsCreate(cx1client, logger, t.Name, &tests.Results)
	ReportTestsCreate(cx1client, logger, t.Name, &tests.Reports)*/
}
func (t *TestSet) TestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	logger.Tracef("Read tests: %v", t.Name)

	/*FlagTestsRead(cx1client, logger, t.Name, &tests.Flags)
	ImportTestsRead(cx1client, logger, t.Name, &tests.Imports)
	GroupTestsRead(cx1client, logger, t.Name, &tests.Groups)
	ApplicationTestsRead(cx1client, logger, t.Name, &tests.Applications)
	ProjectTestsRead(cx1client, logger, t.Name, &tests.Projects)
	RoleTestsRead(cx1client, logger, t.Name, &tests.Roles)
	AccessTestsRead(cx1client, logger, t.Name, &tests.AccessAssignments)
	UserTestsRead(cx1client, logger, t.Name, &tests.Users)
	QueryTestsRead(cx1client, logger, t.Name, &tests.Queries)
	PresetTestsRead(cx1client, logger, t.Name, &tests.Presets)
	ScanTestsRead(cx1client, logger, t.Name, &tests.Scans)
	ResultTestsRead(cx1client, logger, t.Name, &tests.Results)
	ReportTestsRead(cx1client, logger, t.Name, &tests.Reports)*/
}
func (t *TestSet) TestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	logger.Tracef("Update tests: %v", t.Name)

	/*FlagTestsUpdate(cx1client, logger, t.Name, &tests.Flags)
	ImportTestsUpdate(cx1client, logger, t.Name, &tests.Imports)
	GroupTestsUpdate(cx1client, logger, t.Name, &tests.Groups)
	ApplicationTestsUpdate(cx1client, logger, t.Name, &tests.Applications)
	ProjectTestsUpdate(cx1client, logger, t.Name, &tests.Projects)
	RoleTestsUpdate(cx1client, logger, t.Name, &tests.Roles)
	AccessTestsUpdate(cx1client, logger, t.Name, &tests.AccessAssignments)
	UserTestsUpdate(cx1client, logger, t.Name, &tests.Users)
	QueryTestsUpdate(cx1client, logger, t.Name, &tests.Queries)
	PresetTestsUpdate(cx1client, logger, t.Name, &tests.Presets)
	ScanTestsUpdate(cx1client, logger, t.Name, &tests.Scans)
	ResultTestsUpdate(cx1client, logger, t.Name, &tests.Results)
	ReportTestsUpdate(cx1client, logger, t.Name, &tests.Reports)*/
}
func (t *TestSet) TestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	logger.Tracef("Delete tests: %v", t.Name)

	/*FlagTestsDelete(cx1client, logger, t.Name, &tests.Flags)
	ImportTestsDelete(cx1client, logger, t.Name, &tests.Imports)
	GroupTestsDelete(cx1client, logger, t.Name, &tests.Groups)
	ApplicationTestsDelete(cx1client, logger, t.Name, &tests.Applications)
	ProjectTestsDelete(cx1client, logger, t.Name, &tests.Projects)
	RoleTestsDelete(cx1client, logger, t.Name, &tests.Roles)
	AccessTestsDelete(cx1client, logger, t.Name, &tests.AccessAssignments)
	UserTestsDelete(cx1client, logger, t.Name, &tests.Users)
	QueryTestsDelete(cx1client, logger, t.Name, &tests.Queries)
	PresetTestsDelete(cx1client, logger, t.Name, &tests.Presets)
	ScanTestsDelete(cx1client, logger, t.Name, &tests.Scans)
	ResultTestsDelete(cx1client, logger, t.Name, &tests.Results)
	ReportTestsDelete(cx1client, logger, t.Name, &tests.Reports)*/
}

func LogStart(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string) {
	logger.Infof("")
	if failTest {
		logger.Infof("Starting %v %v Negative-Test '%v' #%d - %v", CRUD, Module, testName, testId, testObject)
	} else {
		logger.Infof("Starting %v %v Test '%v' #%d - %v", CRUD, Module, testName, testId, testObject)
	}
}

func LogPass(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Errorf("FAIL [%.3fs]: %v %v Negative-Test '%v' #%d (%v) - %v", duration, CRUD, Module, testName, testId, testObject, "test passed unexpectedly")
		/*TestResults = append(TestResults, TestResult{
			failTest, TST_FAIL, CRUD, Module, duration, testName, testId, testObject, "test passed unexpectedly", testSource,
		})*/
	} else {
		logger.Infof("PASS [%.3fs]: %v %v Test '%v' #%d (%v)", duration, CRUD, Module, testName, testId, testObject)
		/*TestResults = append(TestResults, TestResult{
			failTest, TST_PASS, CRUD, Module, duration, testName, testId, testObject, "", testSource,
		})*/
	}
}
func LogSkip(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string, reason string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Warnf("SKIP [%.3fs]: %v %v Negative-Test '%v' #%d - %v", duration, CRUD, Module, testName, testId, reason)
	} else {
		logger.Warnf("SKIP [%.3fs]: %v %v Test '%v' #%d - %v", duration, CRUD, Module, testName, testId, reason)
	}
	/*TestResults = append(TestResults, TestResult{
		failTest, TST_SKIP, CRUD, Module, duration, testName, testId, testObject, reason, testSource,
	})*/
}
func LogFail(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string, reason error) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Infof("PASS [%.3fs]: %v %v Negative-Test '%v' #%d (%v)", duration, CRUD, Module, testName, testId, testObject)
		/*TestResults = append(TestResults, TestResult{
			failTest, TST_PASS, CRUD, Module, duration, testName, testId, testObject, "", testSource,
		})*/
	} else {
		logger.Errorf("FAIL [%.3fs]: %v %v Test '%v' #%d (%v) - %s", duration, CRUD, Module, testName, testId, testObject, reason)
		/*TestResults = append(TestResults, TestResult{
			failTest, TST_FAIL, CRUD, Module, duration, testName, testId, testObject, reason.Error(), testSource,
		})*/
	}
}

func LogResult(logger *logrus.Logger, result TestResult) {
	testType := "Test"
	if result.FailTest {
		testType = "Negative-Test"
	}
	switch result.Result {
	case TST_FAIL:
		logger.Errorf("FAIL [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
	case TST_SKIP:
		logger.Warnf("SKIP [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
	case TST_PASS:
		logger.Infof("PASS [%.3fs]: %v %v %v '%v' (%v)", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject)
	}
}

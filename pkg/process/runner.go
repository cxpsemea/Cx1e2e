package process

import (
	"fmt"
	"os/exec"
	"slices"
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
	IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, testType string, Engines *types.EnabledEngines) error
	IsNegative() bool
	GetSource() string
	GetID() uint
	GetModule() string
	GetFlags() []string
	GetVersion() types.ProductVersion
	GetVersionStr() string
	GetCurrentThread() int
	OnFail() types.FailAction

	RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, Engines *types.EnabledEngines) error
	RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, Engines *types.EnabledEngines) error
	RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, Engines *types.EnabledEngines) error
	RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, Engines *types.EnabledEngines) error
}

func MakeResult(test TestRunner) TestResult {
	return TestResult{
		FailTest:   test.IsNegative(),
		Result:     TST_SKIP,
		Module:     test.GetModule(),
		Duration:   0,
		Id:         test.GetID(),
		TestObject: test.String(),
		TestSource: test.GetSource(),
	}
}

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig, threads int) uint {
	startTime := time.Now()
	all_results := []TestResult{}
	dir := NewDirector(Config)

	if !Config.MultiThreadable && threads > 1 {
		logger.Warnf("Configuration does not allow multi-threading tests while cx1e2e was run with threads=%d - resetting to 1", threads)
		threads = 1
	}

	out_channels := make(chan *[]TestResult, threads)
	for i := range threads {
		go NewRunner(i+1, &dir, cx1client, logger, Config, out_channels)
	}

	for range threads {
		results := <-out_channels
		all_results = append(all_results, *results...)
	}

	close(out_channels)
	endTime := time.Now()

	// tests are finished running, so do some cleanup
	if types.ASM != nil {
		types.ASM.Clear(cx1client, logger)
	}

	// the test-results may be unsorted due to threading, sort them
	if threads > 1 {
		slices.SortFunc(all_results, func(a, b TestResult) int {
			return int(a.Id - b.Id)
		})
	}

	status, err := GenerateReport(&all_results, logger, Config, startTime, endTime, threads)
	if err != nil {
		logger.Errorf("Failed to generate the report: %s", err)
	}
	logger.Infof("Test complete")

	return status
}

func (t *TestSet) RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, Config *TestConfig, testSetFail error) []TestResult {
	logger.Tracef("Running test set: %v [%v]", t.Name, t.TestSource)

	var err error = testSetFail
	var testClient *Cx1ClientGo.Cx1Client

	if t.Wait > 0 {
		logger.Infof("Waiting for %d seconds", t.Wait)
		time.Sleep(time.Duration(t.Wait) * time.Second)
	}

	all_results := []TestResult{}

	if err == nil {
		if t.OtherUser() {
			logger.Infof("Test is configured to run as other user")

			for counter := 3; counter > 0 && testClient == nil; counter-- {
				testClient, err = t.GetOtherClient(cx1client, logger, Config)

				if err != nil {
					logger.Errorf("Failed to get new Cx1 client for test set %v: %s", t.Name, err)
				} else {
					logger.Infof("Created new Cx1 client for test set %v: %s", t.Name, testClient.String())
					break
				}
				logger.Infof("Waiting 15 seconds to retry")
				time.Sleep(time.Second * time.Duration(15))
			}
		} else {
			testClient = cx1client
		}
	}

	var results []TestResult

	if len(t.SubTests) > 0 {
		for id := range t.SubTests {
			results = t.SubTests[id].RunTests(testClient, logger, Config, err)
			all_results = append(all_results, results...)
		}
	} else {
		results, err = t.Run(testClient, logger, types.OP_CREATE, Config, err)
		all_results = append(all_results, results...)
		results, err = t.Run(testClient, logger, types.OP_READ, Config, err)
		all_results = append(all_results, results...)
		results, err = t.Run(testClient, logger, types.OP_UPDATE, Config, err)
		all_results = append(all_results, results...)
		results, _ = t.Run(testClient, logger, types.OP_DELETE, Config, err)
		all_results = append(all_results, results...)
	}

	return all_results
}

func (t *TestSet) Run(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, CRUD string, Config *TestConfig, TestSetFail error) ([]TestResult, error) {
	results := []TestResult{}
	var TestSetFailError error
	TestSetFailError = TestSetFail

	// for CRU operations, follow this order, but for Delete do the reverse
	if CRUD != types.OP_DELETE {
		for id := range t.Flags {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Flags[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Analytics {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Analytics[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Imports {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Imports[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Groups {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Groups[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Applications {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Applications[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Projects {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Projects[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Roles {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Roles[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Users {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Users[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Clients {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Clients[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.AccessAssignments {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.AccessAssignments[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Queries {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Queries[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Presets {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Presets[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Scans {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Scans[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Branches {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Branches[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Results {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Results[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Reports {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Reports[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
	} else { // in reverse order for DELETE
		for id := range t.Scans {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Scans[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Presets {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Presets[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Queries {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Queries[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.AccessAssignments {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.AccessAssignments[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Clients {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Clients[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Users {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Users[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Roles {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Roles[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Projects {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Projects[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Applications {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Applications[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
		for id := range t.Groups {
			err := RunTest(cx1client, logger, CRUD, t.Name, &(t.Groups[id]), &results, Config, TestSetFailError)
			if err != nil && TestSetFailError == nil {
				TestSetFailError = err
			}
		}
	}

	return results, TestSetFailError
}

func RunTest(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, CRUD, testName string, test TestRunner, results *[]TestResult, Config *TestConfig, failSet error) error {
	if test.IsType(CRUD) {
		var result TestResult
		failAction := test.OnFail()

		if failSet != nil {
			result = MakeResult(test)
			result.CRUD = CRUD
			result.Name = testName
			result.Duration = 0
			result.Reason = []string{failSet.Error()}
			result.Result = TST_FAIL
			logger.Warnf("Test for %v %v prerequisite failed: %s", CRUD, test.String(), failSet)
		} else {
			err := test.IsSupported(cx1client, logger, CRUD, &Config.Engines)

			if err == nil && !CheckFlags(cx1client, logger, test) {
				err = fmt.Errorf("test requires feature flag(s): %v ", strings.Join(test.GetFlags(), ","))
			}

			if err == nil && !CheckVersion(cx1client, logger, test) {
				v, _ := cx1client.GetVersion()
				err = fmt.Errorf("test expects %v, current version is %v", test.GetVersionStr(), v.String())
			}

			if err != nil && !test.IsForced() { // if an error prevents us from running the test, and the test isn't a Forced test, skip
				result = MakeResult(test)
				result.CRUD = CRUD
				result.Name = testName
				result.Duration = 0
				result.Reason = []string{err.Error()}
				result.Result = TST_SKIP
				logger.Warnf("Test for %v %v will be skipped. Reason: %s", CRUD, test.String(), err)
			} else { // test can run
				result = Run(cx1client, logger, CRUD, testName, test, Config)
				if failAction.RetryCount > 0 && result.Result == TST_FAIL {
					for count := 1; count <= (int)(failAction.RetryCount) && result.Result == TST_FAIL; count++ {
						logger.Infof("Test for %v %v failed: %v, waiting %d seconds for retry %d of %d", CRUD, test.String(), result.Reason[0], failAction.RetryDelay, count, failAction.RetryCount)
						time.Sleep(time.Duration(failAction.RetryDelay) * time.Second)
						result = Run(cx1client, logger, CRUD, testName, test, Config)
					}

					if result.Result == TST_FAIL {
						result.Reason = append(result.Reason, fmt.Sprintf(" (with %d retries)", failAction.RetryCount))
					}
				}
			}
		}

		LogResult(logger, result)
		*results = append(*results, result)

		if result.Result == TST_FAIL {
			if len(failAction.Commands) > 0 {
				logger.Debugf("Failed test includes %d post-fail commands", len(failAction.Commands))
				for id, command := range failAction.Commands {
					logger.Debugf("Running command %d: %v", id, command)
					parts := strings.Split(command, " ")
					cmd := exec.Command(parts[0], parts[1:]...)
					output, err := cmd.CombinedOutput()
					if err != nil {
						logger.Errorf("Command failed: %s", err)
						result.Reason = append(result.Reason, fmt.Sprintf("OnFail command #%d '%v' failed: %s", id, parts[0], err))
					} else {
						str := fmt.Sprintf("OnFail command #%d '%v' returned: %v", id, parts[0], string(output))
						logger.Info(str)
						result.Reason = append(result.Reason, str)
					}
				}
			}

			if failAction.FailSet {
				err := FailError(result)
				return err
			}
		}
	}

	return nil
}

func Run(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, CRUD, testName string, test TestRunner, Config *TestConfig) TestResult {
	//logger.Infof("Running test: %v %v", CRUD, test.String())
	LogStart(logger, test, CRUD, testName)
	result := MakeResult(test)
	result.CRUD = CRUD
	result.Name = testName

	err := test.Validate(CRUD)
	if err != nil {
		result.Result = TST_SKIP
		result.Reason = []string{err.Error()}
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
		result.Reason = []string{err.Error()}
		if test.IsNegative() { // negative test with error = pass
			result.Result = TST_PASS
			return result
		} else {
			result.Result = TST_FAIL
			result.Reason = []string{err.Error()}
			return result
		}
	} else {
		if test.IsNegative() { // negative test with no error = fail
			result.Result = TST_FAIL
			result.Reason = []string{"action succeeded but should have failed"}
			return result
		} else {
			result.Result = TST_PASS
			return result
		}
	}
}

func CheckFlags(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, test TestRunner) bool {
	for _, flag := range test.GetFlags() {
		negative := false
		if flag[0] == '!' {
			negative = true
			flag = flag[1:]
		}

		val, err := cx1client.CheckFlag(flag)
		if err != nil {
			logger.Errorf("Failed to check flag %v: %s", flag, err)
			return false
		}

		if !val && !negative {
			logger.Debugf("Test requires feature flag %v but it is disabled", flag)
			return false
		}
		if val && negative {
			logger.Debugf("Test requires absence of feature flag %v but it is enabled", flag)
			return false
		}
	}

	return true
}

func CheckVersion(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, test TestRunner) bool {
	cur, _ := cx1client.GetVersion()

	pv := test.GetVersion()

	if pv.CxOne.IsSet() {
		if pv.CxOne.Min != "" {
			check, err := cur.CheckCxOne(pv.CxOne.Min)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check < 0 {
				return false
			}
		}
		if pv.CxOne.Max != "" {
			check, err := cur.CheckCxOne(pv.CxOne.Max)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check >= 0 {
				return false
			}
		}
	}

	if pv.SAST.IsSet() {
		if pv.SAST.Min != "" {
			check, err := cur.CheckSAST(pv.SAST.Min)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check < 0 {
				return false
			}
		}
		if pv.SAST.Max != "" {
			check, err := cur.CheckSAST(pv.SAST.Max)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check >= 0 {
				return false
			}
		}
	}

	if pv.IAC.IsSet() {
		if pv.IAC.Min != "" {
			check, err := cur.CheckIAC(pv.IAC.Min)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check < 0 {
				return false
			}
		}
		if pv.IAC.Max != "" {
			check, err := cur.CheckIAC(pv.IAC.Max)
			if err != nil {
				logger.Errorf("Failed to check version: %s", err)
			}
			if check >= 0 {
				return false
			}
		}
	}

	return true
}

func LogStart(logger *types.ThreadLogger, test TestRunner, CRUD, testName string) {
	logger.Infof("")
	testType := "Test"
	if test.IsNegative() {
		testType = "Negative-Test"
	}

	logger.Infof("Starting test #%d - %v %v %v '%v' - %v [%v]", test.GetID(), CRUD, test.GetModule(), testType, testName, test.String(), test.GetSource())
}

func LogResult(logger *types.ThreadLogger, result TestResult) {
	testType := "Test"
	if result.FailTest {
		testType = "Negative-Test"
	}
	switch result.Result {
	case TST_FAIL:
		logger.Errorf("FAIL [%.3fs]: %v %v %v '%v' (%v) [%v]", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject, result.TestSource)
		if result.Attempts > 1 {
			logger.Errorf("Failure reason: %v (with %d attempts)", result.Reason[0], result.Attempts)
		} else {
			logger.Errorf("Failure reason: %v", result.Reason[0])
		}
	case TST_SKIP:
		logger.Warnf("SKIP [%.3fs]: %v %v %v '%v' (%v) [%v]", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject, result.TestSource)
		logger.Warnf("Skip reason: %v", result.Reason)
	case TST_PASS:
		if result.Attempts > 1 {
			logger.Infof("PASS [%.3fs]: %v %v %v '%v' (%v) [%v] - took %d attempts", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject, result.TestSource, result.Attempts)
		} else {
			logger.Infof("PASS [%.3fs]: %v %v %v '%v' (%v) [%v]", result.Duration, result.CRUD, result.Module, testType, result.Name, result.TestObject, result.TestSource)
		}
	}
}

func FailError(result TestResult) error {
	testType := "Test"
	if result.FailTest {
		testType = "Negative-Test"
	}
	return fmt.Errorf("previous test %v %v %v '%v' (%v) failed: %v", result.CRUD, result.Module, testType, result.Name, result.TestObject, result.Reason[0])
}

func (t TestSet) OtherUser() bool {
	return t.RunAs.APIKey != "" || (t.RunAs.ClientID != "" && t.RunAs.ClientSecret != "") || t.RunAs.OIDCClient != ""
}

func (t TestSet) GetOtherClient(cx1client *Cx1ClientGo.Cx1Client, logger *types.ThreadLogger, config *TestConfig) (*Cx1ClientGo.Cx1Client, error) {
	httpClient, err := config.CreateHTTPClient(logger.GetLogger())
	if err != nil {
		return nil, err
	}

	if t.RunAs.APIKey != "" {
		return Cx1ClientGo.NewAPIKeyClient(httpClient, config.Cx1URL, config.IAMURL, config.Tenant, t.RunAs.APIKey, logger.GetLogger())
	}

	if t.RunAs.ClientID != "" && t.RunAs.ClientSecret != "" {
		return Cx1ClientGo.NewOAuthClient(httpClient, config.Cx1URL, config.IAMURL, config.Tenant, t.RunAs.ClientID, t.RunAs.ClientSecret, logger.GetLogger())
	}

	if t.RunAs.OIDCClient != "" {
		client, err := cx1client.GetClientByName(t.RunAs.OIDCClient)
		if err != nil {
			return nil, err
		}

		secret, err := cx1client.GetClientSecret(&client)
		if err != nil {
			return nil, err
		}

		return Cx1ClientGo.NewOAuthClient(httpClient, config.Cx1URL, config.IAMURL, config.Tenant, client.ClientID, secret, logger.GetLogger())
	}

	return nil, fmt.Errorf("no credentials provided in test %v step, file %v", t.Name, t.File)
}

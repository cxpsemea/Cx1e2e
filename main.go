package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"gopkg.in/yaml.v2"
)

var TestResults []TestResult

const (
	OP_CREATE = "Create"
	OP_READ   = "Read"
	OP_UPDATE = "Update"
	OP_DELETE = "Delete"
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

const (
	TST_FAIL = 0
	TST_PASS = 1
	TST_SKIP = 2
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	myformatter := &easy.Formatter{}
	myformatter.TimestampFormat = "2006-01-02 15:04:05.000"
	myformatter.LogFormat = "[%lvl%][%time%] %msg%\n"
	logger.SetFormatter(myformatter)
	logger.SetOutput(os.Stdout)

	testConfig := flag.String("config", "", "Path to a test config.yaml")
	APIKey := flag.String("apikey", "", "CheckmarxOne API Key (if not using client id/secret)")
	ClientID := flag.String("client", "", "CheckmarxOne Client ID (if not using API Key)")
	ClientSecret := flag.String("secret", "", "CheckmarxOne Client Secret (if not using API Key)")
	Cx1URL := flag.String("cx1", "", "Optional: CheckmarxOne platform URL, if not defined in the test config.yaml")
	IAMURL := flag.String("iam", "", "Optional: CheckmarxOne IAM URL, if not defined in the test config.yaml")
	Tenant := flag.String("tenant", "", "Optional: CheckmarxOne tenant, if not defined in the test config.yaml")

	flag.Parse()

	if *testConfig == "" || (*APIKey == "" && (*ClientID == "" || *ClientSecret == "")) {
		logger.Info("The purpose of this tool is to automate testing of the API for various workflows based on the yaml configuration. For help run: cx1e2e.exe -h")
		logger.Fatalf("Test configuration yaml or authentication (API Key or client+secret) not provided.")
	}

	var err error
	Config, err := LoadConfig(logger, *testConfig)
	if err != nil {
		logger.Fatalf("Failed to load configuration file %v: %s", *testConfig, err)
		return
	}

	switch strings.ToUpper(Config.LogLevel) {
	case "":
		logger.SetLevel(logrus.InfoLevel)
	case "TRACE":
		logger.SetLevel(logrus.TraceLevel)
	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		logger.SetLevel(logrus.InfoLevel)
	case "WARNING":
		logger.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logger.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logger.SetLevel(logrus.FatalLevel)
	}

	var cx1client *Cx1ClientGo.Cx1Client
	httpClient := &http.Client{}

	if Config.ProxyURL != "" {
		proxyURL, err := url.Parse(Config.ProxyURL)
		if err != nil {
			logger.Fatalf("Failed to parse specified proxy address %v: %s", Config.ProxyURL, err)
			return
		}
		transport := &http.Transport{}
		transport.Proxy = http.ProxyURL(proxyURL)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		httpClient.Transport = transport
		logger.Infof("Running with proxy: %v", Config.ProxyURL)
	}

	if *Tenant != "" {
		Config.Tenant = *Tenant
	}
	if *Cx1URL != "" {
		Config.Cx1URL = *Cx1URL
	}
	if *IAMURL != "" {
		Config.IAMURL = *IAMURL
	}

	if *APIKey != "" {
		cx1client, err = Cx1ClientGo.NewAPIKeyClient(httpClient, Config.Cx1URL, Config.IAMURL, Config.Tenant, *APIKey, logger)
	} else {
		cx1client, err = Cx1ClientGo.NewOAuthClient(httpClient, Config.Cx1URL, Config.IAMURL, Config.Tenant, *ClientID, *ClientSecret, logger)
	}

	if err != nil {
		logger.Fatalf("Failed to create Cx1 client: %s", err)
		return
	}

	logger.Infof("Created Cx1 client %s", cx1client.String())

	RunTests(cx1client, logger, &Config)

	logger.Infof("Test result summary:\n")
	count_failed := 0
	count_passed := 0
	count_skipped := 0

	for _, result := range TestResults {
		switch result.Result {
		case 1:
			fmt.Printf("PASS %v - %v %v: %v\n", result.Name, result.CRUD, result.Module, result.TestObject)
			count_passed++
		case 0:
			fmt.Printf("FAIL %v - %v %v: %v\n", result.Name, result.CRUD, result.Module, result.TestObject)
			count_failed++
		case 2:
			fmt.Printf("SKIP %v - %v %v: %v\n", result.Name, result.CRUD, result.Module, result.TestObject)
			count_skipped++
		}
	}
	fmt.Println("")
	fmt.Printf("Ran %d tests\n", (count_failed + count_passed + count_skipped))
	if count_failed > 0 {
		fmt.Printf("FAILED %d tests\n", count_failed)
	}
	if count_skipped > 0 {
		fmt.Printf("SKIPPED %d tests\n", count_skipped)
	}
	if count_passed > 0 {
		fmt.Printf("PASSED %d tests\n", count_passed)
	}

	err = GenerateReport(&TestResults, &Config)
	if err != nil {
		logger.Errorf("Failed to generate HTML report: %s", err)
	}
}

func LoadConfig(logger *logrus.Logger, configPath string) (TestConfig, error) {
	var conf TestConfig

	file, err := os.Open(configPath)
	if err != nil {
		return conf, err
	}

	conf.ConfigPath, _ = filepath.Abs(file.Name())
	currentRoot := filepath.Dir(file.Name())

	defer file.Close()
	d := yaml.NewDecoder(file)

	err = d.Decode(&conf)
	if err != nil {
		return conf, err
	}

	testSet := make([]TestSet, 0)

	// propagate the filename to sub-tests
	// TODO: refactor this to use generics?
	for id := range conf.Tests {
		for id2 := range conf.Tests[id].AccessAssignments {
			conf.Tests[id].AccessAssignments[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Applications {
			conf.Tests[id].Applications[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Flags {
			conf.Tests[id].Flags[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Groups {
			conf.Tests[id].Groups[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Imports {
			conf.Tests[id].Imports[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Presets {
			conf.Tests[id].Presets[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Projects {
			conf.Tests[id].Projects[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Queries {
			conf.Tests[id].Queries[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Reports {
			conf.Tests[id].Reports[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Results {
			conf.Tests[id].Results[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Roles {
			conf.Tests[id].Roles[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Scans {
			conf.Tests[id].Scans[id2].TestSource = configPath
		}
		for id2 := range conf.Tests[id].Users {
			conf.Tests[id].Users[id2].TestSource = configPath
		}
	}

	for _, set := range conf.Tests {
		logger.Tracef("Checking TestSet %v for file references", set.Name)
		if set.File != "" {
			configPath, err := getFilePath(currentRoot, set.File)
			if err != nil {
				return conf, err
			}

			conf2, err := LoadConfig(logger, configPath)
			if err != nil {
				return conf, fmt.Errorf("error loading sub-test %v: %s", set.File, err)
			}
			logger.Debugf("Loaded sub-config from %v", conf2.ConfigPath)
			testSet = append(testSet, conf2.Tests...)
		} else {
			for id, scan := range set.Scans {
				logger.Tracef(" - Checking Scan TestSet %v for file references", set.Name)
				if scan.ZipFile != "" {
					filePath, err := getFilePath(currentRoot, scan.ZipFile)
					if err != nil {
						return conf, fmt.Errorf("error locating scan zipfile %v", scan.ZipFile)
					}
					set.Scans[id].ZipFile = filePath
				}
			}
			for id, imp := range set.Imports {
				logger.Tracef(" - Checking Import TestSet %v for file references", set.Name)
				if imp.ZipFile != "" {
					filePath, err := getFilePath(currentRoot, imp.ZipFile)
					if err != nil {
						return conf, fmt.Errorf("error locating import zipfile %v", imp.ZipFile)
					}
					set.Imports[id].ZipFile = filePath
				}
				if imp.ProjectMapFile != "" {
					filePath, err := getFilePath(currentRoot, imp.ProjectMapFile)
					if err != nil {
						return conf, fmt.Errorf("error locating import ProjectMapFile %v", imp.ProjectMapFile)
					}
					set.Imports[id].ProjectMapFile = filePath
				}
			}
			testSet = append(testSet, set)
		}
	}
	conf.Tests = testSet

	return conf, nil
}

func getFilePath(currentRoot, file string) (string, error) {
	osPath := filepath.FromSlash(file)
	//logger.Debugf("Trying to find config file %v, current root is %v", osPath, currentRoot)
	if _, err := os.Stat(osPath); err == nil {
		return filepath.Clean(osPath), nil
	} else {
		testPath := filepath.Join(currentRoot, osPath)
		//logger.Debugf("File doesn't exist, testing: %v", testPath)
		if _, err := os.Stat(testPath); err == nil {
			return filepath.Clean(testPath), nil
		} else {
			return "", fmt.Errorf("unable to find configuration file %v", testPath)
		}
	}
}

func GenerateReport(tests *[]TestResult, Config *TestConfig) error {
	report, err := os.Create("cx1e2e_result.html")
	if err != nil {
		return err
	}

	defer report.Close()

	var Access, Application, Flag, Group, Import, Preset, Project, Query, Result, Report, Role, Scan, User CounterSet
	for _, r := range *tests {
		var set *CounterSet
		switch r.Module {
		case MOD_ACCESS:
			set = &Access
		case MOD_APPLICATION:
			set = &Application
		case MOD_FLAG:
			set = &Flag
		case MOD_GROUP:
			set = &Group
		case MOD_IMPORT:
			set = &Import
		case MOD_PRESET:
			set = &Preset
		case MOD_PROJECT:
			set = &Project
		case MOD_QUERY:
			set = &Query
		case MOD_RESULT:
			set = &Result
		case MOD_REPORT:
			set = &Report
		case MOD_ROLE:
			set = &Role
		case MOD_SCAN:
			set = &Scan
		case MOD_USER:
			set = &User
		}

		var count *Counter

		switch r.CRUD {
		case OP_CREATE:
			count = &(set.Create)
		case OP_READ:
			count = &(set.Read)
		case OP_UPDATE:
			count = &(set.Update)
		case OP_DELETE:
			count = &(set.Delete)
		}

		switch r.Result {
		case TST_PASS:
			count.Pass++
		case TST_FAIL:
			count.Fail++
		case TST_SKIP:
			count.Skip++
		}
	}

	_, err = report.WriteString(fmt.Sprintf("<html><head><title>%v tenant %v test - %v</title></head><body>", Config.Cx1URL, Config.Tenant, time.Now().String()))
	if err != nil {
		return err
	}
	report.WriteString("<h2>Summary</h2>")
	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th rowspan=2>Area</th><th colspan=3>Create</th><th colspan=3>Read</th><th colspan=3>Update</th><th colspan=3>Delete</th></tr>\n")
	report.WriteString("<tr><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th></tr>\n")
	writeCounterSet(report, "Access Assignment", &Access)
	writeCounterSet(report, "Application", &Application)
	writeCounterSet(report, "Flag", &Flag)
	writeCounterSet(report, "Group", &Group)
	writeCounterSet(report, "Import", &Import)
	writeCounterSet(report, "Preset", &Preset)
	writeCounterSet(report, "Project", &Project)
	writeCounterSet(report, "Query", &Query)
	writeCounterSet(report, "Result", &Result)
	writeCounterSet(report, "Report", &Report)
	writeCounterSet(report, "Role", &Role)
	writeCounterSet(report, "Scan", &Scan)
	writeCounterSet(report, "User", &User)
	report.WriteString("</table><br>")

	report.WriteString("<h2>Details</h2>")
	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th>Test Set</th><th>Test</th><th>Duration (sec)</th><th>Result</th></tr>\n")

	for _, t := range *tests {
		result := "<span style='color:green'>PASS</span>"
		if t.Result == TST_FAIL {
			result = fmt.Sprintf("<span style='color:red'>FAIL: %v</span>", t.Reason)
		} else if t.Result == TST_SKIP {
			result = fmt.Sprintf("<span style='color:red'>SKIP: %v</span>", t.Reason)
		}
		report.WriteString(fmt.Sprintf("<tr><td>%v<br>(%v)</td><td>%v %v: %v</td><td>%.2f</td><td>%v</td></tr>\n", t.Name, t.TestSource, t.CRUD, t.Module, t.TestObject, t.Duration, result))
	}

	report.WriteString("</table>\n")

	_, err = report.WriteString("</body></html>")
	if err != nil {
		return err
	}

	return report.Sync()
}

func writeCell(report *os.File, count uint, good bool) {
	if count == 0 {
		report.WriteString("<td>&nbsp;</td>")
	} else if good {
		report.WriteString(fmt.Sprintf("<td style='color:green;text-align:center;'>%d</td>", count))
	} else {
		report.WriteString(fmt.Sprintf("<td style='color:red;text-align:center;'>%d</td>", count))
	}
}

func writeCounterSet(report *os.File, module string, count *CounterSet) {
	report.WriteString(fmt.Sprintf("<tr><td>%v</td>", module))

	writeCell(report, count.Create.Pass, true)
	writeCell(report, count.Create.Fail, false)
	writeCell(report, count.Create.Skip, false)
	writeCell(report, count.Read.Pass, true)
	writeCell(report, count.Read.Fail, false)
	writeCell(report, count.Read.Skip, false)
	writeCell(report, count.Update.Pass, true)
	writeCell(report, count.Update.Fail, false)
	writeCell(report, count.Update.Skip, false)
	writeCell(report, count.Delete.Pass, true)
	writeCell(report, count.Delete.Fail, false)
	writeCell(report, count.Delete.Skip, false)

	report.WriteString("</tr>\n")
}

func IsCreate(test string) bool {
	return strings.Contains(test, "C")
}
func IsRead(test string) bool {
	return strings.Contains(test, "R")
}
func IsUpdate(test string) bool {
	return strings.Contains(test, "U")
}
func IsDelete(test string) bool {
	return strings.Contains(test, "D")
}

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig) {
	for _, t := range Config.Tests {
		logger.Tracef("Running test set: %v", t.Name)

		if t.Wait > 0 {
			logger.Infof("Waiting for %d seconds", t.Wait)
			time.Sleep(time.Duration(t.Wait) * time.Second)
		}
		TestCreate(cx1client, logger, t.Name, &t)
		TestRead(cx1client, logger, t.Name, &t)
		TestUpdate(cx1client, logger, t.Name, &t)
		TestDelete(cx1client, logger, t.Name, &t)
	}
}

func TestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	logger.Tracef("Create tests: %v", testname)

	FlagTestsCreate(cx1client, logger, testname, &tests.Flags)
	ImportTestsCreate(cx1client, logger, testname, &tests.Imports)
	GroupTestsCreate(cx1client, logger, testname, &tests.Groups)
	ApplicationTestsCreate(cx1client, logger, testname, &tests.Applications)
	ProjectTestsCreate(cx1client, logger, testname, &tests.Projects)
	RoleTestsCreate(cx1client, logger, testname, &tests.Roles)
	UserTestsCreate(cx1client, logger, testname, &tests.Users)
	AccessTestsCreate(cx1client, logger, testname, &tests.AccessAssignments)
	QueryTestsCreate(cx1client, logger, testname, &tests.Queries)
	PresetTestsCreate(cx1client, logger, testname, &tests.Presets)
	ScanTestsCreate(cx1client, logger, testname, &tests.Scans)
	ResultTestsCreate(cx1client, logger, testname, &tests.Results)
	ReportTestsCreate(cx1client, logger, testname, &tests.Reports)
}
func TestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	logger.Tracef("Read tests: %v", testname)

	FlagTestsRead(cx1client, logger, testname, &tests.Flags)
	ImportTestsRead(cx1client, logger, testname, &tests.Imports)
	GroupTestsRead(cx1client, logger, testname, &tests.Groups)
	ApplicationTestsRead(cx1client, logger, testname, &tests.Applications)
	ProjectTestsRead(cx1client, logger, testname, &tests.Projects)
	RoleTestsRead(cx1client, logger, testname, &tests.Roles)
	AccessTestsRead(cx1client, logger, testname, &tests.AccessAssignments)
	UserTestsRead(cx1client, logger, testname, &tests.Users)
	QueryTestsRead(cx1client, logger, testname, &tests.Queries)
	PresetTestsRead(cx1client, logger, testname, &tests.Presets)
	ScanTestsRead(cx1client, logger, testname, &tests.Scans)
	ResultTestsRead(cx1client, logger, testname, &tests.Results)
	ReportTestsRead(cx1client, logger, testname, &tests.Reports)
}
func TestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	logger.Tracef("Update tests: %v", testname)

	FlagTestsUpdate(cx1client, logger, testname, &tests.Flags)
	ImportTestsUpdate(cx1client, logger, testname, &tests.Imports)
	GroupTestsUpdate(cx1client, logger, testname, &tests.Groups)
	ApplicationTestsUpdate(cx1client, logger, testname, &tests.Applications)
	ProjectTestsUpdate(cx1client, logger, testname, &tests.Projects)
	RoleTestsUpdate(cx1client, logger, testname, &tests.Roles)
	AccessTestsUpdate(cx1client, logger, testname, &tests.AccessAssignments)
	UserTestsUpdate(cx1client, logger, testname, &tests.Users)
	QueryTestsUpdate(cx1client, logger, testname, &tests.Queries)
	PresetTestsUpdate(cx1client, logger, testname, &tests.Presets)
	ScanTestsUpdate(cx1client, logger, testname, &tests.Scans)
	ResultTestsUpdate(cx1client, logger, testname, &tests.Results)
	ReportTestsUpdate(cx1client, logger, testname, &tests.Reports)
}
func TestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	logger.Tracef("Delete tests: %v", testname)

	FlagTestsDelete(cx1client, logger, testname, &tests.Flags)
	ImportTestsDelete(cx1client, logger, testname, &tests.Imports)
	GroupTestsDelete(cx1client, logger, testname, &tests.Groups)
	ApplicationTestsDelete(cx1client, logger, testname, &tests.Applications)
	ProjectTestsDelete(cx1client, logger, testname, &tests.Projects)
	RoleTestsDelete(cx1client, logger, testname, &tests.Roles)
	AccessTestsDelete(cx1client, logger, testname, &tests.AccessAssignments)
	UserTestsDelete(cx1client, logger, testname, &tests.Users)
	QueryTestsDelete(cx1client, logger, testname, &tests.Queries)
	PresetTestsDelete(cx1client, logger, testname, &tests.Presets)
	ScanTestsDelete(cx1client, logger, testname, &tests.Scans)
	ResultTestsDelete(cx1client, logger, testname, &tests.Results)
	ReportTestsDelete(cx1client, logger, testname, &tests.Reports)
}

func LogStart(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string) {
	logger.Infof("")
	logger.Infof("Starting %v %v Test '%v' #%d - %v", CRUD, Module, testName, testId, testObject)
}

func LogPass(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Errorf("FAIL [%.3fs]: %v %v FailTest '%v' #%d (%v) - %v", duration, CRUD, Module, testName, testId, testObject, "test passed but was expected to fail")
		TestResults = append(TestResults, TestResult{
			failTest, TST_FAIL, CRUD, Module, duration, testName, testId, testObject, "test passed but was expected to fail", testSource,
		})
	} else {
		logger.Infof("PASS [%.3fs]: %v %v Test '%v' #%d (%v)", duration, CRUD, Module, testName, testId, testObject)
		TestResults = append(TestResults, TestResult{
			failTest, TST_PASS, CRUD, Module, duration, testName, testId, testObject, "", testSource,
		})
	}
}
func LogSkip(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string, reason string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	logger.Warnf("SKIP [%.3fs]: %v %v Test '%v' #%d - %v", duration, CRUD, Module, testName, testId, reason)
	TestResults = append(TestResults, TestResult{
		failTest, TST_SKIP, CRUD, Module, duration, testName, testId, testObject, reason, testSource,
	})
}
func LogFail(failTest bool, logger *logrus.Logger, CRUD string, Module string, start int64, testName string, testId int, testObject string, testSource string, reason error) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Infof("PASS [%.3fs]: %v %v FailTest '%v' #%d (%v)", duration, CRUD, Module, testName, testId, testObject)
		TestResults = append(TestResults, TestResult{
			failTest, TST_PASS, CRUD, Module, duration, testName, testId, testObject, "", testSource,
		})
	} else {
		logger.Errorf("FAIL [%.3fs]: %v %v Test '%v' #%d (%v) - %s", duration, CRUD, Module, testName, testId, testObject, reason)
		TestResults = append(TestResults, TestResult{
			failTest, TST_FAIL, CRUD, Module, duration, testName, testId, testObject, reason.Error(), testSource,
		})
	}
}

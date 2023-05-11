package main

import (
	"crypto/tls"
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

var logger *logrus.Logger

var Config TestConfig
var TestResults []TestResult

func main() {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	myformatter := &easy.Formatter{}
	myformatter.TimestampFormat = "2006-01-02 15:04:05.000"
	myformatter.LogFormat = "[%lvl%][%time%] %msg%\n"
	logger.SetFormatter(myformatter)
	logger.SetOutput(os.Stdout)

	if len(os.Args) != 3 && len(os.Args) != 6 {
		logger.Info("The purpose of this tool is to automate testing of the API for various workflows based on the yaml configuration.")
		logger.Info("Expected arguments not provided. Usage:\n1)\tcx1e2e <test definition yaml file> <APIKey>\n")
		logger.Info("2)\tcx1e2e <test definition yaml file> <APIKey> <Cx1 URL> <IAM URL> <Tenant>\n")
		logger.Info("Note: API Key authentication is currently required and OIDC client/secret authentication is not supported.\n")
		return
	}

	var err error
	Config, err = LoadConfig(logger, os.Args[1])
	if err != nil {
		logger.Fatalf("Failed to load configuration file %v: %s", os.Args[1], err)
		return
	}

	switch strings.ToUpper(Config.LogLevel) {
	case "":
		break
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

	if len(os.Args) == 6 {
		Config.Cx1URL = os.Args[3]
		Config.IAMURL = os.Args[4]
		Config.Tenant = os.Args[5]
	}

	cx1client, err = Cx1ClientGo.NewAPIKeyClient(httpClient, Config.Cx1URL, Config.IAMURL, Config.Tenant, os.Args[2], logger)

	if err != nil {
		logger.Fatalf("Failed to create Cx1 client: %s", err)
		return
	}

	logger.Infof("Created Cx1 client %s", cx1client.String())

	RunTests(cx1client, logger)

	logger.Infof("Test result summary:\n")
	failed_test := false
	for _, result := range TestResults {
		switch result.Result {
		case 1:
			fmt.Printf("PASS %v - %v: %v\n", result.Name, result.CRUD, result.TestObject)
		case 0:
			fmt.Printf("FAIL %v - %v: %v\n", result.Name, result.CRUD, result.TestObject)
			failed_test = true
		case 2:
			fmt.Printf("SKIP %v - %v: %v\n", result.Name, result.CRUD, result.TestObject)
		}
	}
	fmt.Println("")
	if failed_test {
		logger.Errorf("Some tests failed")
	} else {
		logger.Infof("All tests passed")
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

	for _, set := range conf.Tests {
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
				if scan.ZipFile != "" {
					filePath, err := getFilePath(currentRoot, scan.ZipFile)
					if err != nil {
						return conf, fmt.Errorf("error locating scan zipfile %v", scan.ZipFile)
					}
					set.Scans[id].ZipFile = filePath
				}
			}
			testSet = append(testSet, set)
		}
	}
	conf.Tests = testSet

	return conf, nil
}

func getFilePath(currentRoot, file string) (string, error) {
	logger.Debugf("Trying to find config file %v, current root is %v", file, currentRoot)
	if _, err := os.Stat(file); err == nil {
		return filepath.Clean(file), nil
	} else {
		testPath := fmt.Sprintf("%v\\%v", currentRoot, file)
		logger.Debugf("File doesn't exist, testing: %v", testPath)
		if _, err := os.Stat(testPath); err == nil {
			return filepath.Clean(testPath), nil
		} else {
			return "", fmt.Errorf("unable to find configuration file %v", file)
		}
	}
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

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	for _, t := range Config.Tests {
		TestCreate(cx1client, logger, t.Name, &t)
		TestRead(cx1client, logger, t.Name, &t)
		TestUpdate(cx1client, logger, t.Name, &t)
		TestDelete(cx1client, logger, t.Name, &t)
	}
}

func TestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	if len(tests.Groups) > 0 {
		GroupTestsCreate(cx1client, logger, testname, &tests.Groups)
	}
	if len(tests.Applications) > 0 {
		ApplicationTestsCreate(cx1client, logger, testname, &tests.Applications)
	}
	if len(tests.Projects) > 0 {
		ProjectTestsCreate(cx1client, logger, testname, &tests.Projects)
	}
	if len(tests.Roles) > 0 {
		RoleTestsCreate(cx1client, logger, testname, &tests.Roles)
	}
	if len(tests.Users) > 0 {
		UserTestsCreate(cx1client, logger, testname, &tests.Users)
	}
	if len(tests.Queries) > 0 {
		QueryTestsCreate(cx1client, logger, testname, &tests.Queries)
	}
	if len(tests.Presets) > 0 {
		PresetTestsCreate(cx1client, logger, testname, &tests.Presets)
	}
	if len(tests.Scans) > 0 {
		ScanTestsCreate(cx1client, logger, testname, &tests.Scans)
	}
	if len(tests.Results) > 0 {
		ResultTestsCreate(cx1client, logger, testname, &tests.Results)
	}
}
func TestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	if len(tests.Groups) > 0 {
		GroupTestsRead(cx1client, logger, testname, &tests.Groups)
	}
	if len(tests.Applications) > 0 {
		ApplicationTestsRead(cx1client, logger, testname, &tests.Applications)
	}
	if len(tests.Projects) > 0 {
		ProjectTestsRead(cx1client, logger, testname, &tests.Projects)
	}
	if len(tests.Roles) > 0 {
		RoleTestsRead(cx1client, logger, testname, &tests.Roles)
	}
	if len(tests.Users) > 0 {
		UserTestsRead(cx1client, logger, testname, &tests.Users)
	}
	if len(tests.Queries) > 0 {
		QueryTestsRead(cx1client, logger, testname, &tests.Queries)
	}
	if len(tests.Presets) > 0 {
		PresetTestsRead(cx1client, logger, testname, &tests.Presets)
	}
	if len(tests.Scans) > 0 {
		ScanTestsRead(cx1client, logger, testname, &tests.Scans)
	}
	if len(tests.Results) > 0 {
		ResultTestsRead(cx1client, logger, testname, &tests.Results)
	}
}
func TestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	if len(tests.Groups) > 0 {
		GroupTestsUpdate(cx1client, logger, testname, &tests.Groups)
	}
	if len(tests.Applications) > 0 {
		ApplicationTestsUpdate(cx1client, logger, testname, &tests.Applications)
	}
	if len(tests.Projects) > 0 {
		ProjectTestsUpdate(cx1client, logger, testname, &tests.Projects)
	}
	if len(tests.Roles) > 0 {
		RoleTestsUpdate(cx1client, logger, testname, &tests.Roles)
	}
	if len(tests.Users) > 0 {
		UserTestsUpdate(cx1client, logger, testname, &tests.Users)
	}
	if len(tests.Queries) > 0 {
		QueryTestsUpdate(cx1client, logger, testname, &tests.Queries)
	}
	if len(tests.Presets) > 0 {
		PresetTestsUpdate(cx1client, logger, testname, &tests.Presets)
	}
	if len(tests.Scans) > 0 {
		ScanTestsUpdate(cx1client, logger, testname, &tests.Scans)
	}
	if len(tests.Results) > 0 {
		ResultTestsUpdate(cx1client, logger, testname, &tests.Results)
	}
}
func TestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) {
	if len(tests.Groups) > 0 {
		GroupTestsDelete(cx1client, logger, testname, &tests.Groups)
	}
	if len(tests.Applications) > 0 {
		ApplicationTestsDelete(cx1client, logger, testname, &tests.Applications)
	}
	if len(tests.Projects) > 0 {
		ProjectTestsDelete(cx1client, logger, testname, &tests.Projects)
	}
	if len(tests.Roles) > 0 {
		RoleTestsDelete(cx1client, logger, testname, &tests.Roles)
	}
	if len(tests.Users) > 0 {
		UserTestsDelete(cx1client, logger, testname, &tests.Users)
	}
	if len(tests.Queries) > 0 {
		QueryTestsDelete(cx1client, logger, testname, &tests.Queries)
	}
	if len(tests.Presets) > 0 {
		PresetTestsDelete(cx1client, logger, testname, &tests.Presets)
	}
	if len(tests.Scans) > 0 {
		ScanTestsDelete(cx1client, logger, testname, &tests.Scans)
	}
	if len(tests.Results) > 0 {
		ResultTestsDelete(cx1client, logger, testname, &tests.Results)
	}
}

func LogStart(failTest bool, logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string) {
	logger.Infof("")
	logger.Infof("Starting %v Test '%v' #%d - %v", CRUD, testName, testId, testObject)
}

func LogPass(failTest bool, logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Errorf("FAIL [%.3fs]: %v FailTest '%v' #%d (%v) - %v", duration, CRUD, testName, testId, testObject, "test passed but was expected to fail")
		TestResults = append(TestResults, TestResult{
			failTest, 0, CRUD, duration, testName, testId, testObject, "test passed but was expected to fail",
		})
	} else {
		logger.Infof("PASS [%.3fs]: %v Test '%v' #%d (%v)", duration, CRUD, testName, testId, testObject)
		TestResults = append(TestResults, TestResult{
			failTest, 1, CRUD, duration, testName, testId, testObject, "",
		})
	}
}
func LogSkip(failTest bool, logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string, reason string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	logger.Warnf("SKIP [%.3fs]: %v Test '%v' #%d - %v", duration, CRUD, testName, testId, reason)
	TestResults = append(TestResults, TestResult{
		failTest, 1, CRUD, duration, testName, testId, testObject, "",
	})
}
func LogFail(failTest bool, logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string, reason error) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	if failTest {
		logger.Infof("PASS [%.3fs]: %v FailTest '%v' #%d (%v)", duration, CRUD, testName, testId, testObject)
		TestResults = append(TestResults, TestResult{
			failTest, 1, CRUD, duration, testName, testId, testObject, "",
		})
	} else {
		logger.Errorf("FAIL [%.3fs]: %v Test '%v' #%d (%v) - %s", duration, CRUD, testName, testId, testObject, reason)
		TestResults = append(TestResults, TestResult{
			failTest, 0, CRUD, duration, testName, testId, testObject, reason.Error(),
		})
	}
}

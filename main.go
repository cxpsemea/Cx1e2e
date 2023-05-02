package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"gopkg.in/yaml.v2"
)

var logger *logrus.Logger

var Config TestConfig

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
	err = LoadConfig(os.Args[1])
	if err != nil {
		logger.Fatalf("Failed to load configuration file %v: %s", os.Args[1], err)
		return
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

	if RunTests(cx1client, logger) {
		logger.Info("All tests PASS")
	} else {
		logger.Error("Some tests FAIL")
	}
}

func LoadConfig(testconfig string) error {
	file, err := os.Open(testconfig)

	if err != nil {
		return err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)

	err = d.Decode(&Config)
	if err != nil {
		return err
	}

	return nil
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

func RunTests(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) bool {
	tests := true

	for _, t := range Config.Tests {
		if !TestCreate(cx1client, logger, t.Name, &t) {
			tests = false
		}
		if !TestRead(cx1client, logger, t.Name, &t) {
			tests = false
		}
		if !TestUpdate(cx1client, logger, t.Name, &t) {
			tests = false
		}
		if !TestDelete(cx1client, logger, t.Name, &t) {
			tests = false
		}
	}

	return tests
}

func TestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) bool {
	result := true
	if len(tests.Groups) > 0 && !GroupTestsCreate(cx1client, logger, testname, &tests.Groups) {
		result = false
	}

	if len(tests.Applications) > 0 && !ApplicationTestsCreate(cx1client, logger, testname, &tests.Applications) {
		result = false
	}

	if len(tests.Projects) > 0 && !ProjectTestsCreate(cx1client, logger, testname, &tests.Projects) {
		result = false
	}

	if len(tests.Roles) > 0 && !RoleTestsCreate(cx1client, logger, testname, &tests.Roles) {
		result = false
	}

	if len(tests.Users) > 0 && !UserTestsCreate(cx1client, logger, testname, &tests.Users) {
		result = false
	}

	if len(tests.Queries) > 0 && !QueryTestsCreate(cx1client, logger, testname, &tests.Queries) {
		result = false
	}

	if len(tests.Presets) > 0 && !PresetTestsCreate(cx1client, logger, testname, &tests.Presets) {
		result = false
	}

	if len(tests.Scans) > 0 && !ScanTestsCreate(cx1client, logger, testname, &tests.Scans) {
		result = false
	}

	if len(tests.Results) > 0 && !ResultTestsCreate(cx1client, logger, testname, &tests.Results) {
		result = false
	}

	return result
}
func TestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) bool {
	result := true
	if len(tests.Groups) > 0 && !GroupTestsRead(cx1client, logger, testname, &tests.Groups) {
		result = false
	}

	if len(tests.Applications) > 0 && !ApplicationTestsRead(cx1client, logger, testname, &tests.Applications) {
		result = false
	}

	if len(tests.Projects) > 0 && !ProjectTestsRead(cx1client, logger, testname, &tests.Projects) {
		result = false
	}

	if len(tests.Roles) > 0 && !RoleTestsRead(cx1client, logger, testname, &tests.Roles) {
		result = false
	}

	if len(tests.Users) > 0 && !UserTestsRead(cx1client, logger, testname, &tests.Users) {
		result = false
	}

	if len(tests.Queries) > 0 && !QueryTestsRead(cx1client, logger, testname, &tests.Queries) {
		result = false
	}

	if len(tests.Presets) > 0 && !PresetTestsRead(cx1client, logger, testname, &tests.Presets) {
		result = false
	}

	if len(tests.Scans) > 0 && !ScanTestsRead(cx1client, logger, testname, &tests.Scans) {
		result = false
	}

	if len(tests.Results) > 0 && !ResultTestsRead(cx1client, logger, testname, &tests.Results) {
		result = false
	}

	return result
}
func TestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) bool {
	result := true
	if len(tests.Groups) > 0 && !GroupTestsUpdate(cx1client, logger, testname, &tests.Groups) {
		result = false
	}

	if len(tests.Applications) > 0 && !ApplicationTestsUpdate(cx1client, logger, testname, &tests.Applications) {
		result = false
	}

	if len(tests.Projects) > 0 && !ProjectTestsUpdate(cx1client, logger, testname, &tests.Projects) {
		result = false
	}

	if len(tests.Roles) > 0 && !RoleTestsUpdate(cx1client, logger, testname, &tests.Roles) {
		result = false
	}

	if len(tests.Users) > 0 && !UserTestsUpdate(cx1client, logger, testname, &tests.Users) {
		result = false
	}

	if len(tests.Queries) > 0 && !QueryTestsUpdate(cx1client, logger, testname, &tests.Queries) {
		result = false
	}

	if len(tests.Presets) > 0 && !PresetTestsUpdate(cx1client, logger, testname, &tests.Presets) {
		result = false
	}

	if len(tests.Scans) > 0 && !ScanTestsUpdate(cx1client, logger, testname, &tests.Scans) {
		result = false
	}

	if len(tests.Results) > 0 && !ResultTestsUpdate(cx1client, logger, testname, &tests.Results) {
		result = false
	}

	return result
}
func TestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, tests *TestSet) bool {
	result := true
	if len(tests.Groups) > 0 && !GroupTestsDelete(cx1client, logger, testname, &tests.Groups) {
		result = false
	}

	if len(tests.Applications) > 0 && !ApplicationTestsDelete(cx1client, logger, testname, &tests.Applications) {
		result = false
	}

	if len(tests.Projects) > 0 && !ProjectTestsDelete(cx1client, logger, testname, &tests.Projects) {
		result = false
	}

	if len(tests.Roles) > 0 && !RoleTestsDelete(cx1client, logger, testname, &tests.Roles) {
		result = false
	}

	if len(tests.Users) > 0 && !UserTestsDelete(cx1client, logger, testname, &tests.Users) {
		result = false
	}

	if len(tests.Queries) > 0 && !QueryTestsDelete(cx1client, logger, testname, &tests.Queries) {
		result = false
	}

	if len(tests.Presets) > 0 && !PresetTestsDelete(cx1client, logger, testname, &tests.Presets) {
		result = false
	}

	if len(tests.Scans) > 0 && !ScanTestsDelete(cx1client, logger, testname, &tests.Scans) {
		result = false
	}

	if len(tests.Results) > 0 && !ResultTestsDelete(cx1client, logger, testname, &tests.Results) {
		result = false
	}

	return result
}

func LogPass(logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	logger.Infof("PASS [%.3fs]: %v Test '%v' #%d (%v)", duration, CRUD, testName, testId, testObject)
}
func LogSkip(logger *logrus.Logger, CRUD string, start int64, testName string, testId int, reason string) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	logger.Warnf("SKIP [%.3fs]: %v Test '%v' #%d - %v", duration, CRUD, testName, testId, reason)
}
func LogFail(logger *logrus.Logger, CRUD string, start int64, testName string, testId int, testObject string, reason error) {
	duration := float64(time.Now().UnixNano()-start) / float64(time.Second)
	logger.Errorf("FAIL [%.3fs]: %v Test '%v' #%d (%v) - %s", duration, CRUD, testName, testId, testObject, reason)
}

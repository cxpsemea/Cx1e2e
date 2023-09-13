package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"gopkg.in/yaml.v2"
)

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
	retval := run()

	if retval == 0 {
		os.Exit(1) // all tests failed
	}

	if retval == 1 {
		os.Exit(0) // all tests passed
	}

	os.Exit(2) // partial success
}

func run() float32 {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)
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
		return 0
	}

	/*switch strings.ToUpper(Config.LogLevel) {
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
	}*/

	var cx1client *Cx1ClientGo.Cx1Client
	httpClient := &http.Client{}

	if Config.ProxyURL != "" {
		proxyURL, err := url.Parse(Config.ProxyURL)
		if err != nil {
			logger.Fatalf("Failed to parse specified proxy address %v: %s", Config.ProxyURL, err)
			return 0
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
		return 0
	}

	logger.Infof("Created Cx1 client %s", cx1client.String())

	TestResults := RunTests(cx1client, logger, &Config)

	logger.Infof("Test result summary:\n")
	count_failed := 0
	count_passed := 0
	count_skipped := 0

	for _, result := range TestResults {
		var testtype = "Test"
		if result.FailTest {
			testtype = "Negative-Test"
		}
		switch result.Result {
		case 1:
			fmt.Printf("PASS %v - %v %v %v: %v\n", result.Name, result.CRUD, result.Module, testtype, result.TestObject)
			count_passed++
		case 0:
			fmt.Printf("FAIL %v - %v %v %v: %v\n", result.Name, result.CRUD, result.Module, testtype, result.TestObject)
			count_failed++
		case 2:
			fmt.Printf("SKIP %v - %v %v %v: %v\n", result.Name, result.CRUD, result.Module, testtype, result.TestObject)
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

	return float32(count_passed) / float32(count_failed+count_passed+count_skipped)
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return conf, err
	}

	re := regexp.MustCompile(`%([0-9a-zA-Z_]+)%`)
	fileContents := string(fileBytes)
	for matches := re.FindStringSubmatch(fileContents); len(matches) > 0; matches = re.FindStringSubmatch(fileContents) {
		fileContents = strings.ReplaceAll(fileContents, fmt.Sprintf("%%%v%%", matches[1]), os.Getenv(matches[1]))
	}

	d := yaml.NewDecoder(strings.NewReader(fileContents))

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

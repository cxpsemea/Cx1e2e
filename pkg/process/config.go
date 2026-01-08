package process

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cxpsemea/cx1e2e/pkg/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var LastTestID uint = 0

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

	//testSet := make([]TestSet, 0)

	for tid := range conf.Tests {
		conf.Tests[tid].TestSource = configPath
		set := &conf.Tests[tid]
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
			//testSet = append(testSet, conf2.Tests...)
			conf.Tests[tid].SubTests = conf2.Tests
			conf.Tests[tid].Thread = set.Thread
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
			//testSet = append(testSet, set)
		}

		conf.Tests[tid].Init()
	}

	return conf, nil
}

func (t *TestConfig) IsValid(logger *logrus.Logger) bool {
	failedTests := false
	for _, test := range t.Tests {
		if !test.IsValid(logger) {
			failedTests = true
		}
	}
	return !failedTests
}

func (t *TestConfig) InitTestIDs() {
	t.TestCount = t.GetTestCount()

	// IDs are set in the same order as execution in Runner
	// 1st: subtests
	// 2nd: CRU ops in order
	// last: D ops in reverse order
	for id := range t.Tests {
		t.Tests[id].InitTestIDs()
	}
}

func (t *TestSet) InitTestIDs() {
	// 1st: subtests
	// 2nd: CRU ops in order
	// last: D ops in reverse order
	for id := range t.SubTests {
		t.SubTests[id].InitTestIDs()
	}

	t.InitTestIDsCRUD(types.OP_CREATE)
	t.InitTestIDsCRUD(types.OP_READ)
	t.InitTestIDsCRUD(types.OP_UPDATE)
	t.InitTestIDsCRUD(types.OP_DELETE)
}

func (c TestConfig) PrintTests() {
	for id := range c.Tests {
		c.Tests[id].PrintTests()
	}
}

type AllCRUD interface {
	types.AccessAssignmentCRUD | types.AnalyticsCRUD
}

func (t TestSet) PrintTests() {
	for id := range t.SubTests {
		t.SubTests[id].PrintTests()
	}

	for id2 := range t.AccessAssignments {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.AccessAssignments[id2].TestID, t.AccessAssignments[id2].String())
	}
	for id2 := range t.Analytics {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Analytics[id2].TestID, t.Analytics[id2].String())
	}
	for id2 := range t.Applications {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Applications[id2].TestID, t.Applications[id2].String())
	}
	for id2 := range t.Branches {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Branches[id2].TestID, t.Branches[id2].String())
	}
	for id2 := range t.Clients {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Clients[id2].TestID, t.Clients[id2].String())
	}
	for id2 := range t.Flags {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Flags[id2].TestID, t.Flags[id2].String())
	}
	for id2 := range t.Groups {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Groups[id2].TestID, t.Groups[id2].String())
	}
	for id2 := range t.Imports {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Imports[id2].TestID, t.Imports[id2].String())
	}
	for id2 := range t.Presets {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Presets[id2].TestID, t.Presets[id2].String())
	}
	for id2 := range t.Projects {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Projects[id2].TestID, t.Projects[id2].String())
	}
	for id2 := range t.Queries {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Queries[id2].TestID, t.Queries[id2].String())
	}
	for id2 := range t.Reports {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Reports[id2].TestID, t.Reports[id2].String())
	}
	for id2 := range t.Results {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Results[id2].TestID, t.Results[id2].String())
	}
	for id2 := range t.Roles {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Roles[id2].TestID, t.Roles[id2].String())
	}
	for id2 := range t.Scans {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Scans[id2].TestID, t.Scans[id2].String())
	}
	for id2 := range t.Users {
		fmt.Printf("[%d] [%d] %v\n", t.Thread, t.Users[id2].TestID, t.Users[id2].String())
	}
}

func (t *TestSet) Init() {
	for id2 := range t.AccessAssignments {
		t.AccessAssignments[id2].TestSource = t.TestSource
		t.AccessAssignments[id2].Thread = t.Thread
	}
	for id2 := range t.Analytics {
		t.Analytics[id2].TestSource = t.TestSource
		t.Analytics[id2].Thread = t.Thread
	}
	for id2 := range t.Applications {
		t.Applications[id2].TestSource = t.TestSource
		t.Applications[id2].Thread = t.Thread
	}
	for id2 := range t.Branches {
		t.Branches[id2].TestSource = t.TestSource
		t.Branches[id2].Thread = t.Thread
	}
	for id2 := range t.Clients {
		t.Clients[id2].TestSource = t.TestSource
		t.Clients[id2].Thread = t.Thread
	}
	for id2 := range t.Flags {
		t.Flags[id2].TestSource = t.TestSource
		t.Flags[id2].Thread = t.Thread
	}
	for id2 := range t.Groups {
		t.Groups[id2].TestSource = t.TestSource
		t.Groups[id2].Thread = t.Thread
	}
	for id2 := range t.Imports {
		t.Imports[id2].TestSource = t.TestSource
		t.Imports[id2].Thread = t.Thread
	}
	for id2 := range t.Presets {
		t.Presets[id2].TestSource = t.TestSource
		t.Presets[id2].Thread = t.Thread
	}
	for id2 := range t.Projects {
		t.Projects[id2].TestSource = t.TestSource
		t.Projects[id2].Thread = t.Thread
	}
	for id2 := range t.Queries {
		t.Queries[id2].TestSource = t.TestSource
		t.Queries[id2].Thread = t.Thread
	}
	for id2 := range t.Reports {
		t.Reports[id2].TestSource = t.TestSource
		t.Reports[id2].Thread = t.Thread
	}
	for id2 := range t.Results {
		t.Results[id2].TestSource = t.TestSource
		t.Results[id2].Thread = t.Thread
		if t.Results[id2].Number == 0 {
			t.Results[id2].Number = 1
		}
	}
	for id2 := range t.Roles {
		t.Roles[id2].TestSource = t.TestSource
		t.Roles[id2].Thread = t.Thread
	}
	for id2 := range t.Scans {
		t.Scans[id2].TestSource = t.TestSource
		t.Scans[id2].Thread = t.Thread
	}
	for id2 := range t.Users {
		t.Users[id2].TestSource = t.TestSource
		t.Users[id2].Thread = t.Thread
	}

	for id2 := range t.SubTests {
		if t.Thread != 0 {
			t.SubTests[id2].Thread = t.Thread
		}
		t.SubTests[id2].Init()
	}
}

func (t *TestSet) InitTestIDsCRUD(CRUD string) {
	if CRUD != types.OP_DELETE {
		for id := range t.Flags {
			if t.Flags[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Flags[id].TestID = LastTestID
			}
		}
		for id := range t.Analytics {
			if t.Analytics[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Analytics[id].TestID = LastTestID
			}
		}
		for id := range t.Imports {
			if t.Imports[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Imports[id].TestID = LastTestID
			}
		}
		for id := range t.Groups {
			if t.Groups[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Groups[id].TestID = LastTestID
			}
		}
		for id := range t.Applications {
			if t.Applications[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Applications[id].TestID = LastTestID
			}
		}
		for id := range t.Projects {
			if t.Projects[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Projects[id].TestID = LastTestID
			}
		}
		for id := range t.Roles {
			if t.Roles[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Roles[id].TestID = LastTestID
			}
		}
		for id := range t.Users {
			if t.Users[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Users[id].TestID = LastTestID
			}
		}
		for id := range t.Clients {
			if t.Clients[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Clients[id].TestID = LastTestID
			}
		}
		for id := range t.AccessAssignments {
			if t.AccessAssignments[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.AccessAssignments[id].TestID = LastTestID
			}
		}
		for id := range t.Queries {
			if t.Queries[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Queries[id].TestID = LastTestID
			}
		}
		for id := range t.Presets {
			if t.Presets[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Presets[id].TestID = LastTestID
			}
		}
		for id := range t.Scans {
			if t.Scans[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Scans[id].TestID = LastTestID
			}
		}
		for id := range t.Branches {
			if t.Branches[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Branches[id].TestID = LastTestID
			}
		}
		for id := range t.Results {
			if t.Results[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Results[id].TestID = LastTestID
			}
		}
		for id := range t.Reports {
			if t.Reports[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Reports[id].TestID = LastTestID
			}
		}
	} else { // in reverse order for DELETE
		for id := range t.Scans {
			if t.Scans[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Scans[id].TestID = LastTestID
			}
		}
		for id := range t.Presets {
			if t.Presets[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Presets[id].TestID = LastTestID
			}
		}
		for id := range t.Queries {
			if t.Queries[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Queries[id].TestID = LastTestID
			}
		}
		for id := range t.AccessAssignments {
			if t.AccessAssignments[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.AccessAssignments[id].TestID = LastTestID
			}
		}
		for id := range t.Clients {
			if t.Clients[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Clients[id].TestID = LastTestID
			}
		}
		for id := range t.Users {
			if t.Users[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Users[id].TestID = LastTestID
			}
		}
		for id := range t.Roles {
			if t.Roles[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Roles[id].TestID = LastTestID
			}
		}
		for id := range t.Projects {
			if t.Projects[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Projects[id].TestID = LastTestID
			}
		}
		for id := range t.Applications {
			if t.Applications[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Applications[id].TestID = LastTestID
			}
		}
		for id := range t.Groups {
			if t.Groups[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Groups[id].TestID = LastTestID
			}
		}
	}
}

func (t *TestSet) SetActiveThread(thread int) {
	t.ActiveThread = thread
	for id2 := range t.AccessAssignments {
		t.AccessAssignments[id2].ActiveThread = thread
	}
	for id2 := range t.Analytics {
		t.Analytics[id2].ActiveThread = thread
	}
	for id2 := range t.Applications {
		t.Applications[id2].ActiveThread = thread
	}
	for id2 := range t.Branches {
		t.Branches[id2].ActiveThread = thread
	}
	for id2 := range t.Clients {
		t.Clients[id2].ActiveThread = thread
	}
	for id2 := range t.Flags {
		t.Flags[id2].ActiveThread = thread
	}
	for id2 := range t.Groups {
		t.Groups[id2].ActiveThread = thread
	}
	for id2 := range t.Imports {
		t.Imports[id2].ActiveThread = thread
	}
	for id2 := range t.Presets {
		t.Presets[id2].ActiveThread = thread
	}
	for id2 := range t.Projects {
		t.Projects[id2].ActiveThread = thread
	}
	for id2 := range t.Queries {
		t.Queries[id2].ActiveThread = thread
	}
	for id2 := range t.Reports {
		t.Reports[id2].ActiveThread = thread
	}
	for id2 := range t.Results {
		t.Results[id2].ActiveThread = thread
	}
	for id2 := range t.Roles {
		t.Roles[id2].ActiveThread = thread
	}
	for id2 := range t.Scans {
		t.Scans[id2].ActiveThread = thread
	}
	for id2 := range t.Users {
		t.Users[id2].ActiveThread = thread
	}

	for id2 := range t.SubTests {
		t.SubTests[id2].SetActiveThread(thread)
	}
}

func (t TestSet) IsValid(logger *logrus.Logger) bool {
	failedTests := false

	for id2 := range t.AccessAssignments {
		if !isTestValid(&t.AccessAssignments[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Analytics {
		if !isTestValid(&t.Analytics[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Applications {
		if !isTestValid(&t.Applications[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Branches {
		if !isTestValid(&t.Branches[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Clients {
		if !isTestValid(&t.Clients[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Flags {
		if !isTestValid(&t.Flags[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Groups {
		if !isTestValid(&t.Groups[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Imports {
		if !isTestValid(&t.Imports[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Presets {
		if !isTestValid(&t.Presets[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Projects {
		if !isTestValid(&t.Projects[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Queries {
		if !isTestValid(&t.Queries[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Reports {
		if !isTestValid(&t.Reports[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Results {
		if !isTestValid(&t.Results[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Roles {
		if !isTestValid(&t.Roles[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Scans {
		if !isTestValid(&t.Scans[id2], logger) {
			failedTests = true
		}
	}
	for id2 := range t.Users {
		if !isTestValid(&t.Users[id2], logger) {
			failedTests = true
		}
	}

	for id2 := range t.SubTests {
		if !t.SubTests[id2].IsValid(logger) {
			failedTests = true
		}
	}

	return !failedTests
}

func isTestValid(runner TestRunner, logger *logrus.Logger) bool {
	failedTest := false
	for _, test := range []string{types.OP_CREATE, types.OP_READ, types.OP_UPDATE, types.OP_DELETE} {
		if runner.IsType(test) {
			if err := runner.Validate(test); err != nil {
				logger.Infof("Test [%v] %v %v is invalid: %v", runner.GetSource(), runner.String(), test, err)
				failedTest = true
			}
		}
	}

	return !failedTest
}

func (t TestSet) GetTestCount() int {
	var count int = 0

	count += len(t.AccessAssignments)
	count += len(t.Applications)
	count += len(t.Branches)
	count += len(t.Clients)
	count += len(t.Flags)
	count += len(t.Groups)
	count += len(t.Imports)
	count += len(t.Presets)
	count += len(t.Projects)
	count += len(t.Queries)
	count += len(t.Reports)
	count += len(t.Results)
	count += len(t.Roles)
	count += len(t.Scans)
	count += len(t.Users)

	for id := range t.SubTests {
		count += t.SubTests[id].GetTestCount()
	}

	return count
}

func (o TestConfig) GetTestCount() int {
	var count int = 0
	for id := range o.Tests {
		count += o.Tests[id].GetTestCount()
	}
	return count
}

func (o TestConfig) CreateHTTPClient(logger *logrus.Logger) (*http.Client, error) {
	httpClient := &http.Client{}
	transport := &http.Transport{}

	if o.ProxyURL != "" {
		proxyURL, err := url.Parse(o.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse specified proxy address %v: %s", o.ProxyURL, err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		logger.Infof("Running with proxy: %v", o.ProxyURL)
	}

	if o.NoTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		logger.Warn("Running without TLS verification")
	}

	if o.IPv4 {
		logger.Infof("Running with IPv4 only")
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("tcp4", addr)
		}
	} else if o.IPv6 {
		logger.Infof("Running with IPv6 only")
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("tcp6", addr)
		}
	}

	httpClient.Transport = transport
	return httpClient, nil
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

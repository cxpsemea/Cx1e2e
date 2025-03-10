package process

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cxpsemea/cx1e2e/pkg/types"
	"github.com/hashicorp/go-retryablehttp"
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

	// propagate the filename to sub-tests
	// TODO: refactor this to use generics?
	for id := range conf.Tests {
		conf.Tests[id].TestSource = configPath
		conf.Tests[id].Init()
	}

	for tid := range conf.Tests {
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
	}
	//conf.Tests = testSet
	return conf, nil
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

func (t *TestSet) Init() {
	for id2 := range t.AccessAssignments {
		t.AccessAssignments[id2].TestSource = t.TestSource
	}
	for id2 := range t.Applications {
		t.Applications[id2].TestSource = t.TestSource
	}
	for id2 := range t.Clients {
		t.Clients[id2].TestSource = t.TestSource
	}
	for id2 := range t.Flags {
		t.Flags[id2].TestSource = t.TestSource
	}
	for id2 := range t.Groups {
		t.Groups[id2].TestSource = t.TestSource
	}
	for id2 := range t.Imports {
		t.Imports[id2].TestSource = t.TestSource
	}
	for id2 := range t.Presets {
		t.Presets[id2].TestSource = t.TestSource
	}
	for id2 := range t.Projects {
		t.Projects[id2].TestSource = t.TestSource
	}
	for id2 := range t.Queries {
		t.Queries[id2].TestSource = t.TestSource
	}
	for id2 := range t.Reports {
		t.Reports[id2].TestSource = t.TestSource
	}
	for id2 := range t.Results {
		t.Results[id2].TestSource = t.TestSource
		if t.Results[id2].Number == 0 {
			t.Results[id2].Number = 1
		}
	}
	for id2 := range t.Roles {
		t.Roles[id2].TestSource = t.TestSource
	}
	for id2 := range t.Scans {
		t.Scans[id2].TestSource = t.TestSource
	}
	for id2 := range t.Users {
		t.Users[id2].TestSource = t.TestSource
	}
}

func (t *TestSet) InitTestIDsCRUD(CRUD string) {
	debug := true
	if CRUD != types.OP_DELETE {
		for id := range t.Flags {
			if t.Flags[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Flags[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Flags[id].String())
				}
			}
		}
		for id := range t.Imports {
			if t.Imports[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Imports[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Imports[id].String())
				}
			}
		}
		for id := range t.Groups {
			if t.Groups[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Groups[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Groups[id].String())
				}
			}
		}
		for id := range t.Applications {
			if t.Applications[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Applications[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Applications[id].String())
				}
			}
		}
		for id := range t.Projects {
			if t.Projects[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Projects[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Projects[id].String())
				}
			}
		}
		for id := range t.Roles {
			if t.Roles[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Roles[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Roles[id].String())
				}
			}
		}
		for id := range t.Users {
			if t.Users[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Users[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Users[id].String())
				}
			}
		}
		for id := range t.Clients {
			if t.Clients[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Clients[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Clients[id].String())
				}
			}
		}
		for id := range t.AccessAssignments {
			if t.AccessAssignments[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.AccessAssignments[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.AccessAssignments[id].String())
				}
			}
		}
		for id := range t.Queries {
			if t.Queries[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Queries[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Queries[id].String())
				}
			}
		}
		for id := range t.Presets {
			if t.Presets[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Presets[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Presets[id].String())
				}
			}
		}
		for id := range t.Scans {
			if t.Scans[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Scans[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Scans[id].String())
				}
			}
		}
		for id := range t.Results {
			if t.Results[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Results[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Results[id].String())
				}
			}
		}
		for id := range t.Reports {
			if t.Reports[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Reports[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Reports[id].String())
				}
			}
		}
	} else { // in reverse order for DELETE
		for id := range t.Scans {
			if t.Scans[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Scans[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Scans[id].String())
				}
			}
		}
		for id := range t.Presets {
			if t.Presets[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Presets[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Presets[id].String())
				}
			}
		}
		for id := range t.Queries {
			if t.Queries[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Queries[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Queries[id].String())
				}
			}
		}
		for id := range t.AccessAssignments {
			if t.AccessAssignments[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.AccessAssignments[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.AccessAssignments[id].String())
				}
			}
		}
		for id := range t.Clients {
			if t.Clients[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Clients[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Clients[id].String())
				}
			}
		}
		for id := range t.Users {
			if t.Users[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Users[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Users[id].String())
				}
			}
		}
		for id := range t.Roles {
			if t.Roles[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Roles[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Roles[id].String())
				}
			}
		}
		for id := range t.Projects {
			if t.Projects[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Projects[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Projects[id].String())
				}
			}
		}
		for id := range t.Applications {
			if t.Applications[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Applications[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Applications[id].String())
				}
			}
		}
		for id := range t.Groups {
			if t.Groups[id].CRUDTest.IsType(CRUD) {
				LastTestID++
				t.Groups[id].TestID = LastTestID
				if debug {
					fmt.Printf("%d: %v %v\n", LastTestID, CRUD, t.Groups[id].String())
				}
			}
		}
	}
}

func (t TestSet) GetTestCount() int {
	var count int = 0

	count += len(t.AccessAssignments)
	count += len(t.Applications)
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
	leveledlogger := LeveledLogger{logger: logger}
	cx1retryclient := retryablehttp.NewClient()
	cx1retryclient.RetryMax = 3
	cx1retryclient.Logger = leveledlogger
	httpClient := cx1retryclient.StandardClient()

	if o.ProxyURL != "" {
		proxyURL, err := url.Parse(o.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse specified proxy address %v: %s", o.ProxyURL, err)
		}
		transport := &http.Transport{}
		transport.Proxy = http.ProxyURL(proxyURL)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		httpClient.Transport = transport
		logger.Infof("Running with proxy: %v", o.ProxyURL)
	} else if o.NoTLS {
		transport := &http.Transport{}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		httpClient.Transport = transport
		logger.Info("Running without TLS verification")
	}

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

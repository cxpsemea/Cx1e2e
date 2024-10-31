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

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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

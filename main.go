package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/cxpsemea/cx1e2e/pkg/process"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
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
	Config, err := process.LoadConfig(logger, *testConfig)
	if err != nil {
		logger.Fatalf("Failed to load configuration file %v: %s", *testConfig, err)
		return 0
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

	TestResults := process.RunTests(cx1client, logger, &Config)

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

	err = process.GenerateReport(&TestResults, &Config)
	if err != nil {
		logger.Errorf("Failed to generate HTML report: %s", err)
	}

	return float32(count_passed) / float32(count_failed+count_passed+count_skipped)
}

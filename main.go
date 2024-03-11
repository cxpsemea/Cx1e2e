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
	LogLevel := flag.String("log", "INFO", "Log level: TRACE, DEBUG, INFO, WARNING, ERROR, FATAL")
	ReportType := flag.String("report-type", "html,json", "Report output format: html or json")
	ReportName := flag.String("report-name", "cx1e2e_result", "Report output base name")
	Engines := flag.String("engines", "sast,sca,kics,apisec", "Run tests only for these engines")

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

	if *LogLevel == "" {
		*LogLevel = Config.LogLevel
	}

	switch strings.ToUpper(*LogLevel) {
	case "TRACE":
		logger.Info("Setting log level to TRACE")
		logger.SetLevel(logrus.TraceLevel)
	case "DEBUG":
		logger.Info("Setting log level to DEBUG")
		logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		logger.Info("Setting log level to INFO")
		logger.SetLevel(logrus.InfoLevel)
	case "WARNING":
		logger.Info("Setting log level to WARNING")
		logger.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logger.Info("Setting log level to ERROR")
		logger.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logger.Info("Setting log level to FATAL")
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.Info("Log level set to default: INFO")
	}

	if *ReportName != "" {
		Config.ReportName = *ReportName
	}
	if Config.ReportName == "" {
		Config.ReportName = "cx1e2e_result"
	}

	if *ReportType != "" {
		Config.ReportType = strings.ToLower(*ReportType)
	}
	if Config.ReportType == "" {
		Config.ReportType = "html,json"
	} else {
		if Config.ReportType != "html" && Config.ReportType != "json" && Config.ReportType != "html,json" {
			logger.Errorf("Supplied report type (%v) is invalid, using default", Config.ReportType)
			Config.ReportType = "html,json"
		}
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
		Config.AuthType = fmt.Sprintf("APIKey %v", Cx1ClientGo.ShortenGUID(*APIKey))
	} else {
		cx1client, err = Cx1ClientGo.NewOAuthClient(httpClient, Config.Cx1URL, Config.IAMURL, Config.Tenant, *ClientID, *ClientSecret, logger)
		Config.AuthType = fmt.Sprintf("OAuth client %v", *ClientID)
	}

	if err != nil {
		logger.Fatalf("Failed to create Cx1 client: %s", err)
		return 0
	}

	logger.Infof("Created Cx1 client %s", cx1client.String())
	currentUser, err := cx1client.Whoami()
	if err != nil {
		logger.Fatalf("Failed to get cx1 client current user: %s", err)
	}
	Config.AuthUser = currentUser.String()
	Config.EnvironmentVersion, err = cx1client.GetVersion()
	if err != nil {
		logger.Errorf("Failed to get version info: %s", err)
	}
	logger.Infof("Cx1 version: %v", Config.EnvironmentVersion.String())

	EngineList := strings.Split(strings.ToLower(*Engines), ",")
	for _, e := range EngineList {
		switch e {
		case "sast":
			Config.Engines.SAST = true
		case "sca":
			Config.Engines.SCA = true
		case "kics":
			Config.Engines.KICS = true
		case "apisec":
			Config.Engines.APISEC = true
		}
	}

	return process.RunTests(cx1client, logger, &Config)
}

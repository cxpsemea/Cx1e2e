package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/cxpsemea/cx1e2e/pkg/process"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func main() {
	os.Exit(int(run())) // returns the number of tests that failed
}

func run() uint {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	myformatter := &easy.Formatter{}
	myformatter.TimestampFormat = "2006-01-02 15:04:05.000"
	myformatter.LogFormat = "[%lvl%][%time%] %msg%\n"
	logger.SetFormatter(myformatter)
	logger.SetOutput(os.Stdout)

	logger.Warnf("Notice: the cx1e2e repo is likely to become internal in the future. If you require continued access, please reach out to your CSM or Checkmarx contact.")

	testConfig := flag.String("config", "", "Path to a test config.yaml")
	APIKey := flag.String("apikey", "", "CheckmarxOne API Key (if not using client id/secret)")
	ClientID := flag.String("client", "", "CheckmarxOne Client ID (if not using API Key)")
	ClientSecret := flag.String("secret", "", "CheckmarxOne Client Secret (if not using API Key)")
	Cx1URL := flag.String("cx1", "", "Optional: CheckmarxOne platform URL, if not defined in the test config.yaml")
	IAMURL := flag.String("iam", "", "Optional: CheckmarxOne IAM URL, if not defined in the test config.yaml")
	Tenant := flag.String("tenant", "", "Optional: CheckmarxOne tenant, if not defined in the test config.yaml")
	LogLevel := flag.String("log", "", "Log level: TRACE, DEBUG, INFO, WARNING, ERROR, FATAL")
	ReportType := flag.String("report-type", "html,json", "Report output format: html or json")
	ReportName := flag.String("report-name", "cx1e2e_result", "Report output base name")
	Engines := flag.String("engines", "sast,sca,iac,apisec,2ms,containers", "Run tests only for these engines")
	Proxy := flag.String("proxy", "", "Optional: Proxy to use when connecting to CheckmarxOne")
	NoTLS := flag.Bool("notls", false, "Optional: Disable TLS verification")
	Threads := flag.Int("threads", 1, "How many concurrent tests to run")
	LogFile := flag.String("logfile", "", "Optional: output log to file")
	InlineReport := flag.Bool("inline-report", false, "Print the report (json/html) contents at the end of execution")

	flag.Parse()

	if *LogFile != "" {
		file, err := os.OpenFile(*LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			logger.Errorf("Failed to create log file '%v': %v", *LogFile, err)
		}
		mw := io.MultiWriter(os.Stdout, file)
		logger.SetOutput(mw)
		logger.Infof("Logging to file %v", *LogFile)
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

	if *testConfig == "" || (*APIKey == "" && (*ClientID == "" || *ClientSecret == "")) {
		logger.Info("The purpose of this tool is to automate testing of the API for various workflows based on the yaml configuration. For help run: cx1e2e.exe -h")
		logger.Errorf("Test configuration yaml or authentication (API Key or client+secret) not provided.")
		return 1
	}

	var err error
	Config, err := process.LoadConfig(logger, *testConfig)
	if err != nil {
		logger.Errorf("Failed to load configuration file %v: %s", *testConfig, err)
		return 1
	}

	if !Config.IsValid(logger) {
		logger.Errorf("Test configuration failed to validate - review the logs and update the YAMLs")
		return 1
	}

	if *LogLevel == "" && Config.LogLevel != "" {
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
		}
		*LogLevel = Config.LogLevel
	}

	if *Threads <= 0 {
		*Threads = 1
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

	if *Proxy != "" {
		Config.ProxyURL = *Proxy
	}

	if *NoTLS {
		Config.NoTLS = true
	}

	httpClient, err := Config.CreateHTTPClient(logger)
	if err != nil {
		logger.Errorf("Failed to create HTTP client: %s", err)
		return 1
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
		logger.Errorf("Failed to create Cx1 client: %s", err)
		return 1
	}

	logger.Infof("Created Cx1 client %s", cx1client.String())
	if cx1client.IsUser {
		currentUser, err := cx1client.GetCurrentUser()
		if err != nil {
			logger.Errorf("Failed to get cx1 client current user: %s", err)
		} else {
			Config.AuthUser = currentUser.String()
		}
	} else {
		currentClient, err := cx1client.GetCurrentClient()
		if err != nil {
			logger.Errorf("Failed to get cx1 client current OIDC client: %s", err)
		} else {
			Config.AuthUser = currentClient.String()
		}
	}

	// both currentUser/currentClient checks may fail in which case:
	if Config.AuthUser == "" {
		claim := cx1client.GetClaims()
		if claim.Username != "" {
			Config.AuthUser = fmt.Sprintf("%v (%v)", claim.Username, claim.Email)
		} else if claim.ClientID != "" {
			Config.AuthUser = fmt.Sprintf("%v (%v)", claim.ClientID, claim.Email)
		} else {
			Config.AuthUser = claim.Email
		}
	}

	Config.EnvironmentVersion, err = cx1client.GetVersion()
	if err != nil {
		logger.Errorf("Failed to get version info: %s", err)
	}
	logger.Infof("Cx1 version: %v", Config.EnvironmentVersion.String())
	Config.InlineReport = *InlineReport

	EngineList := strings.Split(strings.ToLower(*Engines), ",")
	for _, e := range EngineList {
		switch e {
		case "sast":
			Config.Engines.SAST = true
		case "sca":
			Config.Engines.SCA = true
		case "kics", "iac":
			Config.Engines.IAC = true
		case "apisec":
			Config.Engines.APISEC = true
		case "2ms", "secrets":
			Config.Engines.Secrets = true
		case "containers":
			Config.Engines.Containers = true
		}
	}

	Config.InitTestIDs()

	return process.RunTests(cx1client, logger, &Config, *Threads)
}

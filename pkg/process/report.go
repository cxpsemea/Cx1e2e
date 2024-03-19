package process

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cxpsemea/cx1e2e/pkg/types"
	"github.com/sirupsen/logrus"
)

func prepareReportData(tests *[]TestResult, Config *TestConfig) Report {
	var report Report
	report.Settings.Target = fmt.Sprintf("%v tenant %v", Config.Cx1URL, Config.Tenant)
	report.Settings.Auth = fmt.Sprintf("%v user %v", Config.AuthType, Config.AuthUser)
	report.Settings.Config = Config.ConfigPath
	report.Settings.Timestamp = time.Now().String()
	report.Settings.E2ESuffix = os.Getenv("E2E_RUN_SUFFIX")

	for _, r := range *tests {
		report.AddTest(&r)
	}

	return report
}

func (c *Counter) AddTest(t *TestResult) {
	switch t.Result {
	case TST_PASS:
		c.Pass++
	case TST_FAIL:
		c.Fail++
	case TST_SKIP:
		c.Skip++
	}
}

func (c *CounterSet) AddTest(t *TestResult) {
	switch t.CRUD {
	case types.OP_CREATE:
		c.Create.AddTest(t)
	case types.OP_READ:
		c.Read.AddTest(t)
	case types.OP_UPDATE:
		c.Update.AddTest(t)
	case types.OP_DELETE:
		c.Delete.AddTest(t)
	}
}

func (s *ReportSummary) AddTest(t *TestResult) {
	switch t.Module {
	case types.MOD_ACCESS:
		s.Area.Access.AddTest(t)
	case types.MOD_APPLICATION:
		s.Area.Application.AddTest(t)
	case types.MOD_FLAG:
		s.Area.Flag.AddTest(t)
	case types.MOD_GROUP:
		s.Area.Group.AddTest(t)
	case types.MOD_IMPORT:
		s.Area.Import.AddTest(t)
	case types.MOD_PRESET:
		s.Area.Preset.AddTest(t)
	case types.MOD_PROJECT:
		s.Area.Project.AddTest(t)
	case types.MOD_QUERY:
		s.Area.Query.AddTest(t)
	case types.MOD_RESULT:
		s.Area.Result.AddTest(t)
	case types.MOD_REPORT:
		s.Area.Report.AddTest(t)
	case types.MOD_ROLE:
		s.Area.Role.AddTest(t)
	case types.MOD_SCAN:
		s.Area.Scan.AddTest(t)
	case types.MOD_USER:
		s.Area.User.AddTest(t)
	}

	switch t.Result {
	case TST_PASS:
		s.Total.Pass++
	case TST_SKIP:
		s.Total.Skip++
	case TST_FAIL:
		s.Total.Fail++
	}
}

func (r *Report) AddTest(t *TestResult) {
	r.Summary.AddTest(t)

	testtype := "Test"
	if t.FailTest {
		testtype = "Negative-Test"
	}

	details := ReportTestDetails{
		Name:       t.Name,
		Source:     t.TestSource,
		Test:       fmt.Sprintf("%v %v %v: %v", t.CRUD, t.Module, testtype, t.TestObject),
		Duration:   t.Duration,
		ResultType: t.Result,
	}

	switch t.Result {
	case TST_PASS:
		details.Result = "PASS"
	case TST_FAIL:
		details.Result = fmt.Sprintf("FAIL: %v", t.Reason)
	case TST_SKIP:
		details.Result = fmt.Sprintf("SKIP: %v", t.Reason)
	}

	r.Details = append(r.Details, details)
}

func (d ReportTestDetails) String() string {
	result := "PASS"
	switch d.ResultType {
	case TST_FAIL:
		result = "FAIL"
	case TST_SKIP:
		result = "SKIP"
	}

	return fmt.Sprintf("%v %v - %v", result, d.Name, d.Test)
}

func OutputSummaryConsole(reportData *Report, logger *logrus.Logger) {
	fmt.Println("Test result summary:")
	for _, r := range reportData.Details {
		fmt.Println(r.String())
	}

	fmt.Println("")
	fmt.Printf("Ran %d tests\n", (reportData.Summary.Total.Fail + reportData.Summary.Total.Pass + reportData.Summary.Total.Skip))
	if reportData.Summary.Total.Fail > 0 {
		fmt.Printf("FAILED %d tests\n", reportData.Summary.Total.Fail)
	}
	if reportData.Summary.Total.Skip > 0 {
		fmt.Printf("SKIPPED %d tests\n", reportData.Summary.Total.Skip)
	}
	if reportData.Summary.Total.Pass > 0 {
		fmt.Printf("PASSED %d tests\n", reportData.Summary.Total.Pass)
	}

}

func OutputReportHTML(reportName string, reportData *Report, Config *TestConfig) error {
	report, err := os.Create(reportName)
	if err != nil {
		return err
	}

	defer report.Close()
	_, err = report.WriteString(fmt.Sprintf("<html><head><title>%v tenant %v test - %v</title></head><body>", Config.Cx1URL, Config.Tenant, time.Now().String()))
	if err != nil {
		return err
	}

	report.WriteString("<h2>Settings</h2>")
	report.WriteString(fmt.Sprintf("Running end to end tests against %v<br>", reportData.Settings.Target))
	report.WriteString(fmt.Sprintf("Target versions are: %v", reportData.Settings.Version.String()))
	report.WriteString(fmt.Sprintf("Authenticated using %v<br>", reportData.Settings.Auth))
	report.WriteString(fmt.Sprintf("Test set defined in configuration %v<br>", reportData.Settings.Config))
	report.WriteString(fmt.Sprintf("Execution timestamp: %v.<br>", reportData.Settings.Timestamp))
	if os.Getenv("E2E_RUN_SUFFIX") == "" {
		report.WriteString(fmt.Sprintf("Default object name suffix %%E2E_RUN_SUFFIX%% environment variable is blank. Objects created by cx1e2e will use default names.<br>"))
	} else {
		report.WriteString(fmt.Sprintf("Default object name suffix %%E2E_RUN_SUFFIX%% environment variable is set to %v. Objects created by cx1e2e will use this suffix in the name.<br>", os.Getenv("E2E_RUN_SUFFIX")))
	}

	report.WriteString("<h2>Summary</h2>")

	report.WriteString(fmt.Sprintf("<p>Test status:<br>FAIL: %d<br>SKIP: %d<br>PASS:%d<br></p>", reportData.Summary.Total.Fail, reportData.Summary.Total.Skip, reportData.Summary.Total.Pass))

	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th rowspan=2>Area</th><th colspan=3>Create</th><th colspan=3>Read</th><th colspan=3>Update</th><th colspan=3>Delete</th></tr>\n")
	report.WriteString("<tr><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th></tr>\n")
	writeCounterSet(report, "Access Assignment", &reportData.Summary.Area.Access)
	writeCounterSet(report, "Application", &reportData.Summary.Area.Application)
	writeCounterSet(report, "Flag", &reportData.Summary.Area.Flag)
	writeCounterSet(report, "Group", &reportData.Summary.Area.Group)
	writeCounterSet(report, "Import", &reportData.Summary.Area.Import)
	writeCounterSet(report, "Preset", &reportData.Summary.Area.Preset)
	writeCounterSet(report, "Project", &reportData.Summary.Area.Project)
	writeCounterSet(report, "Query", &reportData.Summary.Area.Query)
	writeCounterSet(report, "Result", &reportData.Summary.Area.Result)
	writeCounterSet(report, "Report", &reportData.Summary.Area.Report)
	writeCounterSet(report, "Role", &reportData.Summary.Area.Role)
	writeCounterSet(report, "Scan", &reportData.Summary.Area.Scan)
	writeCounterSet(report, "User", &reportData.Summary.Area.User)
	report.WriteString("</table><br>")

	report.WriteString("<h2>Details</h2>")
	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th>Test Set</th><th>Test</th><th>Duration (sec)</th><th>Result</th></tr>\n")

	for _, t := range reportData.Details {
		switch t.ResultType {
		case TST_PASS:
			report.WriteString(fmt.Sprintf("<tr><td>%v<br>(%v)</td><td>%v</td><td>%.2f</td><td><span style='color:green'>%v</span></td></tr>\n", t.Name, t.Source, t.Test, t.Duration, t.Result))
		case TST_SKIP:
			report.WriteString(fmt.Sprintf("<tr><td>%v<br>(%v)</td><td>%v</td><td>%.2f</td><td><span style='color:orange'>%v</span></td></tr>\n", t.Name, t.Source, t.Test, t.Duration, t.Result))
		case TST_FAIL:
			report.WriteString(fmt.Sprintf("<tr><td>%v<br>(%v)</td><td>%v</td><td>%.2f</td><td><span style='color:red'>%v</span></td></tr>\n", t.Name, t.Source, t.Test, t.Duration, t.Result))
		}
	}

	report.WriteString("</table>\n")

	_, err = report.WriteString("</body></html>")
	if err != nil {
		return err
	}

	return report.Sync()
}

func OutputReportJSON(reportName string, reportData *Report) error {
	report, err := os.Create(reportName)
	if err != nil {
		return err
	}

	defer report.Close()

	json, err := json.Marshal(*reportData)
	if err != nil {
		return err
	}
	_, err = report.Write(json)
	if err != nil {
		return err
	}

	return report.Sync()
}

func GenerateReport(tests *[]TestResult, logger *logrus.Logger, Config *TestConfig) (float32, error) {
	reportData := prepareReportData(tests, Config)
	OutputSummaryConsole(&reportData, logger)

	if strings.Contains(Config.ReportType, "html") {
		err := OutputReportHTML(fmt.Sprintf("%v.html", Config.ReportName), &reportData, Config)
		if err != nil {
			logger.Errorf("Failed to write HTML report to %v.html: %s", Config.ReportName, err)
		}
	}

	if strings.Contains(Config.ReportType, "json") {
		err := OutputReportJSON(fmt.Sprintf("%v.json", Config.ReportName), &reportData)
		if err != nil {
			logger.Errorf("Failed to write JSON report to %v.json: %s", Config.ReportName, err)
		}
	}

	status := float32(reportData.Summary.Total.Pass) / float32(reportData.Summary.Total.Skip+reportData.Summary.Total.Fail+reportData.Summary.Total.Pass)

	return status, nil
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

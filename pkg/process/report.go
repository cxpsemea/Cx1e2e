package process

import (
	"fmt"
	"os"
	"time"

	"github.com/cxpsemea/cx1e2e/pkg/types"
	"github.com/sirupsen/logrus"
)

func GenerateReport(tests *[]TestResult, logger *logrus.Logger, Config *TestConfig) (float32, error) {
	count_failed := 0
	count_passed := 0
	count_skipped := 0

	logger.Infof("Test result summary:\n")

	var Access, Application, Flag, Group, Import, Preset, Project, Query, Result, Report, Role, Scan, User CounterSet
	for _, r := range *tests {
		var set *CounterSet
		switch r.Module {
		case types.MOD_ACCESS:
			set = &Access
		case types.MOD_APPLICATION:
			set = &Application
		case types.MOD_FLAG:
			set = &Flag
		case types.MOD_GROUP:
			set = &Group
		case types.MOD_IMPORT:
			set = &Import
		case types.MOD_PRESET:
			set = &Preset
		case types.MOD_PROJECT:
			set = &Project
		case types.MOD_QUERY:
			set = &Query
		case types.MOD_RESULT:
			set = &Result
		case types.MOD_REPORT:
			set = &Report
		case types.MOD_ROLE:
			set = &Role
		case types.MOD_SCAN:
			set = &Scan
		case types.MOD_USER:
			set = &User
		}

		var count *Counter

		switch r.CRUD {
		case types.OP_CREATE:
			count = &(set.Create)
		case types.OP_READ:
			count = &(set.Read)
		case types.OP_UPDATE:
			count = &(set.Update)
		case types.OP_DELETE:
			count = &(set.Delete)
		}

		switch r.Result {
		case TST_PASS:
			count.Pass++
		case TST_FAIL:
			count.Fail++
		case TST_SKIP:
			count.Skip++
		}

		var testtype = "Test"
		if r.FailTest {
			testtype = "Negative-Test"
		}
		switch r.Result {
		case 1:
			fmt.Printf("PASS %v - %v %v %v: %v\n", r.Name, r.CRUD, r.Module, testtype, r.TestObject)
			count_passed++
		case 0:
			fmt.Printf("FAIL %v - %v %v %v: %v\n", r.Name, r.CRUD, r.Module, testtype, r.TestObject)
			count_failed++
		case 2:
			fmt.Printf("SKIP %v - %v %v %v: %v\n", r.Name, r.CRUD, r.Module, testtype, r.TestObject)
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
	status := float32(count_passed) / float32(count_failed+count_passed+count_skipped)

	report, err := os.Create("cx1e2e_result.html")
	if err != nil {
		return status, err
	}

	defer report.Close()
	_, err = report.WriteString(fmt.Sprintf("<html><head><title>%v tenant %v test - %v</title></head><body>", Config.Cx1URL, Config.Tenant, time.Now().String()))
	if err != nil {
		return status, err
	}

	report.WriteString("<h2>Settings</h2>")
	report.WriteString(fmt.Sprintf("Running end to end tests against %v tenant %v<br>", Config.Cx1URL, Config.Tenant))
	report.WriteString(fmt.Sprintf("Authenticated using %v<br>", Config.AuthType))
	report.WriteString(fmt.Sprintf("Test set defined in configuration %v<br>", Config.ConfigPath))
	report.WriteString(fmt.Sprintf("Execution timestamp: %v.<br>", time.Now().String()))
	if os.Getenv("E2E_RUN_SUFFIX") == "" {
		report.WriteString(fmt.Sprintf("Default object name suffix %%E2E_RUN_SUFFIX%% environment variable is blank. Objects created by cx1e2e will use default names.<br>"))
	} else {
		report.WriteString(fmt.Sprintf("Default object name suffix %%E2E_RUN_SUFFIX%% environment variable is set to %v. Objects created by cx1e2e will use this suffix in the name.<br>", os.Getenv("E2E_RUN_SUFFIX")))
	}

	report.WriteString("<h2>Summary</h2>")

	report.WriteString(fmt.Sprintf("<p>Test status:<br>FAIL: %d<br>SKIP: %d<br>PASS:%d<br></p>", count_failed, count_skipped, count_passed))

	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th rowspan=2>Area</th><th colspan=3>Create</th><th colspan=3>Read</th><th colspan=3>Update</th><th colspan=3>Delete</th></tr>\n")
	report.WriteString("<tr><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th><th>Pass</th><th>Fail</th><th>Skip</th></tr>\n")
	writeCounterSet(report, "Access Assignment", &Access)
	writeCounterSet(report, "Application", &Application)
	writeCounterSet(report, "Flag", &Flag)
	writeCounterSet(report, "Group", &Group)
	writeCounterSet(report, "Import", &Import)
	writeCounterSet(report, "Preset", &Preset)
	writeCounterSet(report, "Project", &Project)
	writeCounterSet(report, "Query", &Query)
	writeCounterSet(report, "Result", &Result)
	writeCounterSet(report, "Report", &Report)
	writeCounterSet(report, "Role", &Role)
	writeCounterSet(report, "Scan", &Scan)
	writeCounterSet(report, "User", &User)
	report.WriteString("</table><br>")

	report.WriteString("<h2>Details</h2>")
	report.WriteString("<table border=1 style='border:1px solid black' cellpadding=2 cellspacing=0><tr><th>Test Set</th><th>Test</th><th>Duration (sec)</th><th>Result</th></tr>\n")

	for _, t := range *tests {
		result := "<span style='color:green'>PASS</span>"
		if t.Result == TST_FAIL {
			result = fmt.Sprintf("<span style='color:red'>FAIL: %v</span>", t.Reason)
		} else if t.Result == TST_SKIP {
			result = fmt.Sprintf("<span style='color:red'>SKIP: %v</span>", t.Reason)
		}
		testtype := "Test"
		if t.FailTest {
			testtype = "Negative-Test"
		}
		report.WriteString(fmt.Sprintf("<tr><td>%v<br>(%v)</td><td>%v %v %v: %v</td><td>%.2f</td><td>%v</td></tr>\n", t.Name, t.TestSource, t.CRUD, t.Module, testtype, t.TestObject, t.Duration, result))
	}

	report.WriteString("</table>\n")

	_, err = report.WriteString("</body></html>")
	if err != nil {
		return status, err
	}

	err = report.Sync()
	if err != nil {
		return status, err
	}

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

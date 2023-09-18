package process

import (
	"fmt"
	"os"
	"time"

	"github.com/cxpsemea/cx1e2e/pkg/types"
)

func GenerateReport(tests *[]TestResult, Config *TestConfig) error {
	report, err := os.Create("cx1e2e_result.html")
	if err != nil {
		return err
	}

	defer report.Close()

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
	}

	_, err = report.WriteString(fmt.Sprintf("<html><head><title>%v tenant %v test - %v</title></head><body>", Config.Cx1URL, Config.Tenant, time.Now().String()))
	if err != nil {
		return err
	}
	report.WriteString("<h2>Summary</h2>")
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
		return err
	}

	return report.Sync()
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

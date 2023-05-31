package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t ReportCRUD) IsValid() bool {
	if t.ProjectName == "" || t.Number == 0 || t.Format == "" {
		return false
	}

	return true
}

func ReportTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Reports *[]ReportCRUD) bool {
	result := true
	for id := range *Reports {
		t := &(*Reports)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing project name, scan number, or report format)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				err := ReportTestCreate(cx1client, logger, testname, &(*Reports)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_CREATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func ReportTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ReportCRUD) error {
	project, err := cx1client.GetProjectByName(t.ProjectName)
	if err != nil {
		return err
	}

	var filter Cx1ClientGo.ScanFilter
	if t.Branch != "" {
		filter.Branches = []string{t.Branch}
	}
	if t.Status != "" {
		filter.Statuses = []string{t.Status}
	}
	filter.Limit = int(t.Number)

	scans, err := cx1client.GetLastScansByIDFiltered(project.ProjectID, filter)
	if err != nil {
		return err
	}

	for id := range scans {
		if uint(id+1) == t.Number {
			t.Scan = &scans[id]
			break
		}
	}

	if t.Scan == nil {
		return fmt.Errorf("specified scan not found")
	}

	reportID, err := cx1client.RequestNewReportByID(t.Scan.ScanID, project.ProjectID, t.Branch, t.Format)
	if err != nil {
		return err
	}

	reportURL, err := cx1client.ReportPollingByID(reportID)
	if err != nil {
		return err
	}

	_, err = cx1client.DownloadReport(reportURL)
	if err != nil {
		return err
	}

	return nil
}

func ReportTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Reports *[]ReportCRUD) bool {
	result := true
	for id := range *Reports {
		t := &(*Reports)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_READ, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing project)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				err := ReportTestRead(cx1client, logger, testname, &(*Reports)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_READ, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func ReportTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ReportCRUD) error {
	return fmt.Errorf("not supported")
}

func ReportTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Reports *[]ReportCRUD) bool {
	result := true
	for id := range *Reports {
		t := &(*Reports)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Scan == nil {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				err := ReportTestUpdate(cx1client, logger, testname, &(*Reports)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func ReportTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ReportCRUD) error {
	return fmt.Errorf("not supported")
}

func ReportTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Reports *[]ReportCRUD) bool {
	result := true
	for id := range *Reports {
		t := &(*Reports)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Scan == nil {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before deleting)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				err := ReportTestDelete(cx1client, logger, testname, &(*Reports)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_DELETE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_REPORT, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func ReportTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ReportCRUD) error {
	return fmt.Errorf("not supported")
}

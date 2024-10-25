package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ReportCRUD) Validate(CRUD string) error {
	if CRUD != OP_CREATE {
		return fmt.Errorf("test type is not supported")
	}

	if t.ProjectName == "" {
		return fmt.Errorf("project name is missing")
	}
	if t.Number == 0 {
		return fmt.Errorf("scan number is missing (starting from 1)")
	}
	if t.Format == "" {
		return fmt.Errorf("report type is missing")
	}

	return nil
}

func (t *ReportCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	if CRUD != OP_CREATE {
		return fmt.Errorf("can only create a report")
	}
	return nil
}

func (t *ReportCRUD) GetModule() string {
	return MOD_REPORT
}

func (t *ReportCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
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
	filter.Limit = uint64(t.Number)

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

func (t *ReportCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *ReportCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *ReportCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

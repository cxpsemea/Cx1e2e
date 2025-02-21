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

	if t.ReportType != "scan" && t.ReportType != "project" {
		return fmt.Errorf("report type must be 'scan' or 'project'")
	}

	if len(t.Scanners) == 0 {
		return fmt.Errorf("report scanners must have more than one scanner, eg: sast, sca, kics")
	}

	if t.ReportType == "scan" {
		if t.ReportVersion < 1 || t.ReportVersion > 2 {
			return fmt.Errorf("scan report version can only be version 1 or 2")
		}
		if len(t.ProjectNames) != 1 || t.ProjectNames[0] == "" {
			return fmt.Errorf("single project name is missing")
		}
		if t.Number == 0 {
			return fmt.Errorf("scan number is missing (starting from 1)")
		}
	} else {
		if t.ReportVersion != 2 {
			return fmt.Errorf("project report version can only be version 2")
		}
		if len(t.ProjectNames) == 0 {
			return fmt.Errorf("list of project names is missing")
		}
	}

	if t.Format == "" {
		return fmt.Errorf("report format is missing")
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

func (t *ReportCRUD) createScanReport(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) (string, error) {
	project, err := cx1client.GetProjectByName(t.ProjectNames[0])
	if err != nil {
		return "", err
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
		return "", err
	}

	for id := range scans {
		if uint(id+1) == t.Number {
			t.Scan = &scans[id]
			break
		}
	}

	if t.Scan == nil {
		return "", fmt.Errorf("specified scan not found")
	}

	if t.ReportVersion == 1 {
		if version, err := cx1client.GetVersion(); err == nil && version.CheckCxOne("3.20.0") >= 0 && version.CheckCxOne("3.21.0") == -1 {
			// version is somewhere in 3.20.x - regular PDF report-gen is broken
			logger.Debugf("Switching from report v1 to report v2 due to Cx1 version %v", version.CxOne)
			t.ReportVersion = 2
		}
	}

	if t.ReportVersion == 2 {
		return cx1client.RequestNewReportByScanIDv2(t.Scan.ScanID, t.Scanners, []string{}, []string{}, t.Format) // todo: generate an all-engine report?
	} else {
		logger.Infof("Using v1 report-gen API")
		return cx1client.RequestNewReportByID(t.Scan.ScanID, project.ProjectID, t.Branch, t.Format, t.Scanners, []string{"ScanSummary", "ExecutiveSummary", "ScanResults"})
	}
}

func (t *ReportCRUD) createProjectReport(cx1client *Cx1ClientGo.Cx1Client) (string, error) {
	var projectIDs []string

	for _, pname := range t.ProjectNames {
		project, err := cx1client.GetProjectByName(pname)
		if err != nil {
			return "", err
		}
		projectIDs = append(projectIDs, project.ProjectID)
	}

	return cx1client.RequestNewReportByProjectIDv2(projectIDs, t.Scanners, []string{}, []string{}, t.Format)
}

func (t *ReportCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	var reportID string
	var err error

	if t.ReportType == "scan" {
		reportID, err = t.createScanReport(cx1client, logger)
	} else {
		reportID, err = t.createProjectReport(cx1client)
	}

	if err != nil {
		return err
	}

	var reportURL string
	if t.Timeout > 0 {
		reportURL, err = cx1client.ReportPollingByIDWithTimeout(reportID, cx1client.GetClientVars().ReportPollingDelaySeconds, t.Timeout)
	} else {
		reportURL, err = cx1client.ReportPollingByID(reportID)
	}

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

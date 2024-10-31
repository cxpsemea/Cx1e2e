package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ScanCRUD) Validate(CRUD string) error {
	if t.Project == "" {
		return fmt.Errorf("project name is missing")
	}
	if CRUD == OP_CREATE && ((t.Repository == "" && t.Branch == "") && t.ZipFile == "") {
		return fmt.Errorf("project repository and branch or zip file is missing")
	}

	return nil
}

func (t *ScanCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	if CRUD == OP_UPDATE {
		return fmt.Errorf("updating a scan is not supported")
	}

	return nil
}

func (t *ScanCRUD) GetModule() string {
	return MOD_SCAN
}

func (t *ScanCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}

	scanConfigs := []Cx1ClientGo.ScanConfiguration{}

	scanDelay := cx1client.GetClientVars().ScanPollingDelaySeconds

	requested_engines := strings.Split(t.Engine, " ")
	engines := make([]string, 0)

	for _, e := range requested_engines {
		if !cx1client.IsEngineAllowed(e) && !t.IsForced() {
			logger.Warnf("Requested to run a scan with engine %v but this is not supported in the license and will be skipped", e)
		} else if !Engines.IsEnabled(e) && !t.IsForced() {
			logger.Warnf("Requested to run a scan with engine %v but this was disabled for this test execution", e)
		} else {
			engines = append(engines, e)
			scanConfig := Cx1ClientGo.ScanConfiguration{}
			scanConfig.ScanType = e
			if e == "sast" {
				scanConfig.Values = map[string]string{"incremental": strconv.FormatBool(t.Incremental), "presetName": t.Preset}
			}
			scanConfigs = append(scanConfigs, scanConfig)
		}
	}

	var test_Scan Cx1ClientGo.Scan

	if t.ZipFile == "" {
		test_Scan, err = cx1client.ScanProjectGitByID(project.ProjectID, t.Repository, t.Branch, scanConfigs, map[string]string{})
		if err != nil {
			return err
		}
	} else {
		uploadURL, err := cx1client.GetUploadURL()
		if err != nil {
			return err
		}

		_, err = cx1client.PutFile(uploadURL, t.ZipFile)
		if err != nil {
			return err
		}

		test_Scan, err = cx1client.ScanProjectZipByID(project.ProjectID, uploadURL, t.Branch, scanConfigs, map[string]string{})
		if err != nil {
			return err
		}
	}

	t.Scan = &test_Scan
	if t.WaitForEnd {
		test_Scan, err = cx1client.ScanPollingWithTimeout(&test_Scan, true, scanDelay, t.Timeout)
		if err != nil {
			if err.Error()[:12] == "scan polling" && t.Cancel {
				logger.Infof("Scan %v took too long and will be canceled", test_Scan.String())
				err = cx1client.CancelScanByID(test_Scan.ScanID)
				if err != nil {
					return err
				}
				test_Scan, err = cx1client.ScanPollingWithTimeout(&test_Scan, true, 30, 300) // allow up to 5 minutes to cancel.
				if err == nil {
					return fmt.Errorf("scan took too long and was canceled")
				} else {
					return fmt.Errorf("scan took too long and the attempt to cancel failed with error: %s", err)
				}
			} else {
				return err
			}
		}

		t.Scan = &test_Scan

		expectedResult := "Completed"
		if t.Status != "" {
			expectedResult = t.Status
		}

		getWorkflow := false

		missingEngines := make([]string, 0)

		if test_Scan.Status == "Canceled" {
			return fmt.Errorf("scan was canceled")
		} else if (test_Scan.Status == "Completed" || test_Scan.Status == "Failed" || test_Scan.Status == "Partial") && test_Scan.Status != expectedResult {
			// the scan was not cancelled but did not have the expected result, so get the details
			getWorkflow = true
		} else if test_Scan.Status == "Completed" && expectedResult == "Completed" {
			// expected the scan to complete, so make sure that all of the requested engines had this status.
			for _, eng := range engines {
				matched := false
				for _, status := range test_Scan.StatusDetails {
					if strings.EqualFold(status.Name, eng) {
						matched = true
						if status.Status != expectedResult {
							// user asked for this engine to ran and it ran, but the result was unexpected
							getWorkflow = true
						}
					}
				}
				if !matched {
					// user asked for this engine to run, but it did not. the workflow won't say anything about that.
					missingEngines = append(missingEngines, eng)
				}
			}
		}

		if len(missingEngines) > 0 {
			err := fmt.Errorf("scan finished with expected status %v however the following requested scan engines did not run: %v", test_Scan.Status, strings.Join(missingEngines, ", "))
			logger.Infof("Engines didn't return a status: %s", err)
			return err
		}

		if getWorkflow {
			workflow, err := cx1client.GetScanWorkflowByID(test_Scan.ScanID)
			if err != nil {
				logger.Errorf("Failed to get workflow update for scan %v: %s", test_Scan.ScanID, err)
				return fmt.Errorf("scan finished with status '%v' but %v was expected", test_Scan.Status, expectedResult)
			} else {
				if len(workflow) == 0 {
					return fmt.Errorf("scan finished with status '%v' but %v was expected, there was no workflow log available for additional details", test_Scan.Status, expectedResult)
				} else {
					workflow_index := len(workflow) - 2
					if workflow_index <= 0 {
						workflow_index = 0
					}

					return fmt.Errorf("scan finished with status '%v - %v' but %v was expected", test_Scan.Status, workflow[workflow_index].Info, expectedResult)
				}
			}
		}
	}

	return nil
}

func (t *ScanCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}

	var scans []Cx1ClientGo.Scan

	if t.Filter != nil {
		fmt.Println("filter is: ", t.Filter)
		if t.Filter.Index == 0 {
			t.Filter.Index = 1 //
		}
		t.Cx1ScanFilter = &(Cx1ClientGo.ScanFilter{
			BaseFilter: Cx1ClientGo.BaseFilter{
				Offset: 0,
				Limit:  uint64(t.Filter.Index),
			},
			Statuses: t.Filter.Statuses,
			Branches: t.Filter.Branches,
		})

		scans, err = cx1client.GetLastScansByIDFiltered(project.ProjectID, *t.Cx1ScanFilter)
		if err != nil {
			return err
		}

		if len(scans) != t.Filter.Index {
			return fmt.Errorf("requested %d scans matching filter %v but only received %d", t.Filter.Index, t.Filter.String(), len(scans))
		}

		t.Scan = &scans[len(scans)-1]
	} else {
		scans, err = cx1client.GetLastScansByID(project.ProjectID, 1)
		if err != nil {
			return err
		}
		if len(scans) == 0 {
			return fmt.Errorf("no scan found")
		}

		t.Scan = &scans[0]
	}
	return nil
}

func (t *ScanCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not implemented")
}

func (t *ScanCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Scan == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	return cx1client.DeleteScanByID(t.Scan.ScanID)
}

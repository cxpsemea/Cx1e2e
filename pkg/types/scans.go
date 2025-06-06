package types

import (
	"fmt"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *ScanCRUD) Validate(CRUD string) error {
	if t.Project == "" {
		return fmt.Errorf("project name is missing")
	}
	if CRUD == OP_CREATE && ((t.Repository == "" && t.Branch == "") && t.ZipFile == "") {
		return fmt.Errorf("project repository and branch or zip file is missing")
	}
	if t.Engine == "" {
		return fmt.Errorf("engine is missing")
	}

	return nil
}

func (t *ScanCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	if CRUD == OP_UPDATE {
		return fmt.Errorf("updating a scan is not supported")
	}

	return nil
}

func (t *ScanCRUD) GetModule() string {
	return MOD_SCAN
}

func (t *ScanCRUD) GetLogs(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) error {
	loggedEngines := []string{"sast", "kics"}

	if t.Logs {
		if t.Scan == nil {
			return fmt.Errorf("unable to generate logs: no scan found")
		}

		for _, eng := range loggedEngines {
			for _, run_engines := range t.Scan.StatusDetails {
				if strings.EqualFold(run_engines.Name, eng) {

					bytes, err := cx1client.GetScanLogsByID(t.Scan.ScanID, eng)
					if err != nil {
						return fmt.Errorf("failed to get %v scan logs: %s", eng, err)
					}
					if len(bytes) == 0 {
						return fmt.Errorf("%v scan logs had no data", eng)
					}
				}
			}
		}
	}

	return nil
}

func (t *ScanCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}

	scanConfigSet := Cx1ClientGo.ScanConfigurationSet{}

	//scanConfigs := []Cx1ClientGo.ScanConfiguration{}

	scanDelay := cx1client.GetClientVars().ScanPollingDelaySeconds

	requested_engines := strings.Split(t.Engine, " ")
	engines := make([]string, 0)

	for _, e := range requested_engines {
		if e == "iac" {
			e = "kics"
		}

		if _, ok := cx1client.IsEngineAllowed(e); !ok && !t.IsForced() {
			logger.Warnf("Requested to run a scan with engine %v but this is not supported in the license and will be skipped", e)
		} else if !Engines.IsEnabled(e) && !t.IsForced() {
			logger.Warnf("Requested to run a scan with engine %v but this was disabled for this test execution", e)
		} else {
			engines = append(engines, e)
			//scanConfig := Cx1ClientGo.ScanConfiguration{}
			//scanConfig.ScanType = e
			scanConfigSet.AddConfig(e, "", "")

			if e == "sast" {
				scanConfigSet.AddConfig("sast", "incremental", "false")
				if t.SASTPreset != "" {
					scanConfigSet.AddConfig("sast", "presetName", t.SASTPreset)
				}
				scanConfigSet.AddConfig("sast", "fastScanMode", "false")
				scanConfigSet.AddConfig("sast", "lightQueries", "false")
				//scanConfig.Values = map[string]string{"incremental": strconv.FormatBool(t.Incremental), "presetName": t.Preset}
			} else if e == "kics" {
				if t.IACPreset != "" {
					preset, err := cx1client.GetIACPresetByName(t.IACPreset)
					if err != nil {
						return err
					}
					scanConfigSet.AddConfig("kics", "presetId", preset.PresetID)
				}
			}
			//scanConfigs = append(scanConfigs, scanConfig)
		}
	}

	var test_Scan Cx1ClientGo.Scan

	if t.ZipFile == "" {
		test_Scan, err = cx1client.ScanProjectGitByID(project.ProjectID, t.Repository, t.Branch, scanConfigSet.Configurations, map[string]string{})
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

		test_Scan, err = cx1client.ScanProjectZipByID(project.ProjectID, uploadURL, t.Branch, scanConfigSet.Configurations, map[string]string{})
		if err != nil {
			return err
		}
	}

	t.Scan = &test_Scan
	if t.WaitForEnd {
		test_Scan, err = cx1client.ScanPollingWithTimeout(&test_Scan, true, scanDelay, t.Timeout)
		if err != nil {
			if err.Error()[:4] == "scan" && err.Error()[12:19] == "polling" && t.Cancel {
				logger.Infof("Scan %v took too long and will be canceled", test_Scan.String())
				err = cx1client.CancelScanByID(test_Scan.ScanID)
				if err != nil {
					return err
				}
				test_Scan, err = cx1client.ScanPollingWithTimeout(&test_Scan, true, 30, 300) // allow up to 5 minutes to cancel.
				t.Scan = &test_Scan
				if err == nil {
					return fmt.Errorf("scan took too long and was canceled - status: %v", test_Scan.Status)
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
			var fail_reason string = "Failed"
			{
				var fail_reasons []string
				for _, status := range test_Scan.StatusDetails {
					if status.Status != expectedResult {
						fail_reasons = append(fail_reasons, fmt.Sprintf("%v: %v", status.Name, status.Details))
					}
				}
				fail_reason = strings.Join(fail_reasons, ", ")
			}

			workflow, err := cx1client.GetScanWorkflowByID(test_Scan.ScanID)
			if err != nil {
				logger.Errorf("Failed to get workflow update for scan %v: %s", test_Scan.ScanID, err)
				return fmt.Errorf("scan finished with status '%v' (%v) but %v was expected", test_Scan.Status, fail_reason, expectedResult)
			} else {
				if len(workflow) == 0 {
					return fmt.Errorf("scan finished with status '%v' (%v) but %v was expected, there was no workflow log available for additional details", test_Scan.Status, fail_reason, expectedResult)
				} else {
					workflow_index := len(workflow) - 2
					if workflow_index <= 0 {
						workflow_index = 0
					}

					logger.Debugf("Full workflow: ")
					for id := range workflow {
						logger.Debugf("%d: %v", id, workflow[id].Info)
					}

					return fmt.Errorf("scan finished with status '%v' (%v) but %v was expected", test_Scan.Status, fail_reason, expectedResult)
				}
			}
		}
	}

	return t.GetLogs(cx1client, logger)
}

func (t *ScanCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}

	var scans []Cx1ClientGo.Scan

	if t.Filter != nil {
		if t.Filter.Index == 0 {
			t.Filter.Index = 1 //
		}
		t.Cx1ScanFilter = &(Cx1ClientGo.ScanFilter{
			Statuses:  t.Filter.Statuses,
			Branches:  t.Filter.Branches,
			ProjectID: project.ProjectID,
		})

		engine := t.Engine
		if engine == "iac" {
			engine = "kics"
		}
		scans, err := cx1client.GetLastScansByEngineFiltered(engine, uint64(t.Filter.Index), *t.Cx1ScanFilter)

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

	if t.Summary {
		summary, err := cx1client.GetScanSummariesByID([]string{t.Scan.ScanID})
		if err != nil {
			return fmt.Errorf("failed to get scan summary: %s", err)
		}
		if len(summary) == 0 {
			return fmt.Errorf("scan summary had no data")
		}
	}

	if t.AggregateSummary {
		summary, err := cx1client.GetScanSASTAggregateSummaryByID(t.Scan.ScanID)
		if err != nil {
			return fmt.Errorf("failed to get scan aggregate summary: %s", err)
		}
		if len(summary) == 0 {
			return fmt.Errorf("scan aggregate summary had no data")
		}
	}

	return t.GetLogs(cx1client, logger)
}

func (t *ScanCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not implemented")
}

func (t *ScanCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
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

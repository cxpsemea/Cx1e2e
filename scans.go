package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ScanCRUD) IsValid() bool {
	if t.Project == "" {
		return false
	}

	if (t.Repository == "" || t.Branch == "") && t.ZipFile == "" {
		return false
	}

	return true
}

func ScanTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, scans *[]ScanCRUD) bool {
	result := true
	for id := range *scans {
		t := &(*scans)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(logger, "Create Scan", start, testname, id+1, "invalid test (missing project name, repository, branch, or zipfile)")
			} else {
				err := ScanTestCreate(cx1client, logger, testname, &(*scans)[id])
				if err != nil {
					result = false
					LogFail(logger, "Create Scan", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Create Scan", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ScanTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ScanCRUD) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}
	scanConfig := Cx1ClientGo.ScanConfiguration{}
	scanConfig.ScanType = t.Engine
	scanConfig.Values = map[string]string{"incremental": strconv.FormatBool(t.Incremental), "presetName": t.Preset}

	var test_Scan Cx1ClientGo.Scan

	if t.ZipFile == "" {
		test_Scan, err = cx1client.ScanProjectGitByID(project.ProjectID, t.Repository, t.Branch, []Cx1ClientGo.ScanConfiguration{scanConfig}, map[string]string{})
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

		test_Scan, err = cx1client.ScanProjectZipByID(project.ProjectID, uploadURL, t.Branch, []Cx1ClientGo.ScanConfiguration{scanConfig}, map[string]string{})
		if err != nil {
			return err
		}
	}

	t.Scan = &test_Scan
	if t.WaitForEnd {
		test_Scan, err = cx1client.ScanPolling(&test_Scan)
		if err != nil {
			return err
		}
	}
	t.Scan = &test_Scan

	return nil
}

func ScanTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, scans *[]ScanCRUD) bool {
	result := true
	for id := range *scans {
		t := &(*scans)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Project == "" {
				LogSkip(logger, "Read Scan", start, testname, id+1, "invalid test (missing project)")
			} else {
				err := ScanTestRead(cx1client, logger, testname, &(*scans)[id])
				if err != nil {
					result = false
					LogFail(logger, "Read Scan", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Read Scan", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ScanTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ScanCRUD) error {
	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}
	scans, err := cx1client.GetLastScansByID(project.ProjectID, 1)
	if err != nil {
		return err
	}

	if len(scans) == 0 {
		return fmt.Errorf("no scan found")
	}

	t.Scan = &scans[0]
	return nil
}

func ScanTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, scans *[]ScanCRUD) bool {
	result := true
	for id := range *scans {
		t := &(*scans)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Scan == nil {
				LogSkip(logger, "Update Scan", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := ScanTestUpdate(cx1client, logger, testname, &(*scans)[id])
				if err != nil {
					result = false
					LogFail(logger, "Update Scan", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Update Scan", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ScanTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ScanCRUD) error {
	return fmt.Errorf("not implemented")
}

func ScanTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, scans *[]ScanCRUD) bool {
	result := true
	for id := range *scans {
		t := &(*scans)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Scan == nil {
				LogSkip(logger, "Delete Scan", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := ScanTestDelete(cx1client, logger, testname, &(*scans)[id])
				if err != nil {
					result = false
					LogFail(logger, "Delete Scan", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Delete Scan", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ScanTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ScanCRUD) error {
	return fmt.Errorf("not implemented")
}

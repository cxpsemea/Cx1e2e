package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func ResultTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {

	for id := range *results {
		t := &(*results)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(logger, "Create Result", start, testname, id+1, "action not supported")
		}
	}

	return true
}

func ResultTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {
	result := true
	for id := range *results {
		t := &(*results)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.ProjectName == "" || (t.SimilarityID == 0 && t.ResultHash == "" && t.Number == 0) {
				LogSkip(logger, "Read Result", start, testname, id+1, "invalid test (missing project and finding identifier - similarityId, resultHash, or finding number with optional query identifier)")
			} else {
				err := ResultTestRead(cx1client, logger, testname, &(*results)[id])
				if err != nil {
					result = false
					LogFail(logger, "Read Result", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Read Result", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ResultTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ResultCRUD) error {
	project, err := cx1client.GetProjectByName(t.ProjectName)
	if err != nil {
		return err
	}
	t.Project = &project

	last_scans, err := cx1client.GetLastScansByID(project.ProjectID, 1)
	if err != nil {
		return err
	}
	if len(last_scans) == 0 {
		return fmt.Errorf("no scans run")
	}
	last_scan := last_scans[0]

	results_count, err := cx1client.GetScanResultsCountByID(last_scan.ScanID)
	if err != nil {
		return err
	}

	results, err := cx1client.GetScanResultsByID(last_scan.ScanID, results_count)
	if err != nil {
		return err
	}

	if t.QueryName != "" {
		var counter uint64
		for _, r := range results {
			if r.Data.QueryName == t.QueryName && r.Data.LanguageName == t.QueryLanguage {
				counter++
				if counter == t.Number {
					t.Result = &r
					return nil
				}
			}
		}

		return fmt.Errorf("specified result not found")
	}

	if t.QueryID != 0 {
		var counter uint64
		for _, r := range results {
			logger.Infof("  %d vs %d = %v", t.QueryID, r.Data.QueryID, t.QueryID == r.Data.QueryID)
			if r.Data.QueryID == t.QueryID {
				counter++
				if counter == t.Number {
					t.Result = &r
					return nil
				}
			}
		}
		return fmt.Errorf("specified result not found")
	}

	if t.SimilarityID != 0 {
		for _, r := range results {
			if r.SimilarityID == t.SimilarityID {
				t.Result = &r
				return nil
			}
		}
		return fmt.Errorf("specified result not found")
	}

	if t.ResultHash != "" {
		for _, r := range results {
			if r.Data.ResultHash == t.ResultHash {
				t.Result = &r
				return nil
			}
		}
		return fmt.Errorf("specified result not found")
	}

	var id uint64
	for ; id < results_count; id++ {
		if id+1 == t.Number {
			result := results[id]
			t.Result = &result
			return nil
		}
	}

	return fmt.Errorf("specified result not found")
}

func ResultTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {
	result := true
	for id := range *results {
		t := &(*results)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Result == nil {
				LogSkip(logger, "Update Result", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := ResultTestUpdate(cx1client, logger, testname, t)
				if err != nil {
					result = false
					LogFail(logger, "Update Result", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Update Result", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ResultTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ResultCRUD) error {
	var change Cx1ClientGo.ResultsPredicates
	change.ProjectID = t.Project.ProjectID
	change.SimilarityID = t.Result.SimilarityID
	change.State = t.State
	change.Severity = t.Severity
	change.Comment = t.Comment

	err := cx1client.AddResultsPredicates([]Cx1ClientGo.ResultsPredicates{change})
	return err
}

func ResultTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {
	for id := range *results {
		t := &(*results)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(logger, "Delete Result", start, testname, id+1, "invalid test (action not supported)")
		}
	}
	return true
}

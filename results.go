package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func ResultTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {

	for id := range *results {
		t := &(*results)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_CREATE, MOD_RESULT, start, testname, id+1, t.String(), "action not supported")
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
			if t.ProjectName == "" || t.Number == 0 {
				LogSkip(t.FailTest, logger, OP_READ, MOD_RESULT, start, testname, id+1, t.String(), "invalid test (missing project and finding number with optional filter)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_RESULT, start, testname, id+1, t.String())
				err := ResultTestRead(cx1client, logger, testname, &(*results)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_READ, MOD_RESULT, start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_RESULT, start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func (o ResultFilter) Matches(result Cx1ClientGo.ScanResult) bool {
	if o.QueryID != 0 && o.QueryID != result.Data.QueryID {
		return false
	}
	if o.QueryLanguage != "" && !strings.EqualFold(o.QueryLanguage, result.Data.LanguageName) {
		return false
	}
	if o.QueryGroup != "" && !strings.EqualFold(o.QueryGroup, result.Data.Group) {
		return false
	}
	if o.QueryName != "" && !strings.EqualFold(o.QueryName, result.Data.QueryName) {
		return false
	}
	if o.ResultHash != "" && o.ResultHash != result.Data.ResultHash {
		return false
	}
	if o.Severity != "" && strings.ToUpper(o.Severity) != result.Severity {
		return false
	}
	if o.State != "" && strings.ToUpper(o.State) != result.State {
		return false
	}
	if o.SimilarityID != 0 && o.SimilarityID != result.SimilarityID {
		return false
	}
	return true
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

	filtered_results := make([]Cx1ClientGo.ScanResult, 0)
	for _, r := range results {
		if t.Filter.Matches(r) {
			filtered_results = append(filtered_results, r)
		}
	}

	sort.SliceStable(filtered_results, func(i, j int) bool {
		return filtered_results[i].Data.ResultHash < filtered_results[j].Data.ResultHash
	})

	/*
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
		}*/

	var id uint64
	for id = 0; id < uint64(len(filtered_results)); id++ {
		if id+1 == t.Number {
			result := filtered_results[id]
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
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String(), "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String())
				err := ResultTestUpdate(cx1client, logger, testname, t)
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ResultTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ResultCRUD) error {
	change := t.Result.CreateResultsPredicate(t.Project.ProjectID)
	if t.State != "" {
		change.State = t.State
	}
	if t.Severity != "" {
		change.Severity = t.Severity
	}
	if t.Comment != "" {
		change.Comment = t.Comment
	}

	err := cx1client.AddResultsPredicates([]Cx1ClientGo.ResultsPredicates{change})
	return err
}

func ResultTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, results *[]ResultCRUD) bool {
	for id := range *results {
		t := &(*results)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_DELETE, MOD_RESULT, start, testname, id+1, t.String(), "invalid test (action not supported)")
		}
	}
	return true
}

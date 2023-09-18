package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ResultCRUD) Validate(CRUD string) error {
	if CRUD == OP_UPDATE && t.Result == nil {
		return fmt.Errorf("must read before updating")
	}

	if t.ProjectName == "" {
		return fmt.Errorf("project name is missing")
	}
	if t.Number == 0 {
		return fmt.Errorf("result number is missing (starting from 1)")
	}

	return nil
}

func (t *ResultCRUD) IsSupported(CRUD string) bool {
	return CRUD == OP_UPDATE || CRUD == OP_READ
}

func (t *ResultCRUD) GetModule() string {
	return MOD_RESULT
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

func (t *ResultCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	return fmt.Errorf("not implemented")
}

func (t *ResultCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *ResultCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *ResultCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	return fmt.Errorf("not implemented")
}

package types

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ResultCRUD) Validate(CRUD string) error {
	if t.Type == "" {
		return fmt.Errorf("result type not specified, should be one of: SAST, SCA, KICS")
	}
	if t.ProjectName == "" {
		return fmt.Errorf("project name is missing")
	}
	if t.Number == 0 {
		t.Number = 1
	} else {
		if CRUD != OP_READ {
			return fmt.Errorf("specifying the finding number for any operation other than Read is not supported (results are not always in consistent order)")
		}
	}

	return nil
}

func (t *ResultCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	if !cx1client.IsEngineAllowed(t.Type) {
		return fmt.Errorf("test attempts to access results from engine %v but this is not supported in the license and will be skipped", t.Type)
	}
	if !Engines.IsEnabled(t.Type) {
		return fmt.Errorf("test attempts to access results from engine %v but this was disabled for this test execution", t.Type)
	}

	if CRUD == OP_UPDATE {
		if !(t.Type == "SAST" || t.Type == "KICS") {
			return fmt.Errorf("can't update %v results", t.Type)
		}
	} else if CRUD != OP_READ {
		return fmt.Errorf("can't delete or create results")
	}

	return nil
}

func (t *ResultCRUD) GetModule() string {
	return MOD_RESULT
}

func (o SASTResultFilter) Matches(result *Cx1ClientGo.ScanSASTResult) bool {
	if o.QueryID != "" {
		u, _ := strconv.ParseUint(o.QueryID, 10, 64)
		if u != result.Data.QueryID {
			return false
		}
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
	if o.SimilarityID != "" && o.SimilarityID != result.SimilarityID {
		return false
	}
	return true
}
func (o KICSResultFilter) Matches(result *Cx1ClientGo.ScanKICSResult) bool {
	if o.QueryID != "" && o.QueryID != fmt.Sprintf("%v", result.Data.QueryID) {
		return false
	}
	if o.QueryGroup != "" && !strings.EqualFold(o.QueryGroup, result.Data.Group) {
		return false
	}
	if o.QueryName != "" && !strings.EqualFold(o.QueryName, result.Data.QueryName) {
		return false
	}
	if o.Severity != "" && strings.ToUpper(o.Severity) != result.Severity {
		return false
	}
	if o.State != "" && strings.ToUpper(o.State) != result.State {
		return false
	}
	if o.SimilarityID != "" && o.SimilarityID != result.SimilarityID {
		return false
	}
	return true
}
func (o SCAResultFilter) Matches(result *Cx1ClientGo.ScanSCAResult) bool {
	if o.Severity != "" && strings.ToUpper(o.Severity) != result.Severity {
		return false
	}
	if o.State != "" && strings.ToUpper(o.State) != result.State {
		return false
	}
	if o.SimilarityID != "" && o.SimilarityID != result.SimilarityID {
		return false
	}
	if o.CveName != "" && o.CveName != result.VulnerabilityDetails.CveName {
		return false
	}
	if o.PackageMatch != "" && !strings.Contains(strings.ToUpper(result.Data.PackageIdentifier), strings.ToUpper(o.PackageMatch)) {
		return false
	}
	return true
}

func (t *ResultCRUD) Filter(results *Cx1ClientGo.ScanResultSet) Cx1ClientGo.ScanResultSet {
	var filtered_results Cx1ClientGo.ScanResultSet
	var final_results Cx1ClientGo.ScanResultSet
	switch t.Type {
	case "SAST":
		for id := range results.SAST {
			if t.SASTFilter.Matches(&(results.SAST[id])) {
				filtered_results.SAST = append(filtered_results.SAST, results.SAST[id])
			}
		}
		sort.SliceStable(filtered_results.SAST, func(i, j int) bool {
			return filtered_results.SAST[i].Data.ResultHash < filtered_results.SAST[j].Data.ResultHash
		})

		if t.Number <= uint64(len(filtered_results.SAST)) {
			final_results.SAST = []Cx1ClientGo.ScanSASTResult{filtered_results.SAST[t.Number-1]}
		}
	case "SCA":
		for id := range results.SCA {
			if t.SCAFilter.Matches(&(results.SCA[id])) {
				filtered_results.SCA = append(filtered_results.SCA, results.SCA[id])
			}
		}
		sort.SliceStable(filtered_results.SCA, func(i, j int) bool {
			return filtered_results.SCA[i].SimilarityID < filtered_results.SCA[j].SimilarityID // TODO: Reconsider this sort of sort
		})

		if t.Number <= uint64(len(filtered_results.SCA)) {
			final_results.SCA = []Cx1ClientGo.ScanSCAResult{filtered_results.SCA[t.Number-1]}
		}
	case "KICS":
		for id := range results.KICS {
			if t.KICSFilter.Matches(&(results.KICS[id])) {
				filtered_results.KICS = append(filtered_results.KICS, results.KICS[id])
			}
		}
		sort.SliceStable(filtered_results.KICS, func(i, j int) bool {
			return filtered_results.KICS[i].SimilarityID < filtered_results.KICS[j].SimilarityID // TODO: Check if this sort of sort is sufficient
		})

		if t.Number <= uint64(len(filtered_results.KICS)) {
			final_results.KICS = []Cx1ClientGo.ScanKICSResult{filtered_results.KICS[t.Number-1]}
		}
	}

	return final_results
}

func (t *ResultCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not implemented")
}

func (t *ResultCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
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

	filteredResults := t.Filter(&results)
	t.Results = &filteredResults

	switch t.Type {
	case "SAST":
		if len(t.Results.SAST) == 0 {
			return fmt.Errorf("failed to find SAST finding matching filter %v", t.SASTFilter)
		}
	case "SCA":
		if len(t.Results.SCA) == 0 {
			return fmt.Errorf("failed to find SCA finding matching filter %v", t.SCAFilter)
		}
	case "KICS":
		if len(t.Results.KICS) == 0 {
			return fmt.Errorf("failed to find KICS finding matching filter %v", t.KICSFilter)
		}
	}

	return nil
}

func (t *ResultCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Results == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	switch t.Type {
	case "SAST":
		if len(t.Results.SAST) == 0 {
			return fmt.Errorf("specified SAST result not found")
		}
		change := t.Results.SAST[0].CreateResultsPredicate(t.Project.ProjectID)
		change.Update(t.State, t.Severity, t.Comment)
		err := cx1client.AddSASTResultsPredicates([]Cx1ClientGo.SASTResultsPredicates{change})
		return err
	case "SCA":
		return fmt.Errorf("updating SCA results is not supported")
	case "KICS":
		if len(t.Results.KICS) == 0 {
			return fmt.Errorf("specified KICS result not found")
		}
		change := t.Results.KICS[0].CreateResultsPredicate(t.Project.ProjectID)
		change.Update(t.State, t.Severity, t.Comment)
		err := cx1client.AddKICSResultsPredicates([]Cx1ClientGo.KICSResultsPredicates{change})
		return err
	}

	return fmt.Errorf("unknown type: %v", t.Type)
}

func (t *ResultCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not possible to delete results")
}

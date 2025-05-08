package types

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *ResultCRUD) Validate(CRUD string) error {
	if t.Type == "" {
		return fmt.Errorf("result type not specified, should be one of: SAST, SCA, IAC")
	}
	if t.ProjectName == "" {
		return fmt.Errorf("project name is missing")
	}
	if CRUD != OP_READ && t.Number != 1 {
		return fmt.Errorf("specifying the finding number for any operation other than Read is not supported (results are not always in consistent order)")
	}
	if t.Number == 0 {
		t.Number = 1
	}
	t.Type = strings.ToLower(t.Type)

	return nil
}

func (t *ResultCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	//t.Type = strings.ToLower(t.Type)
	if _, ok := cx1client.IsEngineAllowed(t.Type); !ok {
		return fmt.Errorf("test attempts to access results from engine %v but this is not supported in the license and will be skipped", t.Type)
	}
	if !Engines.IsEnabled(t.Type) {
		return fmt.Errorf("test attempts to access results from engine %v but this was disabled for this test execution", t.Type)
	}

	if CRUD == OP_UPDATE {
		if !(t.Type == "sast" || t.Type == "iac" || t.Type == "kics") {
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
func (o IACResultFilter) Matches(result *Cx1ClientGo.ScanIACResult) bool {
	if o.QueryID != "" && o.QueryID != result.Data.QueryID {
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
	switch strings.ToLower(t.Type) {
	case "sast":
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
	case "sca":
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
	case "iac":
		for id := range results.IAC {
			if t.IACFilter.Matches(&(results.IAC[id])) {
				filtered_results.IAC = append(filtered_results.IAC, results.IAC[id])
			}
		}
		sort.SliceStable(filtered_results.IAC, func(i, j int) bool {
			return filtered_results.IAC[i].SimilarityID < filtered_results.IAC[j].SimilarityID // TODO: Check if this sort of sort is sufficient
		})

		if t.Number <= uint64(len(filtered_results.IAC)) {
			final_results.IAC = []Cx1ClientGo.ScanIACResult{filtered_results.IAC[t.Number-1]}
		}
	}

	return final_results
}

func (t *ResultCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not implemented")
}

func (t *ResultCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	project, err := cx1client.GetProjectByName(t.ProjectName)
	if err != nil {
		return err
	}
	t.Project = &project

	scanFilter := Cx1ClientGo.ScanFilter{
		Statuses:  []string{"Completed"},
		ProjectID: project.ProjectID,
	}

	engine := t.Type
	if engine == "iac" {
		engine = "kics"
	}
	last_scans, err := cx1client.GetLastScansByEngineFiltered(engine, 1, scanFilter)
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
	t.Scan = &last_scan

	switch strings.ToLower(t.Type) {
	case "sast":
		if len(t.Results.SAST) == 0 {
			return fmt.Errorf("failed to find SAST finding matching filter %v", t.SASTFilter.String())
		}
	case "sca":
		if len(t.Results.SCA) == 0 {
			return fmt.Errorf("failed to find SCA finding matching filter %v", t.SCAFilter.String())
		}
	case "iac":
		if len(t.Results.IAC) == 0 {
			return fmt.Errorf("failed to find IAC finding matching filter %v", t.IACFilter.String())
		}
	}

	return nil
}

func (t *ResultCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Results == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	switch strings.ToLower(t.Type) {
	case "sast":
		if len(t.Results.SAST) == 0 {
			return fmt.Errorf("specified SAST result not found")
		}
		change := t.Results.SAST[0].CreateResultsPredicate(t.Project.ProjectID, t.Scan.ScanID)
		change.Update(t.State, t.Severity, t.Comment)
		err := cx1client.AddSASTResultsPredicates([]Cx1ClientGo.SASTResultsPredicates{change})
		return err
	case "sca":
		return fmt.Errorf("updating SCA results is not supported")
	case "iac":
		if len(t.Results.IAC) == 0 {
			return fmt.Errorf("specified IAC result not found")
		}
		change := t.Results.IAC[0].CreateResultsPredicate(t.Project.ProjectID, t.Scan.ScanID)
		change.Update(t.State, t.Severity, t.Comment)
		err := cx1client.AddIACResultsPredicates([]Cx1ClientGo.IACResultsPredicates{change})
		return err
	}

	return fmt.Errorf("unknown type: %v", t.Type)
}

func (t *ResultCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not possible to delete results")
}

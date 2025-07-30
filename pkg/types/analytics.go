package types

import (
	"fmt"
	"slices"

	"github.com/cxpsemea/Cx1ClientGo"
)

var kpiTypes []string = []string{
	"vulnerabilitiesBySeverityTotal",
	"vulnerabilitiesByStateTotal",
	"vulnerabilitiesByStatusTotal",
	"vulnerabilitiesBySeverityAndStateTotal",
	"vulnerabilitiesBySeverityOvertime",
	"fixedVulnerabilitiesBySeverityOvertime",
	"meanTimeToResolution",
	"mostCommonVulnerabilities",
	"mostAgingVulnerabilities",
	"allVulnerabilities",
	"agingTotal",
	"ideTotal",
	"ideOvertime",
}

func (t *AnalyticsCRUD) Validate(CRUD string) error {
	if CRUD != OP_READ {
		return fmt.Errorf("test type is not supported")
	}

	if !slices.Contains(kpiTypes, t.KPI) {
		return fmt.Errorf("kpi '%v' is not supported", t.KPI)
	}

	return nil
}

func (t *AnalyticsCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	if CRUD != OP_READ {
		return fmt.Errorf("can only read from Analytics")
	}
	return nil
}

func (t *AnalyticsCRUD) GetModule() string {
	return MOD_ANALYTICS
}

func (t *AnalyticsCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *AnalyticsCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	var err error

	filter := Cx1ClientGo.AnalyticsFilter{}
	for _, p := range t.Filter.Projects {
		proj, err := cx1client.GetProjectByName(p)
		if err != nil {
			return err
		}
		filter.Projects = append(filter.Projects, proj.ProjectID)
	}

	count := uint64(0)

	switch t.KPI {
	case "vulnerabilitiesBySeverityTotal":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesBySeverityTotal(filter)
		count = results.Total
		err = err2
	case "vulnerabilitiesByStateTotal":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesByStateTotal(filter)
		count = results.Total
		err = err2
	case "vulnerabilitiesByStatusTotal":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesByStatusTotal(filter)
		count = results.Total
		err = err2
	case "vulnerabilitiesBySeverityAndStateTotal":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesBySeverityAndStateTotal(filter)
		count = uint64(len(results))
		err = err2
	case "vulnerabilitiesBySeverityOvertime":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesBySeverityOvertime(filter)
		count = uint64(len(results))
		err = err2
	case "fixedVulnerabilitiesBySeverityOvertime":
		results, err2 := cx1client.GetAnalyticsFixedVulnerabilitiesBySeverityOvertime(filter)
		count = uint64(len(results))
		err = err2
	case "meanTimeToResolution":
		results, err2 := cx1client.GetAnalyticsMeanTimeToResolution(filter)
		count = uint64(results.TotalResults)
		err = err2
	case "mostCommonVulnerabilities":
		results, err2 := cx1client.GetAnalyticsMostCommonVulnerabilities(100, filter)
		count = uint64(len(results))
		err = err2
	case "allVulnerabilities":
		results, err2 := cx1client.GetAnalyticsAllVulnerabilities(1000, 0, filter)
		count = uint64(len(results))
		err = err2
	case "mostAgingVulnerabilities":
		results, err2 := cx1client.GetAnalyticsMostAgingVulnerabilities(100, filter)
		count = uint64(len(results))
		err = err2
	case "agingTotal":
		results, err2 := cx1client.GetAnalyticsVulnerabilitiesByAgingTotal(filter)
		count = uint64(len(results))
		err = err2
	case "ideTotal":
		results, err2 := cx1client.GetAnalyticsIDETotal()
		count = uint64(len(results))
		err = err2
	case "ideOvertime":
		results, err2 := cx1client.GetAnalyticsIDEOverTimeStats()
		count = uint64(len(results))
		err = err2
	}
	if err == nil {
		logger.Infof("Got %d results", count)
	}
	return err
}

func (t *AnalyticsCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *AnalyticsCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

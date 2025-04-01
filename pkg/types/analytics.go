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

	switch t.KPI {
	case "vulnerabilitiesBySeverityTotal":
		_, err = cx1client.GetAnalyticsVulnerabilitiesBySeverityTotal(filter)
	case "vulnerabilitiesByStateTotal":
		_, err = cx1client.GetAnalyticsVulnerabilitiesByStateTotal(filter)
	case "vulnerabilitiesByStatusTotal":
		_, err = cx1client.GetAnalyticsVulnerabilitiesByStatusTotal(filter)
	case "vulnerabilitiesBySeverityAndStateTotal":
		_, err = cx1client.GetAnalyticsVulnerabilitiesByStatusTotal(filter)
	case "vulnerabilitiesBySeverityOvertime":
		_, err = cx1client.GetAnalyticsVulnerabilitiesBySeverityOvertime(filter)
	case "fixedVulnerabilitiesBySeverityOvertime":
		_, err = cx1client.GetAnalyticsFixedVulnerabilitiesBySeverityOvertime(filter)
	case "meanTimeToResolution":
		_, err = cx1client.GetAnalyticsMeanTimeToResolution(filter)
	case "mostCommonVulnerabilities":
		_, err = cx1client.GetAnalyticsMostCommonVulnerabilities(100, filter)
	case "mostAgingVulnerabilities":
		_, err = cx1client.GetAnalyticsMostAgingVulnerabilities(100, filter)
	}

	return err
}

func (t *AnalyticsCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *AnalyticsCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

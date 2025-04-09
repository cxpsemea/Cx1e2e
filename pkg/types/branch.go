package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *BranchCRUD) Validate(CRUD string) error {
	if CRUD != OP_READ {
		return fmt.Errorf("test type is not supported")
	}

	if t.Project == "" {
		return fmt.Errorf("project name must be provided")
	}

	return nil
}

func (t *BranchCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	if CRUD != OP_READ {
		return fmt.Errorf("can only read branches")
	}
	return nil
}

func (t *BranchCRUD) GetModule() string {
	return MOD_BRANCH
}

func (t *BranchCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *BranchCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	var err error
	var filter Cx1ClientGo.ProjectBranchFilter

	project, err := cx1client.GetProjectByName(t.Project)
	if err != nil {
		return err
	}

	filter.ProjectID = project.ProjectID
	if t.Branch != "" {
		filter.Name = t.Branch
	}

	branches, err := cx1client.GetAllProjectBranchesFiltered(filter)
	if err != nil {
		return err
	}

	if len(branches) < int(t.ExpectedCount) {
		return fmt.Errorf("retrieved %d branches matching '%v' in project %v, expected at most %d", len(branches), t.Branch, project.String(), t.ExpectedCount)
	}

	return nil
}

func (t *BranchCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *BranchCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

package types

import (
	"fmt"
	"slices"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *ApplicationCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("application name is missing")
	}

	return nil
}

func (t *ApplicationCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *ApplicationCRUD) GetModule() string {
	return MOD_APPLICATION
}

func updateApplication(cx1client *Cx1ClientGo.Cx1Client, _ *ThreadLogger, t *ApplicationCRUD) error {
	if len(t.Tags) > 0 {
		t.Application.Tags = make(map[string]string)
		for _, tag := range t.Tags {
			t.Application.Tags[tag.Key] = tag.Value
		}
		err := cx1client.UpdateApplication(t.Application)
		if err != nil {
			return err
		}
	}

	if len(t.Rules) > 0 {
		// remove all rules
		t.Application.Rules = make([]Cx1ClientGo.ApplicationRule, 0)
		for _, r := range t.Rules {
			t.Application.AddRule(r.Type, r.Value)
		}

		err := cx1client.UpdateApplication(t.Application)
		if err != nil {
			return err
		}
		updatedApplication, err := cx1client.GetApplicationByID(t.Application.ApplicationID)
		if err != nil {
			return err
		}
		t.Application = &updatedApplication
	}

	if t.Projects != nil {
		newProjects := []string{}
		oldProjects := []string{}
		if t.Application.ProjectIds != nil {
			oldProjects = *t.Application.ProjectIds
		}

		for _, p := range *t.Projects {
			project, err := cx1client.GetProjectByName(p)
			if err != nil {
				return err
			}
			newProjects = append(newProjects, project.ProjectID)
			if !slices.Contains(oldProjects, project.ProjectID) {
				t.Application.AssignProject(&project)
			}
		}

		for _, pid := range oldProjects {
			if !slices.Contains(newProjects, pid) {
				p, err := cx1client.GetProjectByID(pid)
				if err != nil {
					return err
				}
				t.Application.UnassignProject(&p)
			}
		}

		err := cx1client.UpdateApplication(t.Application)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ApplicationCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	/* TODO once apps can be in groups
	group_ids := []string{}

	for _, g := range t.Groups {
		group, err := cx1client.GetGroupByName(g)
		if err != nil {
			return err
		}
		group_ids = append(group_ids, group.GroupID)
	}*/
	test_Application, err := cx1client.CreateApplication(t.Name)
	if err != nil {
		return err
	}
	t.Application = &test_Application

	err = updateApplication(cx1client, logger, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *ApplicationCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	test_Application, err := cx1client.GetApplicationByName(t.Name)
	if err != nil {
		return err
	}
	t.Application = &test_Application
	return nil
}

func (t *ApplicationCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Application == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := updateApplication(cx1client, logger, t)
	if err != nil {
		return err
	}
	return nil
}

func (t *ApplicationCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Application == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeleteApplicationByID(t.Application.ApplicationID)
	if err != nil {
		return err
	}

	t.Application = nil
	return nil
}

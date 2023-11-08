package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ProjectCRUD) Validate(CRUD string) error {
	if (CRUD == OP_UPDATE || CRUD == OP_DELETE) && t.Project == nil {
		return fmt.Errorf("must read before updating or deleting")
	}

	if t.Name == "" {
		return fmt.Errorf("project name is missing")
	}

	return nil
}

func (t *ProjectCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string) bool {
	return true
}

func (t *ProjectCRUD) GetModule() string {
	return MOD_PROJECT
}

func (t *ProjectCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	group_ids := []string{}

	for _, g := range t.Groups {
		group, err := cx1client.GetGroupByName(g)
		if err != nil {
			return err
		}
		group_ids = append(group_ids, group.GroupID)
	}

	tags := make(map[string]string)
	for _, tag := range t.Tags {
		tags[tag.Key] = tag.Value
	}

	test_Project, err := cx1client.CreateProject(t.Name, group_ids, tags)
	if err != nil {
		return err
	}
	t.Project = &test_Project

	if t.Application != "" {
		app, err := cx1client.GetApplicationByName(t.Application)
		if err != nil {
			return err
		}
		app.AssignProject(t.Project)
		err = cx1client.UpdateApplication(&app)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *ProjectCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	test_Project, err := cx1client.GetProjectByName(t.Name)
	if err != nil {
		return err
	}

	t.Project = &test_Project

	if t.Application != "" {
		app, err := cx1client.GetApplicationByName(t.Application)
		if err != nil {
			return err
		}

		for _, p := range app.ProjectIds {
			if p == t.Project.ProjectID {
				return nil
			}
		}

		return fmt.Errorf("expected project %v to live under application %v but it does not", t.Name, t.Application)
	}

	return nil
}

func (t *ProjectCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	if t.Application != "" {
		app, err := cx1client.GetApplicationByName(t.Application)
		if err != nil {
			return err
		}
		app.AssignProject(t.Project)
		err = cx1client.UpdateApplication(&app)
		if err != nil {
			return err
		}
	}

	if len(t.Tags) > 0 {
		t.Project.Tags = make(map[string]string)
		for _, tag := range t.Tags {
			t.Project.Tags[tag.Key] = tag.Value
		}
		err := cx1client.UpdateProject(t.Project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *ProjectCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	err := cx1client.DeleteProject(t.Project)
	if err != nil {
		return err
	}

	t.Project = nil
	return nil
}

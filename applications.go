package main

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ApplicationCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("application name is missing")
	}

	return nil
}
func (t *ApplicationCRUD) GetModule() string {
	return MOD_APPLICATION
}

func updateApplication(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *ApplicationCRUD) error {
	t.Application.Tags = make(map[string]string)
	for _, tag := range t.Tags {
		t.Application.Tags[tag.Key] = tag.Value
	}

	// remove all rules
	t.Application.Rules = make([]Cx1ClientGo.ApplicationRule, 0)
	for _, r := range t.Rules {
		t.Application.AddRule(r.Type, r.Value)
	}

	cx1client.UpdateApplication(t.Application)

	return nil
}

func (t *ApplicationCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *ApplicationCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	test_Application, err := cx1client.GetApplicationByName(t.Name)
	if err != nil {
		return err
	}
	t.Application = &test_Application
	return nil
}

func (t *ApplicationCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	err := updateApplication(cx1client, logger, t)
	if err != nil {
		return err
	}
	return nil
}

func (t *ApplicationCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	err := cx1client.DeleteApplicationByID(t.Application.ApplicationID)
	if err != nil {
		return err
	}

	t.Application = nil
	return nil
}

package main

import (
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func ProjectTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, projects *[]ProjectCRUD) bool {
	result := true
	for id := range *projects {
		t := &(*projects)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Create Project", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := ProjectTestCreate(cx1client, logger, testname, &(*projects)[id])
				if err != nil {
					result = false
					LogFail(logger, "Create Project", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Create Project", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ProjectTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ProjectCRUD) error {
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

func ProjectTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, projects *[]ProjectCRUD) bool {
	result := true
	for id := range *projects {
		t := &(*projects)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Read Project", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := ProjectTestRead(cx1client, logger, testname, &(*projects)[id])
				if err != nil {
					result = false
					LogFail(logger, "Read Project", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Read Project", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ProjectTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ProjectCRUD) error {
	test_Project, err := cx1client.GetProjectByName(t.Name)
	if err != nil {
		return err
	}
	t.Project = &test_Project
	return nil
}

func ProjectTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, projects *[]ProjectCRUD) bool {
	result := true
	for id := range *projects {
		t := &(*projects)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Project == nil {
				LogSkip(logger, "Update Project", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := ProjectTestUpdate(cx1client, logger, testname, &(*projects)[id])
				if err != nil {
					result = false
					LogFail(logger, "Update Project", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Update Project", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ProjectTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ProjectCRUD) error {
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

func ProjectTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, projects *[]ProjectCRUD) bool {
	result := true
	for id := range *projects {
		t := &(*projects)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Project == nil {
				LogSkip(logger, "Delete Project", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := ProjectTestDelete(cx1client, logger, testname, &(*projects)[id])
				if err != nil {
					result = false
					LogFail(logger, "Delete Project", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Delete Project", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ProjectTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ProjectCRUD) error {
	err := cx1client.DeleteProject(t.Project)
	if err != nil {
		return err
	}

	t.Project = nil
	return nil
}

package main

import (
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func ApplicationTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, applications *[]ApplicationCRUD) bool {
	result := true
	for id := range *applications {
		t := &(*applications)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Create Application", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := ApplicationTestCreate(cx1client, logger, testname, &(*applications)[id])
				if err != nil {
					result = false
					LogFail(logger, "Create Application", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Create Application", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
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

func ApplicationTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ApplicationCRUD) error {
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

func ApplicationTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, applications *[]ApplicationCRUD) bool {
	result := true
	for id := range *applications {
		t := &(*applications)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Create", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := ApplicationTestRead(cx1client, logger, testname, &(*applications)[id])
				if err != nil {
					result = false
					LogFail(logger, "Read Application", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Read Application", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ApplicationTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ApplicationCRUD) error {
	test_Application, err := cx1client.GetApplicationByName(t.Name)
	if err != nil {
		return err
	}
	t.Application = &test_Application
	return nil
}

func ApplicationTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, applications *[]ApplicationCRUD) bool {
	result := true
	for id := range *applications {
		t := &(*applications)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Application == nil {
				LogSkip(logger, "Update", start, testname, id+1, "must read before updating")
			} else {
				err := ApplicationTestUpdate(cx1client, logger, testname, &(*applications)[id])
				if err != nil {
					result = false
					LogFail(logger, "Update Application", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Update Application", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ApplicationTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ApplicationCRUD) error {
	err := updateApplication(cx1client, logger, t)
	if err != nil {
		return err
	}
	return nil
}

func ApplicationTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, applications *[]ApplicationCRUD) bool {
	result := true
	for id := range *applications {
		t := &(*applications)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Application == nil {
				LogSkip(logger, "Delete Application", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := ApplicationTestDelete(cx1client, logger, testname, &(*applications)[id])
				if err != nil {
					result = false
					LogFail(logger, "Delete Application", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Delete Application", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func ApplicationTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ApplicationCRUD) error {
	err := cx1client.DeleteApplicationByID(t.Application.ApplicationID)
	if err != nil {
		return err
	}

	t.Application = nil
	return nil
}

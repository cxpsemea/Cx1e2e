package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t AccessAssignmentCRUD) IsValid() bool {
	if t.EntityType == "" || t.EntityName == "" || t.ResourceName == "" || t.ResourceType == "" || len(t.Roles) == 0 {
		return false
	}

	return true
}

func AccessTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, accessAssignments *[]AccessAssignmentCRUD) {
	for id := range *accessAssignments {
		t := &(*accessAssignments)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing entity, resource, or roles)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				err := AccessTestCreate(cx1client, logger, testname, &(*accessAssignments)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_CREATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func prepareAccessAssignment(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *AccessAssignmentCRUD) (Cx1ClientGo.AccessAssignment, error) {
	access := Cx1ClientGo.AccessAssignment{
		TenantID:     cx1client.GetTenantID(),
		EntityType:   t.EntityType,
		ResourceType: t.ResourceType,
	}

	switch t.EntityType {
	case "user":
		user, err := cx1client.GetUserByUserName(t.EntityName)
		if err != nil {
			return access, fmt.Errorf("failed to retrieve user with username %v: %s", t.EntityName, err)
		}
		access.EntityName = user.UserName
		access.EntityID = user.UserID
	case "group":
		group, err := cx1client.GetGroupByName(t.EntityName)
		if err != nil {
			return access, fmt.Errorf("failed to retrieve group named %v: %s", t.EntityName, err)
		}
		access.EntityName = group.Name
		access.EntityID = group.GroupID
	default:
		return access, fmt.Errorf("unknown entitytype %v used for access assignment, options are: user, group", t.EntityType)
	}

	switch t.ResourceType {
	case "tenant":
		access.ResourceName = cx1client.GetTenantName()
		access.ResourceID = cx1client.GetTenantID()
	case "application":
		app, err := cx1client.GetApplicationByName(t.ResourceName)
		if err != nil {
			return access, fmt.Errorf("failed to retrieve application named %v: %s", t.ResourceName, err)
		}
		access.ResourceName = app.Name
		access.ResourceID = app.ApplicationID
	case "project":
		project, err := cx1client.GetProjectByName(t.ResourceName)
		if err != nil {
			return access, fmt.Errorf("failed to retrieve application named %v: %s", t.ResourceName, err)
		}
		access.ResourceName = project.Name
		access.ResourceID = project.ProjectID
	}

	access.EntityRoles = t.Roles

	return access, nil
}

func AccessTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *AccessAssignmentCRUD) error {
	access, err := prepareAccessAssignment(cx1client, logger, t)
	if err != nil {
		return err
	}

	err = cx1client.AddAccessAssignment(access)
	if err != nil {
		return err
	}

	return nil
}

func AccessTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, accessAssignments *[]AccessAssignmentCRUD) {
	for id := range *accessAssignments {
		t := &(*accessAssignments)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_READ, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing entity, resource, or roles)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				err := AccessTestRead(cx1client, logger, testname, &(*accessAssignments)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_READ, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func AccessTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *AccessAssignmentCRUD) error {
	access, err := prepareAccessAssignment(cx1client, logger, t)
	if err != nil {
		return err
	}

	assignment, err := cx1client.GetAccessAssignmentByID(access.EntityID, access.ResourceID)
	if err != nil {
		return fmt.Errorf("no assignment matching %v", t.String())
	}

	for _, rr := range t.Roles {
		hasrole := false
		for _, ar := range assignment.EntityRoles {
			if rr == ar {
				hasrole = true
				break
			}
		}
		if !hasrole {
			return fmt.Errorf("expected %v %v to have role %v on %v %v but it's not there", t.EntityType, t.EntityName, rr, t.ResourceType, t.ResourceName)
		}
	}

	return nil
}

func AccessTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, accessAssignments *[]AccessAssignmentCRUD) {
	for id := range *accessAssignments {
		t := &(*accessAssignments)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing entity, resource, or roles)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				err := AccessTestUpdate(cx1client, logger, testname, &(*accessAssignments)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func AccessTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *AccessAssignmentCRUD) error {
	access, err := prepareAccessAssignment(cx1client, logger, t)
	if err != nil {
		return err
	}

	// TODO: update this once role-assignments are more granular
	err = cx1client.DeleteAccessAssignmentByID(access.EntityID, access.ResourceID)
	if err != nil {
		return fmt.Errorf("failed to remove existing %v %v access to %v %v", t.EntityType, t.EntityName, t.ResourceType, t.ResourceName)
	}

	err = cx1client.AddAccessAssignment(access)
	if err != nil {
		return err
	}

	return nil
}

func AccessTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, accessAssignments *[]AccessAssignmentCRUD) {
	for id := range *accessAssignments {
		t := &(*accessAssignments)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing entity, resource, or roles)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				err := AccessTestDelete(cx1client, logger, testname, &(*accessAssignments)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_DELETE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_ACCESS, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func AccessTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *AccessAssignmentCRUD) error {
	access, err := prepareAccessAssignment(cx1client, logger, t)
	if err != nil {
		return err
	}

	// TODO: update this once role-assignments are more granular
	err = cx1client.DeleteAccessAssignmentByID(access.EntityID, access.ResourceID)
	if err != nil {
		return fmt.Errorf("failed to remove existing %v %v access to %v %v", t.EntityType, t.EntityName, t.ResourceType, t.ResourceName)
	}

	return nil
}

package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func CheckAMFlag(cx1client *Cx1ClientGo.Cx1Client) bool {
	flag, err := cx1client.CheckFlag("ACCESS_MANAGEMENT_ENABLED")
	if err != nil {
		return false
	}
	return flag
}

func (t *AccessAssignmentCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string) bool {
	return true
}

func (t *AccessAssignmentCRUD) Validate(CRUD string) error {
	if t.EntityType == "" || t.EntityName == "" {
		return fmt.Errorf("entity type or name is missing")
	}

	if t.ResourceName == "" || t.ResourceType == "" {
		return fmt.Errorf("resource type or name is missing")
	}

	return nil
}

func (t *AccessAssignmentCRUD) GetModule() string {
	return MOD_ACCESS
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

	access.EntityRoles = make([]Cx1ClientGo.AccessAssignedRole, len(t.Roles) )

	for id, r := t.Roles {
		access.EntityRoles[id].Name = r	
	}

	return access, nil
}

func (t *AccessAssignmentCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *AccessAssignmentCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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
			if rr == ar.Name {
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

func (t *AccessAssignmentCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *AccessAssignmentCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

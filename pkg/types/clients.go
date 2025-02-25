package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *OIDCClientCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("client name is missing")
	}
	return nil
}

func (t *OIDCClientCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *OIDCClientCRUD) GetModule() string {
	return MOD_CLIENT
}

func updateOIDCClientFromConfig(cx1client *Cx1ClientGo.Cx1Client, t *OIDCClientCRUD) error {

	_, err := cx1client.GetUserGroups(t.User)
	if err != nil {
		return err
	}

	for _, g := range t.Groups { // groups to add
		if val, _ := t.User.IsInGroupByName(g); !val {
			group, err := cx1client.GetGroupByName(g)
			if err != nil {
				return fmt.Errorf("failed to find group %v: %s", g, err)
			}
			err = cx1client.AssignUserToGroupByID(t.User, group.GroupID)
			if err != nil {
				return fmt.Errorf("failed to assign client to group %v: %s", g, err)
			}
		}
	}

	for _, g := range t.User.Groups { // groups to remove
		matched := false
		for _, newgroup := range t.Groups {
			if g.Name == newgroup {
				matched = true
				break
			}
		}

		if !matched {
			err = cx1client.RemoveUserFromGroupByID(t.User, g.GroupID)
			if err != nil {
				return fmt.Errorf("failed to remove client from group %v: %s", g.Name, err)
			}
		}
	}

	_, err = cx1client.GetUserRoles(t.User)
	if err != nil {
		return fmt.Errorf("failed to get client's roles: %s", err)
	}

	new_roles := []Cx1ClientGo.Role{}

	fmt.Printf("Expecting to have %d roles\n", len(t.Roles))

	for _, newrole := range t.Roles { // check for roles to add
		if val, _ := t.User.HasRoleByName(newrole); !val {
			role, err := cx1client.GetRoleByName(newrole)
			if err != nil {
				return err
			}

			new_roles = append(new_roles, role)
		}
	}

	if len(new_roles) > 0 {
		err := cx1client.AddUserRoles(t.User, &new_roles)
		if err != nil {
			return fmt.Errorf("failed to grant client %v new roles: %s", t.User.String(), err)
		}
	}

	del_roles := []Cx1ClientGo.Role{}

	for _, oldrole := range t.User.Roles {
		matched := false
		for _, r := range t.Roles {
			if r == oldrole.Name {
				matched = true
			}
		}

		if !matched {
			del_roles = append(del_roles, oldrole)
		}
	}

	if len(del_roles) > 0 {
		err := cx1client.RemoveUserRoles(t.User, &del_roles)
		if err != nil {
			return fmt.Errorf("failed to remove client %v old roles: %s", t.User.String(), err)
		}
	}

	return nil
}

func (t *OIDCClientCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	client, err := cx1client.CreateClient(t.Name, []string{}, 30)
	if err != nil {
		return err
	}
	t.Client = &client

	user, err := cx1client.GetServiceAccountByID(t.Client.ID)
	if err != nil {
		return err
	}
	t.User = &user

	err = updateOIDCClientFromConfig(cx1client, t)
	if err != nil {
		return err
	}
	return nil
}

func (t *OIDCClientCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	test_OIDCClient, err := cx1client.GetClientByName(t.Name)
	if err != nil {
		return err
	}
	t.Client = &test_OIDCClient

	user, err := cx1client.GetServiceAccountByID(t.Client.ID)
	if err != nil {
		return err
	}
	t.User = &user

	return nil
}

func (t *OIDCClientCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Client == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := updateOIDCClientFromConfig(cx1client, t)
	if err != nil {
		return err
	}

	return cx1client.UpdateUser(t.User)
}

func (t *OIDCClientCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Client == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeleteClientByID(t.Client.ID)
	if err != nil {
		return err
	}

	t.Client = nil
	t.User = nil
	return nil
}

package types

import (
	"fmt"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *RoleCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("role name is missing")
	}

	return nil
}

func (t *RoleCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *RoleCRUD) GetModule() string {
	return MOD_ROLE
}

func getRole(cx1client *Cx1ClientGo.Cx1Client, _ *logrus.Logger, roleID string) (*Cx1ClientGo.Role, error) {
	role, err := cx1client.GetRoleByID(roleID)
	if err != nil {
		return nil, err
	}

	sub_roles, err := cx1client.GetRoleComposites(&role)
	if err != nil {
		return &role, err
	}

	role.SubRoles = sub_roles

	return &role, nil
}

func updateRole(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *RoleCRUD) error {
	role, err := getRole(cx1client, logger, t.Role.RoleID)
	if err != nil {
		return err
	}
	t.Role = role

	roles_to_add := make([]Cx1ClientGo.Role, 0)
	for _, r := range t.Permissions {
		if !t.Role.HasRole(r) { // should have the role, but doesn't
			role, err := cx1client.GetRoleByName(r)
			if err != nil {
				return fmt.Errorf("unable to find role %v: %s", r, err)
			}

			roles_to_add = append(roles_to_add, role)
		}
	}

	if len(roles_to_add) > 0 {
		err = cx1client.AddRoleComposites(t.Role, &roles_to_add)
		if err != nil {
			return err
		}
	}

	roles_to_remove := make([]Cx1ClientGo.Role, 0)
	for _, r := range t.Role.SubRoles {
		matched := false
		for _, p := range t.Permissions {
			if p == r.Name {
				matched = true
				break
			}
		}
		if !matched { // has the role, but shouldn't
			roles_to_remove = append(roles_to_remove, r)
		}
	}

	if len(roles_to_remove) > 0 {
		err = cx1client.RemoveRoleComposites(t.Role, &roles_to_remove)
		if err != nil {
			return err
		}
	}

	t.Role, err = getRole(cx1client, logger, t.Role.RoleID)
	if err != nil {
		return err
	}

	return nil
}

func (t *RoleCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	test_Role, err := cx1client.CreateAppRole(t.Name, "cx1e2e test")
	if err != nil {
		return err
	}
	t.Role = &test_Role
	return updateRole(cx1client, logger, t)
}

func (t *RoleCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	test_Role, err := cx1client.GetRoleByName(t.Name)
	if err != nil {
		return err
	}
	sub_roles, err := cx1client.GetRoleComposites(&test_Role)
	if err != nil {
		return err
	}
	test_Role.SubRoles = sub_roles

	if len(t.Filter) > 0 {
		match_count := 0
		missing_roles := []string{}

		for _, filter := range t.Filter {
			matched := false
			for _, sr := range test_Role.SubRoles {
				if strings.EqualFold(filter, sr.Name) {
					match_count++
					matched = true
					break
				}
			}
			if !matched {
				missing_roles = append(missing_roles, filter)
			}
		}

		if match_count != len(t.Filter) {
			return fmt.Errorf("role %v exists but is missing the following sub-roles set in the test filter: %v", test_Role.String(), strings.Join(missing_roles, ", "))
		}
	}

	t.Role = &test_Role
	return nil
}

func (t *RoleCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Role == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	return updateRole(cx1client, logger, t)
}

func (t *RoleCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Role == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	return cx1client.DeleteRoleByID(t.Role.RoleID)
}

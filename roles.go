package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func RoleTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, roles *[]RoleCRUD) bool {
	result := true
	for id := range *roles {
		t := &(*roles)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				err := RoleTestCreate(cx1client, logger, testname, &(*roles)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_CREATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func getRole(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, roleID string) (*Cx1ClientGo.Role, error) {
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

func RoleTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *RoleCRUD) error {
	test_Role, err := cx1client.CreateAppRole(t.Name, "cx1e2e test")
	if err != nil {
		return err
	}
	t.Role = &test_Role
	return updateRole(cx1client, logger, t)
}

func RoleTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, roles *[]RoleCRUD) bool {
	result := true
	for id := range *roles {
		t := &(*roles)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(t.FailTest, logger, OP_READ, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				err := RoleTestRead(cx1client, logger, testname, &(*roles)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_READ, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func RoleTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *RoleCRUD) error {
	test_Role, err := cx1client.GetRoleByName(t.Name)
	if err != nil {
		return err
	}
	t.Role = &test_Role
	return nil
}

func RoleTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, roles *[]RoleCRUD) bool {
	result := true
	for id := range *roles {
		t := &(*roles)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Role == nil {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				err := RoleTestUpdate(cx1client, logger, testname, &(*roles)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func RoleTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *RoleCRUD) error {
	return updateRole(cx1client, logger, t)
}

func RoleTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, roles *[]RoleCRUD) bool {
	result := true
	for id := range *roles {
		t := &(*roles)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Role == nil {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before deleting)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				err := RoleTestDelete(cx1client, logger, testname, &(*roles)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_DELETE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_ROLE, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func RoleTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *RoleCRUD) error {
	return cx1client.DeleteRoleByID(t.Role.RoleID)
}

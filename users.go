package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func UserTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, users *[]UserCRUD) bool {
	result := true
	for id := range *users {
		t := &(*users)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" || t.Email == "" {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name or email)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				err := UserTestCreate(cx1client, logger, testname, &(*users)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_CREATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func updateUserFromConfig(cx1client *Cx1ClientGo.Cx1Client, t *UserCRUD) error {
	_, err := cx1client.GetUserGroups(t.User)
	if err != nil {
		return err
	}

	for _, g := range t.Groups { // groups to add
		if !t.User.IsInGroupByName(g) {
			group, err := cx1client.GetGroupByName(g)
			if err != nil {
				return fmt.Errorf("failed to find group %v: %s", g, err)
			}
			err = cx1client.AssignUserToGroupByID(t.User, group.GroupID)
			if err != nil {
				return fmt.Errorf("failed to assign user to group %v: %s", g, err)
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
				return fmt.Errorf("failed to remove user from group %v: %s", g.Name, err)
			}
		}
	}

	_, err = cx1client.GetUserRoles(t.User)
	if err != nil {
		return fmt.Errorf("failed to get user's roles: %s", err)
	}

	new_roles := []Cx1ClientGo.Role{}

	for _, newrole := range t.Roles { // check for roles to add
		if !t.User.HasRoleByName(newrole) {
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
			return fmt.Errorf("failed to grant user %v new roles: %s", t.User.String(), err)
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
			return fmt.Errorf("failed to remove user %v old roles: %s", t.User.String(), err)
		}
	}

	return nil
}

func UserTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *UserCRUD) error {
	var test_User Cx1ClientGo.User
	test_User.UserName = t.Name
	test_User.Email = t.Email

	test_User, err := cx1client.CreateUser(test_User)
	if err != nil {
		return err
	}

	t.User = &test_User

	err = updateUserFromConfig(cx1client, t)
	if err != nil {
		return err
	}
	return nil
}

func UserTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, users *[]UserCRUD) bool {
	result := true
	for id := range *users {
		t := &(*users)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" && t.Email == "" {
				LogSkip(t.FailTest, logger, OP_READ, MOD_USER, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name or email)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				err := UserTestRead(cx1client, logger, testname, &(*users)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_READ, MOD_USER, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func UserTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *UserCRUD) error {
	test_User, err := cx1client.GetUserByUserName(t.Name)
	if err != nil {
		return err
	}
	t.User = &test_User
	return nil
}

func UserTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, users *[]UserCRUD) bool {
	result := true
	for id := range *users {
		t := &(*users)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.User == nil {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				err := UserTestUpdate(cx1client, logger, testname, &(*users)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func UserTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *UserCRUD) error {
	err := updateUserFromConfig(cx1client, t)
	if err != nil {
		return err
	}

	return cx1client.UpdateUser(t.User)
}

func UserTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, users *[]UserCRUD) bool {
	result := true
	for id := range *users {
		t := &(*users)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.User == nil {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before deleting)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				err := UserTestDelete(cx1client, logger, testname, &(*users)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_DELETE, MOD_USER, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_USER, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func UserTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *UserCRUD) error {
	err := cx1client.DeleteUser(t.User)
	if err != nil {
		return err
	}

	t.User = nil
	return nil
}

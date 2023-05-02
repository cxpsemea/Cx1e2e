package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func GroupTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, groups *[]GroupCRUD) bool {
	result := true
	for id := range *groups {
		t := &(*groups)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Create Group", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := GroupTestCreate(cx1client, logger, testname, &(*groups)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Create Group", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Create Group", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func updateGroup(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *GroupCRUD) error {
	var err error
	if len(t.ClientRoles) > 0 {
		if len(t.Group.ClientRoles) == 0 {
			t.Group.ClientRoles = make(map[string][]string, 0)
		}
		for _, c := range t.ClientRoles {
			t.Group.ClientRoles[c.Client] = c.Roles
		}
		err = cx1client.UpdateGroup(t.Group)
		if err != nil {
			return fmt.Errorf("failed to set roles for group %v: %s", t.Name, err)
		}
	}

	if t.Parent != "" {
		parent, err := cx1client.GetGroupByName(t.Parent)
		if err != nil {
			return err
		}

		err = cx1client.SetGroupParent(t.Group, &parent)
		if err != nil {
			return fmt.Errorf("failed to set group %v as child under %v: %s", t.Group.GroupID, parent.GroupID, err)
		}
	}

	return nil
}

func GroupTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *GroupCRUD) error {
	test_Group, err := cx1client.CreateGroup(t.Name)
	if err != nil {
		return err
	}
	test_Group, err = cx1client.GetGroupByID(test_Group.GroupID)
	if err != nil {
		return err
	}
	t.Group = &test_Group

	err = updateGroup(cx1client, logger, t)
	if err != nil {
		return err
	}

	return nil
}

func GroupTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, groups *[]GroupCRUD) bool {
	result := true
	for id := range *groups {
		t := &(*groups)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(logger, "Read Group", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := GroupTestRead(cx1client, logger, testname, &(*groups)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Read Group", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Read Group", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func GroupTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *GroupCRUD) error {
	test_Group, err := cx1client.GetGroupByName(t.Name)
	if err != nil {
		return err
	}

	test_Group, err = cx1client.GetGroupByID(test_Group.GroupID)
	if err != nil {
		return err
	}

	t.Group = &test_Group
	return nil
}

func GroupTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, groups *[]GroupCRUD) bool {
	result := true
	for id := range *groups {
		t := &(*groups)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Group == nil {
				LogSkip(logger, "Update Group", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := GroupTestUpdate(cx1client, logger, testname, &(*groups)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Update Group", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Update Group", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func GroupTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *GroupCRUD) error {
	err := updateGroup(cx1client, logger, t)
	if err != nil {
		return err
	}

	return nil
}

func GroupTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, groups *[]GroupCRUD) bool {
	result := true
	for id := range *groups {
		t := &(*groups)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Group == nil {
				LogSkip(logger, "Delete Group", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := GroupTestDelete(cx1client, logger, testname, &(*groups)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Delete Group", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Delete Group", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func GroupTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *GroupCRUD) error {
	err := cx1client.DeleteGroup(t.Group)
	if err != nil {
		return err
	}

	t.Group = nil
	return nil
}

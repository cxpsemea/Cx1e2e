package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *GroupCRUD) Validate(CRUD string) error {
	if (CRUD == OP_UPDATE || CRUD == OP_DELETE) && t.Group == nil {
		return fmt.Errorf("must read before updating or deleting")
	}

	if t.Name == "" {
		return fmt.Errorf("group name is missing")
	}

	return nil
}

func (t *GroupCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string) bool {
	return true
}

func (t *GroupCRUD) GetModule() string {
	return MOD_GROUP
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

func (t *GroupCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *GroupCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
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

func (t *GroupCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	err := updateGroup(cx1client, logger, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *GroupCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	err := cx1client.DeleteGroup(t.Group)
	if err != nil {
		return err
	}

	t.Group = nil
	return nil
}

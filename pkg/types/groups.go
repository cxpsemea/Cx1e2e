package types

import (
	"fmt"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *GroupCRUD) Validate(CRUD string) error {
	if t.Name == "" && t.Path == "" {
		return fmt.Errorf("group name or path is missing")
	}

	if t.Name == "" && t.Path != "" {
		parts := strings.Split(t.Path, "/")
		t.Name = parts[len(parts)-1]
	}

	return nil
}

func (t *GroupCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *GroupCRUD) GetModule() string {
	return MOD_GROUP
}

func setGroupRoles(cx1client *Cx1ClientGo.Cx1Client, t *GroupCRUD) error {
	if len(t.ClientRoles) > 0 {
		if len(t.Group.ClientRoles) == 0 {
			t.Group.ClientRoles = make(map[string][]string, 0)
		}
		for _, c := range t.ClientRoles {
			t.Group.ClientRoles[c.Client] = c.Roles
		}
		err := cx1client.UpdateGroup(t.Group)
		if err != nil {
			return fmt.Errorf("failed to set roles for group %v: %s", t.Name, err)
		}
	}
	return nil
}

func updateGroup(cx1client *Cx1ClientGo.Cx1Client, t *GroupCRUD) error {
	var err error
	if err = setGroupRoles(cx1client, t); err != nil {
		return err
	}

	if t.Parent != "" || t.ParentPath != "" {
		var parent Cx1ClientGo.Group
		if t.Parent != "" {
			parent, err = cx1client.GetGroupByName(t.Parent)
		} else {
			parent, err = cx1client.GetGroupByPath(t.ParentPath)
		}

		if err != nil {
			return err
		}

		if parent.GroupID != t.Group.ParentID { // group needs to move
			err = cx1client.SetGroupParent(t.Group, &parent)
			if err != nil {
				return fmt.Errorf("failed to set group %v as child under %v: %s", t.Group.GroupID, parent.GroupID, err)
			}
		}
	}

	return nil
}

func (t *GroupCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	var err error
	var test_Group Cx1ClientGo.Group
	if t.Parent == "" && (t.Path == "" || t.Path == ("/"+t.Name)) && (t.ParentPath == "" || t.ParentPath == "/") {
		test_Group, err = cx1client.CreateGroup(t.Name)
	} else {
		var parent Cx1ClientGo.Group
		if t.Parent != "" {
			parent, err = cx1client.GetGroupByName(t.Parent)
			if err != nil {
				return fmt.Errorf("failed to get parent group %v: %s", t.Parent, err)
			}
		} else if t.ParentPath != "" && t.ParentPath != "/" {
			parent, err = cx1client.GetGroupByPath(t.ParentPath)
			if err != nil {
				return fmt.Errorf("failed to get parent group %v: %s", t.ParentPath, err)
			}
		} else {
			parts := strings.Split(t.Path, "/")
			parentPath := strings.Join(parts[:len(parts)-1], "/")
			parent, err = cx1client.GetGroupByPath(parentPath)
			if err != nil {
				return fmt.Errorf("failed to get parent group %v: %s", parentPath, err)
			}
		}
		test_Group, err = cx1client.CreateChildGroup(&parent, t.Name)
	}

	if err != nil {
		return err
	}

	test_Group, err = cx1client.GetGroupByID(test_Group.GroupID)
	if err != nil {
		return err
	}
	t.Group = &test_Group

	err = updateGroup(cx1client, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *GroupCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	var err error
	var test_Group Cx1ClientGo.Group

	if t.Path != "" {
		test_Group, err = cx1client.GetGroupByPath(t.Path)

		if err != nil {
			return fmt.Errorf("failed to get group %v: %s", t.Path, err)
		}
	} else if t.Parent != "" || (t.ParentPath != "" && t.ParentPath != "/") {
		var parent Cx1ClientGo.Group
		if t.Parent != "" {
			parent, err = cx1client.GetGroupByName(t.Parent)
			if err != nil {
				return fmt.Errorf("failed to get parent group %v: %s", t.Parent, err)
			}
		} else if t.ParentPath != "" {
			parent, err = cx1client.GetGroupByPath(t.ParentPath)
			if err != nil {
				return fmt.Errorf("failed to get parent group %v: %s", t.ParentPath, err)
			}
		}

		if _, err = cx1client.GetGroupChildren(&parent); err != nil {
			return fmt.Errorf("failed to get parent group %v's children: %s", parent.String(), err)
		}

		if test_Group, err = parent.FindSubgroupByName(t.Name); err != nil {
			return fmt.Errorf("failed to find parent group %v's child %v", parent.String(), t.Name)
		}
	} else {
		test_Group, err = cx1client.GetGroupByName(t.Name)
		if err != nil {
			return err
		}
	}

	test_Group, err = cx1client.GetGroupByID(test_Group.GroupID) // TODO: review if this remains necessary in 3.25+
	if err != nil {
		return err
	}

	t.Group = &test_Group
	return nil
}

func (t *GroupCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Group == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := updateGroup(cx1client, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *GroupCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Group == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeleteGroup(t.Group)
	if err != nil {
		return err
	}

	t.Group = nil
	return nil
}

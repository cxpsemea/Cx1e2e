package types

import (
	"fmt"
	"slices"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *ProjectCRUD) Validate(CRUD string) error {
	/*if (CRUD == OP_UPDATE || CRUD == OP_DELETE) && t.Project == nil {
		return fmt.Errorf("must read before updating or deleting")
	}*/

	if t.Name == "" {
		return fmt.Errorf("project name is missing")
	}

	if CRUD == OP_CREATE && t.Applications != nil && len(*t.Applications) > 1 {
		return fmt.Errorf("cannot create a project inside multiple applications - create it in one application, then update it to add others")
	}

	if t.Application != "" {
		return fmt.Errorf("the configuration %v test %v includes a project with an 'Application' set. This has been replaced by the array 'Applications' - please update your configuration", t.TestSource, t.Name)
	}

	return nil
}

func (t *ProjectCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *ProjectCRUD) GetModule() string {
	return MOD_PROJECT
}

func (t *ProjectCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	group_ids := []string{}

	if t.Groups != nil {
		for _, g := range *t.Groups {
			group, err := cx1client.GetGroupByName(g)
			if err != nil {
				return err
			}
			group_ids = append(group_ids, group.GroupID)
		}
	}

	tags := make(map[string]string)
	for _, tag := range t.Tags {
		tags[tag.Key] = tag.Value
	}

	if t.Applications == nil {
		test_Project, err := cx1client.CreateProject(t.Name, group_ids, tags)
		if err != nil {
			return err
		}
		t.Project = &test_Project
	} else {
		app, err := cx1client.GetApplicationByName((*t.Applications)[0])
		if err != nil {
			return err
		}
		test_Project, err := cx1client.CreateProjectInApplication(t.Name, group_ids, tags, app.ApplicationID)
		if err != nil {
			return err
		}
		t.Project = &test_Project
	}

	if t.Project != nil {
		if t.Preset != "" {
			projConfig := Cx1ClientGo.ConfigurationSetting{
				Key:           "scan.config.sast.presetName",
				Name:          "presetName",
				Category:      "sast",
				AllowOverride: true,
				Value:         t.Preset,
			}

			err := cx1client.UpdateProjectConfiguration(t.Project, []Cx1ClientGo.ConfigurationSetting{projConfig})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *ProjectCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	test_Project, err := cx1client.GetProjectByName(t.Name)
	if err != nil {
		return err
	}

	t.Project = &test_Project

	if t.Applications != nil && t.IsType(OP_READ) { // we only want to validate on read (read op can be called from update, which could be adding apps)
		for _, appName := range *t.Applications {
			match := false
			app, err := cx1client.GetApplicationByName(appName)
			if err != nil {
				return err
			}

			for _, p := range *app.ProjectIds {
				if p == t.Project.ProjectID {
					match = true
					break
				}
			}

			if !match {
				return fmt.Errorf("expected project %v to live under application %v but it does not", t.Name, appName)
			}
		}
	}

	return nil
}

func (t *ProjectCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Project == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	if t.Applications != nil {
		// this flow has to change based on direct_app_association
		if flag, _ := cx1client.CheckFlag("DIRECT_APP_ASSOCIATION_ENABLED"); flag {
			// with direct app association, we can do all of the project-to-app assignment via the project update call
			newApps := []string{}
			for _, appName := range *t.Applications {
				app, err := cx1client.GetApplicationByName(appName)
				if err != nil {
					return err
				}
				newApps = append(newApps, app.ApplicationID)
				if !slices.Contains(*t.Project.Applications, app.ApplicationID) {
					t.Project.AssignApplicationByID(app.ApplicationID)
				}
			}
			currentApps := *t.Project.Applications
			for _, appId := range currentApps {
				if !slices.Contains(newApps, appId) {
					t.Project.RemoveApplicationByID(appId)
				}
			}

			err := cx1client.UpdateProject(t.Project)
			if err != nil {
				return err
			}
		} else {
			// without direct app assignment, all changes must be done via rules through the application
			newApps := []string{}
			for _, appName := range *t.Applications {
				app, err := cx1client.GetApplicationByName(appName)
				if err != nil {
					return err
				}
				newApps = append(newApps, app.ApplicationID)
				app.AssignProject(t.Project)
				err = cx1client.UpdateApplication(&app)
				if err != nil {
					return fmt.Errorf("failed to add project %v to application %v: %v", t.Project.String(), app.String(), err)
				}
			}
			//t.Project.Applications = &newApps

			for _, appId := range *t.Project.Applications {
				if !slices.Contains(newApps, appId) {
					app, err := cx1client.GetApplicationByID(appId)
					if err != nil {
						return err
					}
					app.UnassignProject(t.Project)
					err = cx1client.UpdateApplication(&app)
					if err != nil {
						return fmt.Errorf("failed to remove project %v from application %v: %v", t.Project.String(), app.String(), err)
					}
				}
			}

			project, err := cx1client.GetProjectByID(t.Project.ProjectID)
			if err != nil {
				return err
			}
			t.Project = &project
		}
	}

	if len(t.Tags) > 0 {
		t.Project.Tags = make(map[string]string)
		for _, tag := range t.Tags {
			t.Project.Tags[tag.Key] = tag.Value
		}
		err := cx1client.UpdateProject(t.Project)
		if err != nil {
			return err
		}
	}

	if t.Preset != "" {
		projConfig := Cx1ClientGo.ConfigurationSetting{
			Key:           "scan.config.sast.presetName",
			Name:          "presetName",
			Category:      "sast",
			AllowOverride: true,
			Value:         t.Preset,
		}

		err := cx1client.UpdateProjectConfiguration(t.Project, []Cx1ClientGo.ConfigurationSetting{projConfig})
		if err != nil {
			return err
		}
	}

	if t.Groups != nil { // the groups field is set in the yaml, so we want to change/update the groups
		group_ids := []string{}

		changed := false
		for _, g := range *t.Groups {
			group, err := cx1client.GetGroupByName(g)
			if err != nil {
				return err
			}
			group_ids = append(group_ids, group.GroupID)
			if !slices.Contains(t.Project.Groups, group.GroupID) {
				changed = true
			}
		}

		for _, g := range t.Project.Groups {
			if !slices.Contains(group_ids, g) {
				changed = true
			}
		}

		if changed {
			t.Project.Groups = group_ids
			if err := cx1client.UpdateProject(t.Project); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *ProjectCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Project == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeleteProject(t.Project)
	if err != nil {
		return err
	}

	t.Project = nil
	return nil
}

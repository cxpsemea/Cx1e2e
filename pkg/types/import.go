package types

import (
	"fmt"
	"os"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *ImportCRUD) Validate(CRUD string) error {
	if CRUD != OP_CREATE {
		return fmt.Errorf("test type is not supported")
	}

	if t.Name == "" {
		return fmt.Errorf("import name is missing")
	}

	if t.ZipFile == "" || t.EncryptionKey == "" {
		return fmt.Errorf("missing zipfile or encryption key")
	}

	return nil
}

func (t *ImportCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	if CRUD != OP_CREATE {
		return fmt.Errorf("can only create an import")
	}
	return nil
}

func (t *ImportCRUD) GetModule() string {
	return MOD_IMPORT
}

func (t *ImportCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	fileContents, err := os.ReadFile(t.ZipFile)
	if err != nil {
		return fmt.Errorf("failed to read %v: %s", t.ZipFile, err)
	}

	projectMapping := []byte{}
	if t.ProjectMapFile != "" {
		projectMapping, err = os.ReadFile(t.ProjectMapFile)
		if err != nil {
			return fmt.Errorf("failed to read %v: %s", t.ProjectMapFile, err)
		}
	}

	importID, err := cx1client.StartMigration(fileContents, projectMapping, t.EncryptionKey) // no project-to-app mapping
	if err != nil {
		return fmt.Errorf("failed to start import: %v", err)
	}

	var result string
	if t.TimeoutSeconds != 0 {
		cvars := cx1client.GetClientVars()
		result, err = cx1client.ImportPollingByIDWithTimeout(importID, cvars.MigrationPollingDelaySeconds, t.TimeoutSeconds)
	} else {
		result, err = cx1client.ImportPollingByID(importID)
	}
	if err != nil {
		return fmt.Errorf("failed during import: %s", err)
	}

	if result == "Failed" {
		return fmt.Errorf("import failed")
	}

	return nil
}

func (t *ImportCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *ImportCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *ImportCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

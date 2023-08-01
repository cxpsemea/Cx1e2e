package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func ImportTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Imports *[]ImportCRUD) bool {
	result := true
	for id := range *Imports {
		t := &(*Imports)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" || t.ZipFile == "" || t.EncryptionKey == "" {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name, zipfile, or encryption key)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource)
				err := ImportTestCreate(cx1client, logger, testname, &(*Imports)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_CREATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func ImportTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *ImportCRUD) error {
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

	result, err := cx1client.ImportPollingByID(importID)
	if err != nil {
		return fmt.Errorf("failed during import: %s", err)
	}

	if result == "Failed" {
		return fmt.Errorf("import failed")
	}

	return nil
}

func ImportTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Imports *[]ImportCRUD) bool {
	for id := range *Imports {
		t := &(*Imports)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_READ, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
	return true
}

func ImportTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Imports *[]ImportCRUD) bool {
	for id := range *Imports {
		t := &(*Imports)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_UPDATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
	return true
}

func ImportTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Imports *[]ImportCRUD) bool {
	for id := range *Imports {
		t := &(*Imports)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_UPDATE, MOD_IMPORT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
	return true
}

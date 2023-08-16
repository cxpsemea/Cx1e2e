package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func FlagTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Flags *[]FlagCRUD) {
	for id := range *Flags {
		t := &(*Flags)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_CREATE, MOD_RESULT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
}

func FlagTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Flags *[]FlagCRUD) {
	for id := range *Flags {
		t := &(*Flags)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(t.FailTest, logger, OP_READ, MOD_FLAG, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_FLAG, start, testname, id+1, t.String(), t.TestSource)
				err := FlagTestRead(cx1client, logger, testname, &(*Flags)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_READ, MOD_FLAG, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_FLAG, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func FlagTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *FlagCRUD) error {
	test_Flag, err := cx1client.CheckFlag(t.Name)
	if err != nil {
		return err
	}

	logger.Debugf("Flag %v is set to %v", t.Name, test_Flag)

	if !test_Flag {
		return fmt.Errorf("flag %v set to false", t.Name)
	}

	return nil
}

func FlagTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Flags *[]FlagCRUD) {
	for id := range *Flags {
		t := &(*Flags)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
}

func FlagTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, Flags *[]FlagCRUD) {
	for id := range *Flags {
		t := &(*Flags)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			LogSkip(t.FailTest, logger, OP_UPDATE, MOD_RESULT, start, testname, id+1, t.String(), t.TestSource, "action not supported")
		}
	}
}

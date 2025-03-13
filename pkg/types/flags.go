package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *FlagCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("flag name is missing")
	}

	return nil
}

func (t *FlagCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	if CRUD != OP_READ {
		return fmt.Errorf("can only read flags")
	}
	return nil
}

func (t *FlagCRUD) GetModule() string {
	return MOD_FLAG
}

func (t *FlagCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *FlagCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
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

func (t *FlagCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

func (t *FlagCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	return fmt.Errorf("not supported")
}

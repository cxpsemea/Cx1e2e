package types

import (
	"fmt"
	"strings"

	"github.com/cxpsemea/Cx1ClientGo"
)

func (t *PresetCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("preset name is missing")
	}
	if t.Engine == "" {
		return fmt.Errorf("engine name is missing")
	}

	t.Engine = strings.ToLower(t.Engine)
	if t.Engine != "sast" && t.Engine != "iac" {
		return fmt.Errorf("engine must be 'sast' or 'iac'")
	}

	return nil
}

func (t *PresetCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *PresetCRUD) GetModule() string {
	return MOD_PRESET
}

func getQueryCollection(cx1client *Cx1ClientGo.Cx1Client, _ *ThreadLogger, t *PresetCRUD) (Cx1ClientGo.SASTQueryCollection, error) {
	collection := Cx1ClientGo.SASTQueryCollection{}
	qc, err := cx1client.GetQueries()
	if err != nil {
		return collection, fmt.Errorf("failed to retrieve query collection: %s", err)
	}

	for _, q := range t.SASTQueries {
		qq := qc.GetQueryByName(q.QueryLanguage, q.QueryGroup, q.QueryName)
		if qq == nil {
			return collection, fmt.Errorf("failed to find query %v -> %v -> %v", q.QueryLanguage, q.QueryGroup, q.QueryName)
		}

		collection.AddQuery(*qq)
	}
	return collection, nil
}

func (t *PresetCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	query_ids, err := getQueryCollection(cx1client, logger, t)
	if err != nil {
		return err
	}

	test_Preset, err := cx1client.CreateSASTPreset(t.Name, t.Description, query_ids)
	if err != nil {
		return err
	}
	t.Preset = &test_Preset
	return nil
}

func (t *PresetCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	test_Preset, err := cx1client.GetPresetByName(t.Engine, t.Name)
	if err != nil {
		return err
	}
	t.Preset = &test_Preset
	return nil
}

func (t *PresetCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Preset == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	queryCollection, err := getQueryCollection(cx1client, logger, t)
	if err != nil {
		return err
	}

	//t.Preset.QueryIDs = query_ids
	t.Preset.UpdateQueries(queryCollection)
	err = cx1client.UpdateSASTPreset(*t.Preset)
	return err
}

func (t *PresetCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Preset == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeletePreset(*t.Preset)
	if err != nil {
		return err
	}

	t.Preset = nil
	return nil
}

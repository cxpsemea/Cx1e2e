package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (t *PresetCRUD) Validate(CRUD string) error {
	if t.Name == "" {
		return fmt.Errorf("preset name is missing")
	}

	return nil
}

func (t *PresetCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *PresetCRUD) GetModule() string {
	return MOD_PRESET
}

func getQueryIDs(cx1client *Cx1ClientGo.Cx1Client, _ *logrus.Logger, t *PresetCRUD) ([]uint64, error) {
	query_ids := make([]uint64, len(t.Queries))

	qc, err := cx1client.GetQueries()
	if err != nil {
		return query_ids, fmt.Errorf("failed to retrieve query collection: %s", err)
	}

	for id, q := range t.Queries {
		qq := qc.GetQueryByName(q.QueryLanguage, q.QueryGroup, q.QueryName)
		if qq == nil {
			return query_ids, fmt.Errorf("failed to find query %v -> %v -> %v", q.QueryLanguage, q.QueryGroup, q.QueryName)
		}
		query_ids[id] = qq.QueryID
	}
	return query_ids, nil
}

func (t *PresetCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	query_ids, err := getQueryIDs(cx1client, logger, t)
	if err != nil {
		return err
	}

	test_Preset, err := cx1client.CreatePreset(t.Name, t.Description, query_ids)
	if err != nil {
		return err
	}
	t.Preset = &test_Preset
	return nil
}

func (t *PresetCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	test_Preset, err := cx1client.GetPresetByName(t.Name)
	if err != nil {
		return err
	}
	t.Preset = &test_Preset
	return nil
}

func (t *PresetCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Preset == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	query_ids, err := getQueryIDs(cx1client, logger, t)
	if err != nil {
		return err
	}

	t.Preset.QueryIDs = query_ids
	err = cx1client.UpdatePreset(t.Preset)
	return err
}

func (t *PresetCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.Preset == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	err := cx1client.DeletePreset(t.Preset)
	if err != nil {
		return err
	}

	t.Preset = nil
	return nil
}

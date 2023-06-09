package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func PresetTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, presets *[]PresetCRUD) bool {
	result := true
	for id := range *presets {
		t := &(*presets)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				err := PresetTestCreate(cx1client, logger, testname, &(*presets)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_CREATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				}
			}

		}
	}
	return result
}

func getQueryIDs(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *PresetCRUD) ([]uint64, error) {
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

func PresetTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *PresetCRUD) error {
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

func PresetTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, presets *[]PresetCRUD) bool {
	result := true
	for id := range *presets {
		t := &(*presets)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if t.Name == "" {
				LogSkip(t.FailTest, logger, OP_READ, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				err := PresetTestRead(cx1client, logger, testname, &(*presets)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_READ, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func PresetTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *PresetCRUD) error {
	test_Preset, err := cx1client.GetPresetByName(t.Name)
	if err != nil {
		return err
	}
	t.Preset = &test_Preset
	return nil
}

func PresetTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, presets *[]PresetCRUD) bool {
	result := true
	for id := range *presets {
		t := &(*presets)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Preset == nil {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				err := PresetTestUpdate(cx1client, logger, testname, &(*presets)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				}
			}

		}
	}
	return result
}

func PresetTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *PresetCRUD) error {
	query_ids, err := getQueryIDs(cx1client, logger, t)
	if err != nil {
		return err
	}

	t.Preset.QueryIDs = query_ids
	err = cx1client.UpdatePreset(t.Preset)
	return err
}

func PresetTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, presets *[]PresetCRUD) bool {
	result := true
	for id := range *presets {
		t := &(*presets)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Preset == nil {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before deleting)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				err := PresetTestDelete(cx1client, logger, testname, &(*presets)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, OP_DELETE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_PRESET, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
	return result
}

func PresetTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *PresetCRUD) error {
	err := cx1client.DeletePreset(t.Preset)
	if err != nil {
		return err
	}

	t.Preset = nil
	return nil
}

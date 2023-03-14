package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

/*
	These are for CxQL (CxAudit) endpoints specifically.
	Currently not implemented in the Cx1Client
*/

func (q *CxQLCRUD) IsValidCxQL() bool {
	return q.QueryID != 0 || (q.QueryLanguage != "" && q.QueryGroup != "" && q.QueryName != "")
}

func CxQLTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidCxQL() {
				LogSkip(logger, "Create CxQL", start, testname, id+1, "invalid test (missing CxQL identifier)")
			} else {
				err := CxQLTestCreate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(logger, "Create CxQL", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Create CxQL", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func CxQLTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	/*	test_CxQL, err := cx1client.CreateCxQL(t.CxQLName)
		if err != nil {
			return err
		}
		t.CxQL = &test_CxQL*/

	return fmt.Errorf("not implemented")
}

func CxQLTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidCxQL() {
				LogSkip(logger, "Read CxQL", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := CxQLTestRead(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(logger, "Read CxQL", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Read CxQL", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

/*
func getCxQL(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) *Cx1ClientGo.Query {
	if t.QueryID == 0 {
		qc, err := cx1client.GetQueries()

		if err != nil {
			return nil
		}

		return qc.GetQueryByName(t.QueryLanguage, t.QueryGroup, t.QueryName)
	} else {
		q, _ := cx1client.GetQueryByID(t.QueryID)
		return &q
	}
}
*/

func CxQLTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	/*	t.Query = getCxQL(cx1client, t)

		if t.Query == nil {
			return fmt.Errorf("no such Query %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}*/

	return fmt.Errorf("not implemented")
}

func CxQLTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(logger, "Update CxQL", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := CxQLTestUpdate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(logger, "Update CxQL", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Update CxQL", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func CxQLTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	// TODO
	return fmt.Errorf("not implemented")
}

func CxQLTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(logger, "Delete CxQL", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := CxQLTestDelete(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(logger, "Delete CxQL", start, testname, id+1, t.String(), err)
				} else {
					LogPass(logger, "Delete CxQL", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func CxQLTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	return fmt.Errorf("not implemented")
}

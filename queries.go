package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

/*
	The /api/queries api is expected to be replaced by cxaudit api, hence this set of tests will be changed over to the cxaudit endpoints.
*/

func (q *QueryCRUD) IsValidQuery() bool {
	return q.QueryID != 0 || (q.QueryLanguage != "" && q.QueryGroup != "" && q.QueryName != "")
}

func QueryTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]QueryCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidQuery() {
				LogSkip(logger, "Create Query", start, testname, id+1, "invalid test (missing query identifier)")
			} else {
				err := QueryTestCreate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Create Query", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Create Query", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func QueryTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *QueryCRUD) error {
	/*	test_Query, err := cx1client.CreateQuery(t.QueryName)
		if err != nil {
			return err
		}
		t.Query = &test_Query*/

	return fmt.Errorf("not implemented")
}

func QueryTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]QueryCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidQuery() {
				LogSkip(logger, "Read Query", start, testname, id+1, "invalid test (missing name)")
			} else {
				err := QueryTestRead(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Read Query", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Read Query", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func getQuery(cx1client *Cx1ClientGo.Cx1Client, t *QueryCRUD) *Cx1ClientGo.Query {
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

func QueryTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *QueryCRUD) error {
	t.Query = getQuery(cx1client, t)

	if t.Query == nil {
		return fmt.Errorf("no such query %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	return fmt.Errorf("not implemented")
}

func QueryTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]QueryCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(logger, "Update Query", start, testname, id+1, "invalid test (must read before updating)")
			} else {
				err := QueryTestUpdate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Update Query", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Update Query", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func QueryTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *QueryCRUD) error {
	// TODO
	return fmt.Errorf("not implemented")
}

func QueryTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]QueryCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(logger, "Delete Query", start, testname, id+1, "invalid test (must read before deleting)")
			} else {
				err := QueryTestDelete(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					result = false
					LogFail(t.FailTest, logger, "Delete Query", start, testname, id+1, t.String(), err)
				} else {
					LogPass(t.FailTest, logger, "Delete Query", start, testname, id+1, t.String())
				}
			}
		}
	}
	return result
}

func QueryTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *QueryCRUD) error {
	return fmt.Errorf("not implemented")
}

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (q *CxQLCRUD) IsValidQuery() bool {
	return q.QueryLanguage != "" && q.QueryGroup != "" && q.QueryName != "" && q.Scope.Project != ""
}

func QueryTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidQuery() {
				LogSkip(t.FailTest, logger, "Create Query", start, testname, id+1, t.String(), "invalid test (missing query identifier)")
			} else {
				LogStart(t.FailTest, logger, "Create Query", start, testname, id+1, t.String())
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

func getAuditSession(cx1client *Cx1ClientGo.Cx1Client, projectId string) (string, error) {
	lastscans, err := cx1client.GetLastScansByStatusAndID(projectId, 1, []string{"Completed"})
	if err != nil {
		logger.Errorf("Error getting last successful scan for project %v: %s", projectId, err)
		return "", err
	}

	if len(lastscans) == 0 {
		return "", fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", projectId)
	}

	lastscan := lastscans[0]

	return cx1client.GetAuditSessionByID(projectId, lastscan.ScanID, true)
}

func getQueryScope(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) (string, error) {
	scope := "Corp"
	if !t.Scope.Corp {
		if t.Scope.Application != "" {
			app, err := cx1client.GetApplicationByName(t.Scope.Application)
			if err != nil {
				return "", fmt.Errorf("failed to find application named %v", t.Scope.Application)
			}
			scope = app.ApplicationID
		} else {
			proj, err := cx1client.GetProjectByName(t.Scope.Project)
			if err != nil {
				return "", fmt.Errorf("failed to find project named %v", t.Scope.Project)
			}
			scope = proj.ProjectID
		}
	}
	return scope, nil
}

func getQuery(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *CxQLCRUD) *Cx1ClientGo.AuditQuery {
	scope, err := getQueryScope(cx1client, t)
	if err != nil {
		logger.Errorf("Error with query scope: %v", err)
		return nil
	}

	auditQuery, err := cx1client.GetQueryByName(scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	if err != nil {
		logger.Warnf("Error getting query %v: %s", t.String(), err)
		return nil
	}

	logger.Debugf("Found query %v", auditQuery.String())

	return &auditQuery
}

func updateQuery(cx1client *Cx1ClientGo.Cx1Client, query *Cx1ClientGo.AuditQuery, projectId string, t *CxQLCRUD) error {
	session, err := getAuditSession(cx1client, projectId)
	if err != nil {
		return err
	}

	switch strings.ToUpper(t.Severity) {
	case "INFO":
		query.Severity = 0
	case "INFORMATION":
		query.Severity = 0
	case "LOW":
		query.Severity = 1
	case "MEDIUM":
		query.Severity = 2
	case "HIGH":
		query.Severity = 3
	}

	if t.Source != "" {
		query.Source = t.Source
	}

	if t.Compile {
		err = cx1client.AuditCompileQuery(session, *query)
		if err != nil {
			logger.Errorf("Error triggering query compile: %s", err)
			return err
		}

		err = cx1client.AuditCompilePollingByID(session)
		if err != nil {
			logger.Errorf("Error while polling compiler: %s", err)
			return err
		}
	}

	err = cx1client.UpdateQuery(session, *query)
	if err != nil {
		return err
	}

	return nil
}

func QueryTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	query := getQuery(cx1client, logger, t)
	proj, err := cx1client.GetProjectByName(t.Scope.Project)
	if err != nil {
		return err
	}

	if query != nil {
		logger.Infof("Found query: %v", query.String())
		return updateQuery(cx1client, query, proj.ProjectID, t)
	} else {
		// query does not exist at all so needs to be created on corp level
		// Second query: create new corp/tenant query
		newQuery, err := cx1client.AuditCreateQuery(t.QueryLanguage, t.QueryGroup, t.QueryName)
		if err != nil {
			return err
		}
		newQuery.Source = t.Source
		return updateQuery(cx1client, &newQuery, proj.ProjectID, t)
	}
}

func QueryTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValidQuery() {
				LogSkip(t.FailTest, logger, "Read Query", start, testname, id+1, t.String(), "invalid test (missing name)")
			} else {
				LogStart(t.FailTest, logger, "Read Query", start, testname, id+1, t.String())
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

func QueryTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	t.Query = getQuery(cx1client, logger, t)

	if t.Query == nil {
		return fmt.Errorf("no such query %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	return nil
}

func QueryTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(t.FailTest, logger, "Update Query", start, testname, id+1, t.String(), "invalid test (must read before updating)")
			} else {
				LogStart(t.FailTest, logger, "Update Query", start, testname, id+1, t.String())
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

func QueryTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	proj, err := cx1client.GetProjectByName(t.Scope.Project)
	if err != nil {
		return err
	}

	return updateQuery(cx1client, t.Query, proj.ProjectID, t)
}

func QueryTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) bool {
	result := true
	for id := range *queries {
		t := &(*queries)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(t.FailTest, logger, "Delete Query", start, testname, id+1, t.String(), "invalid test (must read before deleting)")
			} else {
				LogStart(t.FailTest, logger, "Delete Query", start, testname, id+1, t.String())
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

func QueryTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	return cx1client.DeleteQuery(*t.Query)
}

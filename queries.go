package main

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

func (q *CxQLCRUD) IsValid() bool {
	return q.QueryLanguage != "" && q.QueryGroup != "" && q.QueryName != "" && (q.Scope.Project != "" || q.Scope.Application != "")
}

func CheckALQFlag(cx1client *Cx1ClientGo.Cx1Client) bool {
	appLevelQueries, err := cx1client.CheckFlag("AUDIT_APPLICATION_LEVEL_ENABLED")
	if err != nil {
		return false
	}
	return appLevelQueries
}

func QueryTestsCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) {
	for id := range *queries {
		t := &(*queries)[id]
		if IsCreate(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing query identifier)")
			} else if !CheckALQFlag(cx1client) && t.Scope.Application != "" {
				LogSkip(t.FailTest, logger, OP_CREATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (application-level queries are not enabled)")
			} else {
				LogStart(t.FailTest, logger, OP_CREATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				err := QueryTestCreate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_CREATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_CREATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func getAuditSession(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) (string, error) {
	if t.LastScan == nil {
		proj, err := cx1client.GetProjectByName(t.Scope.Project)
		if err != nil {
			return "", err
		}

		lastscans, err := cx1client.GetLastScansByStatusAndID(proj.ProjectID, 1, []string{"Completed"})
		if err != nil {
			return "", fmt.Errorf("error getting last successful scan for project %v: %s", proj.ProjectID, err)
		}

		if len(lastscans) == 0 {
			return "", fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", proj.ProjectID)
		}

		t.LastScan = &lastscans[0]
	}

	return cx1client.GetAuditSessionByID(t.LastScan.ProjectID, t.LastScan.ScanID, true)
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

	t.ScopeID = scope

	auditQuery, err := cx1client.GetQueryByName(scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	if err != nil {
		logger.Warnf("Error getting query %v: %s", t.String(), err)
		return nil
	}

	return &auditQuery
}

func compileQuery(cx1client *Cx1ClientGo.Cx1Client, query *Cx1ClientGo.AuditQuery, t *CxQLCRUD) error {
	session, err := getAuditSession(cx1client, t)
	if err != nil {
		return err
	}

	err = cx1client.AuditCompileQuery(session, *query)
	if err != nil {
		return fmt.Errorf("error triggering query compile: %s", err)
	}

	err = cx1client.AuditCompilePollingByID(session)
	if err != nil {
		return fmt.Errorf("error while polling compiler: %s", err)
	}
	return nil
}

func updateQuery(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) error {
	t.Query.Severity = cx1client.GetSeverityID(t.Severity)

	if t.Source != "" {
		t.Query.Source = t.Source
	}

	t.Query.IsExecutable = t.IsExecutable

	if t.Compile {
		err := compileQuery(cx1client, t.Query, t)
		if err != nil {
			return err
		}
	}

	err := cx1client.UpdateQuery(*t.Query)
	if err != nil {
		return err
	}

	return nil
}

func QueryTestCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	t.Query = getQuery(cx1client, logger, t)

	if t.Query != nil {
		logger.Debugf("Found query: %v", t.Query.String())

		if t.Scope.Corp {
			//logger.Info("Will create corp override")
			newq := t.Query.CreateTenantOverride()
			t.Query = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq := t.Query.CreateApplicationOverrideByID(t.ScopeID)
				t.Query = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq := t.Query.CreateProjectOverrideByID(t.ScopeID)
				t.Query = &newq
			}
		}

		logger.Debugf("Updating query %v", t.Query.String())
		return updateQuery(cx1client, t)
	} else {
		// query does not exist at all so needs to be created on corp level
		// Second query: create new corp/tenant query

		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		newQuery, err := cx1client.AuditNewQuery(t.QueryLanguage, t.QueryGroup, t.QueryName)
		if err != nil {
			return err
		}
		newQuery.Source = t.Source
		newQuery.Severity = cx1client.GetSeverityID(t.Severity)
		newQuery.IsExecutable = t.IsExecutable

		if t.Compile {
			err = compileQuery(cx1client, &newQuery, t)
			if err != nil {
				return err
			}
		}

		session, err := getAuditSession(cx1client, t)
		if err != nil {
			return err
		}

		newQuery, err = cx1client.AuditCreateCorpQuery(session, newQuery)
		if err != nil {
			return err
		}
		t.Query = &newQuery

		return nil
	}
}

func QueryTestsRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) {
	for id := range *queries {
		t := &(*queries)[id]
		if IsRead(t.Test) {
			start := time.Now().UnixNano()
			if !t.IsValid() {
				LogSkip(t.FailTest, logger, OP_READ, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (missing name)")
			} else if !CheckALQFlag(cx1client) && t.Scope.Application != "" {
				LogSkip(t.FailTest, logger, OP_READ, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (application-level queries are not enabled)")
			} else {
				LogStart(t.FailTest, logger, OP_READ, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				err := QueryTestRead(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_READ, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_READ, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func QueryTestRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	t.Query = getQuery(cx1client, logger, t)

	if t.Query == nil {
		return fmt.Errorf("no such query %v: %v -> %v -> %v exists", t.Scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	return nil
}

func QueryTestsUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) {
	for id := range *queries {
		t := &(*queries)[id]
		if IsUpdate(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before updating)")
			} else if !CheckALQFlag(cx1client) && t.Scope.Application != "" {
				LogSkip(t.FailTest, logger, OP_UPDATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (application-level queries are not enabled)")
			} else {
				LogStart(t.FailTest, logger, OP_UPDATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				err := QueryTestUpdate(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_UPDATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_UPDATE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func QueryTestUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	return updateQuery(cx1client, t)
}

func QueryTestsDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, queries *[]CxQLCRUD) {
	for id := range *queries {
		t := &(*queries)[id]
		if IsDelete(t.Test) {
			start := time.Now().UnixNano()
			if t.Query == nil {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (must read before deleting)")
			} else if !CheckALQFlag(cx1client) && t.Scope.Application != "" {
				LogSkip(t.FailTest, logger, OP_DELETE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, "invalid test (application-level queries are not enabled)")
			} else {
				LogStart(t.FailTest, logger, OP_DELETE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				err := QueryTestDelete(cx1client, logger, testname, &(*queries)[id])
				if err != nil {
					LogFail(t.FailTest, logger, OP_DELETE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource, err)
				} else {
					LogPass(t.FailTest, logger, OP_DELETE, MOD_QUERY, start, testname, id+1, t.String(), t.TestSource)
				}
			}
		}
	}
}

func QueryTestDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, testname string, t *CxQLCRUD) error {
	return cx1client.DeleteQuery(*t.Query)
}

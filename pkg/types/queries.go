package types

import (
	"fmt"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
)

// var auditSession *Cx1ClientGo.AuditSession
var ASM *AuditSessionManager

func init() {
	ASM = NewAuditSessionManager()
}

func (t *CxQLCRUD) Validate(CRUD string) error {
	if t.QueryLanguage == "" || t.QueryGroup == "" || t.QueryName == "" {
		return fmt.Errorf("query language, group, or name is missing")
	}

	if t.Scope.Project == "" {
		return fmt.Errorf("project name is missing")
	}

	return nil
}

func (t *CxQLCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	return nil
}

func (t *CxQLCRUD) GetModule() string {
	return MOD_QUERY
}

func CheckALQFlag(cx1client *Cx1ClientGo.Cx1Client) bool {
	appLevelQueries, err := cx1client.CheckFlag("AUDIT_APPLICATION_LEVEL_ENABLED")
	if err != nil {
		return false
	}
	return appLevelQueries
}

func getAuditSession(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.AuditSession, error) {
	if t.LastScan == nil {
		proj, err := cx1client.GetProjectByName(t.Scope.Project)
		if err != nil {
			return nil, err
		}

		lastscans, err := cx1client.GetLastScansByStatusAndID(proj.ProjectID, 1, []string{"Completed"})
		if err != nil {
			return nil, fmt.Errorf("error getting last successful scan for project %v: %s", proj.ProjectID, err)
		}

		if len(lastscans) == 0 {
			return nil, fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", proj.ProjectID)
		}

		t.LastScan = &lastscans[0]
	}

	return ASM.GetOrCreateSession(t.Scope, t.QueryLanguage, t.LastScan, cx1client, logger)
}

/*
func getAuditSession_old(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	if auditSession != nil {
		if (t.Scope.Corp || auditSession.ProjectID == t.Scope.ProjectID) && auditSession.HasLanguage(t.QueryLanguage) {
			err := cx1client.AuditSessionKeepAlive(auditSession)
			if err != nil {
				auditSession = nil
				logger.Warningf("Tried to reuse existing audit session but it couldn't be refreshed")
			} else {
				scope := "Tenant"
				if !t.Scope.Corp {
					if t.Scope.Application != "" {
						scope = fmt.Sprintf("application %v", t.Scope.Application)
					} else {
						scope = fmt.Sprintf("project %v", t.Scope.Project)
					}
				}
				logger.Warningf("Reusing existing %v (scope: %v, project ID: %v, language: %v)", auditSession.String(), scope, t.Scope.ProjectID, t.QueryLanguage)
				return nil
			}
		} else {
			logger.Warningf("Existing audit session is not suitable (corp? %v, has %v? %v, is project id %v? %v)", t.Scope.Corp, t.QueryLanguage, auditSession.HasLanguage(t.QueryLanguage), t.Scope.ProjectID, auditSession.ProjectID)
		}
	}

	if t.LastScan == nil {
		proj, err := cx1client.GetProjectByName(t.Scope.Project)
		if err != nil {
			return err
		}

		lastscans, err := cx1client.GetLastScansByStatusAndID(proj.ProjectID, 1, []string{"Completed"})
		if err != nil {
			return fmt.Errorf("error getting last successful scan for project %v: %s", proj.ProjectID, err)
		}

		if len(lastscans) == 0 {
			return fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", proj.ProjectID)
		}

		t.LastScan = &lastscans[0]
	}

	session, err := cx1client.GetAuditSessionByID("sast", t.LastScan.ProjectID, t.LastScan.ScanID)
	if err == nil {
		auditSession = &session
	}

	return err
}
*/

func getQueryScope(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) (string, string, error) {
	scope := "Tenant"
	scopeStr := cx1client.QueryTypeTenant()
	if !t.Scope.Corp {
		proj, err := cx1client.GetProjectByName(t.Scope.Project)
		if err != nil {
			return "", "", fmt.Errorf("failed to find project named %v", t.Scope.Project)
		}

		t.Scope.ProjectID = proj.ProjectID

		if t.Scope.Application != "" {
			app, err := cx1client.GetApplicationByName(t.Scope.Application)
			if err != nil {
				return "", "", fmt.Errorf("failed to find application named %v", t.Scope.Application)
			}
			scope = app.ApplicationID
			scopeStr = cx1client.QueryTypeApplication()
		} else {
			scope = proj.ProjectID
			scopeStr = cx1client.QueryTypeProject()
		}
	}
	return scope, scopeStr, nil
}

func getQuery(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.Query, *Cx1ClientGo.Query) {
	scope, scopeStr, err := getQueryScope(cx1client, t)
	if err != nil {
		logger.Errorf("Error with query scope: %v", err)
		return nil, nil
	}

	t.ScopeID = scope
	t.ScopeStr = scopeStr

	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		logger.Errorf("Failed to get audit session: %s", err)
		return nil, nil
	}

	queries, err := cx1client.GetQueries()
	if err != nil {
		logger.Errorf("Failed to get query collection from CheckmarxOne: %s", err)
		return nil, nil
	}

	var paQueries []Cx1ClientGo.Query

	// sometimes this fails with a 404 for some reason
	// quick-and-dirty retry

	maxRetry := 3
	retryDelay := 30
	for i := 0; i < maxRetry; i++ {
		if t.Scope.Corp {
			paQueries, err = cx1client.GetAuditQueriesByLevelID(auditSession, scopeStr, scope)
		} else {
			paQueries, err = cx1client.GetAuditQueriesByLevelID(auditSession, cx1client.QueryTypeProject(), t.Scope.ProjectID)
		}
		if err == nil {
			break
		} else {
			logger.Warnf("Attempt %d/%d to get %v-level queries failed with error '%s', waiting %d sec to retry...", i, maxRetry, scopeStr, err, retryDelay)
			if err = cx1client.AuditSessionKeepAlive(auditSession); err != nil {
				logger.Errorf("%v has expired, generating a new session", auditSession.String())
				auditSession = nil
				auditSession, err = getAuditSession(cx1client, logger, t)
				if err != nil {
					logger.Errorf("Failed to get audit session: %s", err)
					return nil, nil
				}
			}
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}
	if err != nil {
		logger.Errorf("Failed to get %v-level queries for project %v: %s", scopeStr, t.ScopeID, err)
	}

	queries.AddQueries(&paQueries)

	var query *Cx1ClientGo.Query
	logger.Debugf("Trying to find query on scope %v: %v -> %v -> %v", scopeStr, t.QueryLanguage, t.QueryGroup, t.QueryName)
	query = queries.GetQueryByLevelAndName(scopeStr, scope, t.QueryLanguage, t.QueryGroup, t.QueryName)

	if query != nil {
		logger.Debugf("Found query: %v", query.StringDetailed())
	} else {
		logger.Debugf("Query doesn't exist")
	}

	baseQuery := queries.GetQueryByName(t.QueryLanguage, t.QueryGroup, t.QueryName) // TODO: this needs better logic - what if base == project level?

	return query, baseQuery
}

func getQuery_old(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.Query, *Cx1ClientGo.Query) {
	scope, scopeStr, err := getQueryScope(cx1client, t)
	if err != nil {
		logger.Errorf("Error with query scope: %v", err)
		return nil, nil
	}

	t.ScopeID = scope

	var queries []Cx1ClientGo.AuditQuery_v310

	if t.Scope.Corp {
		scopeStr = cx1client.QueryTypeTenant()
		queries, err = cx1client.GetQueriesByLevelID_v310(scopeStr, scope)
	} else {
		queries, err = cx1client.GetQueriesByLevelID_v310(scopeStr, t.Scope.ProjectID)
	}

	if err != nil {
		logger.Errorf("Failed to get queries: %s", err)
		return nil, nil
	}

	var newQuery, baseQuery *Cx1ClientGo.Query

	auditQuery, err := cx1client.FindQueryByName_v310(queries, scopeStr, t.QueryLanguage, t.QueryGroup, t.QueryName)

	if err != nil {
		logger.Warnf("Error getting %v-level query %v: %s", scopeStr, t.String(), err)
	} else {
		query := auditQuery.ToQuery()
		newQuery = &query
	}

	bq, err := cx1client.FindQueryByName_v310(queries, Cx1ClientGo.AUDIT_QUERY_PRODUCT, t.QueryLanguage, t.QueryGroup, t.QueryName)
	if err != nil {
		logger.Warnf("Error getting product-level query %v: %s", t.String(), err)
	} else {
		query := bq.ToQuery()
		baseQuery = &query
	}

	return newQuery, baseQuery
}

func updateQuery(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		return nil
	}

	t.Query.Severity = t.Severity

	if t.Source != "" {
		t.Query.Source = t.Source
	}

	t.Query.IsExecutable = t.IsExecutable

	_, err = cx1client.UpdateQuery(auditSession, t.Query)
	return err
}

func updateQuery_old(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) error {
	t.Query.Severity = t.Severity

	if t.Source != "" {
		t.Query.Source = t.Source
	}

	t.Query.IsExecutable = t.IsExecutable

	query := t.Query.ToAuditQuery_v310()
	return cx1client.UpdateQuery_v310(query)
}

func (t *CxQLCRUD) TerminateSession(session_source string, cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) {
	// a session can be created by the automatic Read operation inserted prior to an Update or Delete operation.
	// in that case, the session would be created & deleted during the RunRead part, and no longer exist when the Update/Delete executes
	// so we only want to terminate the session if it was created during the same operation as the test
	if t.DeleteSession && t.CRUDTest.IsType(session_source) {
		auditSession, err := ASM.GetSession(t.Scope, t.QueryLanguage, t.LastScan, cx1client, logger)
		if err != nil {
			logger.Errorf("Failed to get audit session: %s", err)
			return
		}

		if auditSession != nil {
			err = ASM.DeleteSession(auditSession, cx1client, logger)
			if err != nil {
				logger.Errorf("Failed to delete Audit session %v: %s", auditSession.ID, err)
			}
		}
	}
}

func create(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		logger.Errorf("Failed to get audit session: %s", err)
		return nil
	}

	var baseQuery *Cx1ClientGo.Query
	t.Query, baseQuery = getQuery(cx1client, logger, t)

	if t.Query != nil {
		logger.Debugf("Query already exists in target scope: %v", t.Query.StringDetailed())
		return updateQuery(cx1client, logger, t)
	} else if baseQuery != nil {
		logger.Debugf("Found base query: %v", baseQuery.String())

		if t.Scope.Corp {
			newq, err := cx1client.CreateQueryOverride(auditSession, cx1client.QueryTypeTenant(), baseQuery)
			if err != nil {
				return fmt.Errorf("failed to create tenant override of %v: %s", baseQuery.StringDetailed(), err)
			}
			t.Query = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq, err := cx1client.CreateQueryOverride(auditSession, cx1client.QueryTypeApplication(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.Query = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq, err := cx1client.CreateQueryOverride(auditSession, cx1client.QueryTypeProject(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.Query = &newq
			}
		}

		logger.Debugf("Updating query %v", t.Query.String())
		return updateQuery(cx1client, logger, t)
	} else {
		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		newQuery := Cx1ClientGo.Query{
			Level:        cx1client.QueryTypeTenant(),
			LevelID:      cx1client.QueryTypeTenant(),
			Source:       t.Source,
			Name:         t.QueryName,
			Group:        t.QueryGroup,
			Language:     t.QueryLanguage,
			Severity:     t.Severity,
			IsExecutable: t.IsExecutable,
			Custom:       true,
		}

		newQuery, _, err := cx1client.CreateNewQuery(auditSession, newQuery)
		if err != nil {
			return err
		}

		t.Query = &newQuery

		return nil
	}
}

func create_old(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	var err error
	var baseQuery *Cx1ClientGo.Query

	t.Query, baseQuery = getQuery_old(cx1client, logger, t)

	if t.Query != nil {
		logger.Debugf("Updating query %v", t.Query.String())
		err = updateQuery_old(cx1client, t)
		return err
	} else {
		// query does not exist at all so needs to be created

		if baseQuery == nil {
			if !t.Scope.Corp {
				return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
			} else {
				return fmt.Errorf("creating a new Tenant-level query is no longer possible with the old API")
			}
		} else {
			logger.Debugf("Found base query: %v", baseQuery.String())

			if t.Scope.Corp {
				logger.Debugf("Will create corp override of %v", baseQuery.String())
				newq := baseQuery.ToAuditQuery_v310().CreateTenantOverride().ToQuery()
				t.Query = &newq
			} else {
				if t.Scope.Application != "" {
					logger.Debugf("Will create application override on %v", t.Scope.Application)
					newq := baseQuery.ToAuditQuery_v310().CreateApplicationOverrideByID(t.ScopeID).ToQuery()
					t.Query = &newq
				} else {
					logger.Debugf("Will create project override on %v", t.Scope.Project)
					newq := baseQuery.ToAuditQuery_v310().CreateProjectOverrideByID(t.ScopeID).ToQuery()
					t.Query = &newq
				}
			}
			err = updateQuery_old(cx1client, t)
			return err
		}
	}
}

func (t *CxQLCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.OldAPI {
		return create_old(cx1client, logger, t)
	} else {
		defer t.TerminateSession(OP_CREATE, cx1client, logger)
		return create(cx1client, logger, t)
	}
}

func (t *CxQLCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	var query *Cx1ClientGo.Query
	if t.OldAPI {
		query, _ = getQuery_old(cx1client, logger, t)
	} else {
		query, _ = getQuery(cx1client, logger, t)
	}

	defer t.TerminateSession(OP_READ, cx1client, logger)

	if query == nil {
		return fmt.Errorf("no such query %v: %v -> %v -> %v exists", t.Scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	if t.Scope.Corp {
		if query.Level != cx1client.QueryTypeTenant() {
			return fmt.Errorf("no Corp-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	} else if t.Scope.Application != "" {
		if query.Level != cx1client.QueryTypeApplication() {
			return fmt.Errorf("no Application-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	} else if t.Scope.Project != "" {
		if query.Level != cx1client.QueryTypeProject() {
			return fmt.Errorf("no Project-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	}

	t.Query = query

	return nil
}

func (t *CxQLCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Query == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	if t.OldAPI {
		return updateQuery_old(cx1client, t)
	} else {
		defer t.TerminateSession(OP_UPDATE, cx1client, logger)
		err := updateQuery(cx1client, logger, t)
		return err
	}

}

func (t *CxQLCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Query == nil {
		if t.CRUDTest.IsType(OP_READ) { // already tried to read
			return fmt.Errorf("read operation failed")
		} else {
			if err := t.RunRead(cx1client, logger, Engines); err != nil {
				return fmt.Errorf("read operation failed: %s", err)
			}
		}
	}

	if t.OldAPI {
		return cx1client.DeleteQuery_v310(t.Query.ToAuditQuery_v310())
	}

	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		return nil
	}
	defer t.TerminateSession(OP_DELETE, cx1client, logger)

	if t.Query.EditorKey == "" {
		logger.Warnf("Editor key for query %v is empty - attempting to calculate", t.Query.StringDetailed())
		t.Query.CalculateEditorKey()
	}

	return cx1client.DeleteQueryOverrideByKey(auditSession, t.Query.EditorKey)
}

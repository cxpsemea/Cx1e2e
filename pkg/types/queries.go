package types

import (
	"fmt"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

var auditSession *Cx1ClientGo.AuditSession

func (t *CxQLCRUD) Validate(CRUD string) error {
	if (CRUD == OP_UPDATE || CRUD == OP_DELETE) && t.Query == nil {
		return fmt.Errorf("must read before updating or deleting")
	}

	if t.QueryLanguage == "" || t.QueryGroup == "" || t.QueryName == "" {
		return fmt.Errorf("query language, group, or name is missing")
	}

	if t.Scope.Project == "" {
		return fmt.Errorf("project name is missing")
	}

	return nil
}

func (t *CxQLCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, CRUD string, Engines *EnabledEngines) error {
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

func getAuditSession(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *CxQLCRUD) (Cx1ClientGo.AuditSession, error) {
	var session Cx1ClientGo.AuditSession

	if auditSession != nil {
		if (t.Scope.Corp || auditSession.ProjectID == t.ScopeID) && auditSession.HasLanguage(t.QueryLanguage) {
			err := cx1client.AuditSessionKeepAlive(auditSession)
			if err != nil {
				auditSession = nil
				logger.Warningf("Tried to reuse existing audit session but it couldn't be refreshed")
			} else {
				logger.Warningf("Reusing existing audit session %v", auditSession.ID)
				return *auditSession, nil
			}
		} else {
			logger.Warningf("Existing audit session is not suitable (corp? %v, has %v? %v, is project id %v? %v)", t.Scope.Corp, t.QueryLanguage, auditSession.HasLanguage(t.QueryLanguage), t.ScopeID, auditSession.ProjectID)
		}
	}

	if t.LastScan == nil {
		proj, err := cx1client.GetProjectByName(t.Scope.Project)
		if err != nil {
			return session, err
		}

		lastscans, err := cx1client.GetLastScansByStatusAndID(proj.ProjectID, 1, []string{"Completed"})
		if err != nil {
			return session, fmt.Errorf("error getting last successful scan for project %v: %s", proj.ProjectID, err)
		}

		if len(lastscans) == 0 {
			return session, fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", proj.ProjectID)
		}

		t.LastScan = &lastscans[0]
	}

	session, err := cx1client.GetAuditSessionByID("sast", t.LastScan.ProjectID, t.LastScan.ScanID)
	if err == nil {
		auditSession = &session
	}

	return session, err
}

/*
there is no more old audit session

func getAuditSession_old(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) (string, error) {
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

	return cx1client.GetAuditSessionByID_v310(t.LastScan.ProjectID, t.LastScan.ScanID, true)
}
*/

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

func getQuery(cx1client *Cx1ClientGo.Cx1Client, session *Cx1ClientGo.AuditSession, logger *logrus.Logger, t *CxQLCRUD) (*Cx1ClientGo.Query, *Cx1ClientGo.Query) {
	scope, err := getQueryScope(cx1client, t)
	if err != nil {
		logger.Errorf("Error with query scope: %v", err)
		return nil, nil
	}

	t.ScopeID = scope

	queries, err := cx1client.GetQueries()
	if err != nil {
		logger.Errorf("Failed to get query collection from CheckmarxOne: %s", err)
		return nil, nil
	}

	var paQueries []Cx1ClientGo.Query
	if t.Scope.Corp {
		paQueries, err = cx1client.GetAuditQueriesByLevelID(session, Cx1ClientGo.AUDIT_QUERY_TENANT, Cx1ClientGo.AUDIT_QUERY_TENANT)
	} else {
		paQueries, err = cx1client.GetAuditQueriesByLevelID(session, Cx1ClientGo.AUDIT_QUERY_PROJECT, t.ScopeID)
	}
	if err != nil {
		logger.Errorf("Failed to get project-level queries for project %v: %s", t.ScopeID, err)
	}
	queries.AddQueries(&paQueries)

	var query *Cx1ClientGo.Query
	if t.Scope.Corp {
		logger.Debugf("Trying to find corp query on scope %v: %v -> %v -> %v", Cx1ClientGo.AUDIT_QUERY_TENANT, t.QueryLanguage, t.QueryGroup, t.QueryName)
		query = queries.GetQueryByLevelAndName(Cx1ClientGo.AUDIT_QUERY_TENANT, Cx1ClientGo.AUDIT_QUERY_TENANT, t.QueryLanguage, t.QueryGroup, t.QueryName)
	} else if t.Scope.Application != "" {
		logger.Debugf("Trying to find application query on scope %v: %v -> %v -> %v", t.ScopeID, t.QueryLanguage, t.QueryGroup, t.QueryName)
		query = queries.GetQueryByLevelAndName(Cx1ClientGo.AUDIT_QUERY_APPLICATION, t.ScopeID, t.QueryLanguage, t.QueryGroup, t.QueryName)
	} else {
		logger.Debugf("Trying to find project query on scope %v: %v -> %v -> %v", t.ScopeID, t.QueryLanguage, t.QueryGroup, t.QueryName)
		query = queries.GetQueryByLevelAndName(Cx1ClientGo.AUDIT_QUERY_PROJECT, t.ScopeID, t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	if query != nil {
		logger.Debugf("Found query: %v", query.StringDetailed())
	} else {
		logger.Debugf("Query doesn't exist")
	}

	baseQuery := queries.GetQueryByName(t.QueryLanguage, t.QueryGroup, t.QueryName)

	return query, baseQuery
}

func getQuery_old(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *CxQLCRUD) *Cx1ClientGo.Query {
	scope, err := getQueryScope(cx1client, t)
	if err != nil {
		logger.Errorf("Error with query scope: %v", err)
		return nil
	}

	t.ScopeID = scope

	scopeStr := ""
	if t.Scope.Corp {
		scopeStr = Cx1ClientGo.AUDIT_QUERY_TENANT
	} else if t.Scope.Application != "" {
		scopeStr = Cx1ClientGo.AUDIT_QUERY_APPLICATION
	} else {
		scopeStr = Cx1ClientGo.AUDIT_QUERY_PROJECT
	}

	auditQuery, err := cx1client.GetQueryByName_v310(scopeStr, scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	if err != nil {
		logger.Warnf("Error getting query %v: %s", t.String(), err)
		return nil
	}

	query := auditQuery.ToQuery()

	return &query
}

func updateQuery(cx1client *Cx1ClientGo.Cx1Client, session *Cx1ClientGo.AuditSession, t *CxQLCRUD) error {
	t.Query.Severity = t.Severity

	if t.Source != "" {
		t.Query.Source = t.Source
	}

	t.Query.IsExecutable = t.IsExecutable

	return cx1client.UpdateQuery(session, t.Query)
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

func (t *CxQLCRUD) TerminateSession(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, session *Cx1ClientGo.AuditSession) {
	if t.DeleteSession {
		err := cx1client.AuditDeleteSession(session)
		if err != nil {
			logger.Errorf("Failed to delete Audit session %v: %s", session.ID, err)
		}
	}
}

func create(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *CxQLCRUD) error {
	var session Cx1ClientGo.AuditSession
	var err error

	if t.Compile {
		session, err = getAuditSession(cx1client, logger, t)
		if err != nil {
			return err
		}
	}
	defer t.TerminateSession(cx1client, logger, &session)

	var baseQuery *Cx1ClientGo.Query
	t.Query, baseQuery = getQuery(cx1client, &session, logger, t)

	if t.Query != nil {
		logger.Debugf("Query already exists in target scope: %v", t.Query.StringDetailed())
		return updateQuery(cx1client, &session, t)
	} else if baseQuery != nil {
		logger.Debugf("Found base query: %v", baseQuery.String())

		if t.Scope.Corp {
			newq, err := cx1client.CreateQueryOverride(&session, Cx1ClientGo.AUDIT_QUERY_TENANT, baseQuery)
			if err != nil {
				return fmt.Errorf("failed to create tenant override of %v: %s", baseQuery.StringDetailed(), err)
			}
			t.Query = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq, err := cx1client.CreateQueryOverride(&session, Cx1ClientGo.AUDIT_QUERY_APPLICATION, baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.Query = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq, err := cx1client.CreateQueryOverride(&session, Cx1ClientGo.AUDIT_QUERY_PROJECT, baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.Query = &newq
			}
		}

		logger.Debugf("Updating query %v", t.Query.String())
		return updateQuery(cx1client, &session, t)
	} else {
		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		newQuery := Cx1ClientGo.Query{
			Level:        Cx1ClientGo.AUDIT_QUERY_TENANT,
			LevelID:      Cx1ClientGo.AUDIT_QUERY_TENANT,
			Source:       t.Source,
			Name:         t.QueryName,
			Group:        t.QueryGroup,
			Language:     t.QueryLanguage,
			Severity:     t.Severity,
			IsExecutable: t.IsExecutable,
			Custom:       true,
		}

		newQuery, err := cx1client.CreateNewQuery(&session, newQuery)
		if err != nil {
			return err
		}

		t.Query = &newQuery

		return nil
	}
}

func create_old(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, t *CxQLCRUD) error {
	var err error

	t.Query = getQuery_old(cx1client, logger, t)

	if t.Query != nil {
		logger.Debugf("Found query: %v", t.Query.String())

		if t.Scope.Corp {
			//logger.Info("Will create corp override")
			newq := t.Query.ToAuditQuery_v310().CreateTenantOverride().ToQuery()
			t.Query = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq := t.Query.ToAuditQuery_v310().CreateApplicationOverrideByID(t.ScopeID).ToQuery()
				t.Query = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq := t.Query.ToAuditQuery_v310().CreateProjectOverrideByID(t.ScopeID).ToQuery()
				t.Query = &newq
			}
		}

		logger.Debugf("Updating query %v", t.Query.String())
		err = updateQuery_old(cx1client, t)
		return err
	} else {
		// query does not exist at all so needs to be created on corp level
		// Second query: create new corp/tenant query

		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		return fmt.Errorf("creating a new Tenant-level query is no longer possible with the old API")
	}
}

func (t *CxQLCRUD) RunCreate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.OldAPI {
		return create_old(cx1client, logger, t)
	} else {
		return create(cx1client, logger, t)
	}
}

func (t *CxQLCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	var query *Cx1ClientGo.Query
	if t.OldAPI {
		query = getQuery_old(cx1client, logger, t)
	} else {
		session, err := getAuditSession(cx1client, logger, t)
		if err != nil {
			return err
		}

		query, _ = getQuery(cx1client, &session, logger, t)

		t.TerminateSession(cx1client, logger, &session)
	}

	if query == nil {
		return fmt.Errorf("no such query %v: %v -> %v -> %v exists", t.Scope, t.QueryLanguage, t.QueryGroup, t.QueryName)
	}

	if t.Scope.Corp {
		if query.Level != Cx1ClientGo.AUDIT_QUERY_TENANT {
			return fmt.Errorf("no Corp-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	} else if t.Scope.Application != "" {
		if query.Level != Cx1ClientGo.AUDIT_QUERY_APPLICATION {
			return fmt.Errorf("no Application-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	} else if t.Scope.Project != "" {
		if query.Level != Cx1ClientGo.AUDIT_QUERY_PROJECT {
			return fmt.Errorf("no Project-level query override for %v -> %v -> %v exists", t.QueryLanguage, t.QueryGroup, t.QueryName)
		}
	}

	t.Query = query

	return nil
}

func (t *CxQLCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.OldAPI {
		return updateQuery_old(cx1client, t)
	} else {
		session, err := getAuditSession(cx1client, logger, t)
		if err != nil {
			return err
		}
		defer t.TerminateSession(cx1client, logger, &session)
		err = updateQuery(cx1client, &session, t)
		return err
	}

}

func (t *CxQLCRUD) RunDelete(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Engines *EnabledEngines) error {
	if t.OldAPI {
		return cx1client.DeleteQuery_v310(t.Query.ToAuditQuery_v310())
	}

	session, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		return err
	}
	defer t.TerminateSession(cx1client, logger, &session)

	return cx1client.DeleteQueryOverrideByKey(&session, t.Query.EditorKey)
}

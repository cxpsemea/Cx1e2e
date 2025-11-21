package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
)

// var auditSession *Cx1ClientGo.AuditSession
var ASM *AuditSessionManager

func init() {
	ASM = NewAuditSessionManager()
}

func (t *CxQLCRUD) Validate(CRUD string) error {
	if t.QueryGroup == "" || t.QueryName == "" || (t.QueryLanguage == "" && t.QueryPlatform == "") {
		return fmt.Errorf("query language|platform, group, or name is missing")
	}
	if t.Engine == "" {
		return fmt.Errorf("engine is missing")
	}
	t.Engine = strings.ToLower(t.Engine)
	if t.Engine != "sast" && t.Engine != "iac" {
		return fmt.Errorf("engine must be 'sast' or 'iac'")
	}

	if t.Scope.Project == "" {
		return fmt.Errorf("project name is missing")
	}

	return nil
}

func (t *CxQLCRUD) IsSupported(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, CRUD string, Engines *EnabledEngines) error {
	if _, ok := cx1client.IsEngineAllowed(t.Engine); !ok {
		return fmt.Errorf("test attempts to access queries for engine %v but this is not supported in the license and will be skipped", t.Engine)
	}
	if !Engines.IsEnabled(t.Engine) {
		return fmt.Errorf("test attempts to access queries for engine %v but this was disabled for this test execution", t.Engine)
	}
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

		scanFilter := Cx1ClientGo.ScanFilter{
			Statuses:  []string{"Completed"},
			ProjectID: proj.ProjectID,
		}

		engine := t.Engine
		if engine == "iac" {
			engine = "kics"
		}
		lastscans, err := cx1client.GetLastScansByEngineFiltered(engine, 1, scanFilter)
		if err != nil {
			return nil, fmt.Errorf("error getting last successful scan for project %v: %s", proj.ProjectID, err)
		}

		if len(lastscans) == 0 {
			return nil, fmt.Errorf("unable to create audit session: no Completed scans exist for project %v", proj.ProjectID)
		}

		t.LastScan = &lastscans[0]
	}

	return ASM.GetOrCreateSession(t.ActiveThread, t.Scope, t.Engine, t.QueryPlatform, t.QueryLanguage, t.LastScan, cx1client, logger)

}

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

func getSASTQuery(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.SASTQuery, *Cx1ClientGo.SASTQuery) {
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

	queries, err := cx1client.GetSASTQueryCollection()
	if err != nil {
		logger.Errorf("Failed to get query collection from CheckmarxOne: %s", err)
		return nil, nil
	}

	var paQueries Cx1ClientGo.SASTQueryCollection

	// sometimes this fails with a 404 for some reason
	// quick-and-dirty retry

	maxRetry := 3
	retryDelay := 30
	for i := 0; i < maxRetry; i++ {
		if t.Scope.Corp {
			paQueries, err = cx1client.GetAuditSASTQueriesByLevelID(auditSession, scopeStr, scope)
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

	queries.AddCollection(&paQueries)

	var query *Cx1ClientGo.SASTQuery
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

func getIACQuery(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.IACQuery, *Cx1ClientGo.IACQuery) {
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

	queries, err := cx1client.GetIACQueryCollection()
	if err != nil {
		logger.Errorf("Failed to get query collection from CheckmarxOne: %s", err)
		return nil, nil
	}

	var paQueries Cx1ClientGo.IACQueryCollection

	// sometimes this fails with a 404 for some reason
	// quick-and-dirty retry

	maxRetry := 3
	retryDelay := 30
	for i := 0; i < maxRetry; i++ {
		if t.Scope.Corp {
			paQueries, err = cx1client.GetAuditIACQueriesByLevelID(auditSession, scopeStr, scope)
		} else {
			paQueries, err = cx1client.GetAuditIACQueriesByLevelID(auditSession, cx1client.QueryTypeProject(), t.Scope.ProjectID)
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

	queries.AddCollection(&paQueries)

	var query *Cx1ClientGo.IACQuery
	logger.Debugf("Trying to find query on scope %v: %v -> %v -> %v", scopeStr, t.QueryPlatform, t.QueryGroup, t.QueryName)
	query = queries.GetQueryByLevelAndName(scopeStr, scope, t.QueryPlatform, t.QueryGroup, t.QueryName)

	if query != nil {
		logger.Debugf("Found query: %v", query.StringDetailed())
	} else {
		logger.Debugf("Query doesn't exist")
	}

	baseQuery := queries.GetQueryByName(t.QueryPlatform, t.QueryGroup, t.QueryName) // TODO: this needs better logic - what if base == project level?

	return query, baseQuery
}

func getQuery_old(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) (*Cx1ClientGo.SASTQuery, *Cx1ClientGo.SASTQuery) {
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

	var newQuery, baseQuery *Cx1ClientGo.SASTQuery

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

	if t.Engine == "sast" {
		var new_query Cx1ClientGo.SASTQuery
		meta := t.SASTQuery.GetMetadata()

		if t.Severity != "" {
			meta.Severity = t.Severity
		}

		if t.IsExecutable != nil {
			if *t.IsExecutable != meta.IsExecutable {
				logger.Warnf("Attempting to change IsExecutable from %v to %v for query %v", meta.IsExecutable, *t.IsExecutable, t.SASTQuery.StringDetailed())
				meta.IsExecutable = *t.IsExecutable
			}
		}
		if t.CWE != "" {
			cweID, err := strconv.ParseInt(t.CWE, 10, 64)
			if err != nil {
				logger.Errorf("CWE %v is not a valid int64 value: %v", t.CWE, err)
			} else {
				meta.Cwe = cweID
			}
		}

		if t.DescriptionID != 0 {
			meta.CxDescriptionID = t.DescriptionID
		}

		if t.SASTQuery.MetadataDifferent(meta) {
			new_query, err = cx1client.UpdateSASTQueryMetadata(auditSession, *t.SASTQuery, meta)
			if err != nil {
				if strings.Contains(err.Error(), "query not found") {
					logger.Errorf("Failed to update metadata for query: %v. Will pause and retry.", err)
					time.Sleep(15 * time.Second)
					new_query, err = cx1client.UpdateSASTQueryMetadata(auditSession, *t.SASTQuery, meta)
				}
				if err != nil {
					return err
				}
			}
		}

		if t.Source != "" {
			t.SASTQuery.Source = t.Source
			new_query, _, err = cx1client.UpdateSASTQuerySource(auditSession, *t.SASTQuery, t.Source)
			if err != nil {
				return err
			}
		}

		t.SASTQuery = &new_query
	} else if t.Engine == "iac" {
		var new_query Cx1ClientGo.IACQuery
		if t.Severity != t.IACQuery.Severity {
			meta := t.IACQuery.GetMetadata()
			meta.Severity = t.Severity
			new_query, err = cx1client.UpdateIACQueryMetadata(auditSession, *t.IACQuery, meta)
			if err != nil {
				return err
			}
		}

		if t.Source != "" {
			new_query, _, err = cx1client.UpdateIACQuerySource(auditSession, *t.IACQuery, t.Source)
		}
		t.IACQuery = &new_query
	}

	return err
}

func updateQuery_old(cx1client *Cx1ClientGo.Cx1Client, t *CxQLCRUD) error {
	t.SASTQuery.Severity = t.Severity

	if t.Source != "" {
		t.SASTQuery.Source = t.Source
	}

	if t.IsExecutable != nil {
		t.SASTQuery.IsExecutable = *t.IsExecutable
	}

	query := t.SASTQuery.ToAuditQuery_v310()
	return cx1client.UpdateQuery_v310(query)
}

func (t *CxQLCRUD) TerminateSession(session_source string, cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) {
	// a session can be created by the automatic Read operation inserted prior to an Update or Delete operation.
	// in that case, the session would be created & deleted during the RunRead part, and no longer exist when the Update/Delete executes
	// so we only want to terminate the session if it was created during the same operation as the test
	if t.DeleteSession && t.CRUDTest.IsType(session_source) && t.LastScan != nil {
		auditSession, err := ASM.GetSession(t.ActiveThread, t.Scope, t.Engine, t.QueryPlatform, t.QueryLanguage, t.LastScan, cx1client, logger)
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

func createSAST(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		logger.Errorf("Failed to get audit session: %s", err)
		return err
	}

	var baseQuery *Cx1ClientGo.SASTQuery
	t.SASTQuery, baseQuery = getSASTQuery(cx1client, logger, t)

	if t.SASTQuery != nil {
		logger.Debugf("Query already exists in target scope: %v", t.SASTQuery.StringDetailed())
		return updateQuery(cx1client, logger, t)
	} else if baseQuery != nil {
		logger.Debugf("Found base query: %v", baseQuery.String())

		if t.Scope.Corp {
			newq, err := cx1client.CreateSASTQueryOverride(auditSession, cx1client.QueryTypeTenant(), baseQuery)
			if err != nil {
				return fmt.Errorf("failed to create tenant override of %v: %s", baseQuery.StringDetailed(), err)
			}
			t.SASTQuery = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq, err := cx1client.CreateSASTQueryOverride(auditSession, cx1client.QueryTypeApplication(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.SASTQuery = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq, err := cx1client.CreateSASTQueryOverride(auditSession, cx1client.QueryTypeProject(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create project override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.SASTQuery = &newq
			}
		}

		logger.Debugf("Updating query %v", t.SASTQuery.String())
		return updateQuery(cx1client, logger, t)
	} else {
		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		if t.IsExecutable == nil {
			return fmt.Errorf("cannot create a new corp query without specifying if it is executable")
		}

		newQuery := Cx1ClientGo.SASTQuery{
			Level:        cx1client.QueryTypeTenant(),
			LevelID:      cx1client.QueryTypeTenant(),
			Source:       t.Source,
			Name:         t.QueryName,
			Group:        t.QueryGroup,
			Language:     t.QueryLanguage,
			Severity:     t.Severity,
			IsExecutable: *t.IsExecutable,
			Custom:       true,
		}

		if t.CWE != "" {
			cweID, err := strconv.ParseInt(t.CWE, 10, 64)
			if err != nil {
				cweID = 0
			}
			newQuery.CweID = cweID
		}
		if t.DescriptionID != 0 {
			newQuery.QueryDescriptionId = t.DescriptionID
		}

		newQuery, _, err := cx1client.CreateNewSASTQuery(auditSession, newQuery)
		if err != nil {
			return err
		}

		t.SASTQuery = &newQuery

		return nil
	}
}

func createIAC(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	auditSession, err := getAuditSession(cx1client, logger, t)
	if err != nil {
		logger.Errorf("Failed to get audit session: %s", err)
		return err
	}

	var baseQuery *Cx1ClientGo.IACQuery
	t.IACQuery, baseQuery = getIACQuery(cx1client, logger, t)

	if t.IACQuery != nil {
		logger.Debugf("Query already exists in target scope: %v", t.IACQuery.StringDetailed())
		return updateQuery(cx1client, logger, t)
	} else if baseQuery != nil {
		logger.Debugf("Found base query: %v", baseQuery.String())

		if t.Scope.Corp {
			newq, err := cx1client.CreateIACQueryOverride(auditSession, cx1client.QueryTypeTenant(), baseQuery)
			if err != nil {
				return fmt.Errorf("failed to create tenant override of %v: %s", baseQuery.StringDetailed(), err)
			}
			t.IACQuery = &newq
		} else {
			if t.Scope.Application != "" {
				logger.Debugf("Will create application override on %v", t.Scope.Application)
				newq, err := cx1client.CreateIACQueryOverride(auditSession, cx1client.QueryTypeApplication(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create application override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.IACQuery = &newq
			} else {
				logger.Debugf("Will create project override on %v", t.Scope.Project)
				newq, err := cx1client.CreateIACQueryOverride(auditSession, cx1client.QueryTypeProject(), baseQuery)
				if err != nil {
					return fmt.Errorf("failed to create project override of %v: %s", baseQuery.StringDetailed(), err)
				}
				t.IACQuery = &newq
			}
		}

		logger.Debugf("Updating query %v", t.IACQuery.String())
		return updateQuery(cx1client, logger, t)
	} else {
		if !t.Scope.Corp {
			return fmt.Errorf("query %v does not exist and must be created at Tenant level before it can be created on a Project or Application level", t.String())
		}

		newQuery := Cx1ClientGo.IACQuery{
			Level:          cx1client.QueryTypeTenant(),
			LevelID:        cx1client.QueryTypeTenant(),
			Source:         t.Source,
			Name:           t.QueryName,
			Group:          t.QueryGroup,
			Platform:       t.QueryPlatform,
			Severity:       t.Severity,
			Custom:         true,
			Description:    t.Description,
			DescriptionURL: t.DescriptionURL,
			Category:       t.Category,
			CWE:            t.CWE,
		}

		newQuery, _, err := cx1client.CreateNewIACQuery(auditSession, newQuery)
		if err != nil {
			return err
		}

		t.IACQuery = &newQuery

		return nil
	}
}

func create_old(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, t *CxQLCRUD) error {
	var err error
	var baseQuery *Cx1ClientGo.SASTQuery

	t.SASTQuery, baseQuery = getQuery_old(cx1client, logger, t)

	if t.SASTQuery != nil {
		logger.Debugf("Updating query %v", t.SASTQuery.String())
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
				t.SASTQuery = &newq
			} else {
				if t.Scope.Application != "" {
					logger.Debugf("Will create application override on %v", t.Scope.Application)
					newq := baseQuery.ToAuditQuery_v310().CreateApplicationOverrideByID(t.ScopeID).ToQuery()
					t.SASTQuery = &newq
				} else {
					logger.Debugf("Will create project override on %v", t.Scope.Project)
					newq := baseQuery.ToAuditQuery_v310().CreateProjectOverrideByID(t.ScopeID).ToQuery()
					t.SASTQuery = &newq
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
		if t.Engine == "sast" {
			return createSAST(cx1client, logger, t)
		} else if t.Engine == "iac" {
			return createIAC(cx1client, logger, t)
		}
	}
	return fmt.Errorf("unknown engine")
}

func (t *CxQLCRUD) RunRead(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.Engine == "sast" {
		var query *Cx1ClientGo.SASTQuery
		if t.OldAPI {
			query, _ = getQuery_old(cx1client, logger, t)
		} else {
			query, _ = getSASTQuery(cx1client, logger, t)
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

		t.SASTQuery = query
	} else if t.Engine == "iac" {
		var query *Cx1ClientGo.IACQuery
		query, _ = getIACQuery(cx1client, logger, t)

		defer t.TerminateSession(OP_READ, cx1client, logger)

		if query == nil {
			return fmt.Errorf("no such query %v: %v -> %v -> %v exists", t.Scope, t.QueryPlatform, t.QueryGroup, t.QueryName)
		}

		if t.Scope.Corp {
			if query.Level != cx1client.QueryTypeTenant() {
				return fmt.Errorf("no Corp-level query override for %v -> %v -> %v exists", t.QueryPlatform, t.QueryGroup, t.QueryName)
			}
		} else if t.Scope.Application != "" {
			if query.Level != cx1client.QueryTypeApplication() {
				return fmt.Errorf("no Application-level query override for %v -> %v -> %v exists", t.QueryPlatform, t.QueryGroup, t.QueryName)
			}
		} else if t.Scope.Project != "" {
			if query.Level != cx1client.QueryTypeProject() {
				return fmt.Errorf("no Project-level query override for %v -> %v -> %v exists", t.QueryPlatform, t.QueryGroup, t.QueryName)
			}
		}

		t.IACQuery = query
	}

	return nil
}

func (t *CxQLCRUD) RunUpdate(cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger, Engines *EnabledEngines) error {
	if t.SASTQuery == nil {
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
	if t.Engine == "sast" {

		if t.SASTQuery == nil {
			if t.CRUDTest.IsType(OP_READ) { // already tried to read
				return fmt.Errorf("read operation failed")
			} else {
				if err := t.RunRead(cx1client, logger, Engines); err != nil {
					return fmt.Errorf("read operation failed: %s", err)
				}
			}
		}

		if t.OldAPI {
			return cx1client.DeleteQuery_v310(t.SASTQuery.ToAuditQuery_v310())
		}

		auditSession, err := getAuditSession(cx1client, logger, t)
		if err != nil {
			return nil
		}
		defer t.TerminateSession(OP_DELETE, cx1client, logger)

		if t.SASTQuery.EditorKey == "" {
			logger.Warnf("Editor key for query %v is empty - attempting to calculate", t.SASTQuery.StringDetailed())
			t.SASTQuery.CalculateEditorKey()
		}

		return cx1client.DeleteQueryOverrideByKey(auditSession, t.SASTQuery.EditorKey)
	} else if t.Engine == "iac" {
		if t.IACQuery == nil {
			if t.CRUDTest.IsType(OP_READ) { // already tried to read
				return fmt.Errorf("read operation failed")
			} else {
				if err := t.RunRead(cx1client, logger, Engines); err != nil {
					return fmt.Errorf("read operation failed: %s", err)
				}
			}
		}

		auditSession, err := getAuditSession(cx1client, logger, t)
		if err != nil {
			return nil
		}
		defer t.TerminateSession(OP_DELETE, cx1client, logger)
		return cx1client.DeleteQueryOverrideByKey(auditSession, t.IACQuery.QueryID)
	}
	return fmt.Errorf("unknown engine")
}

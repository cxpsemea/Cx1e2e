package types

import (
	"fmt"
	"sync"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

type AuditSessionManager struct {
	Lock     sync.Mutex
	Sessions []Cx1ClientGo.AuditSession
}

func NewAuditSessionManager() *AuditSessionManager {
	return &AuditSessionManager{
		Sessions: []Cx1ClientGo.AuditSession{},
	}
}

func (m *AuditSessionManager) GetOrCreateSession(scope CxQLScope, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) (*Cx1ClientGo.AuditSession, error) {
	session, err := m.GetSession(scope, language, lastScan, cx1client, logger)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	// no session found so we need to create a new one
	new_session, err := cx1client.GetAuditSessionByID("sast", lastScan.ProjectID, lastScan.ScanID)
	if err != nil {
		return nil, err
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.Sessions = append(m.Sessions, new_session)
	return &m.Sessions[len(m.Sessions)-1], nil
}

func (m *AuditSessionManager) GetSession(scope CxQLScope, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) (*Cx1ClientGo.AuditSession, error) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for id := range m.Sessions {
		if (m.Sessions[id].ProjectID == scope.ProjectID || scope.Corp) && m.Sessions[id].HasLanguage(language) {
			if err := cx1client.AuditSessionKeepAlive(&m.Sessions[id]); err != nil {
				logger.Warningf("Tried to reuse existing audit session but it couldn't be refreshed")
				break
			} else {
				scopeStr := "Tenant"
				if !scope.Corp {
					if scope.Application != "" {
						scopeStr = fmt.Sprintf("application %v", scope.Application)
					} else {
						scopeStr = fmt.Sprintf("project %v", scope.Project)
					}
				}
				logger.Warningf("Reusing existing %v (scope: %v, project ID: %v, language: %v)", m.Sessions[id].String(), scopeStr, scope.ProjectID, language)
				return &m.Sessions[id], nil
			}
		}
	}

	return nil, nil
}

func (m *AuditSessionManager) Clear(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for _, s := range m.Sessions {
		err := cx1client.AuditDeleteSession(&s)
		if err != nil {
			logger.Errorf("Failed to terminate audit session: %s", err)
		}
	}
	m.Sessions = []Cx1ClientGo.AuditSession{}
}

package types

import (
	"slices"
	"sync"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

const AuditSessionTimeoutMinutes = 5

type AuditSessionManager struct {
	Lock     sync.Mutex
	Sessions []*Cx1ClientGo.AuditSession
}

func NewAuditSessionManager() *AuditSessionManager {
	return &AuditSessionManager{
		Sessions: []*Cx1ClientGo.AuditSession{},
	}
}

func (m *AuditSessionManager) GetOrCreateSession(scope CxQLScope, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) (*Cx1ClientGo.AuditSession, error) {
	logger.Debugf("Audit Session Manager: get or create session for scope %v, language %v, project %v", scope.String(), language, lastScan.ProjectID)
	session, err := m.GetSession(scope, language, lastScan, cx1client, logger)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	// no session found so we need to create a new one
	logger.Debugf("Audit Session Manager: create session for scope %v, language %v, project %v", scope, language, lastScan.ProjectID)
	new_session, err := cx1client.GetAuditSessionByID("sast", lastScan.ProjectID, lastScan.ScanID)
	if err != nil {
		return nil, err
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.Sessions = append(m.Sessions, &new_session)
	logger.Debugf("Audit Session Manager: created session %v", new_session.String())
	return m.Sessions[len(m.Sessions)-1], nil
}

func (m *AuditSessionManager) GetSession(scope CxQLScope, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) (*Cx1ClientGo.AuditSession, error) {
	logger.Debugf("Audit Session Manager: get session for scope %v, language %v, project %v", scope.String(), language, lastScan.ProjectID)
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.PrintSessions(logger)

	for id := range m.Sessions {
		if (m.Sessions[id].ProjectID == scope.ProjectID || scope.Corp) && m.Sessions[id].HasLanguage(language) {
			if time.Since(m.Sessions[id].LastHearbeat) < AuditSessionTimeoutMinutes*time.Minute {
				if err := cx1client.AuditSessionKeepAlive(m.Sessions[id]); err != nil {
					logger.Warningf("Tried to refresh existing audit session %v but failed: %s", m.Sessions[id].String(), err)
					_ = cx1client.AuditDeleteSession(m.Sessions[id])
					m.Sessions = slices.Delete(m.Sessions, id, id+1)
					return nil, nil
				} else {
					logger.Warningf("Found existing audit session %v (scope: %v, language: %v)", m.Sessions[id].String(), scope.String(), language)
					return m.Sessions[id], nil
				}
			} else {
				logger.Warningf("Found existing audit session %v but it was created more than %d minutes ago (%v) and may have expired", m.Sessions[id].String(), AuditSessionTimeoutMinutes, m.Sessions[id].CreatedAt.String())
				_ = cx1client.AuditDeleteSession(m.Sessions[id])
				m.Sessions = slices.Delete(m.Sessions, id, id+1)
				return nil, nil
			}
		}
	}

	return nil, nil
}

func (m *AuditSessionManager) Clear(cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) {
	if len(m.Sessions) == 0 {
		return
	}
	logger.Debug("Audit Session Manager: clearing sessions")
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for _, s := range m.Sessions {
		err := cx1client.AuditDeleteSession(s)
		if err != nil {
			logger.Errorf("Failed to terminate audit session: %s", err)
		} else {
			logger.Warningf("Terminated left-over audit session %v (created: %v)", s.String(), s.CreatedAt.String())
		}
	}
	m.Sessions = []*Cx1ClientGo.AuditSession{}
}

func (m *AuditSessionManager) DeleteSession(session *Cx1ClientGo.AuditSession, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger) error {
	logger.Debugf("Audit Session Manager: delete session %v", session.String())

	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.PrintSessions(logger)

	theSession := *session

	for id := range m.Sessions {
		if m.Sessions[id].ID == session.ID {
			m.Sessions = slices.Delete(m.Sessions, id, id+1)
			break
		}
	}

	err := cx1client.AuditDeleteSession(&theSession)

	m.PrintSessions(logger)

	return err
}

func (m *AuditSessionManager) PrintSessions(logger *logrus.Logger) {
	logger.Tracef("Listing AuditSessionManager's %d active sessions", len(m.Sessions))
	for id, s := range m.Sessions {
		logger.Tracef(" - %d: %v", id, s.String())
	}
}

package types

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

const AuditSessionTimeoutMinutes = 5

type AuditSessionManager struct {
	Lock     sync.Mutex
	Sessions []AuditSessionWrapper
}

type AuditSessionWrapper struct {
	Thread  int
	Session *Cx1ClientGo.AuditSession
}

func (w *AuditSessionWrapper) String() string {
	return fmt.Sprintf("[T%d] %s", w.Thread, w.Session.String())
}

func (w *AuditSessionWrapper) ID() string {
	return w.Session.ID
}

func NewAuditSessionManager() *AuditSessionManager {
	return &AuditSessionManager{
		Sessions: []AuditSessionWrapper{},
	}
}

func (m *AuditSessionManager) GetOrCreateSession(thread int, scope CxQLScope, engine, platform, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) (*Cx1ClientGo.AuditSession, error) {
	logger.Debugf("Audit Session Manager: get or create session for engine %v, scope %v, platform %v, language %v, project %v", engine, scope.String(), platform, language, lastScan.ProjectID)
	session, err := m.GetSession(thread, scope, engine, platform, language, lastScan, cx1client, logger)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	// no session found so we need to create a new one
	logger.Debugf("Audit Session Manager: create session for engine %v, scope %v, platform %v, language %v, project %v", engine, scope, platform, language, lastScan.ProjectID)
	new_session, err := cx1client.GetAuditSessionByID(engine, lastScan.ProjectID, lastScan.ScanID)
	if err != nil {
		return nil, err
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.Sessions = append(m.Sessions, AuditSessionWrapper{Session: &new_session, Thread: thread})
	logger.Debugf("Audit Session Manager: created session %v", new_session.String())
	return &new_session, nil
}

func (m *AuditSessionManager) GetSession(thread int, scope CxQLScope, engine, platform, language string, lastScan *Cx1ClientGo.Scan, cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) (*Cx1ClientGo.AuditSession, error) {
	logger.Debugf("Audit Session Manager: get session for engine %v, scope %v, platform %v, language %v, project %v", engine, scope.String(), platform, language, lastScan.ProjectID)
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.PrintSessions(logger)

	for id := range m.Sessions {
		session := m.Sessions[id].Session
		logger.Debugf("Checking session: %v", session.String())
		if m.Sessions[id].Thread == thread && (session.ProjectID == scope.ProjectID || scope.Corp) && (session.HasLanguage(language) || session.HasPlatform(platform)) && session.Engine == engine {
			if time.Since(session.LastHeartbeat) < AuditSessionTimeoutMinutes*time.Minute {
				if err := cx1client.AuditSessionKeepAlive(session); err != nil {
					logger.Warnf("Tried to refresh existing audit session %v but failed: %s", session.String(), err)
					_ = cx1client.AuditDeleteSession(session)
					m.Sessions = slices.Delete(m.Sessions, id, id+1)
					return nil, nil
				} else {
					logger.Debugf("Found existing audit session %v", session.String())
					return session, nil
				}
			} else {
				logger.Warnf("Found existing audit session %v but it was created more than %d minutes ago (%v) and may have expired", session.String(), AuditSessionTimeoutMinutes, session.CreatedAt.String())
				_ = cx1client.AuditDeleteSession(session)
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
		err := cx1client.AuditDeleteSession(s.Session)
		if err != nil {
			logger.Errorf("Failed to terminate audit session: %s", err)
		} else {
			logger.Warningf("Terminated left-over audit session %v (created: %v)", s.String(), s.Session.CreatedAt.String())
		}
	}
	m.Sessions = []AuditSessionWrapper{}
}

func (m *AuditSessionManager) DeleteSession(session *Cx1ClientGo.AuditSession, cx1client *Cx1ClientGo.Cx1Client, logger *ThreadLogger) error {
	logger.Debugf("Audit Session Manager: delete session %v", session.String())

	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.PrintSessions(logger)

	theSession := *session

	for id := range m.Sessions {
		if m.Sessions[id].Session.ID == session.ID {
			m.Sessions = slices.Delete(m.Sessions, id, id+1)
			break
		}
	}

	err := cx1client.AuditDeleteSession(&theSession)

	m.PrintSessions(logger)

	return err
}

func (m *AuditSessionManager) PrintSessions(logger *ThreadLogger) {
	logger.Tracef("Listing AuditSessionManager's %d active sessions", len(m.Sessions))
	for id, s := range m.Sessions {
		logger.Tracef(" - %d: %v", id, s.String())
	}
}

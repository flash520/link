/**
 * @Author: koulei
 * @Description:
 * @File: manager
 * @Version: 1.0.0
 * @Date: 2023/3/4 18:18
 */

package link

import "sync"

const sessionMapNum = 32

type Manager struct {
	sessionMaps  [sessionMapNum]sessionMap
	disposedOnce sync.Once
	disposedWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
	disposed bool
}

// NewManager 创建管理器并初始化多片sessions
func NewManager() *Manager {
	manager := &Manager{}

	for i := 0; i < sessionMapNum; i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return manager
}

// NewSession 创建新session
func (manager *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := newSession(manager, codec, sendChanSize)
	manager.putSession(session)
	return session
}

// GetSession 根据sessionID返回session
func (manager *Manager) GetSession(sessionID uint64) *Session {
	sMap := &manager.sessionMaps[sessionID%sessionMapNum]

	sMap.RLock()
	defer sMap.RUnlock()

	return sMap.sessions[sessionID]
}

// Dispose 关闭全部session
func (manager *Manager) Dispose() {
	manager.disposedOnce.Do(func() {
		for i := 0; i < sessionMapNum; i++ {
			sMap := &manager.sessionMaps[i]
			sMap.Lock()
			sMap.disposed = true
			for _, session := range sMap.sessions {
				session.Close()
			}
			sMap.Unlock()
		}
		manager.disposedWait.Wait()
	})
}

// putSession 添加session
func (manager *Manager) putSession(session *Session) {
	sMap := &manager.sessionMaps[session.id%sessionMapNum]

	sMap.Lock()
	defer sMap.Unlock()

	if sMap.disposed {
		session.Close()
		return
	}

	sMap.sessions[session.id] = session
	manager.disposedWait.Add(1)
}

// delSession 删除session
func (manager *Manager) delSession(session *Session) {
	sMap := &manager.sessionMaps[session.id%sessionMapNum]

	sMap.Lock()
	defer sMap.Unlock()

	if _, ok := sMap.sessions[session.ID()]; ok {
		delete(sMap.sessions, session.id)
		manager.disposedWait.Done()
	}
}

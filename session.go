/**
 * @Author: koulei
 * @Description:
 * @File: session
 * @Version: 1.0.0
 * @Date: 2023/3/4 11:42
 */

package link

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	globalSessionID     uint64
	SessionBlockedError = errors.New("session blocked")
	SessionClosedError  = errors.New("session closed")
)

type Session struct {
	id        uint64
	manager   *Manager
	codec     Codec
	sendChan  chan interface{}
	recvMutex sync.Mutex
	sendMutex sync.RWMutex

	closeFlag  int32
	closeChan  chan int
	closeMutex sync.RWMutex
}

func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

// newSession 创建新session
func newSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	session := &Session{
		manager:   manager,
		codec:     codec,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionID, 1),
	}
	if sendChanSize > 0 {
		session.sendChan = make(chan interface{}, sendChanSize)
		go session.sendLoop()
	}
	return session
}

// ID 返回sessionID
func (session *Session) ID() uint64 {
	return session.id
}

// Codec 返回编解码器
func (session *Session) Codec() Codec {
	return session.codec
}

// Send 发送消息
func (session *Session) Send(msg interface{}) error {
	if session.sendChan == nil {
		if session.IsClosed() {
			return SessionClosedError
		}
		session.sendMutex.Lock()
		defer session.sendMutex.Unlock()
		if err := session.codec.Send(msg); err != nil {
			session.Close()
			return err
		}
		return nil
	}

	session.sendMutex.RLock()
	select {
	case session.sendChan <- msg:
		session.sendMutex.RUnlock()
		return nil
	default:
		session.sendMutex.RUnlock()
		return SessionBlockedError
	}
}

// Receive 接收消息
func (session *Session) Receive() (interface{}, error) {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	msg, err := session.codec.Receive()
	if err != nil {
		session.Close()
		return nil, err
	}
	return msg, nil
}

// sendLoop 循环发送消息
func (session *Session) sendLoop() {
	defer session.Close()
	for {
		select {
		case msg, ok := <-session.sendChan:
			if !ok || session.codec.Send(msg) != nil {
				return
			}
		case <-session.closeChan:
			return
		}
	}
}

// IsClosed 返回session是否关闭
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

// Close 关闭session
func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		close(session.closeChan)
		if session.sendChan != nil {
			session.closeMutex.Lock()
			close(session.sendChan)
			session.closeMutex.Unlock()
		}

		err := session.codec.Close()
		go func() {
			if session.manager != nil {
				session.manager.delSession(session)
			}
		}()
		return err
	}

	return SessionClosedError
}

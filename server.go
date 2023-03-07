/**
 * @Author: koulei
 * @Description:
 * @File: server
 * @Version: 1.0.0
 * @Date: 2023/3/5 11:11
 */

package link

import "net"

type Server struct {
	manager      *Manager
	listener     net.Listener
	protocol     Protocol
	handler      Handler
	sendChanSize int
}

type Handler interface {
	HandleSession(session *Session)
}

var _ = HandlerFunc(nil)

type HandlerFunc func(session *Session)

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

func NewServer(listener net.Listener, protocol Protocol, sendChanSize int, handler Handler) *Server {
	return &Server{
		manager:      NewManager(),
		listener:     listener,
		protocol:     protocol,
		handler:      handler,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Serve() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			codec, err := server.protocol.NewCodec(conn)
			if err != nil {
				conn.Close()
				return
			}
			session := server.manager.NewSession(codec, server.sendChanSize)
			server.handler.HandleSession(session)
		}()
	}
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}

func (server *Server) Stop() {
	server.listener.Close()
	server.manager.Dispose()
}

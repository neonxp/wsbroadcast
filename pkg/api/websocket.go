/*
Copyright © 2020 Alexander Kiryukhin <a.kiryukhin@mail.ru>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package api

import (
	"io"
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"

	"backend/pkg/model"
)

func (s *Server) websocketHandler(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		room := c.Param("room")
		rid := atomic.AddUint64(&s.lsn, 1)
		defer func() {
			s.mu.Lock()
			c := s.connections[room][rid]
			delete(s.connections[room], rid)
			e := model.Event{
				From: c.ID,
				Type: "disconnected",
			}
			for _, conns := range s.connections[room] {
				conns.msg <- e
			}
			s.mu.Unlock()
		}()
		s.mu.Lock()
		if _, ok := s.connections[room]; !ok {
			s.connections[room] = map[uint64]*connection{}
		}
		s.connections[room][rid] = &connection{
			ID:    rid,
			State: "",
			msg:   make(chan model.Event),
		}
		e := model.Event{
			From: rid,
			Type: "connected",
		}
		for _, conns := range s.connections[room] {
			conns.msg <- e
		}
		s.mu.Unlock()
		go func() {
			for e := range s.connections[room][rid].msg {
				if err := websocket.JSON.Send(ws, e); err != nil {
					if err == io.EOF {
						return
					}
					continue
				}
			}
		}()
		for {
			e := model.Event{}
			if err := websocket.JSON.Receive(ws, &e); err != nil {
				if err == io.EOF {
					return
				}
				continue
			}
			s.mu.Lock()
			e.From = rid
			if e.Type == "setState" {
				s.connections[room][rid].State = string(e.Data)
			}
			for id, conns := range s.connections[room] {
				if id != rid {
					conns.msg <- e
				}
			}
			s.mu.Unlock()
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

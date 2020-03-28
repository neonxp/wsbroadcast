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
	"encoding/json"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/neonxp/wsbroadcast/pkg/model"
)

func (s *Server) createChannel(c echo.Context) error {
	req := &struct {
		Payload json.RawMessage `json:"payload"`
	}{}
	if err := c.Bind(req); err != nil {
		return err
	}
	m := &model.Channel{
		Payload:   req.Payload,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}
	id, err := s.channelManager.Add(m)
	if err != nil {
		return err
	}
	var resp = struct {
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}{
		ID:      id.Hex(),
		Payload: req.Payload,
	}
	return c.JSON(201, resp)
}

func (s *Server) getChannel(c echo.Context) error {
	room := c.Param("channel")
	roomID, err := primitive.ObjectIDFromHex(room)
	if err != nil {
		return err
	}
	r := new(model.Channel)
	if err := s.channelManager.FindOne(bson.M{"_id": roomID}, r); err != nil {
		return err
	}
	connections := []*connection{}
	for _, cn := range s.connections[room] {
		connections = append(connections, cn)
	}
	var resp = struct {
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
		Members []*connection   `json:"members"`
	}{
		ID:      r.ID.Hex(),
		Payload: r.Payload,
		Members: connections,
	}
	return c.JSON(200, resp)
}

func (s *Server) updateChannel(c echo.Context) error {
	room := c.Param("channel")
	roomID, err := primitive.ObjectIDFromHex(room)
	if err != nil {
		return err
	}
	req := &struct {
		Payload json.RawMessage `json:"payload"`
	}{}
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := s.channelManager.Update(roomID, bson.M{"payload": req.Payload}); err != nil {
		return err
	}
	var resp = struct {
		ID      string          `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}{
		ID:      roomID.Hex(),
		Payload: req.Payload,
	}
	return c.JSON(200, resp)
}

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
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo/v4"

	"backend/pkg/config"
	"backend/pkg/db"
)

type Server struct {
	db             *mongo.Database
	cfg            *config.Config
	channelManager *db.Manager
	e              *echo.Echo
	connections    map[string]map[uint64]*connection
	mu             sync.Mutex
	lsn            uint64
}

func NewServer(mongo *mongo.Database, cfg *config.Config) (*Server, error) {
	roomManager, err := db.NewManager(mongo.Collection("channels"), nil)
	if err != nil {
		return nil, err
	}
	return &Server{
		db:             mongo,
		cfg:            cfg,
		channelManager: roomManager,
		e:              echo.New(),
		connections:    map[string]map[uint64]*connection{},
		mu:             sync.Mutex{},
		lsn:            0,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		if err := s.Close(); err != nil {
			log.Println(err)
		}
	}()
	s.e.Use(middleware.CORS())
	s.e.POST("/channel", s.createChannel)
	s.e.GET("/channel/:channel", s.getChannel)
	s.e.PUT("/channel/:channel", s.updateChannel)
	s.e.GET("/channel/:channel/ws", s.websocketHandler)
	if err := s.e.Start(s.cfg.Listen); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Close() error {
	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.e.Shutdown(stopCtx)
}

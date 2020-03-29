package main

import (
	"log"
	"time"

	"github.com/neguse/gomelift/pkg/proto/pbuffer"

	"github.com/neguse/gomelift/pkg/gamelift"
	glog "github.com/neguse/gomelift/pkg/log"
)

type Handler struct {
	c gamelift.Client
}

func (h *Handler) StartGameSession(event *pbuffer.ActivateGameSession) {
	log.Println("StartGameSession", event)

	if err := h.c.ActivateGameSession(&pbuffer.GameSessionActivate{
		GameSessionId: event.GetGameSession().GetGameSessionId(),
		MaxPlayers:    event.GetGameSession().GetMaxPlayers(),
		Port:          event.GetGameSession().GetPort(),
	}); err != nil {
		log.Panic(err)
	}
	log.Println("ActivateGameSession complete. sessionID:", *h.c.GetGameSessionId())
}

func (h *Handler) UpdateGameSession(event *pbuffer.UpdateGameSession) {
	log.Println("UpdateGameSession", event)
}

func (h *Handler) ProcessTerminate(event *pbuffer.TerminateProcess) {
	log.Println("ProcessTerminate", event)
}

func (h *Handler) HealthCheck() bool {
	log.Println("HealthCheck")
	return true
}

func main() {
	logger := &glog.StandardLogger{}
	c := gamelift.NewClient(logger)
	c.Handle(&Handler{c: c})
	err := c.Open()
	if err != nil {
		log.Panic(err)
	}
	if err := c.ProcessReady(&pbuffer.ProcessReady{
		LogPathsToUpload:          []string{},
		Port:                      7777,
		MaxConcurrentGameSessions: 2,
	}); err != nil {
		log.Panic(err)
	}
	res, err := c.GetInstanceCertificate(&pbuffer.GetInstanceCertificate{})
	log.Println(res, err)

	time.Sleep(time.Second * 120)

	if err := c.ProcessEnding(&pbuffer.ProcessEnding{}); err != nil {
		log.Panic(err)
	}
}

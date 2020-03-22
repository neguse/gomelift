package main

import (
	"log"
	"time"

	"github.com/neguse/gomelift/pkg/gamelift"
	"github.com/neguse/gomelift/pkg/proto/pbuffer"
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
	log.Println("ActivateGameSession complete")
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
	c := gamelift.NewClient()
	c.Handle(&Handler{c: c})
	c.Open()
	// TODO: 接続成功したことをハンドリングできるようにするべき
	time.Sleep(time.Second * 10000)
	if err := c.ProcessReady(&pbuffer.ProcessReady{
		LogPathsToUpload:          []string{},
		Port:                      7777,
		MaxConcurrentGameSessions: 2,
	}); err != nil {
		log.Panic(err)
	}

	log.Println("aaa")
	for {
		time.Sleep(time.Second * 60)
	}

	log.Println("bbb")
	if err := c.ProcessEnding(&pbuffer.ProcessEnding{}); err != nil {
		log.Panic(err)
	}
	log.Println("ccc")
}

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/neguse/gomelift/pkg/gamelift"
	glog "github.com/neguse/gomelift/pkg/log"
	"github.com/neguse/gomelift/pkg/proto/pbuffer"
)

var upgrader = websocket.Upgrader{} // use default options

type Handler struct {
	c    gamelift.Client
	port int
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

func (h *Handler) AcceptPlayerHandler(w http.ResponseWriter, r *http.Request) {
	psess := r.URL.Query().Get("psess")
	if err := h.c.AcceptPlayerSession(&pbuffer.AcceptPlayerSession{
		GameSessionId:   *h.c.GetGameSessionId(),
		PlayerSessionId: psess,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemovePlayerHandler(w http.ResponseWriter, r *http.Request) {
	psess := r.URL.Query().Get("psess")
	if err := h.c.RemovePlayerSession(&pbuffer.RemovePlayerSession{
		GameSessionId:   *h.c.GetGameSessionId(),
		PlayerSessionId: psess,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TerminateHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.c.TerminateGameSession(&pbuffer.GameSessionTerminate{
		GameSessionId: *h.c.GetGameSessionId(),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// This is hidoi.
	log.Panic("terminate called")

	w.WriteHeader(http.StatusNoContent)
}

// OpenFreeUDPPort opens free UDP port.
// This example does not actually use UDP ports,
// but to avoid port collisions with the HTTP server,
// it binds the same number of UDP port in advance.
func OpenFreeUDPPort(portBase int, num int) (net.PacketConn, int, error) {
	for i := 0; i < num; i++ {
		port := portBase + i
		conn, err := net.ListenPacket("udp", fmt.Sprint(":", port))
		if err != nil {
			continue
		}
		return conn, port, nil
	}
	return nil, 0, errors.New("failed to open free port")
}

func main() {
	conn, port, err := OpenFreeUDPPort(9000, 100)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	logger := &glog.StandardLogger{}
	c := gamelift.NewClient(logger)
	h := &Handler{c: c}
	c.Handle(h)
	if err := c.Open(); err != nil {
		log.Panic(err)
	}
	if err := c.ProcessReady(&pbuffer.ProcessReady{
		LogPathsToUpload: []string{},
		Port:             int32(port),
		// MaxConcurrentGameSessions: 0, // not set in original ServerSDK
	}); err != nil {
		log.Panic(err)
	}
	res, err := c.GetInstanceCertificate(&pbuffer.GetInstanceCertificate{})
	log.Println(res, err)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.HandleFunc("/acceptPlayer", func(w http.ResponseWriter, r *http.Request) {
		h.AcceptPlayerHandler(w, r)
	})
	r.HandleFunc("/removePlayer", func(w http.ResponseWriter, r *http.Request) {
		h.RemovePlayerHandler(w, r)
	})
	r.HandleFunc("/terminate", func(w http.ResponseWriter, r *http.Request) {
		h.TerminateHandler(w, r)
	})
	if err := http.ListenAndServeTLS(fmt.Sprint(":", port), res.CertificatePath, res.PrivateKeyPath, r); err != nil {
		log.Panic(err)
	}
}

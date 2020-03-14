package gamelift

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/neguse/gomelift/pkg/proto/pbuffer"
	"github.com/neguse/gomelift/pkg/socketio"
)

const (
	healthCheckTimeout = 60
)

type Handler interface {
	StartGameSession(event *pbuffer.ActivateGameSession)
	UpdateGameSession(event *pbuffer.UpdateGameSession)
	ProcessTerminate(event *pbuffer.TerminateProcess)
	HealthCheck() bool
}

type Client interface {
	Handle(h Handler)

	Open() error
	ProcessReady(event *pbuffer.ProcessReady) error
	ProcessEnding(event *pbuffer.ProcessEnding) error
	ActivateGameSession(event *pbuffer.GameSessionActivate) error
	TerminateGameSession(event *pbuffer.GameSessionTerminate) error
	StartMatchBackfill(event *pbuffer.BackfillMatchmakingRequest) (*pbuffer.BackfillMatchmakingResponse, error)
	StopMatchBackfill(event *pbuffer.StopMatchmakingRequest) error
	UpdatePlayerSessionCreationPolicy(event *pbuffer.UpdatePlayerSessionCreationPolicy) error
	AcceptPlayerSession(event *pbuffer.AcceptPlayerSession) error
	RemovePlayerSession(event *pbuffer.RemovePlayerSession) error
	DescribePlayerSessions(event *pbuffer.DescribePlayerSessionsRequest) (*pbuffer.DescribePlayerSessionsResponse, error)
	GetInstanceCertificate(event *pbuffer.GetInstanceCertificate) (*pbuffer.GetInstanceCertificateResponse, error)

	GetGameSessionId() *string
	GetTerminationTime() *time.Time
}

type client struct {
	client  *socketio.Client
	handler Handler
	isReady bool
}

func NewClient() Client {
	return &client{}
}

func (c *client) Handle(h Handler) {
	c.handler = h
}

func (c *client) Open() error {
	q := url.Values{}
	if ppid := os.Getenv("MAIN_PID"); ppid != "" {
		q.Set("pID", ppid)
	} else {
		q.Set("pID", fmt.Sprint(os.Getpid()))
	}
	q.Set("sdkVersion", "3.4.0")
	q.Set("sdkLanguage", "Go")
	u := "http://127.0.0.1:5757/socket.io/?" + q.Encode()
	c.client = socketio.NewClient(u)
	c.client.HandleFunc(func(p *socketio.Packet) {
		name := string(p.Data[0].(json.RawMessage))
		var str string
		err := json.Unmarshal([]byte(p.Data[1].(json.RawMessage)), &str)
		if err != nil {
			log.Panic(err)
		}
		switch name {

		case `"StartGameSession"`:
			msg := &pbuffer.ActivateGameSession{}
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				log.Panic(err)
			}
			log.Println("handled StartGameSession", msg)
			ackPacket := socketio.NewAckPacket(p, []interface{}{true})
			if err := c.client.SendPacket(ackPacket); err != nil {
				log.Panic(err)
			}
			c.handler.StartGameSession(msg)
		case `"UpdateGameSession"`:
			msg := &pbuffer.UpdateGameSession{}
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				log.Panic(err)
			}
			log.Println("handled UpdateGameSession", msg)
			ackPacket := socketio.NewAckPacket(p, []interface{}{true})
			if err := c.client.SendPacket(ackPacket); err != nil {
				log.Panic(err)
			}
			c.handler.UpdateGameSession(msg)
		case `"TerminateProcess"`:
			msg := &pbuffer.TerminateProcess{}
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				log.Panic(err)
			}
			log.Println("handled TerminateProcess", msg)
			ackPacket := socketio.NewAckPacket(p, []interface{}{true})
			if err := c.client.SendPacket(ackPacket); err != nil {
				log.Panic(err)
			}
			c.handler.ProcessTerminate(msg)
		default:
			log.Println("unhandled packet", name)
		}
	})

	return c.client.Open()
}

func (c *client) ReportHealth() {
	// TODO: nonblocking
	health := c.handler.HealthCheck()
	event := &pbuffer.ReportHealth{HealthStatus: health}
	data, err := proto.Marshal(event)
	if err != nil {
		log.Panic(err)
	}
	var rmsg []interface{}
	rmsg = append(rmsg, proto.MessageName(event), data)
	c.client.Send(rmsg)
}

func (c *client) ProcessReady(event *pbuffer.ProcessReady) error {
	data, err := proto.Marshal(event)
	if err != nil {
		log.Panic(err)
	}
	var rmsg []interface{}
	rmsg = append(rmsg, proto.MessageName(event), data)
	c.client.Send(rmsg)
	c.isReady = true

	// wake healthcheck goroutine
	go func() {
		for c.isReady {
			c.ReportHealth()
			time.Sleep(time.Second * healthCheckTimeout)
		}
	}()
	return nil
}

func (c *client) ProcessEnding(event *pbuffer.ProcessEnding) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) ActivateGameSession(event *pbuffer.GameSessionActivate) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) TerminateGameSession(event *pbuffer.GameSessionTerminate) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) StartMatchBackfill(event *pbuffer.BackfillMatchmakingRequest) (*pbuffer.BackfillMatchmakingResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (c *client) StopMatchBackfill(event *pbuffer.StopMatchmakingRequest) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) UpdatePlayerSessionCreationPolicy(event *pbuffer.UpdatePlayerSessionCreationPolicy) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) AcceptPlayerSession(event *pbuffer.AcceptPlayerSession) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) RemovePlayerSession(event *pbuffer.RemovePlayerSession) error {
	panic("not implemented") // TODO: Implement
}

func (c *client) DescribePlayerSessions(event *pbuffer.DescribePlayerSessionsRequest) (*pbuffer.DescribePlayerSessionsResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (c *client) GetInstanceCertificate(event *pbuffer.GetInstanceCertificate) (*pbuffer.GetInstanceCertificateResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (c *client) GetGameSessionId() *string {
	panic("not implemented") // TODO: Implement
}

func (c *client) GetTerminationTime() *time.Time {
	panic("not implemented") // TODO: Implement
}

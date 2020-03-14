package main

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

type StartGameSession struct {
	GameSession struct {
		GameSessionID string `json:"gameSessionId"`
		FleetID       string `json:"fleetId"`
		MaxPlayers    int    `json:"maxPlayers"`
		IPAddress     string `json:"ipAddress"`
		Port          int    `json:"port"`
		DNSName       string `json:"dnsName"`
	} `json:"gameSession"`
}

func main() {
	q := url.Values{}
	if ppid := os.Getenv("MAIN_PID"); ppid != "" {
		q.Set("pID", ppid)
	} else {
		q.Set("pID", fmt.Sprint(os.Getpid()))
	}
	q.Set("sdkVersion", "3.4.0")
	q.Set("sdkLanguage", "GomeLift")
	u := "http://127.0.0.1:5757/socket.io/?" + q.Encode()
	// u := "http://127.0.0.1:3000/socket.io/?" + q.Encode()

	c := socketio.NewClient(u)
	c.HandleFunc(func(p *socketio.Packet) {
		name := string(p.Data[0].(json.RawMessage))
		var str string
		err := json.Unmarshal([]byte(p.Data[1].(json.RawMessage)), &str)
		if err != nil {
			log.Panic(err)
		}
		switch name {
		case `"StartGameSession"`:
			msg := &StartGameSession{}
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				log.Panic(err)
			}
			log.Println("handled StartGameSession", msg)
			ackPacket := socketio.NewAckPacket(p, []interface{}{true})
			if err := c.SendPacket(ackPacket); err != nil {
				log.Panic(err)
			}
		// case `"hello"`:
		// 	ackPacket := socketio.NewAckPacket(p, []interface{}{true})
		// 	if err := c.SendPacket(ackPacket); err != nil {
		// 		log.Panic(err)
		// 	}
		default:
			log.Println("unhandled packet", name)
		}
	})

	err := c.Open()
	log.Println(c, err)

	// TODO: 接続成功したことをハンドリングできるようにするべき
	time.Sleep(time.Second)

	{
		msg := &pbuffer.ProcessReady{Port: 7777, MaxConcurrentGameSessions: 10}
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Panic(err)
		}
		var rmsg []interface{}
		rmsg = append(rmsg, proto.MessageName(msg), data)
		c.Send(rmsg)
		log.Println("ProcessReady")
	}
	for {
		msg := &pbuffer.ReportHealth{HealthStatus: true}
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Panic(err)
		}
		var rmsg []interface{}
		rmsg = append(rmsg, proto.MessageName(msg), data)
		c.Send(rmsg)
		log.Println("Health")
		time.Sleep(time.Second * 6)
	}
}

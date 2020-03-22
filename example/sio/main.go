package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/neguse/gomelift/pkg/socketio"
)

func main() {
	c := socketio.NewClient("http://127.0.0.1:3000/socket.io/")
	c.HandleFunc(func(p *socketio.Packet) {
		log.Println("handle", p)

		var s string
		if err := json.Unmarshal(p.Data[0].(json.RawMessage), &s); err != nil {
			log.Panic(err)
		}
		if s == `ccc` {
			ack := socketio.NewAckPacket(p, []interface{}{interface{}("ddd")})
			if err := c.SendPacket(ack); err != nil {
				log.Panic(err)
			}
		}
	})

	if err := c.Open(); err != nil {
		log.Panic(err)
	}

	res, err := c.SendAck([]interface{}{"aaa", "bbb"})
	if err != nil {
		log.Panic(err)
	}
	log.Println("received ack", res)

	for {
		time.Sleep(time.Second * 100)
	}

}

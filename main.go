package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/neguse/gomelift/pkg/proto/pbuffer"
	"github.com/neguse/gomelift/pkg/socketio"
)

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

	c := socketio.NewClient(u)
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

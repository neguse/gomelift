package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/gamelift"
)

func stringAddr(s string) *string {
	return &s
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	gl := gamelift.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))
	res, err := gl.StartMatchmaking(&gamelift.StartMatchmakingInput{
		ConfigurationName: stringAddr("TestConfig"),
		Players: []*gamelift.Player{
			&gamelift.Player{
				PlayerId: stringAddr("player1"),
				Team:     stringAddr("TestTeam"),
			},
		},
	})

	if err != nil {
		log.Panic(err)
	}
	log.Println(res)

	ticketID := *res.MatchmakingTicket.TicketId

	var (
		addr            string
		playerSessionID string
	)

	for {
		res, err := gl.DescribeMatchmaking(&gamelift.DescribeMatchmakingInput{
			TicketIds: []*string{&ticketID},
		})

		if err != nil {
			continue
		}
		log.Println(res)
		if *res.TicketList[0].Status == "COMPLETED" {
			connInfo := res.TicketList[0].GameSessionConnectionInfo
			addr = fmt.Sprint(*connInfo.DnsName, ":", *connInfo.Port)
			playerSessionID = *res.TicketList[0].GameSessionConnectionInfo.MatchedPlayerSessions[0].PlayerSessionId
			break
		}
		time.Sleep(time.Second * 5)
	}

	log.Println("connecting to ", addr)

	// send accept request
	{
		url := fmt.Sprint("https://", addr, "/acceptPlayer?psess=", playerSessionID)
		resp, err := http.Get(url)
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}
		log.Print(data)
	}

	time.Sleep(time.Second * 500)

	// send remove request
	{
		url := fmt.Sprint("https://", addr, "/removePlayer?psess=", playerSessionID)
		resp, err := http.Get(url)
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}
		log.Print(data)
	}

	// send terminate request
	{
		url := fmt.Sprint("https://", addr, "/terminate")
		resp, err := http.Get(url)
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}
		log.Print(data)
	}
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
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

	f := func(n int, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		res, err := gl.StartMatchmaking(&gamelift.StartMatchmakingInput{
			ConfigurationName: stringAddr("TestConfig"),
			Players: []*gamelift.Player{
				&gamelift.Player{
					PlayerId: stringAddr(fmt.Sprint("player", n)),
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
			status := *res.TicketList[0].Status
			if status == "COMPLETED" {
				connInfo := res.TicketList[0].GameSessionConnectionInfo
				addr = fmt.Sprint(*connInfo.DnsName, ":", *connInfo.Port)
				playerSessionID = *res.TicketList[0].GameSessionConnectionInfo.MatchedPlayerSessions[0].PlayerSessionId
				break
			} else if status == "TIMED_OUT" {
				log.Println("timed out")
				return
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
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		go f(i, &wg)
		time.Sleep(time.Second * 8)
	}
	wg.Wait()
}

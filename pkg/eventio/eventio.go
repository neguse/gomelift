package eventio

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OpenResponse struct {
	Sid          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

type PacketType int

const (
	Open PacketType = iota
	Close
	Ping
	Pong
	Message
	Upgrade
	Noop
)

type Packet struct {
	Type PacketType
	Data string
}

var (
	ErrorEmptyPacket     = errors.New("Packet length should be at least 1 byte")
	ErrorHttpStatusNotOk = errors.New("HTTP Status Not OK")
)

func ParsePacket(packet string) (Packet, error) {
	b64 := false
	if len(packet) == 0 {
		return Packet{}, ErrorEmptyPacket
	}
	if packet[0] == 'b' {
		b64 = true
		packet = packet[1:]
	}
	t, err := strconv.Atoi(packet[0:1])
	if err != nil {
		return Packet{}, err
	}
	var data string
	if b64 {
		datab, err := base64.StdEncoding.DecodeString(packet[1:len(packet)])
		if err != nil {
			return Packet{}, err
		}
		data = string(datab)
	} else {
		data = packet[1:len(packet)]
	}
	return Packet{
		Type: PacketType(t),
		Data: data,
	}, nil
}

func EncodePacket(p Packet) ([]byte, error) {
	s := fmt.Sprintf("%d%v", p.Type, p.Data)
	return []byte(s), nil
}

func ParsePayloads(data string) ([]Packet, error) {
	var packets []Packet
	for len(data) > 0 {
		log.Println([]byte(data))
		log.Println(data)

		var p Packet
		n := strings.Index(data, ":")
		var (
			err error
			l   int
		)
		if l, err = strconv.Atoi(data[:n]); err != nil {
			return nil, err
		}
		p, err = ParsePacket(data[n+1 : n+1+l])
		if err != nil {
			return nil, err
		}
		data = data[n+1+l : len(data)]
		packets = append(packets, p)
	}
	return packets, nil
}

func EncodePayloads(packets []Packet) ([]byte, error) {
	var buf []byte
	for _, packet := range packets {
		p, err := EncodePacket(packet)
		if err != nil {
			return nil, err
		}
		s := fmt.Sprintf("%d:%s", len(p), p)
		buf = append(buf, []byte(s)...)
	}
	return buf, nil
}

func ParseResponse(resp *http.Response) ([]Packet, error) {
	if resp.StatusCode != http.StatusOK {
		log.Println(resp.StatusCode)
		return nil, ErrorHttpStatusNotOk
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParsePayloads(string(data))
}

type Handler interface {
	HandleMessage(msg string)
}

type nullHandler struct{}

func (h nullHandler) HandleMessage(msg string) {
}

type Client struct {
	url          string
	sid          string
	pingInterval int
	pingTimeout  int
	upgrades     []string
	sendCh       chan Packet
	handler      Handler
}

func NewClient(url string) *Client {
	return &Client{
		url:     url,
		sendCh:  make(chan Packet, 100),
		handler: nullHandler{},
	}
}

func (c *Client) FullUrl() string {
	u, err := url.Parse(c.url)
	if err != nil {
		log.Panic(err)
	}
	v := u.Query()
	v.Add("transport", "polling")
	v.Add("b64", "1")
	if c.sid != "" {
		v.Add("sid", c.sid)
	}
	u.RawQuery = v.Encode()
	return u.String()
}

func (c *Client) HandleOpen(r OpenResponse) error {
	c.sid = r.Sid
	c.pingInterval = r.PingInterval
	c.pingTimeout = r.PingTimeout
	c.upgrades = r.Upgrades
	return nil
}

func (c *Client) poll() error {
	resp, err := http.Get(c.FullUrl())
	if err != nil {
		return err
	}
	packets, err := ParseResponse(resp)
	if err != nil {
		return err
	}
	for _, packet := range packets {
		if err := c.HandlePacket(packet); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) loop(pingTick <-chan time.Time) error {
	f := func() error {
		for {
			log.Println("tick", time.Millisecond*time.Duration(c.pingInterval))
			select {
			case p := <-c.sendCh:
				log.Println("sending", p)
				data, err := EncodePayloads([]Packet{p})
				if err != nil {
					return err
				}
				log.Println("sending", c.FullUrl(), string(data))
				{
					resp, err := http.Post(c.FullUrl(), "text/plain;charset=UTF-8", strings.NewReader(string(data)))
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					data, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					log.Println(string(data))
				}
			case <-pingTick:
				c.sendPing("probe")
			}
		}
	}
	if err := f(); err != nil {
		log.Panic(err)
		return err
	}
	return nil
}

func (c *Client) sendPacket(p Packet) {
	log.Println("send", p)
	c.sendCh <- p
}

func (c *Client) SendMessage(m string) {
	p := Packet{
		Type: Message,
		Data: m,
	}
	c.sendPacket(p)
}

func (c *Client) sendPing(m string) {
	p := Packet{
		Type: Ping,
		Data: m,
	}
	c.sendPacket(p)
}

func (c *Client) HandlePacket(p Packet) error {
	log.Println("recv", p)
	switch p.Type {
	case Open:
		var r OpenResponse
		if err := json.Unmarshal([]byte(p.Data), &r); err != nil {
			log.Panic(err)
		}
		log.Println(r)
		return c.HandleOpen(r)
	case Close:
	case Ping:
	case Pong:
	case Message:
		c.handler.HandleMessage(p.Data)
		return nil
	case Upgrade:
	case Noop:
	}
	return nil
}

func (c *Client) Open() error {
	err := c.poll()
	if err != nil {
		log.Panic(err)
	}
	go func() {
		for {
			err := c.poll()
			if err != nil {
				log.Panic(err)
			}
		}
	}()
	go c.loop(time.Tick(time.Second * time.Duration(c.pingInterval)))
	return nil
}

func (c *Client) Send(msg string) {
	p := Packet{Type: Message, Data: msg}
	c.sendCh <- p
}

func (c *Client) Handle(h Handler) {
	c.handler = h
}

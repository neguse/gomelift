package eventio

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/neguse/gomelift/pkg/log"
)

type OpenResponse struct {
	Sid          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

type PacketType int

func (pt PacketType) String() string {
	switch pt {
	case Open:
		return "Open"
	case Close:
		return "Close"
	case Ping:
		return "Ping"
	case Pong:
		return "Pong"
	case Message:
		return "Message"
	case Upgrade:
		return "Upgrade"
	case Noop:
		return "Noop"
	default:
		return "Unknown"
	}
}

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
	c            *websocket.Conn
	logger       log.Logger
}

func NewClient(url string, logger log.Logger) *Client {
	return &Client{
		url:     url,
		sendCh:  make(chan Packet, 100),
		handler: nullHandler{},
		logger:  logger,
	}
}

func (c *Client) FullUrl() string {
	u, err := url.Parse(c.url)
	if err != nil {
		c.logger.Panic("failed to parse", err)
	}
	v := u.Query()
	v.Add("transport", "websocket")
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
	typ, data, err := c.c.ReadMessage()
	if err != nil {
		return err
	}
	if typ != websocket.TextMessage {
		c.logger.Panic("unsupported message type", typ)
	}
	packet, err := ParsePacket(string(data))
	if err != nil {
		return err
	}

	if err := c.HandlePacket(packet); err != nil {
		return err
	}

	return nil
}

func (c *Client) loop() error {
	f := func() error {
		pingTick := time.NewTicker(time.Millisecond * time.Duration(c.pingInterval))
		for {
			select {
			case <-pingTick.C:
				c.sendPing("probe")

			case p := <-c.sendCh:
				c.logger.Log("sending", p.Type)
				data, err := EncodePacket(p)
				if err != nil {
					return err
				}
				{
					err := func() error {
						err := c.c.WriteMessage(websocket.TextMessage, data)
						if err != nil {
							return err
						}
						return nil
					}()
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := f(); err != nil {
		c.logger.Panic("error occurred in eventio.Client.loop()", err)
		return err
	}
	return nil
}

func (c *Client) sendPacket(p Packet) {
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
	c.logger.Log("recv", p.Type)
	switch p.Type {
	case Open:
		var r OpenResponse
		if err := json.Unmarshal([]byte(p.Data), &r); err != nil {
			c.logger.Panic("error occurrend in eventio.Client.HandlePacket()", err)
		}
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
	wsConn, _, err := websocket.DefaultDialer.Dial(c.FullUrl(), nil)
	if err != nil {
		return err
	}
	c.c = wsConn

	err = c.poll()
	if err != nil {
		c.logger.Log("error occurred in eventio.Client.Open()", err)
	}

	go func() {
		for {
			err := c.poll()
			if err != nil {
				c.logger.Log("error occurred in eventio.Client.Open()", err)
			}
		}
	}()
	go c.loop()
	return nil
}

func (c *Client) Send(msg string) {
	p := Packet{Type: Message, Data: msg}
	c.sendCh <- p
}

func (c *Client) Handle(h Handler) {
	c.handler = h
}

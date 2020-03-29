package socketio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/neguse/gomelift/pkg/eventio"
	"github.com/neguse/gomelift/pkg/log"
)

var (
	ErrorEmptyPacket = errors.New("Packet length should be at least 1 byte")
	ErrorNullPacket  = errors.New("Packet length should be at least 1")
)

type PacketType int

const (
	Connect PacketType = iota
	Disconnect
	Event
	Ack
	Error
	BinaryEvent
	BinaryAck
)

func (pt PacketType) String() string {
	switch pt {
	case Connect:
		return "Connect"
	case Disconnect:
		return "Disconnect"
	case Event:
		return "Event"
	case Ack:
		return "Ack"
	case Error:
		return "Error"
	case BinaryEvent:
		return "BinaryEvent"
	case BinaryAck:
		return "BinaryAck"
	default:
		return "Unknown"
	}
}

type Packet struct {
	Type PacketType
	ID   *int
	Data []interface{}
}

func NewAckPacket(p *Packet, data []interface{}) Packet {
	return Packet{
		Type: Ack,
		ID:   p.ID,
		Data: data,
	}
}

func EncodePacket(p Packet) (string, error) {
	var (
		data []byte
		err  error
	)
	if len(p.Data) > 0 {
		data, err = json.Marshal(p.Data)
		if err != nil {
			return "", err
		}
	} else {
		data = []byte{}
	}
	var idStr string
	if p.ID != nil {
		idStr = fmt.Sprint(*p.ID)
	}
	return fmt.Sprintf("%d%v%v", p.Type, idStr, string(data)), nil
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func DecodePacket(data string) (Packet, error) {
	var p Packet
	if len(data) == 0 {
		return p, ErrorEmptyPacket
	}
	typ, err := strconv.Atoi(data[0:1])
	if err != nil {
		return p, err
	}
	p.Type = PacketType(typ)
	if p.Type != Event && p.Type != Ack && p.Type != Error {
		return p, nil
	}
	var i int
	for i = 1; i < len(data); i++ {
		if !isDigit(data[i]) {
			break
		}
	}
	//log.Println(data, i)
	if i > 1 {
		pid, err := strconv.Atoi(data[1:i])
		if err != nil {
			return p, err
		}
		p.ID = &pid
	}
	//log.Println(data, i, data[i:])
	var msgs []json.RawMessage
	if err := json.Unmarshal([]byte(data[i:]), &msgs); err != nil {
		return p, err
	}
	if len(msgs) == 0 {
		return p, ErrorNullPacket
	}
	//log.Println(msgs)
	for _, msg := range msgs {
		p.Data = append(p.Data, msg)
	}
	return p, nil
}

type HandlerFunc func(packet *Packet)

type Handler interface {
	HandleMessage(packet *Packet)
}

type nullHandler struct{}

func (h nullHandler) HandleMessage(packet *Packet) {
}

func (f HandlerFunc) HandleMessage(packet *Packet) {
	f(packet)
}

type Client struct {
	c       *eventio.Client
	handler Handler
	reqId   int
	ackCh   map[int]chan []interface{}
	ackChMu sync.Mutex
	logger  log.Logger
}

func NewClient(url string, logger log.Logger) *Client {
	ec := eventio.NewClient(url, logger)
	c := &Client{
		c:       ec,
		handler: &nullHandler{},
		reqId:   10000,
		ackCh:   make(map[int]chan []interface{}),
		logger:  logger,
	}
	ec.Handle(c)
	return c
}

func (c *Client) NextReqID() int {
	c.reqId++
	reqID := c.reqId
	return reqID
}

func (c *Client) Handle(h Handler) {
	c.handler = h
}

func (c *Client) HandleFunc(fn func(p *Packet)) {
	c.handler = HandlerFunc(fn)
}

// HandleMessage handles Event.IO Message.
func (c *Client) HandleMessage(msg string) {
	p, err := DecodePacket(msg)
	if err != nil {
		c.logger.Panic("failed to DecodePacket", err)
	}
	c.logger.Log("recv", p.Type)
	switch p.Type {
	case Event:
		c.handler.HandleMessage(&p)
	case Ack:
		c.logger.Log("recv ack id", *p.ID)
		c.ackChMu.Lock()
		if ackCh, ok := c.ackCh[*p.ID]; ok {
			ackCh <- p.Data
			delete(c.ackCh, *p.ID)
		}
		c.ackChMu.Unlock()
	default:
		c.logger.Log("received ignoring type", p.Type)
	}

}

func (c *Client) Open() error {
	return c.c.Open()
}

func (c *Client) SendPacket(p Packet) error {
	s, err := EncodePacket(p)
	if err != nil {
		return err
	}
	c.logger.Log("sending", s)
	c.c.Send(s)
	return nil
}

func (c *Client) SendPacketAck(p Packet) ([]interface{}, error) {
	reqID := c.NextReqID()
	p.ID = &reqID
	s, err := EncodePacket(p)
	if err != nil {
		return nil, err
	}
	c.ackChMu.Lock()
	c.ackCh[reqID] = make(chan []interface{})
	c.ackChMu.Unlock()
	c.logger.Log("sending need ack", reqID, s)
	c.c.Send(s)

	ack := <-c.ackCh[reqID]
	c.logger.Log("received ack", ack)
	return ack, nil
}

func (c *Client) Send(data []interface{}) error {
	p := Packet{
		Data: data,
		ID:   nil,
		Type: Event,
	}
	return c.SendPacket(p)
}

func (c *Client) SendAck(data []interface{}) ([]interface{}, error) {
	p := Packet{
		Data: data,
		ID:   nil,
		Type: Event,
	}
	return c.SendPacketAck(p)
}

package socketio

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/neguse/gomelift/pkg/eventio"
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

type Packet struct {
	Type PacketType
	ID   *int
	Data []interface{}
}

func EncodePacket(p Packet) (string, error) {
	data, err := json.Marshal(p.Data)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(p.Type, string(data)), nil
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
	log.Println(data, i)
	if i > 1 {
		pid, err := strconv.Atoi(data[1:i])
		if err != nil {
			return p, err
		}
		p.ID = &pid
	}
	log.Println(data, i, data[i:])
	var msgs []json.RawMessage
	if err := json.Unmarshal([]byte(data[i:]), &msgs); err != nil {
		return p, err
	}
	if len(msgs) == 0 {
		return p, ErrorNullPacket
	}
	log.Println(msgs)
	for _, msg := range msgs {
		p.Data = append(p.Data, msg)
	}
	return p, nil
}

type Client struct {
	c *eventio.Client
}

func NewClient(url string) *Client {
	ec := eventio.NewClient(url)
	c := &Client{c: ec}
	ec.Handle(c)
	return c
}

func (c *Client) HandleMessage(msg string) {
	log.Println("recv sio msg", msg)
	p, err := DecodePacket(msg)
	if err != nil {
		log.Panic(err)
	}
	log.Println("HandleMessage", p)
}

func (c *Client) Open() error {
	return c.c.Open()
}

func (c *Client) sendPacket(p Packet) error {
	s, err := EncodePacket(p)
	if err != nil {
		return err
	}
	log.Println("sending", s)
	c.c.Send(s)
	return nil
}

func (c *Client) Send(data []interface{}) error {
	p := Packet{
		Data: data,
		ID:   nil,
		Type: Event,
	}
	return c.sendPacket(p)
}

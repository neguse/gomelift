package eventio

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	p := Packet{Type: Message, Data: "data"}
	enc, err := EncodePacket(p)
	if err != nil {
		t.Error(err)
	}
	expect := []byte("4data")
	if !reflect.DeepEqual(enc, expect) {
		t.Error("data mismatch", enc, expect)
	}
}

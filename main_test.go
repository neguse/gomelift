package main

import (
	"encoding/json"
	"testing"
)

func TestDecode(t *testing.T) {
	data := []byte(`"{\n  \"gameSession\": {\n    \"gameSessionId\": \"gsess-abd\",\n    \"fleetId\": \"fleet-123\",\n    \"maxPlayers\": 2,\n    \"ipAddress\": \"127.0.0.1\",\n    \"port\": 7777,\n    \"dnsName\": \"localhost\"\n  }\n}"`)
	str := ""
	msg := &StartGameSession{}
	err := json.Unmarshal(data, &str)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
	err = json.Unmarshal([]byte(str), msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}

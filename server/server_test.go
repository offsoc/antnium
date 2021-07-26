package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/dobin/antnium/client"
	"github.com/dobin/antnium/model"
)

func TestServerClientIntegration(t *testing.T) {
	port := "55001"
	packetId := "packetid-42"
	computerId := "computerid-23"
	s := NewServer("127.0.0.1:" + port)

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	s.packetDb.add(packetInfo)

	// make server go
	go s.Serve()

	// create client, receive the packet we added above
	// This tests most of the stuff (encryption, encoding, campaign data, server paths and more)
	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Config.ComputerId = computerId
	packet, err := c.GetPacket()
	if err != nil {
		t.Errorf("Error when receiving packet: " + err.Error())
	}
	if packet.PacketId != packetId {
		t.Errorf("Packet received, but wrong packetid: %s", packet.PacketId)
	}
	if packet.Arguments["arg0"] != "value0" {
		t.Errorf("Packet received, but wrong args: %v", packet.Arguments)
	}
}

func TestServerAuthAdmin(t *testing.T) {
	var err error

	// Start server in the background
	port := "55002"
	s := NewServer("127.0.0.1:" + port)
	go s.Serve()

	// Create a default (non authenticated) HTTP client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Test Admin
	r, _ := http.NewRequest("GET", "http://127.0.0.1:55002/admin/packets", nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 200 {
		t.Errorf("Could access admin API without authentication")
	}
}

func TestServerAuthClient(t *testing.T) {
	var err error
	var url string
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Start server in the background
	port := "55000"
	s := NewServer("127.0.0.1:" + port)

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	s.packetDb.add(packetInfo)

	go s.Serve()

	// Create a default (non authenticated) HTTP client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Config.ComputerId = computerId

	// Test Client: No key
	url = c.PacketGetUrl()
	r, _ := http.NewRequest("GET", url, nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		t.Errorf("Error accessing server api with url: " + url)
	}
	if resp.StatusCode == 200 {
		t.Errorf("Could access server API though i should not: " + url)
	}

	// Test Client: Correct key
	packet, err = c.GetPacket()
	if err != nil {
		t.Errorf("Could not get packet: " + err.Error())
	}
	if packet.PacketType != "test" {
		t.Errorf("Recv packet err")
	}
	if packet.ComputerId != computerId {
		t.Errorf("Recv packet err")
	}
	if packet.PacketId != packetId {
		t.Errorf("Recv packet err")
	}

	// Test: Static
	/*
		url = c.PacketGetUrl()
		r, _ = http.NewRequest("GET", url, nil)
		resp, err = unauthHttp.Do(r)
		if err != nil {
			t.Errorf("Error accessing static with url: " + url)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Could access static: " + url)
		}
	*/

	// Test: Upload?
}

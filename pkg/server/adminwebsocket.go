package server

/* Mostly based on
   https://rogerwelin.github.io/golang/websockets/gorilla/2018/03/13/golang-websockets.html
*/

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebsocketData is just a wrapper for PacketInfo atm
type WebsocketData struct {
	PacketInfo PacketInfo `json:"PacketInfo"`
}

type AuthToken string

type AdminWebSocket struct {
	clients     map[*websocket.Conn]bool
	adminapiKey string

	channel    chan *WebsocketData
	wsUpgrader websocket.Upgrader
}

func MakeAdminWebSocket(adminApiKey string) AdminWebSocket {
	a := AdminWebSocket{
		make(map[*websocket.Conn]bool),
		adminApiKey,
		make(chan *WebsocketData),
		websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return a
}

// wsHandler is the entry point for new websocket connections
func (a *AdminWebSocket) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := a.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("AdminWebsocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	// first message should be the AdminApiKey
	var authToken AuthToken
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("AdminWebsocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Warnf("AdminWebsocket: could not decode auth: %v", message)
		return
	}
	if string(authToken) == a.adminapiKey {
		// register client as auth succeeded
		a.clients[ws] = true
	} else {
		log.Warn("AdminWebsocket: incorrect key: " + authToken)
	}
}

func (a *AdminWebSocket) broadcastPacket(packetInfo PacketInfo) {
	websocketData := WebsocketData{
		packetInfo,
	}
	a.channel <- &websocketData
}

// Distributor is a Thread which distributes data to all connected websocket clients. Lifetime: app
func (a *AdminWebSocket) Distributor() {
	for {
		guiData := <-a.channel

		data, err := json.Marshal(guiData)
		if err != nil {
			log.Error("Could not JSON marshal")
		}

		// send to every client that is currently connected
		for client := range a.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("AdminWebsocket error: %s", err)
				client.Close()
				delete(a.clients, client)
			}
		}
	}
}

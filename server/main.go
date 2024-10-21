package main

import (
	"dummy-endpoints-ws/structs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handler function that returns the current port via WebSocket
func portHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		var out []structs.Response
		for i := 0; i < structs.ResponseRowsPerServer; i++ {
			res := structs.Response{
				Message:   fmt.Sprintf("This is port: %d", port),
				TimeStamp: time.Now().Format(time.RFC3339),
				Price:     structs.RandomInt(1, 100),
				Address:   "0x" + fmt.Sprintf("%d", i),
			}
			out = append(out, res)
		}

		message, err := json.Marshal(out)
		if err != nil {
			log.Printf("Error marshalling response: %v", err)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing WebSocket message: %v", err)
		}
	}
}

func main() {
	beginPort := structs.GetPorts().Min
	endPort := structs.GetPorts().Max
	fmt.Println("Total number of ports(servers): ", endPort-beginPort+1)

	if beginPort > endPort {
		log.Fatalf("Begin port should be less than or equal to end port")
	}

	for port := beginPort; port <= endPort; port++ {
		if !structs.Contains(structs.GetPorts().Failed, port) {
			go func(p int) {
				mux := http.NewServeMux()
				mux.HandleFunc("/ws", portHandler(p))

				addr := fmt.Sprintf(":%d", p)
				log.Printf("Starting WebSocket server on port %d", p)

				if err := http.ListenAndServe(addr, mux); err != nil {
					log.Fatalf("Failed to start server on port %d: %v", p, err)
				}
			}(port)
		}
	}

	select {}
}

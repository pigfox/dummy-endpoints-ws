package requester

import (
	"dummy-endpoints-ws/structs"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

func MakeWebSocketRequest(url string) ([]structs.Response, error) {
	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to WebSocket server: %w", err)
	}
	defer conn.Close()

	// Set timeout
	conn.SetReadDeadline(time.Now().Add(structs.RequestTimeOut * time.Millisecond))

	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading WebSocket message: %w", err)
	}

	// Parse the JSON response
	var responses []structs.Response
	if err := json.Unmarshal(message, &responses); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return responses, nil
}

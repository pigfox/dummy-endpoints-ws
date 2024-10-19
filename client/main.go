package main

import (
	"dummy-endpoints-ws/requester"
	"dummy-endpoints-ws/structs"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

func main() {
	beginTime := time.Now()
	beginPort := structs.GetPorts().Min
	endPort := structs.GetPorts().Max

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResponses := []structs.Response{}

	// Iterate through each port and establish WebSocket connections concurrently
	for port := beginPort; port <= endPort; port++ {
		wg.Add(1)

		go func(port int) {
			defer wg.Done()

			url := fmt.Sprintf("ws://localhost:%d/ws", port)
			responses, err := requester.MakeWebSocketRequest(url)
			if err != nil {
				log.Printf("Error for port %d: %v", port, err)
				return
			}

			// Lock before modifying the shared slice
			mu.Lock()
			allResponses = append(allResponses, responses...)
			mu.Unlock()

		}(port)
	}

	// Wait for all Go routines to complete
	wg.Wait()

	// Sort the responses by the Address field
	sort.Slice(allResponses, func(i, j int) bool {
		return allResponses[i].Address < allResponses[j].Address
	})

	// Group responses by Address for price comparison
	groupedByAddress := make(map[string][]structs.Response)

	for _, response := range allResponses {
		groupedByAddress[response.Address] = append(groupedByAddress[response.Address], response)
	}

	// Compare prices by address and check if the difference exceeds the threshold
	priceDifferenceThreshold := structs.PriceDifferencePct / 100.0

	for address, responses := range groupedByAddress {
		if len(responses) < 2 {
			// Not enough responses to compare
			continue
		}

		// Sort responses by price within the group
		sort.Slice(responses, func(i, j int) bool {
			return responses[i].Price < responses[j].Price
		})

		// Compare each consecutive pair of prices
		for i := 1; i < len(responses); i++ {
			price1 := float64(responses[i-1].Price)
			price2 := float64(responses[i].Price)
			diffPct := math.Abs(price2-price1) / price1

			if diffPct > priceDifferenceThreshold {
				port1 := responses[i-1].Message
				port2 := responses[i].Message
				// Add output indication from which ports the prices are coming from
				fmt.Printf("Significant price difference found at Address: %s\n", address)
				fmt.Printf("Port1: %s, Price1: %d, Port2: %s, Price2: %d, Difference: %.2f%%\n",
					port1, int(price1), port2, int(price2), diffPct*100)
			}
		}
	}
	fmt.Println("Total time taken: ", time.Since(beginTime))
	fmt.Println("Total number of responses: ", len(allResponses))
	fmt.Println("Total number of ports(servers): ", endPort-beginPort+1)
}

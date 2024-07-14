package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
		Port 							= 8000
		WebSocketRoute    = "/stream"
		recordedStreamsFilepath = "../recordedStreams/recordedStream.json"
)

type callStreamRecord struct {
	CallDurationMs		 	int64      				 `json:"callDurationMs"`
	MediaPayloadSize 		int                `json:"mediaPayloadSize"`
	MediaPayloadCount		int                `json:"mediaPayloadCount"`
	MediaPayloadsPerSec	int                `json:"mediaPayloadsPerSec"`
	Messages         		[]WebSocketMessage `json:"messages"`
}

type WebSocketMessage struct {
	Event    string          `json:"event"`
	StreamSid string         `json:"streamSid,omitempty"`
	Media    *MediaPayload   `json:"media,omitempty"`
}

type MediaPayload struct {
	Payload string `json:"payload"`
}

func parseRecordedStreamJSON(filepath string) callStreamRecord {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer file.Close()

	// Initialize an instance of callStreamRecord
	callStreamRecord := callStreamRecord{}

	// Decode the JSON data into callStreamRecord struct
	err = json.NewDecoder(file).Decode(&callStreamRecord)
	if err != nil {
		log.Fatal("Failed to decode JSON: ", err)
	}

	return callStreamRecord
}

func main() {
	url := fmt.Sprintf("ws://localhost:%d%s", Port, WebSocketRoute)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
			log.Fatal("Failed to connect to WebSocket server: ", err)
	}
	defer c.Close()

	err = c.WriteJSON(WebSocketMessage{Event: "simulator_start"})
	if err != nil {
			log.Fatal("Failed to send simulator_start message: ", err)
	}
	
	var callStreamRecord = parseRecordedStreamJSON(recordedStreamsFilepath)
	fmt.Println("Parsed call:")
	fmt.Printf("Call duration: %d ms\n", callStreamRecord.CallDurationMs)
	fmt.Printf("Media payload size: %d\n", callStreamRecord.MediaPayloadSize)
	fmt.Printf("Media payload count: %d\n", callStreamRecord.MediaPayloadCount)
	fmt.Printf("Streaming at %d messages per second...\n", callStreamRecord.MediaPayloadsPerSec)

	// Create a ticker that triggers based on MediaPayloadsPerSec
	messageInterval := time.Second / time.Duration(callStreamRecord.MediaPayloadsPerSec)
	ticker := time.NewTicker(messageInterval)
	defer ticker.Stop()

	for _, msg := range callStreamRecord.Messages {
		<-ticker.C // Wait for the next tick

		err := c.WriteJSON(msg)
		if err != nil {
				log.Fatal("Failed to send message: ", err)
		}
	}
}
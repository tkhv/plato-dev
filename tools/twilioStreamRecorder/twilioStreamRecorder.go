package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
		Port 							= 8000
    IncomingCallRoute = "/"
    WebSocketRoute    = "/stream"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

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

func receiveCall(c *gin.Context) {
    if c.Request.Method == http.MethodPost {
        fmt.Println("Incoming call")
        host := c.Request.Host
        xml := fmt.Sprintf(`
					<Response>
							<Connect>
									<Stream url='wss://%s%s' />
							</Connect>
					</Response>`, host, WebSocketRoute)
        c.Data(http.StatusOK, "text/xml", []byte(xml))
    } else {
        c.String(http.StatusOK, "twilioCallRecorder is running.")
    }
}

func callStreamWebSocket(c *gin.Context) {
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        fmt.Println("Failed to set websocket upgrade: ", err)
        return
    }
    defer ws.Close()

		var callStartTime time.Time
		var callDuration time.Duration
		var mediaPayloadSize = 0
		var mediaPayloadCount = 0
		var messages []WebSocketMessage = make([]WebSocketMessage, 0)
		var writeToFile = true

    for {
				// Unmarshal ws message json
        _, message, err := ws.ReadMessage()
        if err != nil {
            fmt.Println("read:", err)
            break
        }
        var data WebSocketMessage
        if err := json.Unmarshal(message, &data); err != nil {
            fmt.Println("json unmarshal error:", err)
            continue
        }

				// Append the message to the messages slice
        messages = append(messages, data)
				
				// ws message handling
        switch data.Event {
					case "simulator_start":
						// To prevent writing to file if the payloads are from the sim.
						fmt.Println("simulator started")
						writeToFile = false
					case "connected":
							fmt.Println("call connected")
					case "start":
							callStartTime = time.Now()
							fmt.Println("twilio started")
					case "media":
							mediaPayloadCount += 1
							mediaPayloadSize = max(mediaPayloadSize, len(data.Media.Payload))

							// Echo the audio back to the caller
							wsMessage := WebSocketMessage{
									Event:    "media",
									StreamSid: data.StreamSid,
									Media: &MediaPayload{
											Payload: data.Media.Payload,
									},
							}
							wsMessageBytes, err := json.Marshal(wsMessage)
							if err != nil {
									log.Println("Marshal error:", err)
									continue
							}
							if err := ws.WriteMessage(websocket.TextMessage, wsMessageBytes); err != nil {
									log.Println("Write error:", err)
									continue
							}
					case "stop":
							callDuration = time.Since(callStartTime)
							fmt.Println("twilio stopped")
        }
    }
		
		// After the loop, create wsStreamRecord instance
    record := callStreamRecord{
			CallDurationMs: 		callDuration.Milliseconds(),
			MediaPayloadCount: 	mediaPayloadCount,
			MediaPayloadSize: 	mediaPayloadSize,
			MediaPayloadsPerSec: int(math.Round((float64(mediaPayloadCount) / callDuration.Seconds()))),
			Messages:         	messages,
		}

		fmt.Println(" === CALL STREAM RECORD === ")
		fmt.Printf("Call duration (ms): %v\n", callDuration.Milliseconds())
		fmt.Printf("Media payload count: %d\n", mediaPayloadCount)
		fmt.Printf("Media payload size: %d\n", mediaPayloadSize)
		fmt.Printf("Media payloads per second: %d\n", record.MediaPayloadsPerSec)
		fmt.Println(" ========================== ")

		if writeToFile {
				// Ensure the directory exists
				dirPath := "../recordedStreams"
				if err := os.MkdirAll(dirPath, 0755); err != nil {
						log.Fatal("Failed to create directory: ", err)
				}
		
				// Now, try to open or create the file
				filePath := dirPath + "/recordedStream.json"
				file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
						log.Fatal("Failed to open file: ", err)
				}
				defer file.Close()

				// Marshal wsStreamRecord to JSON
				recordBytes, err := json.MarshalIndent(record, "", "	")
				if err != nil {
						log.Println("Failed to marshal record: ", err)
						return
				}

				// Write JSON to file
				if _, err := file.Write(recordBytes); err != nil {
						log.Println("Failed to write record to file. \n\n", err)
				} else {
						fmt.Printf("Call stream record written to file %s\n\n", file.Name())
				}
		}
}

func main() {
		gin.SetMode(gin.ReleaseMode)
    router := gin.Default()

    router.Any(IncomingCallRoute, receiveCall)
    router.GET(WebSocketRoute, callStreamWebSocket)

    fmt.Printf("twilioCallRecorder listening on http://localhost:%d\n", Port)
    router.Run(fmt.Sprintf(":%d", Port))
}

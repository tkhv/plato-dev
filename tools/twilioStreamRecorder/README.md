twilioStreamRecorder listens to a twilio call and records all incoming websocket packets to a file in tools/recordedStreams/recordedStream.json, while echoing the audio back to the caller. Recorded streams can be re-streamed to a websocket by twilioStreamSimulator.

Make sure to start ngrok if running locally:

`ngrok http http://localhost:8000`

Then update the webhook URL in your Twilio console.

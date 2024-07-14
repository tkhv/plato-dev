twilioStreamSimulator parses JSON in recordedStreams/recordedStream.json (created after twilioStreamRecorder records a call), and streams them at the same rate to a localhost websocket endpoint at the defined port and path.

Note that the simulated stream directly starts the websocket stream by sending the connected and start events, followed by media. This is different from the actual Twilio stream, which first GETs the root path, which must send out a TwiML response with a bidirectional websocket URL, which then initiates the websocket connection and only then sends the connect, start, and media events.

twilioStreamRecorder can be used to test the simulator by recording a call and then replaying it with the simulator. The simulator sends a start_simulator event initially, which sets a flag in the recorder to not write the recorded stream to disk.

module github.com/srsalisbury/bouncebot

go 1.25.4

replace github.com/srsalisbury/bouncebot/model => ./model

replace github.com/srsalisbury/bouncebot/proto => ./proto

require (
	connectrpc.com/connect v1.19.1
	github.com/lithammer/dedent v1.1.0
	github.com/rs/cors v1.11.1
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

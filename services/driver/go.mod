module github.com/hodeifa/hyperlocal-backend/services/driver

go 1.26.1

require (
	github.com/hodeifa/hyperlocal-backend/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.83.2
)

require (
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/hodeifa/hyperlocal-backend/pkg => ../../pkg

replace github.com/hodeifa/hyperlocal-backend/proto => ../../proto

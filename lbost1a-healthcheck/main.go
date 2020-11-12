package main

import run "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/cmd"

//go:generate mkdir -p grpc-transport
//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ orchestrator-and-healthcheck.proto
//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ healthcheck-and-ipvs.proto
//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ healthcheck-and-dummy.proto

func main() {
	run.Execute()
}

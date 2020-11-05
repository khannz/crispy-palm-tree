package main

// FIXME: tunel also work with routes and links
import run "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/cmd"

//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ orchestrator-and-tunnel.proto

func main() {
	run.Execute()
}

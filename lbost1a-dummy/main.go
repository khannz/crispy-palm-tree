package main

import run "github.com/khannz/crispy-palm-tree/lbost1a-dummy/cmd"

//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ healthcheck-and-dummy.proto

func main() {
	run.Execute()
}

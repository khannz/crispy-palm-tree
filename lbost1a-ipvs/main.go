package main

import run "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/cmd"

//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ healthcheck-and-ipvs.proto

func main() {
	run.Execute()
}

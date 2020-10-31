package main

//go:generate swag init -g application/rest-main.go
//go:generate protoc --go_out=./grpc-transport/ --proto_path=../ --go-grpc_out=./grpc-transport/ --proto_path=../ orchestrator-and-healthcheck.proto

import run "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/cmd"

// @title LBOS tier 1 swagger
// @version 1.2.3
// @description create/delete/moddify tier 1 service
// @Tags New service
// @tag.name Link for docs
// @tag.docs.url http://kb.sdn.sbrf.ru/display/SDN/*
// @tag.docs.description Docs at confluence
// @contact.name Ivan Tikhonov
// @contact.email sdn@sberbank.ru
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	run.Execute()
}

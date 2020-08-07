package main

//go:generate ./swag init -g application/rest.go

import run "github.com/khannz/crispy-palm-tree/cmd"

// @title NLB service swagger
// @version 1.2.3
// @description create/delete nlb service.
// @Tags New nlb service
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

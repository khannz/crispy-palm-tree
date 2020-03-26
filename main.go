package main

//go:generate ./swag init -g application/rest.go

import run "git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/cmd"

func main() {
	run.Execute()
}

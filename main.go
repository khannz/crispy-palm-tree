package main

//go:generate ./swag init -g application/rest.go

import run "github.com/khannz/crispy-palm-tree/cmd"

func main() {
	run.Execute()
}

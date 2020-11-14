package server

import "net/http"

type App struct {
	httpServer *http.Server

	// TODO: Some usecase
}

func NewApp() *App {
	// TODO: bind repos
	return &App{}
}

func (a *App) Run(port string) error {
	return nil
}

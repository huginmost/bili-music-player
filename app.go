package main

import (
	"context"
	"log"

	"github.com/huginmost/bili-music-player/internal/bmserver"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func startAPI() string {
	server, err := bmserver.NewFromEnv()
	if err != nil {
		log.Fatalf("failed to create desktop api: %v", err)
	}

	baseURL, err := server.StartLocalhost()
	if err != nil {
		log.Fatalf("failed to start desktop api: %v", err)
	}

	return baseURL
}

package main

import (
	"context"
)

// App stores the Wails lifecycle context.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved for lifecycle coordination.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called at application termination.
func (a *App) shutdown(ctx context.Context) {
	_ = ctx
}

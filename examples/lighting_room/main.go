package main

import "github.com/sarattha/lumago/engine/app"

func main() {
	game := app.NewGame(app.Config{
		Width:  1280,
		Height: 720,
		Title:  "LumaGo Lighting Room Example",
	})

	// This example will become the first visual lighting test scene.
	// For now, use cmd/sandbox as the runnable scaffold.
	_ = game
}

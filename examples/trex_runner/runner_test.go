package main

import (
	"testing"

	"github.com/sarattha/lumago/engine/graphics"
)

func TestRunnerStartsWithJump(t *testing.T) {
	state := newRunnerState()

	state.Step(1.0/runnerTargetFPS, runnerInput{Start: true, Jump: true})

	if !state.Started {
		t.Fatalf("runner did not start")
	}
	if state.PlayerVelY <= 0 {
		t.Fatalf("player velocity=%.2f, want upward jump", state.PlayerVelY)
	}
}

func TestRunnerDuckUsesLowerHitbox(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	standing := state.playerRect()

	state.Step(1.0/runnerTargetFPS, runnerInput{Duck: true})
	ducking := state.playerRect()

	if !state.Ducking {
		t.Fatalf("runner did not enter ducking state")
	}
	if ducking.H >= standing.H {
		t.Fatalf("ducking hitbox height=%.2f, want less than standing %.2f", ducking.H, standing.H)
	}
	if ducking.W <= standing.W {
		t.Fatalf("ducking hitbox width=%.2f, want wider than standing %.2f", ducking.W, standing.W)
	}
}

func TestRunnerAdvancesScoreAndSpeed(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	state.Obstacles = nil

	for i := 0; i < runnerTargetFPS; i++ {
		state.Step(1.0/runnerTargetFPS, runnerInput{})
	}

	if state.Score <= 0 {
		t.Fatalf("score=%d, want progress after running", state.Score)
	}
	if state.Speed <= runnerStartSpeed {
		t.Fatalf("speed=%.2f, want acceleration from %.2f", state.Speed, float32(runnerStartSpeed))
	}
}

func TestRunnerCollisionEndsRunAndRestartResets(t *testing.T) {
	state := newRunnerState()
	state.Started = true
	state.Obstacles = []runnerObstacle{{Kind: runnerObstacleCactus, X: runnerDinoX}}

	state.Step(1.0/runnerTargetFPS, runnerInput{})

	if !state.GameOver {
		t.Fatalf("collision did not end the run")
	}
	state.Step(1.0/runnerTargetFPS, runnerInput{Restart: true})
	if state.GameOver || !state.Started {
		t.Fatalf("restart failed: started=%t gameOver=%t", state.Started, state.GameOver)
	}
	if state.Score != 0 {
		t.Fatalf("score=%d, want reset", state.Score)
	}
}

func TestRunnerSceneUsesLightingShadowsAndReadableSpriteRoles(t *testing.T) {
	config := defaultRunnerConfig()
	state := newRunnerState()
	world := buildRunnerScene(state, config)

	if len(world.Sprites()) < 90 {
		t.Fatalf("sprites=%d, want composed runner graphics", len(world.Sprites()))
	}
	if len(world.Lights()) != runnerLightCount {
		t.Fatalf("lights=%d, want %d", len(world.Lights()), runnerLightCount)
	}
	if countRunnerShadowLights(world.Lights()) != runnerShadowLightCount {
		t.Fatalf("shadow lights=%d, want %d", countRunnerShadowLights(world.Lights()), runnerShadowLightCount)
	}
	if len(world.Occluders()) < len(state.Obstacles)+3 {
		t.Fatalf("occluders=%d, want ground/player/obstacle shadow casters", len(world.Occluders()))
	}
	counts := countRunnerLayers(world.Sprites())
	if counts[12]+counts[13]+counts[14] < 8 {
		t.Fatalf("dino body sprites too sparse: layers=%v", counts)
	}
	if counts[8]+counts[9]+counts[10] < 8 {
		t.Fatalf("obstacle sprites too sparse: layers=%v", counts)
	}
	if counts[22] < 20 {
		t.Fatalf("score digit sprites=%d, want visible seven-segment score", counts[22])
	}
}

func countRunnerShadowLights(lights []graphics.Light2D) int {
	count := 0
	for _, light := range lights {
		if light.CastShadows {
			count++
		}
	}
	return count
}

func countRunnerLayers(sprites []graphics.SpriteDrawCommand) map[int]int {
	counts := map[int]int{}
	for _, sprite := range sprites {
		counts[sprite.Layer]++
	}
	return counts
}

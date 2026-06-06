package main

import (
	"math"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
	"github.com/sarattha/lumago/engine/scene"
)

const (
	runnerTargetWidth  = 1280
	runnerTargetHeight = 720
	runnerTargetFPS    = 60
	runnerGroundY      = 220
	runnerDinoX        = 210
	runnerGravity      = -3000
	runnerJumpVelocity = 1060
	runnerStartSpeed   = 520
	runnerMaxSpeed     = 930

	runnerLightCount       = 4
	runnerShadowLightCount = 2
)

type runnerInput struct {
	Start   bool
	Jump    bool
	Duck    bool
	Restart bool
}

type runnerState struct {
	Started      bool
	GameOver     bool
	Time         float32
	Distance     float32
	Score        int
	Speed        float32
	PlayerBottom float32
	PlayerVelY   float32
	Ducking      bool
	Obstacles    []runnerObstacle
	Clouds       []runnerCloud
	spawnTimer   float32
	spawnIndex   int
}

type runnerObstacleKind uint8

const (
	runnerObstacleCactus runnerObstacleKind = iota
	runnerObstacleCactusCluster
	runnerObstacleBird
)

type runnerObstacle struct {
	Kind runnerObstacleKind
	X    float32
}

type runnerCloud struct {
	X     float32
	Y     float32
	Scale float32
}

func newRunnerState() runnerState {
	state := runnerState{}
	state.Reset()
	return state
}

func (s *runnerState) Reset() {
	s.Started = true
	s.GameOver = false
	s.Time = 0
	s.Distance = 0
	s.Score = 0
	s.Speed = runnerStartSpeed
	s.PlayerBottom = runnerGroundY
	s.PlayerVelY = 0
	s.Ducking = false
	s.spawnTimer = 1.05
	s.spawnIndex = 0
	s.Obstacles = []runnerObstacle{
		{Kind: runnerObstacleCactus, X: 780},
		{Kind: runnerObstacleBird, X: 1120},
	}
	s.Clouds = []runnerCloud{
		{X: 170, Y: 575, Scale: 1.0},
		{X: 520, Y: 620, Scale: 1.25},
		{X: 915, Y: 550, Scale: 0.9},
	}
}

func (s *runnerState) Step(dt float32, input runnerInput) {
	if dt <= 0 {
		return
	}
	if s.GameOver {
		if input.Restart || input.Start || input.Jump {
			s.Reset()
		}
		return
	}
	if !s.Started {
		s.Ducking = input.Duck
		if input.Start || input.Jump {
			s.Started = true
			if input.Jump {
				s.PlayerVelY = runnerJumpVelocity
			}
		}
		return
	}

	s.Time += dt
	s.Speed = minRunner(runnerMaxSpeed, s.Speed+18*dt)
	s.Distance += s.Speed * dt
	s.Score = int(s.Distance / 12)

	grounded := s.grounded()
	s.Ducking = input.Duck && grounded
	if input.Jump && grounded && !s.Ducking {
		s.PlayerVelY = runnerJumpVelocity
	}
	s.PlayerBottom += s.PlayerVelY * dt
	s.PlayerVelY += runnerGravity * dt
	if s.PlayerBottom <= runnerGroundY {
		s.PlayerBottom = runnerGroundY
		s.PlayerVelY = 0
	}

	for i := range s.Obstacles {
		s.Obstacles[i].X -= s.Speed * dt
	}
	s.trimObstacles()
	s.spawnTimer -= dt
	if s.spawnTimer <= 0 {
		s.spawnObstacle()
	}
	for i := range s.Clouds {
		s.Clouds[i].X -= s.Speed * dt * (0.12 + 0.03*float32(i%2))
		if s.Clouds[i].X < -120 {
			s.Clouds[i].X = runnerTargetWidth + 120 + float32(i*90)
		}
	}

	player := s.playerRect()
	for _, obstacle := range s.Obstacles {
		if rectsOverlap(player, obstacle.Rect()) {
			s.GameOver = true
			s.Started = false
			return
		}
	}
}

func (s *runnerState) grounded() bool {
	return s.PlayerBottom <= runnerGroundY+0.001 && s.PlayerVelY == 0
}

func (s *runnerState) playerRect() lmath.Rect {
	if s.Ducking {
		return lmath.Rect{X: runnerDinoX - 47, Y: s.PlayerBottom, W: 94, H: 50}
	}
	return lmath.Rect{X: runnerDinoX - 34, Y: s.PlayerBottom, W: 68, H: 92}
}

func (s *runnerState) trimObstacles() {
	keep := s.Obstacles[:0]
	for _, obstacle := range s.Obstacles {
		if obstacle.X > -120 {
			keep = append(keep, obstacle)
		}
	}
	s.Obstacles = keep
}

func (s *runnerState) spawnObstacle() {
	kinds := []runnerObstacleKind{
		runnerObstacleCactus,
		runnerObstacleCactusCluster,
		runnerObstacleBird,
		runnerObstacleCactus,
		runnerObstacleCactusCluster,
	}
	kind := kinds[s.spawnIndex%len(kinds)]
	s.spawnIndex++
	s.Obstacles = append(s.Obstacles, runnerObstacle{Kind: kind, X: runnerTargetWidth + 95})
	gap := 0.95 + float32((s.spawnIndex%3))*0.18
	if kind == runnerObstacleBird {
		gap += 0.18
	}
	s.spawnTimer = gap
}

func (o runnerObstacle) Rect() lmath.Rect {
	switch o.Kind {
	case runnerObstacleBird:
		return lmath.Rect{X: o.X - 42, Y: runnerGroundY + 86, W: 84, H: 42}
	case runnerObstacleCactusCluster:
		return lmath.Rect{X: o.X - 40, Y: runnerGroundY, W: 80, H: 88}
	default:
		return lmath.Rect{X: o.X - 24, Y: runnerGroundY, W: 48, H: 78}
	}
}

func buildRunnerScene(state runnerState, config runnerConfig) *scene.Scene {
	world := scene.New()
	world.SetLightingConfig(graphics.LightingConfig2D{
		Ambient:    lmath.Color{R: 0.24, G: 0.25, B: 0.28, A: 1},
		DebugView:  config.DebugView,
		ShadowMode: config.ShadowMode,
	})
	addRunnerSky(world, state)
	addRunnerTrack(world, state)
	addRunnerObstacles(world, state)
	addRunnerDino(world, state)
	addRunnerScore(world, state)
	addRunnerLights(world, state)
	addRunnerOccluders(world, state)
	return world
}

func addRunnerSky(world *scene.Scene, state runnerState) {
	addRunnerRect(world, 640, 360, 1280, 720, lmath.Color{R: 0.08, G: 0.10, B: 0.13, A: 1}, 0, 0)
	addRunnerRect(world, 640, 565, 1280, 310, lmath.Color{R: 0.11, G: 0.12, B: 0.15, A: 1}, 1, 0)
	for i := 0; i < 9; i++ {
		x := float32(70 + i*150)
		y := float32(600 - (i%3)*38)
		addRunnerRect(world, x, y, 4, 4, lmath.Color{R: 0.74, G: 0.82, B: 0.96, A: 1}, 2, 2.2)
	}
	for _, cloud := range state.Clouds {
		addRunnerCloud(world, cloud)
	}
	addRunnerRect(world, 1070, 604, 84, 84, lmath.Color{R: 1.00, G: 0.82, B: 0.46, A: 1}, 2, 2.8)
	addRunnerRect(world, 1070, 604, 48, 48, lmath.Color{R: 1.00, G: 0.94, B: 0.68, A: 1}, 3, 3.5)
}

func addRunnerCloud(world *scene.Scene, cloud runnerCloud) {
	c := lmath.Color{R: 0.39, G: 0.49, B: 0.62, A: 1}
	addRunnerRect(world, cloud.X, cloud.Y, 92*cloud.Scale, 22*cloud.Scale, c, 2, 0.15)
	addRunnerRect(world, cloud.X-28*cloud.Scale, cloud.Y-12*cloud.Scale, 34*cloud.Scale, 28*cloud.Scale, c, 2, 0.15)
	addRunnerRect(world, cloud.X+16*cloud.Scale, cloud.Y-18*cloud.Scale, 44*cloud.Scale, 36*cloud.Scale, c, 2, 0.15)
}

func addRunnerTrack(world *scene.Scene, state runnerState) {
	addRunnerRect(world, 640, runnerGroundY-34, 1280, 68, lmath.Color{R: 0.30, G: 0.27, B: 0.24, A: 1}, 4, 0)
	addRunnerRect(world, 640, runnerGroundY+2, 1280, 8, lmath.Color{R: 0.78, G: 0.66, B: 0.47, A: 1}, 5, 0.3)
	offset := int(state.Distance) % 90
	for x := -offset; x < runnerTargetWidth+120; x += 90 {
		addRunnerRect(world, float32(x+24), runnerGroundY+19, 42, 5, lmath.Color{R: 0.92, G: 0.78, B: 0.52, A: 1}, 6, 0.4)
		addRunnerRect(world, float32(x+70), runnerGroundY-7, 22, 4, lmath.Color{R: 0.55, G: 0.47, B: 0.38, A: 1}, 6, 0.2)
	}
}

func addRunnerObstacles(world *scene.Scene, state runnerState) {
	for _, obstacle := range state.Obstacles {
		switch obstacle.Kind {
		case runnerObstacleBird:
			addRunnerBird(world, obstacle.X, state.Time)
		case runnerObstacleCactusCluster:
			addRunnerCactus(world, obstacle.X-24, 88, 1.0)
			addRunnerCactus(world, obstacle.X+18, 68, 0.82)
			addRunnerCactus(world, obstacle.X+42, 76, 0.92)
		default:
			addRunnerCactus(world, obstacle.X, 78, 1.0)
		}
	}
}

func addRunnerCactus(world *scene.Scene, x, height, scale float32) {
	green := lmath.Color{R: 0.19, G: 0.73, B: 0.38, A: 1}
	dark := lmath.Color{R: 0.08, G: 0.36, B: 0.20, A: 1}
	addRunnerRect(world, x, runnerGroundY+height/2, 24*scale, height, green, 8, 0.2)
	addRunnerRect(world, x-20*scale, runnerGroundY+height-30*scale, 14*scale, 38*scale, green, 8, 0.2)
	addRunnerRect(world, x+20*scale, runnerGroundY+height-44*scale, 14*scale, 32*scale, green, 8, 0.2)
	addRunnerRect(world, x, runnerGroundY+height/2, 5*scale, height-12*scale, dark, 9, 0.05)
}

func addRunnerBird(world *scene.Scene, x, t float32) {
	y := float32(runnerGroundY + 112)
	flap := float32(math.Sin(float64(t*13))) * 10
	body := lmath.Color{R: 0.55, G: 0.50, B: 0.90, A: 1}
	wing := lmath.Color{R: 0.78, G: 0.70, B: 1.00, A: 1}
	addRunnerRect(world, x, y, 58, 24, body, 9, 0.35)
	addRunnerRect(world, x+34, y-3, 26, 16, body, 9, 0.35)
	addRunnerRect(world, x-12, y-22+flap, 52, 12, wing, 10, 0.6)
	addRunnerRect(world, x-10, y+22-flap, 46, 10, wing, 10, 0.6)
	addRunnerRect(world, x+43, y-8, 4, 4, lmath.Black(), 11, 0)
}

func addRunnerDino(world *scene.Scene, state runnerState) {
	rect := state.playerRect()
	x := rect.X + rect.W/2
	body := lmath.Color{R: 0.30, G: 0.86, B: 0.70, A: 1}
	belly := lmath.Color{R: 0.72, G: 1.00, B: 0.88, A: 1}
	dark := lmath.Color{R: 0.06, G: 0.22, B: 0.20, A: 1}
	addRunnerRect(world, x, state.PlayerBottom-10, rect.W*1.15, 16, lmath.Color{R: 0.00, G: 0.00, B: 0.00, A: 0.45}, 7, 0)
	if state.Ducking {
		addRunnerRect(world, x-10, state.PlayerBottom+30, 92, 36, body, 12, 0.45)
		addRunnerRect(world, x+42, state.PlayerBottom+42, 36, 28, body, 13, 0.45)
		addRunnerRect(world, x-28, state.PlayerBottom+22, 38, 18, belly, 14, 0.8)
		addRunnerRect(world, x+53, state.PlayerBottom+48, 5, 5, dark, 15, 0)
		addRunnerRect(world, x-65, state.PlayerBottom+34, 34, 14, body, 12, 0.35)
		addRunnerLegs(world, x-14, state.PlayerBottom, state.Time)
		return
	}
	addRunnerRect(world, x-4, state.PlayerBottom+47, 54, 62, body, 12, 0.45)
	addRunnerRect(world, x+27, state.PlayerBottom+88, 46, 36, body, 13, 0.45)
	addRunnerRect(world, x+53, state.PlayerBottom+76, 26, 16, body, 13, 0.45)
	addRunnerRect(world, x-12, state.PlayerBottom+43, 26, 38, belly, 14, 0.8)
	addRunnerRect(world, x+36, state.PlayerBottom+94, 5, 5, dark, 15, 0)
	addRunnerRect(world, x-44, state.PlayerBottom+50, 36, 16, body, 12, 0.35)
	addRunnerRect(world, x+3, state.PlayerBottom+24, 28, 9, dark, 15, 0)
	addRunnerLegs(world, x, state.PlayerBottom, state.Time)
}

func addRunnerLegs(world *scene.Scene, x, bottom, t float32) {
	leg := lmath.Color{R: 0.20, G: 0.64, B: 0.55, A: 1}
	step := float32(math.Sin(float64(t * 18)))
	addRunnerRect(world, x-16-step*7, bottom+10, 15, 28, leg, 13, 0.2)
	addRunnerRect(world, x+16+step*7, bottom+10, 15, 28, leg, 13, 0.2)
	addRunnerRect(world, x-20-step*7, bottom+1, 28, 9, leg, 13, 0.2)
	addRunnerRect(world, x+19+step*7, bottom+1, 28, 9, leg, 13, 0.2)
}

func addRunnerScore(world *scene.Scene, state runnerState) {
	addRunnerRect(world, 1075, 650, 250, 52, lmath.Color{R: 0.03, G: 0.04, B: 0.05, A: 1}, 20, 0)
	for i, digit := range runnerScoreDigits(state.Score) {
		addRunnerDigit(world, float32(1052+i*32), 650, digit)
	}
	if !state.Started && !state.GameOver {
		addRunnerRect(world, 640, 294, 320, 52, lmath.Color{R: 0.05, G: 0.06, B: 0.08, A: 1}, 20, 0)
		addRunnerRect(world, 640, 294, 258, 8, lmath.Color{R: 0.95, G: 0.90, B: 0.66, A: 1}, 21, 2.4)
		addRunnerRect(world, 640, 313, 180, 6, lmath.Color{R: 0.45, G: 0.85, B: 1.00, A: 1}, 21, 2.0)
	}
	if state.GameOver {
		addRunnerRect(world, 640, 294, 330, 58, lmath.Color{R: 0.12, G: 0.03, B: 0.05, A: 1}, 20, 0)
		addRunnerRect(world, 640, 286, 260, 9, lmath.Color{R: 1.00, G: 0.28, B: 0.25, A: 1}, 21, 2.4)
		addRunnerRect(world, 640, 309, 210, 7, lmath.Color{R: 1.00, G: 0.84, B: 0.48, A: 1}, 21, 2.0)
	}
}

func runnerScoreDigits(score int) [5]int {
	if score < 0 {
		score = 0
	}
	score %= 100000
	digits := [5]int{}
	for i := len(digits) - 1; i >= 0; i-- {
		digits[i] = score % 10
		score /= 10
	}
	return digits
}

func addRunnerDigit(world *scene.Scene, x, y float32, digit int) {
	segments := [10][7]bool{
		{true, true, true, true, true, true, false},
		{false, true, true, false, false, false, false},
		{true, true, false, true, true, false, true},
		{true, true, true, true, false, false, true},
		{false, true, true, false, false, true, true},
		{true, false, true, true, false, true, true},
		{true, false, true, true, true, true, true},
		{true, true, true, false, false, false, false},
		{true, true, true, true, true, true, true},
		{true, true, true, true, false, true, true},
	}
	color := lmath.Color{R: 0.92, G: 0.96, B: 1.00, A: 1}
	dim := lmath.Color{R: 0.17, G: 0.20, B: 0.25, A: 1}
	for i, on := range segments[digit] {
		c := dim
		if on {
			c = color
		}
		switch i {
		case 0:
			addRunnerRect(world, x, y-18, 20, 5, c, 22, 1.8)
		case 1:
			addRunnerRect(world, x+12, y-8, 5, 18, c, 22, 1.8)
		case 2:
			addRunnerRect(world, x+12, y+12, 5, 18, c, 22, 1.8)
		case 3:
			addRunnerRect(world, x, y+22, 20, 5, c, 22, 1.8)
		case 4:
			addRunnerRect(world, x-12, y+12, 5, 18, c, 22, 1.8)
		case 5:
			addRunnerRect(world, x-12, y-8, 5, 18, c, 22, 1.8)
		case 6:
			addRunnerRect(world, x, y+2, 20, 5, c, 22, 1.8)
		}
	}
}

func addRunnerLights(world *scene.Scene, state runnerState) {
	world.SetLights([]graphics.Light2D{
		{Position: lmath.Vec2{X: 1070, Y: 604}, Radius: 760, Color: lmath.Color{R: 1.00, G: 0.84, B: 0.50, A: 1}, Intensity: 1.25, Falloff: 1.55, CastShadows: true},
		{Position: lmath.Vec2{X: runnerDinoX + 38, Y: state.PlayerBottom + 64}, Radius: 230, Color: lmath.Color{R: 0.42, G: 1.00, B: 0.82, A: 1}, Intensity: 0.82, Falloff: 1.25, CastShadows: true},
		{Position: lmath.Vec2{X: 330, Y: runnerGroundY - 10}, Radius: 260, Color: lmath.Color{R: 1.00, G: 0.82, B: 0.42, A: 1}, Intensity: 0.65, Falloff: 1.7},
		{Position: lmath.Vec2{X: 930, Y: runnerGroundY - 10}, Radius: 310, Color: lmath.Color{R: 0.48, G: 0.70, B: 1.00, A: 1}, Intensity: 0.60, Falloff: 1.7},
	})
}

func addRunnerOccluders(world *scene.Scene, state runnerState) {
	world.AddOccluder(graphics.RectOccluder2D(lmath.Rect{X: 0, Y: runnerGroundY - 92, W: runnerTargetWidth, H: 92}, 1))
	player := state.playerRect()
	world.AddOccluder(graphics.RectOccluder2D(player, 2))
	for _, obstacle := range state.Obstacles {
		world.AddOccluder(graphics.RectOccluder2D(obstacle.Rect(), 2))
	}
	world.AddOccluder(graphics.SegmentOccluder2D(lmath.Vec2{X: 150, Y: runnerGroundY - 10}, lmath.Vec2{X: 1130, Y: runnerGroundY - 10}, 1))
}

func addRunnerRect(world *scene.Scene, x, y, w, h float32, color lmath.Color, layer int, emissive float32) {
	world.AddSprite(graphics.SpriteDrawCommand{
		Sprite: graphics.Sprite{
			Material: graphics.Material2D{Roughness: 0.55, Emissive: emissive},
			Src:      lmath.Rect{W: 1, H: 1},
			Color:    color,
		},
		Transform: graphics.Transform2D{
			Position: lmath.Vec2{X: x, Y: y},
			Scale:    lmath.Vec2{X: w, Y: h},
			Z:        float32(layer),
		},
		Layer: layer,
	})
}

func rectsOverlap(a, b lmath.Rect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func minRunner(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

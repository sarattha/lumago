package vulkan

import (
	"math"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

const (
	defaultShadowMapResolution = 256
	defaultSDFCellSize         = 4
	shadowBiasPixels           = 1
)

type shadowSegment struct {
	A     lmath.Vec2
	B     lmath.Vec2
	Layer int
}

type lightShadowMap struct {
	LightIndex  int
	Resolution  int
	Radius      float32
	Segments    []shadowSegment
	Depths      []float32
	HasOccluder bool
}

type sdfTexture struct {
	Width       int
	Height      int
	CellSize    float32
	MaxDistance float32
	Pixels      []float32
	HasGeometry bool
}

func prepareOccluderSegmentsForFrame(dst []shadowSegment, occluders []graphics.Occluder2D, camera graphics.Camera2D) []shadowSegment {
	dst = dst[:0]
	view := camera.ViewMatrix()
	transform := func(point lmath.Vec2) lmath.Vec2 {
		return view.TransformPoint(point)
	}
	for _, occluder := range occluders {
		for _, segment := range occluder.Segments {
			if segment.A == segment.B {
				continue
			}
			dst = append(dst, shadowSegment{A: transform(segment.A), B: transform(segment.B), Layer: occluder.Layer})
		}
		if len(occluder.Points) < 2 {
			continue
		}
		for i := range occluder.Points {
			a := occluder.Points[i]
			b := occluder.Points[(i+1)%len(occluder.Points)]
			if a == b {
				continue
			}
			dst = append(dst, shadowSegment{A: transform(a), B: transform(b), Layer: occluder.Layer})
		}
	}
	return dst
}

func buildStaticSDFTextureFromOccluders(dst sdfTexture, occluders []graphics.Occluder2D, camera graphics.Camera2D, framebufferWidth, framebufferHeight int, cellSize int) sdfTexture {
	if framebufferWidth <= 0 || framebufferHeight <= 0 {
		return sdfTexture{}
	}
	if cellSize <= 0 {
		cellSize = defaultSDFCellSize
	}
	width := (framebufferWidth + cellSize - 1) / cellSize
	height := (framebufferHeight + cellSize - 1) / cellSize
	count := width * height
	if cap(dst.Pixels) < count {
		dst.Pixels = make([]float32, count)
	} else {
		dst.Pixels = dst.Pixels[:count]
	}

	view := camera.ViewMatrix()
	staticPolygons := make([][]lmath.Vec2, 0, len(occluders))
	staticSegments := make([]shadowSegment, 0, len(occluders)*4)
	for _, occluder := range occluders {
		if occluder.Caster.Dynamic {
			continue
		}
		if len(occluder.Points) >= 2 {
			polygon := make([]lmath.Vec2, len(occluder.Points))
			for i, point := range occluder.Points {
				polygon[i] = view.TransformPoint(point)
			}
			staticPolygons = append(staticPolygons, polygon)
			for i := range polygon {
				a := polygon[i]
				b := polygon[(i+1)%len(polygon)]
				if a != b {
					staticSegments = append(staticSegments, shadowSegment{A: a, B: b, Layer: occluder.Layer})
				}
			}
		}
		for _, segment := range occluder.Segments {
			if segment.A == segment.B {
				continue
			}
			staticSegments = append(staticSegments, shadowSegment{A: view.TransformPoint(segment.A), B: view.TransformPoint(segment.B), Layer: occluder.Layer})
		}
	}

	maxDistance := float32(cellSize * 8)
	hasGeometry := len(staticSegments) > 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := lmath.Vec2{
				X: (float32(x) + 0.5) * float32(cellSize),
				Y: (float32(y) + 0.5) * float32(cellSize),
			}
			distance := maxDistance
			for _, segment := range staticSegments {
				if d := distancePointToSegment(pixel, segment.A, segment.B); d < distance {
					distance = d
				}
			}
			for _, polygon := range staticPolygons {
				if pointInPolygon(pixel, polygon) {
					distance = -distance
					break
				}
			}
			dst.Pixels[y*width+x] = clamp32(distance, -maxDistance, maxDistance)
		}
	}

	dst.Width = width
	dst.Height = height
	dst.CellSize = float32(cellSize)
	dst.MaxDistance = maxDistance
	dst.HasGeometry = hasGeometry
	return dst
}

func buildLightShadowMaps(dst []lightShadowMap, lights []graphics.Light2D, segments []shadowSegment, resolution int) []lightShadowMap {
	dst = dst[:0]
	if resolution <= 0 {
		resolution = defaultShadowMapResolution
	}
	for lightIndex, light := range lights {
		if !light.CastShadows || light.Radius <= 0 {
			continue
		}
		culled := cullSegmentsForLight(nil, segments, light)
		depths := make([]float32, resolution)
		for i := range depths {
			depths[i] = 1
		}
		shadowMap := lightShadowMap{
			LightIndex:  lightIndex,
			Resolution:  resolution,
			Radius:      light.Radius,
			Segments:    append([]shadowSegment(nil), culled...),
			Depths:      depths,
			HasOccluder: len(culled) > 0,
		}
		writeLightShadowDepths(&shadowMap, light)
		dst = append(dst, shadowMap)
	}
	return dst
}

func cullSegmentsForLight(dst []shadowSegment, segments []shadowSegment, light graphics.Light2D) []shadowSegment {
	dst = dst[:0]
	radius := max0(light.Radius)
	if radius == 0 {
		return dst
	}
	for _, segment := range segments {
		if distancePointToSegment(light.Position, segment.A, segment.B) <= radius {
			dst = append(dst, segment)
		}
	}
	return dst
}

func writeLightShadowDepths(shadowMap *lightShadowMap, light graphics.Light2D) {
	if shadowMap.Resolution == 0 || len(shadowMap.Depths) == 0 || light.Radius <= 0 {
		return
	}
	for i := 0; i < shadowMap.Resolution; i++ {
		angle := (float64(i)+0.5)/float64(shadowMap.Resolution)*2*math.Pi - math.Pi
		dir := lmath.Vec2{X: float32(math.Cos(angle)), Y: float32(math.Sin(angle))}
		nearest := light.Radius
		for _, segment := range shadowMap.Segments {
			if distance, ok := raySegmentDistance(light.Position, dir, segment.A, segment.B); ok && distance < nearest {
				nearest = distance
			}
		}
		shadowMap.Depths[i] = nearest / light.Radius
	}
}

func shadowFactorForLight(pixel lmath.Vec2, lightIndex int, lights []graphics.Light2D, shadows []lightShadowMap, sdf sdfTexture, mode graphics.ShadowMode2D) float32 {
	if lightIndex < 0 || lightIndex >= len(lights) || !lights[lightIndex].CastShadows {
		return 1
	}
	if mode == graphics.ShadowModeSDFExperimental {
		if lightIndex != firstSDFShadowLight(lights) {
			return 1
		}
		return sampleSDFShadow(pixel, lights[lightIndex], sdf)
	}
	for _, shadowMap := range shadows {
		if shadowMap.LightIndex != lightIndex {
			continue
		}
		return sampleShadowMap(pixel, lights[lightIndex], shadowMap)
	}
	return 1
}

func combinedShadowFactor(pixel lmath.Vec2, lights []graphics.Light2D, shadows []lightShadowMap, sdf sdfTexture, mode graphics.ShadowMode2D) float32 {
	shadowingLights := 0
	sum := float32(0)
	for i, light := range lights {
		if !light.CastShadows {
			continue
		}
		if mode == graphics.ShadowModeSDFExperimental && i != firstSDFShadowLight(lights) {
			continue
		}
		shadowingLights++
		sum += shadowFactorForLight(pixel, i, lights, shadows, sdf, mode)
	}
	if shadowingLights == 0 {
		return 1
	}
	return sum / float32(shadowingLights)
}

func firstSDFShadowLight(lights []graphics.Light2D) int {
	for i, light := range lights {
		if light.CastShadows && light.Radius > 0 {
			return i
		}
	}
	return -1
}

func sampleSDFShadow(pixel lmath.Vec2, light graphics.Light2D, sdf sdfTexture) float32 {
	if !sdf.HasGeometry || sdf.Width <= 0 || sdf.Height <= 0 || len(sdf.Pixels) == 0 || light.Radius <= 0 {
		return 1
	}
	ray := pixel.Sub(light.Position)
	distanceToPixel := vecLength(ray)
	if distanceToPixel <= shadowBiasPixels || distanceToPixel > light.Radius {
		return 1
	}
	dir := ray.MulScalar(1 / distanceToPixel)
	softness := max0(sdf.CellSize * 3)
	if softness == 0 {
		softness = 12
	}
	shadow := float32(1)
	travel := sdfViewportEntryTravel(light.Position, dir, distanceToPixel, sdf)
	if travel < 0 {
		return 1
	}
	for steps := 0; steps < 64 && travel < distanceToPixel; steps++ {
		samplePoint := light.Position.Add(dir.MulScalar(travel))
		distance := sampleSDFTexture(sdf, samplePoint)
		if distance <= sdf.CellSize*0.75 {
			return 0
		}
		if travel > 0 {
			shadow = min32(shadow, clamp32(distance/softness, 0, 1))
		}
		travel += max32(distance*0.8, sdf.CellSize*0.5)
	}
	return clamp32(shadow, 0, 1)
}

func sdfViewportEntryTravel(origin, dir lmath.Vec2, maxTravel float32, sdf sdfTexture) float32 {
	minX := float32(0)
	minY := float32(0)
	maxX := float32(sdf.Width) * sdf.CellSize
	maxY := float32(sdf.Height) * sdf.CellSize
	tMin := float32(0)
	tMax := maxTravel

	if !clipRayAxis(origin.X, dir.X, minX, maxX, &tMin, &tMax) {
		return -1
	}
	if !clipRayAxis(origin.Y, dir.Y, minY, maxY, &tMin, &tMax) {
		return -1
	}
	if tMax < 0 || tMin > maxTravel {
		return -1
	}
	if tMin > 0 {
		tMin += min32(sdf.CellSize*0.25, 1)
	}
	return max32(tMin, shadowBiasPixels)
}

func clipRayAxis(origin, dir, minValue, maxValue float32, tMin, tMax *float32) bool {
	if abs32(dir) < 0.00001 {
		return origin >= minValue && origin <= maxValue
	}
	t1 := (minValue - origin) / dir
	t2 := (maxValue - origin) / dir
	if t1 > t2 {
		t1, t2 = t2, t1
	}
	if t1 > *tMin {
		*tMin = t1
	}
	if t2 < *tMax {
		*tMax = t2
	}
	return *tMin <= *tMax
}

func sampleSDFTexture(sdf sdfTexture, point lmath.Vec2) float32 {
	if point.X < 0 || point.Y < 0 || point.X >= float32(sdf.Width)*sdf.CellSize || point.Y >= float32(sdf.Height)*sdf.CellSize {
		return sdf.MaxDistance
	}
	x := int(point.X / sdf.CellSize)
	y := int(point.Y / sdf.CellSize)
	if x < 0 || y < 0 || x >= sdf.Width || y >= sdf.Height {
		return sdf.MaxDistance
	}
	return sdf.Pixels[y*sdf.Width+x]
}

func pointInPolygon(point lmath.Vec2, polygon []lmath.Vec2) bool {
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		pi := polygon[i]
		pj := polygon[j]
		if ((pi.Y > point.Y) != (pj.Y > point.Y)) && (point.X < (pj.X-pi.X)*(point.Y-pi.Y)/(pj.Y-pi.Y)+pi.X) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func sampleShadowMap(pixel lmath.Vec2, light graphics.Light2D, shadowMap lightShadowMap) float32 {
	if !shadowMap.HasOccluder || shadowMap.Resolution <= 0 || len(shadowMap.Depths) == 0 || light.Radius <= 0 {
		return 1
	}
	delta := pixel.Sub(light.Position)
	distance := vecLength(delta)
	if distance <= shadowBiasPixels {
		return 1
	}
	if distance > light.Radius {
		return 1
	}
	angle := math.Atan2(float64(delta.Y), float64(delta.X))
	u := (angle + math.Pi) / (2 * math.Pi)
	index := int(u * float64(shadowMap.Resolution))
	if index >= shadowMap.Resolution {
		index = shadowMap.Resolution - 1
	}
	if index < 0 {
		index = 0
	}
	nearest := shadowMap.Depths[index] * light.Radius
	if nearest >= light.Radius {
		return 1
	}
	if distance <= nearest+shadowBiasPixels {
		return 1
	}
	return 0
}

func raySegmentDistance(origin, dir, a, b lmath.Vec2) (float32, bool) {
	v1 := origin.Sub(a)
	v2 := b.Sub(a)
	cross := cross2(dir, v2)
	if abs32(cross) < 0.00001 {
		return 0, false
	}
	t := cross2(v2, v1) / cross
	u := cross2(dir, v1) / cross
	if t < 0 || u < 0 || u > 1 {
		return 0, false
	}
	return t, true
}

func distancePointToSegment(point, a, b lmath.Vec2) float32 {
	ab := b.Sub(a)
	lengthSquared := dot2(ab, ab)
	if lengthSquared == 0 {
		return vecLength(point.Sub(a))
	}
	t := dot2(point.Sub(a), ab) / lengthSquared
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closest := a.Add(ab.MulScalar(t))
	return vecLength(point.Sub(closest))
}

func dot2(a, b lmath.Vec2) float32 {
	return a.X*b.X + a.Y*b.Y
}

func cross2(a, b lmath.Vec2) float32 {
	return a.X*b.Y - a.Y*b.X
}

func vecLength(v lmath.Vec2) float32 {
	return float32(math.Sqrt(float64(dot2(v, v))))
}

func abs32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clamp32(value, low, high float32) float32 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

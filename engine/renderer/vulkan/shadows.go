package vulkan

import (
	"math"

	"github.com/sarattha/lumago/engine/graphics"
	lmath "github.com/sarattha/lumago/engine/math"
)

const (
	defaultShadowMapResolution = 256
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

func shadowFactorForLight(pixel lmath.Vec2, lightIndex int, lights []graphics.Light2D, shadows []lightShadowMap) float32 {
	if lightIndex < 0 || lightIndex >= len(lights) || !lights[lightIndex].CastShadows {
		return 1
	}
	for _, shadowMap := range shadows {
		if shadowMap.LightIndex != lightIndex {
			continue
		}
		return sampleShadowMap(pixel, lights[lightIndex], shadowMap)
	}
	return 1
}

func combinedShadowFactor(pixel lmath.Vec2, lights []graphics.Light2D, shadows []lightShadowMap) float32 {
	shadowingLights := 0
	sum := float32(0)
	for i, light := range lights {
		if !light.CastShadows {
			continue
		}
		shadowingLights++
		sum += shadowFactorForLight(pixel, i, lights, shadows)
	}
	if shadowingLights == 0 {
		return 1
	}
	return sum / float32(shadowingLights)
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

package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/veandco/go-sdl2/gfx"

	"github.com/veandco/go-sdl2/sdl"
)

var planets []planet

var winTitle string = "Projekt infa (Go-SDL2)"
var winWidth, winHeight int32 = 800, 600
var window *sdl.Window
var renderer *sdl.Renderer

var mousePos coord
var isCreatingPlanet bool

func run() int {
	var event sdl.Event
	var running bool
	var err error

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		window.Destroy()
	}()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer func() {
		renderer.Destroy()
	}()

	renderer.Clear()
	
	running = true
	isCreatingPlanet = false
	for running {
		if isCreatingPlanet {
			planet := &planets[len(planets)-1]
			planet.radius++
		}

		simulation()

		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				mousePos.x = float64(t.X)
				mousePos.y = float64(t.Y)
			case *sdl.MouseButtonEvent:
				if t.State == 1 {
					isCreatingPlanet = true
					newPlanet(t)
				} else {
					isCreatingPlanet = false
				}
				fmt.Printf("[%d ms] MouseButton\ttype:%d\tid:%d\tx:%d\ty:%d\tbutton:%d\tstate:%d\n",
					t.Timestamp, t.Type, t.Which, t.X, t.Y, t.Button, t.State)
			}
		}
		renderer.Present()

		sdl.Delay(1000 / 30)
	}

	return 0
}

type coord struct {
	x float64
	y float64
}

func (c coord) normalize() coord {
	normalized := coord{
		x: c.x / math.Abs(c.x),
		y: c.y / math.Abs(c.y),
	}

	return normalized
}

type planet struct {
	id        int32
	pos       coord
	vel       coord
	radius    float64
	color     sdl.Color
	destroyed bool
}

func newPlanet(t *sdl.MouseButtonEvent) *planet {
	newPlanet := planet{
		id: int32(len(planets)),
		pos: coord{
			x: float64(t.X),
			y: float64(t.Y),
		},
		vel: coord{
			x: 0,
			y: 0,
		},
		radius: 2,
		color: sdl.Color{
			R: uint8(rand.Intn(255)),
			G: uint8(rand.Intn(255)),
			B: uint8(rand.Intn(255)),
			A: 255,
		},
	}

	planets = append(planets, newPlanet)

	return &newPlanet
}

func (p planet) setRadius(newRadius float64) {
	p.radius = newRadius
}

func (p planet) getMass() float64 {
	mass := (4 / 3) * math.Pi * math.Pow(p.radius, 3) * 5.52 * math.Pow10(3)

	return mass
}

func (p planet) getDistanceTo(target coord) float64 {
	distance := math.Sqrt(math.Pow(target.x-p.pos.x, 2) + math.Pow(target.y-p.pos.y, 2))

	return distance
}

func (p planet) getDirTo(target coord) coord {
	dir := coord{
		x: target.x - p.pos.x,
		y: target.y - p.pos.y,
	}

	return dir
}

func simulation() {
	G_CONST := 6.67e-11

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	for _, pl := range planets {
		p := &planets[pl.id]

		if p.destroyed {
			continue
		}

		if isCreatingPlanet && int(p.id) == len(planets)-1 {
			gfx.AALineRGBA(renderer, int32(p.pos.x), int32(p.pos.y), int32(mousePos.x), int32(mousePos.y), 255, 255, 255, 255)
			p.vel.x = -p.getDirTo(mousePos).x / 100
			p.vel.y = -p.getDirTo(mousePos).y / 100
		}

		for _, ta := range planets {
			t := &planets[ta.id]

			if p.id == t.id || t.destroyed {
				continue
			} else if p.getDistanceTo(t.pos) <= p.radius+t.radius {
				p.destroyed = true
				t.destroyed = true

				newPlanet := planet{
					id: int32(len(planets)),
					pos: coord{
						x: float64(p.pos.x),
						y: float64(p.pos.y),
					},
					vel: coord{
						x: p.vel.x + ((t.getMass() * t.vel.x) / (t.getMass() + p.getMass())),
						y: p.vel.y + ((t.getMass() * t.vel.y) / (t.getMass() + p.getMass())),
					},
					radius: math.Sqrt(math.Pow(p.radius, 2) + math.Pow(t.radius, 2)),
					color: sdl.Color{
						R: p.color.R + t.color.R,
						G: p.color.G + t.color.G,
						B: p.color.B + t.color.B,
						A: 255,
					},
				}

				planets = append(planets, newPlanet)
				continue
			}

			gravForce := G_CONST * ((p.getMass() * t.getMass()) / (math.Pow(p.getDistanceTo(t.pos), 2)))

			p.vel.x += (gravForce / p.getMass()) * p.getDirTo(t.pos).x * 100
			p.vel.y += (gravForce / p.getMass()) * p.getDirTo(t.pos).y * 100
		}

		if !isCreatingPlanet || !(isCreatingPlanet && int(p.id) == len(planets)-1) {
			p.pos.x += p.vel.x
			p.pos.y += p.vel.y
		}

		gfx.FilledCircleColor(renderer, int32(p.pos.x), int32(p.pos.y), int32(p.radius), p.color)
	}
}

func main() {
	var exitcode int
	exitcode = run()

	os.Exit(exitcode)
}

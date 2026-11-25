//go:build ignore

// Breakout in the browser, written in plain Go against the "dom" binding that
// go2js ships (made from TypeScript's lib.dom.d.ts). No JavaScript is written.
//
//	go2js -with dom -o breakout.html examples/browser-game/main.go
//
// Open breakout.html. Left/Right (or A/D) move the paddle, Space starts and
// restarts.
package main

import (
	"dom"
	"fmt"
)

const (
	width, height = 480, 360
	cols, rows    = 8, 5
	brickW        = width / cols
	brickH        = 20
)

type game struct {
	px, bx, by, vx, vy float64
	left, right        bool
	bricks             [rows][cols]bool
	score, lives       int
	state              string // "ready", "play", "over", "won"
}

func (g *game) reset() {
	for r := range g.bricks {
		for c := range g.bricks[r] {
			g.bricks[r][c] = true
		}
	}
	g.score, g.lives = 0, 3
	g.serve()
}

func (g *game) serve() {
	g.px = width/2 - 40
	g.bx, g.by = width/2, height-40
	g.vx, g.vy = 3, -4
	g.state = "ready"
}

func (g *game) step() {
	if g.left {
		g.px -= 6
	}
	if g.right {
		g.px += 6
	}
	g.px = max(0, min(width-80, g.px))
	if g.state != "play" {
		return
	}
	g.bx += g.vx
	g.by += g.vy
	if g.bx < 4 || g.bx > width-4 {
		g.vx = -g.vx
	}
	if g.by < 4 {
		g.vy = -g.vy
	}
	if g.vy > 0 && g.by > height-24 && g.by < height-10 && g.bx > g.px && g.bx < g.px+80 {
		g.vy = -g.vy
		g.vx = (g.bx - (g.px + 40)) / 8
	}
	if r, c := int(g.by)/brickH, int(g.bx)/brickW; r >= 0 && r < rows && c >= 0 && c < cols && g.bricks[r][c] {
		g.bricks[r][c] = false
		g.vy = -g.vy
		g.score += 10
		if g.score == 10*rows*cols {
			g.state = "won"
		}
	}
	if g.by > height {
		g.lives--
		if g.lives == 0 {
			g.state = "over"
		} else {
			g.serve()
		}
	}
}

func (g *game) draw(ctx *dom.CanvasRenderingContext2D) {
	ctx.SetFillStyle("#101820")
	ctx.FillRect(0, 0, width, height)
	colors := []string{"#e63946", "#f4a261", "#e9c46a", "#2a9d8f", "#457b9d"}
	for r := range g.bricks {
		for c := range g.bricks[r] {
			if g.bricks[r][c] {
				ctx.SetFillStyle(colors[r])
				ctx.FillRect(float64(c*brickW+1), float64(r*brickH+1), brickW-2, brickH-2)
			}
		}
	}
	ctx.SetFillStyle("#f1faee")
	ctx.FillRect(g.px, height-16, 80, 8)
	ctx.FillRect(g.bx-4, g.by-4, 8, 8)
	ctx.SetFont("14px sans-serif")
	ctx.SetFillStyle("#f1faee")
	ctx.FillText(fmt.Sprintf("score %d   lives %d", g.score, g.lives), 8, height-4)
	switch g.state {
	case "ready":
		ctx.FillText("press Space", width/2-36, height/2)
	case "over":
		ctx.FillText("game over - press Space", width/2-72, height/2)
	case "won":
		ctx.FillText("you win - press Space", width/2-68, height/2)
	}
}

func main() {
	canvas := dom.DocumentValue.CreateElement("canvas").AsHTMLCanvasElement()
	canvas.SetId("game")
	canvas.SetWidth(width)
	canvas.SetHeight(height)
	dom.DocumentValue.Body().AppendChild(canvas)
	ctx := canvas.GetContextX2d()

	g := &game{}
	g.reset()

	dom.WindowValue.OnKeydown(func(e *dom.KeyboardEvent) {
		switch e.Key() {
		case "ArrowLeft", "a":
			g.left = true
		case "ArrowRight", "d":
			g.right = true
		case " ":
			if g.state == "ready" {
				g.state = "play"
			} else if g.state != "play" {
				g.reset()
			}
		}
	})
	dom.WindowValue.OnKeyup(func(e *dom.KeyboardEvent) {
		switch e.Key() {
		case "ArrowLeft", "a":
			g.left = false
		case "ArrowRight", "d":
			g.right = false
		}
	})

	var frame func(t float64)
	frame = func(t float64) {
		g.step()
		g.draw(ctx)
		println("state", g.state, "score", g.score, "lives", g.lives)
		dom.RequestAnimationFrame(frame)
	}
	dom.RequestAnimationFrame(frame)
}

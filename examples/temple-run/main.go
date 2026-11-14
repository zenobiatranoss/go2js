//go:build ignore

package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 450
	screenHeight = 700
	laneWidth    = 100
)

// إحداثيات المسارات الثلاثة (يسار، وسط، يمين)
var lanes = [3]float32{
	screenWidth/2 - laneWidth,
	screenWidth / 2,
	screenWidth/2 + laneWidth,
}

type Entity struct {
	lane   int
	y      float32
	isCoin bool // true = عملة ذهبية, false = عقبة (صخرة)
}

type Game struct {
	playerLane   int
	playerX      float32
	playerY      float32
	isJumping    bool
	jumpTimer    int
	entities     []Entity
	speed        float32
	score        int
	coins        int
	gameOver     bool
	spawnCounter int
}

func NewGame() *Game {
	return &Game{
		playerLane: 1, // البدء في المسار الأوسط
		playerX:    lanes[1],
		playerY:    screenHeight - 120,
		speed:      6.0,
	}
}

func (g *Game) Update() error {
	// إعادة التشغيل عند الخسارة
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			*g = *NewGame()
		}
		return nil
	}

	// التحكم بالحركة يميناً ويساراً
	if (inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA)) && g.playerLane > 0 {
		g.playerLane--
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD)) && g.playerLane < 2 {
		g.playerLane++
	}

	// التحكم بالقفز لتفادي العقبات
	if (inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) && !g.isJumping {
		g.isJumping = true
		g.jumpTimer = 28 // مدة القفزة بالإطارات
	}

	if g.isJumping {
		g.jumpTimer--
		if g.jumpTimer <= 0 {
			g.isJumping = false
		}
	}

	// حركة سلسة للاعب بين المسارات
	targetX := lanes[g.playerLane]
	g.playerX += (targetX - g.playerX) * 0.25

	// زيادة النقاط والسرعة تدريجياً
	g.score++
	if g.score%300 == 0 {
		g.speed += 0.5
	}

	// توليد العقبات والعملات عشوائياً
	g.spawnCounter++
	spawnRate := int(350 / g.speed)
	if spawnRate < 22 {
		spawnRate = 22
	}
	if g.spawnCounter >= spawnRate {
		g.spawnCounter = 0
		lane := rand.Intn(3)
		isCoin := rand.Float32() < 0.45 // 45% احتمال عملة، 55% عقبة
		g.entities = append(g.entities, Entity{
			lane:   lane,
			y:      -40,
			isCoin: isCoin,
		})
	}

	// تحديث مواقع العقبات والعملات والتحقق من التصادم
	var activeEntities []Entity
	for _, e := range g.entities {
		e.y += g.speed

		// التحقق من التصادم مع اللاعب
		if e.lane == g.playerLane && e.y > g.playerY-25 && e.y < g.playerY+25 {
			if e.isCoin {
				g.coins++
				g.score += 50
				continue // إزالة العملة بعد جمعها
			} else if !g.isJumping {
				// اصطدام بالعقبة واللاعب ليس في حالة قفز
				g.gameOver = true
			}
		}

		if e.y < screenHeight+50 {
			activeEntities = append(activeEntities, e)
		}
	}
	g.entities = activeEntities

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 1. خلفية الغابة/المعبد
	screen.Fill(color.RGBA{34, 48, 38, 255})

	// 2. رسم طريق المعبد الحجري (3 مسارات)
	roadWidth := float32(laneWidth * 3)
	roadX := float32(screenWidth/2) - roadWidth/2
	vector.DrawFilledRect(screen, roadX, 0, roadWidth, screenHeight, color.RGBA{110, 102, 89, 255}, false)

	// خطوط فاصلة بين المسارات
	for i := 1; i < 3; i++ {
		lineX := roadX + float32(i*laneWidth)
		vector.StrokeLine(screen, lineX, 0, lineX, screenHeight, 2, color.RGBA{85, 78, 68, 255}, false)
	}

	// 3. رسم العقبات والعملات
	for _, e := range g.entities {
		ex := lanes[e.lane]
		if e.isCoin {
			// عملة ذهبية
			vector.DrawFilledCircle(screen, ex, e.y, 12, color.RGBA{255, 215, 0, 255}, true)
		} else {
			// عقبة (صخرة حمراء/بنية)
			vector.DrawFilledRect(screen, ex-28, e.y-18, 56, 36, color.RGBA{180, 45, 45, 255}, false)
		}
	}

	// 4. رسم اللاعب (يكبر حجمه عند القفز ليعطي إيحاء الارتفاع)
	playerSize := float32(22)
	playerColor := color.RGBA{50, 180, 255, 255}
	if g.isJumping {
		playerSize = 30
		playerColor = color.RGBA{120, 220, 255, 255}
	}
	// ظل اللاعب
	vector.DrawFilledCircle(screen, g.playerX+4, g.playerY+6, 20, color.RGBA{0, 0, 0, 90}, true)
	// جسم اللاعب
	vector.DrawFilledCircle(screen, g.playerX, g.playerY, playerSize, playerColor, true)

	// 5. واجهة النقاط والتعليمات
	hud := fmt.Sprintf("Score: %d  |  Coins: %d\nControls: Left/Right (A/D) | Jump: Space/Up", g.score/10, g.coins)
	ebitenutil.DebugPrintAt(screen, hud, 15, 15)

	if g.gameOver {
		vector.DrawFilledRect(screen, 50, screenHeight/2-50, screenWidth-100, 100, color.RGBA{0, 0, 0, 210}, false)
		msg := fmt.Sprintf("       GAME OVER!\n   Final Score: %d | Coins: %d\n  Press [SPACE] or [R] to Restart", g.score/10, g.coins)
		ebitenutil.DebugPrintAt(screen, msg, 95, screenHeight/2-25)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Go Temple Run")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenobiatranoss/go2js/compiler"
)

const ebitenProgram = `package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	frames int
	drawn  int
	pos    []color.RGBA
}

func (g *Game) Update() error {
	g.frames++
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		fmt.Println("space")
	}
	if g.frames == 5 {
		return ebiten.Termination
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawn++
	screen.Fill(color.RGBA{10, 20, 30, 255})
	vector.DrawFilledRect(screen, 1, 2, 3, 4, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, 5, 5, 2, color.NRGBA{0, 255, 0, 128}, true)
	ebitenutil.DebugPrintAt(screen, "hi", 0, 0)
}

func (g *Game) Layout(w, h int) (int, int) { return 320, 240 }

func main() {
	g := &Game{}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("test")
	err := ebiten.RunGame(g)
	fmt.Println("frames", g.frames, "drawn", g.drawn, "err", err)
}
`

func TestEbitenProgramRunsWithoutADisplay(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte(ebitenProgram), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	script := filepath.Join(dir, "main.js")

	if err := os.WriteFile(script, []byte(output), 0644); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("node", script)
	command.Env = append(os.Environ(), "GO2JS_EBITEN_FRAMES=100")

	got, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}

	// the game ends itself on its fifth step, and a step that ends the game is
	// not drawn
	if want := "frames 5 drawn 4 err <nil>\n"; string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}

	if strings.Contains(output, "go2jsEbiten") == false {
		t.Fatal("the ebiten runtime was not carried")
	}
}

func TestProgramWithoutEbitenDoesNotCarryItsRuntime(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")

	if err := os.WriteFile(source, []byte("package main\nimport \"fmt\"\nfunc main() { fmt.Println(1) }\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := compiler.CompileFile(source)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "go2jsEbiten") || strings.Contains(output, "go2jsColor") {
		t.Fatal("a program that does not use ebiten carries its runtime")
	}
}

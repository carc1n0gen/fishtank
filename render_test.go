package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func simulationScreen(t *testing.T, width, height int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(width, height)
	return screen
}

func screenRow(screen tcell.Screen, y int) string {
	width, _ := screen.Size()
	var row strings.Builder
	for x := 0; x < width; x++ {
		ch, _, _ := screen.Get(x, y)
		row.WriteString(ch)
	}
	return row.String()
}

func TestDrawASCIIAndClearOldFrame(t *testing.T) {
	screen := simulationScreen(t, 30, 10)
	tank := testTank(30, 10)
	tank.plants = nil
	tank.fish = []fish{{x: 2, depth: 3, kind: 4, direction: 1}}
	tank.draw(screen)
	if got := screenRow(screen, 3); !strings.Contains(got, "><>") {
		t.Fatalf("fish not drawn: %q", got)
	}
	tank.fish[0].x = 15
	tank.draw(screen)
	if got := screenRow(screen, 3); got[2:8] != "      " {
		t.Fatalf("old fish left a trail: %q", got)
	}
	if got := screenRow(screen, 0); got != "+"+strings.Repeat("-", 28)+"+" {
		t.Fatalf("tank does not fill screen width: %q", got)
	}
	for y := 0; y < 10; y++ {
		for x := 0; x < 30; x++ {
			ch, style, width := screen.Get(x, y)
			if len(ch) != 1 || ch[0] < 32 || ch[0] > 126 || width != 1 {
				t.Fatalf("non-ASCII cell at %d,%d: %q", x, y, ch)
			}
			if style != tcell.StyleDefault {
				t.Fatalf("non-default style at %d,%d", x, y)
			}
		}
	}
}

func TestPutTextUnicode(t *testing.T) {
	tests := []struct {
		name  string
		x     int
		text  string
		cells map[int]string
	}{
		{"degree eye right", 1, "><((\u00b0>", map[int]string{0: " ", 1: ">", 5: "\u00b0", 6: ">", 7: " "}},
		{"degree eye left", 1, "<\u00b0))><", map[int]string{1: "<", 2: "\u00b0", 3: ")", 6: "<", 7: " "}},
		{"left clipping", -1, "\u00b0>", map[int]string{0: ">", 1: " "}},
		{"right clipping", 9, "\u00b0>", map[int]string{8: " ", 9: "\u00b0"}},
		{"wide character", 1, "\u754c>", map[int]string{1: "\u754c", 3: ">", 4: " "}},
		{"wide clipping", 9, "\u754c", map[int]string{9: " "}},
		{"combining character", 1, "e\u0301>", map[int]string{1: "e\u0301", 2: ">", 3: " "}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := simulationScreen(t, 10, 1)
			putText(screen, tt.x, 0, tt.text)
			for x, want := range tt.cells {
				if got, _, _ := screen.Get(x, 0); got != want {
					t.Fatalf("column %d: got %q; want %q", x, got, want)
				}
			}
		})
	}
}

func TestResizeRenderingUsesScreenDimensions(t *testing.T) {
	screen := simulationScreen(t, 80, 24)
	tank := testTank(80, 24)
	tank.draw(screen)
	for _, size := range [][2]int{{30, 10}, {40, 4}, {1, 1}, {0, 0}, {100, 30}} {
		screen.SetSize(size[0], size[1])
		// Event dimensions can be stale when several resizes arrive together.
		exit, err := handleEvent(screen, tank, tcell.NewEventResize(999, 999))
		if err != nil || exit {
			t.Fatalf("resize returned exit=%v error=%v", exit, err)
		}
		if tank.width != size[0] || tank.height != size[1] {
			t.Fatalf("did not use actual dimensions: %dx%d", tank.width, tank.height)
		}
		if tank.usable() {
			if got := screenRow(screen, size[1]-1); got != "+"+strings.Repeat("-", size[0]-2)+"+" {
				t.Fatalf("bottom border after resize: %q", got)
			}
		} else if size[0] == 40 {
			if got := screenRow(screen, size[1]/2); !strings.Contains(got, "Resize to at least 20x8") {
				t.Fatalf("missing resize prompt: %q", got)
			}
			if got := screenRow(screen, 0); strings.TrimSpace(got) != "" {
				t.Fatalf("stale tank border in small screen: %q", got)
			}
		}
	}
}

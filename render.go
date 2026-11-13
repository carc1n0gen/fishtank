package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

type canvas interface {
	Size() (int, int)
	Clear()
	SetContent(int, int, rune, []rune, tcell.Style)
}

func putText(screen canvas, x, y int, text string) {
	width, height := screen.Size()
	if y < 0 || y >= height {
		return
	}
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		cellWidth := graphemes.Width()
		if x >= 0 && x+cellWidth <= width {
			runes := graphemes.Runes()
			screen.SetContent(x, y, runes[0], runes[1:], tcell.StyleDefault)
		}
		x += cellWidth
	}
}

func (t *tank) draw(screen canvas) {
	screen.Clear()
	if !t.usable() {
		width, height := screen.Size()
		message := "Resize to at least 20x8"
		putText(screen, max(0, (width-len(message))/2), height/2, message)
		return
	}

	for x := 1; x < t.width-1; x++ {
		putText(screen, x, 0, "-")
		putText(screen, x, t.height-1, "-")
	}
	for y := 1; y < t.height-1; y++ {
		putText(screen, 0, y, "|")
		putText(screen, t.width-1, y, "|")
	}
	putText(screen, 0, 0, "+")
	putText(screen, t.width-1, 0, "+")
	putText(screen, 0, t.height-1, "+")
	putText(screen, t.width-1, t.height-1, "+")

	for _, p := range t.plants {
		for level := 0; level <= p.height; level++ {
			x := t.plantX(p, level)
			ch := "|"
			if level > 0 {
				switch previous := t.plantX(p, level-1); {
				case x < previous:
					ch = "\\"
				case x > previous:
					ch = "/"
				}
			}
			putText(screen, x, t.height-2-level, ch)
		}
	}
	for _, f := range t.fish {
		putText(screen, int(f.x), t.fishRow(f), f.sprite())
	}
}

package main

import (
	"fmt"
	"image"
)

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Editor struct {
	Name   string
	Map    *Map
	Camera image.Rectangle
	Hover  image.Point
	Tile   image.Point // Tile we are hovering
	Cell   Cell
	Scale  int
	Error  error
	Midget Midget // Child mini widgets
}

func (e Editor) Draw(screen *ebiten.Image) {
	if e.Map != nil {
		e.Map.Render(screen, e.Camera)
	}
	if e.Error != nil {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Error: %s", e.Error),
			e.Map.Width*e.Map.Tw, 10,
		)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: (%d,%d): %d %d",
			e.Name, e.Hover.X, e.Hover.Y, e.Cell.Index, e.Cell.Flag,
		), e.Map.Width*e.Map.Tw, 10)
	}
	e.Midget.Draw(screen)
}

func (e Editor) Layout(w, h int) (rw, th int) {
	e.Midget.Layout(w, h)
	return e.Camera.Dx() / e.Scale, e.Camera.Dy() / e.Scale
}

const HELP = `HELP
Mouse: Draw, select, drag pop up panes.
Pause: Exit witout saving.
F1: This help.
F2: Save map in mashite format.
F3: Show tile image, can click to select.
Mouse Wheel: Select tile index.
H, V: Horizontal and Vertical flip
Y: copy hovered tile
Enter: Confirm dialogs.
Esc: Cancel dialogs.
`

func (e *Editor) Update() error {
	var err error
	e.Hover = image.Pt(ebiten.CursorPosition())
	e.Tile = e.Map.ToTile(e.Hover, e.Camera)

	_, wheel := ebiten.Wheel()
	if wheel > 0 {
		e.Cell.Index++
	} else if wheel < 0 {
		e.Cell.Index = max(0, e.Cell.Index-1)
	}

	err = e.Midget.Update()
	if err == nil {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyPause):
			if len(e.Midget.Kids) < 1 {
				e.Midget.YesNo(50, 50, 250, 100, "Quit", "Y",
					func(resp bool) {
						e.Midget.Done = resp
					},
				)
			}
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			e.Cell = e.Map.Get(e.Tile)
		case inpututil.IsKeyJustPressed(ebiten.KeyH):
			e.Cell.Flag ^= FlagHorizontalFlip
		case inpututil.IsKeyJustPressed(ebiten.KeyV):
			e.Cell.Flag ^= FlagVerticalFlip
		case inpututil.IsKeyJustPressed(ebiten.KeyF10):
			e.Error = nil
		case inpututil.IsKeyJustPressed(ebiten.KeyF1):
			e.Midget.Ask(100, 0, 250, 200, HELP, "", func(name string) {})
		case inpututil.IsKeyJustPressed(ebiten.KeyF2):
			e.Midget.Ask(50, 50, 250, 100, "Save As", e.Name,
				func(name string) {
					e.Error = e.Map.Save(name)
					if e.Error == nil {
						e.Name = name
					}
				},
			)
		case inpututil.IsKeyJustPressed(ebiten.KeyF3):
			e.Midget.Tile(200, 100, e.Map.Surface, func(x, y int) {
				_, h := e.Map.Surface.Size()
				idx := x + y*(h/e.Map.Th)
				if idx > 255 {
					idx -= 255
					e.Cell.Flag |= FlagExtended
				}
				e.Cell.Index = byte(max(0, idx))
			})
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
			e.Map.Put(e.Tile, e.Cell)
		default:
		}
	} else {
		if err == MidgetOK { // input handled by some active Midget.
			err = nil
		}
	}

	if e.Midget.Done {
		return Termination
	}
	return err
}

func NewEditor(tm *Map, name string, w, h, scale int) *Editor {
	e := &Editor{Map: tm, Name: name, Camera: image.Rect(0, 0, w, h),
		Scale:  scale,
		Midget: MakeMidget(image.Rect(0, 0, 0, 0)),
	}
	e.Midget.Lock = true
	return e
}

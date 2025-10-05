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
		if e.Tile.In(image.Rect(0, 0, e.Map.Width-1, e.Map.Height-1)) {
			e.Midget.Style.DrawCursor(screen,
				Bounds(e.Tile.X*e.Map.Tw, e.Tile.Y*e.Map.Th, e.Map.Tw, e.Map.Th).Add(e.Camera.Min),
			)
		}
	}

	kl := len(e.Midget.Kids)
	if e.Error != nil {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Error: %s [%d]", e.Error, kl),
			e.Map.Width*e.Map.Tw, 10,
		)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: (%d,%d): %d %d [%d]",
			e.Name, e.Hover.X, e.Hover.Y, e.Cell.Index, e.Cell.Flag, kl,
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
		case inpututil.IsKeyJustPressed(ebiten.KeyN):
			e.Cell.Flag ^= FlagOnTop
		case inpututil.IsKeyJustPressed(ebiten.KeyB):
			e.Cell.Flag ^= FlagSolid
		case inpututil.IsKeyJustPressed(ebiten.KeyG):
			e.Midget.AskFlag(50, 50, 250, 100, "Flag", &e.Cell.Flag)
		case inpututil.IsKeyJustPressed(ebiten.KeyF10):
			e.Error = nil
		case inpututil.IsKeyJustPressed(ebiten.KeyF1):
			e.Midget.Ask(100, 0, 250, 200, HELP, "", Accept)
		case inpututil.IsKeyJustPressed(ebiten.KeyF2):
			e.Midget.Ask(50, 50, 250, 100, "Save As", e.Name,
				func(name string) bool {
					err := e.Map.Save(name)
					e.Error = err
					e.Midget.Error(70, 70, 270, 120, err)
					if e.Error == nil {
						e.Name = name
						return true
					}
					return false
				},
			)
		case inpututil.IsKeyJustPressed(ebiten.KeyF):
			e.Midget.Ask(50, 50, 250, 100, "From", e.Map.From,
				func(name string) bool {
					err := e.Map.LoadSurface(name)
					e.Error = err
					e.Midget.Error(70, 70, 270, 120, err)
					return e.Error == nil
				},
			)
		case inpututil.IsKeyJustPressed(ebiten.KeyP):
			e.Midget.AskString(50, 50, 250, 100, "Prefix", &e.Map.Prefix)
		case inpututil.IsKeyJustPressed(ebiten.KeyO):
			e.Midget.AskInt(50, 50, 250, 100, "Offset", &e.Map.Offset)
		case inpututil.IsKeyJustPressed(ebiten.KeyS):
			e.Midget.AskInt(50, 50, 250, 100, "UI Scale", &e.Scale)
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
		case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
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

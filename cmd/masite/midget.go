package main

import "image"
import "slices"
import "fmt"
import "errors"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/vector"
import "github.com/hajimehoshi/ebiten/v2/inpututil"
import "github.com/hajimehoshi/ebiten/v2/ebitenutil"

// Midget is a small widget.
// It simply implements the ebiten.Game interface
type Midget struct {
	Kids   []Game
	Bounds Rectangle
	Style
	Done bool
	Drag bool
	Lock bool
	From Point
}

func MakeMidget(bounds Rectangle) Midget {
	return Midget{Bounds: bounds, Style: DefaultStyle()}
}

func AbsoluteMouse() Point {
	return image.Pt(ebiten.CursorPosition())
}

func (m *Midget) RelativeMouse() Point {
	return image.Pt(ebiten.CursorPosition()).Sub(m.Bounds.Min)
}

func (m *Midget) IsMouseIn() bool {
	return image.Pt(ebiten.CursorPosition()).In(m.Bounds)
}

func (m *Midget) Update() error {
	err := m.UpdateKids()
	if err != nil {
		return err
	}
	m.UpdateDrag()
	return nil
}

func (m *Midget) Draw(s *Surface) {
	m.Style.DrawBox(s, m.Bounds)
	m.DrawKids(s)
}

func (m *Midget) Add(g Game) Game {
	m.Kids = slices.Insert(m.Kids, 0, g)
	return g
}

func (m *Midget) UpdateKids() error {
	for i, kid := range m.Kids {
		if kid == nil {
			continue
		}
		err := kid.Update()
		if errors.Is(err, Termination) {
			m.Kids = slices.Delete(m.Kids, i, i+1)
		} else if errors.Is(err, MidgetOK) {
			return MidgetOK // handled by toplevel
		} else {
			return err
		}
	}
	return nil
}

func (m *Midget) DrawKids(s *Surface) {
	for i := len(m.Kids) - 1; i >= 0; i-- {
		kid := m.Kids[i]
		if kid != nil {
			kid.Draw(s)
		}
	}
}

func (m *Midget) Layout(w, h int) (rw, rh int) {
	m.Bounds.Max = m.Bounds.Min.Add(image.Pt(w, h))
	return m.Bounds.Dx(), m.Bounds.Dy()
}

// UpdateDrag allows the Midget to be dragged if not locked.
func (m *Midget) UpdateDrag() error {
	if !m.IsMouseIn() || m.Lock {
		m.Drag = false
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		m.Drag = true
		m.From = AbsoluteMouse()
	}
	if m.Drag {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			m.Drag = false
		} else {
			now := AbsoluteMouse()
			delta := now.Sub(m.From)
			m.Bounds = m.Bounds.Add(delta)
			m.From = now
		}
	}
	return nil
}

var _ Game = &Midget{}

// Style is a simplified style
type Style struct {
	Cursor RGBA
	Border RGBA
	Shadow RGBA
	Filled RGBA
	Stroke int
}

func DefaultStyle() Style {
	s := Style{}
	s.Cursor = RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55}
	s.Border = RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
	s.Shadow = RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xaa}
	s.Filled = RGBA{R: 0x00, G: 0x00, B: 0x55, A: 0xaa}
	s.Stroke = 1
	return s
}

func FillRect(Surface *Surface, r Rectangle, col RGBA) {
	vector.DrawFilledRect(
		Surface, float32(r.Min.X), float32(r.Min.Y),
		float32(r.Dx()), float32(r.Dy()),
		col, false,
	)
}

func DrawRect(Surface *Surface, r Rectangle, thick int, col RGBA) {
	vector.StrokeRect(
		Surface, float32(r.Min.X), float32(r.Min.Y),
		float32(r.Dx()), float32(r.Dy()),
		float32(thick), col, false,
	)
}

// DrawsLine draws a line on the diagonal of the Rectangle r.
func DrawLine(Surface *Surface, r Rectangle, thick int, col RGBA) {
	vector.StrokeLine(
		Surface, float32(r.Min.X), float32(r.Min.Y),
		float32(r.Max.X), float32(r.Max.Y),
		float32(thick), col, false,
	)
}

func (s Style) DrawRect(Surface *Surface, r Rectangle) {
	DrawRect(Surface, r, int(s.Stroke), s.Border)
}

func (s Style) DrawCursor(Surface *Surface, r Rectangle) {
	DrawRect(Surface, r, int(s.Stroke), s.Cursor)
}

func (s Style) DrawBox(Surface *Surface, r Rectangle) {
	if s.Shadow.A != 0 {
		shadow := s.Shadow
		shadow.A = (shadow.A / 2) + 1 // make half transparent
		right := image.Rect(r.Max.X+1, r.Min.Y+1, r.Max.X+1, r.Max.Y+1)
		DrawLine(Surface, right, 1, shadow)
		bottom := image.Rect(r.Min.X+1, r.Max.Y+1, r.Max.X+1, r.Max.Y+1)
		DrawLine(Surface, bottom, 1, shadow)
	}

	vector.DrawFilledRect(
		Surface, float32(r.Min.X), float32(r.Min.Y),
		float32(r.Dx()), float32(r.Dy()), s.Filled, false,
	)

	if s.Stroke > 0 {
		vector.StrokeRect(
			Surface, float32(r.Min.X), float32(r.Min.Y),
			float32(r.Dx()), float32(r.Dy()),
			float32(s.Stroke), s.Border, false,
		)
	}
}

func (s Style) DrawCircleInBox(Surface *Surface, box Rectangle) {
	r := box.Dx()
	if box.Dy() < r {
		r = box.Dy()
	}
	r = r / 2
	c := image.Pt((box.Min.X+box.Max.X)/2, (box.Min.Y+box.Max.Y)/2)
	s.DrawCircle(Surface, c, r)
}

func (s Style) DrawCircle(Surface *Surface, c Point, r int) {
	if r < 0 {
		r = 1
	}
	vector.DrawFilledCircle(Surface, float32(c.X), float32(c.Y),
		float32(r), s.Filled, false)

	if s.Stroke > 0 {
		vector.StrokeCircle(
			Surface, float32(c.X), float32(c.Y),
			float32(r), float32(s.Stroke), s.Border, false,
		)
	}
}

type Asker struct {
	Midget
	Prompt string
	Buf    []rune
	On     func(string)
	Cursor int
}

func Ask(bounds Rectangle, prompt, def string, on func(res string)) *Asker {
	return &Asker{Midget: MakeMidget(bounds), Prompt: prompt, On: on, Buf: []rune(def)}
}

func (a *Asker) Update() error {
	var keys []Key
	keys = inpututil.AppendJustPressedKeys(keys)
	for _, key := range keys {
		switch key {
		case ebiten.KeyEnter:
			a.On(string(a.Buf))
			return Termination
		case ebiten.KeyEscape:
			println("esc in ", a.Prompt)
			return Termination
		case ebiten.KeyBackspace:
			if len(a.Buf) > 0 {
				a.Buf = slices.Delete(a.Buf, len(a.Buf)-1, len(a.Buf))
			}
		}
	}

	var chars []rune
	chars = ebiten.AppendInputChars(chars)
	if len(chars) > 0 {
		a.Buf = append(a.Buf, chars...)
	}
	a.Midget.Update()
	return MidgetOK
}

func (a Asker) Draw(s *Surface) {
	a.Midget.Draw(s)
	ebitenutil.DebugPrintAt(s,
		fmt.Sprintf("%s>%s|", a.Prompt, string(a.Buf)),
		a.Bounds.Min.X, a.Bounds.Min.Y)
}

func Bounds(x, y, w, h int) Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

func (m *Midget) Ask(x, y, w, h int, prompt, def string, on func(res string)) *Asker {
	ask := Ask(Bounds(x, y, w, h), prompt, def, on)
	m.Add(ask)
	return ask
}

func (m *Midget) YesNo(x, y, w, h int, prompt, def string, on func(res bool)) *Asker {
	wrap := func(sres string) {
		on(sres == def)
	}
	ask := Ask(Bounds(x, y, w, h), prompt, def, wrap)
	m.Add(ask)
	return ask
}

type Tiler struct {
	Midget
	On       func(x, y int)
	Cursor   Point
	Selected Point
	Tw       int
	Th       int
	Surface  *Surface
}

func Tile(bounds Rectangle, surface *Surface, on func(x, y int)) *Tiler {
	return &Tiler{Midget: MakeMidget(bounds), Surface: surface, On: on, Tw: TW, Th: TH}
}

func (t *Tiler) Update() error {
	var keys []Key
	keys = inpututil.AppendJustPressedKeys(keys)
	for _, key := range keys {
		switch key {
		case ebiten.KeyEnter, ebiten.KeyEscape:
			return Termination
		default:
		}
	}

	if !t.IsMouseIn() {
		return nil
	}
	mouse := t.RelativeMouse()
	tile := image.Pt(mouse.X/t.Tw, mouse.Y/t.Th)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.Selected = tile
		t.On(tile.X, tile.Y)
	}
	t.Cursor = tile

	t.Midget.Update()

	return MidgetOK
}

func (t Tiler) Draw(s *Surface) {
	t.Midget.Draw(s)

	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		float64(t.Bounds.Min.X),
		float64(t.Bounds.Min.Y),
	)

	if t.Surface != nil {
		s.DrawImage(t.Surface, &opts)
	}
	t.Style.DrawCursor(s, Bounds(t.Cursor.X*t.Tw, t.Cursor.Y*t.Th, t.Tw, t.Th).Add(t.Bounds.Min))
	t.Style.DrawCursor(s, Bounds(t.Selected.X*t.Tw, t.Selected.Y*t.Th, t.Tw, t.Th).Add(t.Bounds.Min))
}

func (m *Midget) Tile(x, y int, surface *Surface, on func(x, y int)) *Tiler {
	w, h := surface.Size()
	Tile := Tile(Bounds(x, y, w, h), surface, on)
	m.Add(Tile)
	return Tile
}

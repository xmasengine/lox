package main

import "image"
import "strconv"
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
	Done   bool
	Drag   bool
	Lock   bool
	From   Point
	Frames int
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
	m.Frames++
	err := m.UpdateKids()
	if err != nil {
		return err
	}
	return m.UpdateDrag()
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
			return MidgetOK
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

func (s Style) ForError() Style {
	s.Filled = RGBA{R: 0x55, G: 0x00, B: 0x00, A: 0xaa}
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
	On     func(string) bool
	Cursor int
}

func Ask(bounds Rectangle, prompt, def string, on func(res string) bool) *Asker {
	a := &Asker{Midget: MakeMidget(bounds),
		Prompt: prompt, On: on, Buf: []rune(def),
	}
	a.Cursor = len(a.Buf)
	return a
}

func (a *Asker) Update() error {
	var keys []Key
	keys = inpututil.AppendJustPressedKeys(keys)
	for _, key := range keys {
		switch key {
		case ebiten.KeyEnter:
			done := a.On(string(a.Buf))
			if done {
				return Termination
			}
		case ebiten.KeyEnd:
			a.Cursor = len(a.Buf)
		case ebiten.KeyHome:
			a.Cursor = 0
		case ebiten.KeyLeft:
			a.Cursor = max(0, a.Cursor-1)
		case ebiten.KeyRight:
			a.Cursor = min(a.Cursor+1, len(a.Buf))
		case ebiten.KeyEscape:
			return Termination
		case ebiten.KeyDelete:
			if len(a.Buf) > 0 && a.Cursor < len(a.Buf) {
				a.Buf = slices.Delete(a.Buf, a.Cursor, a.Cursor+1)
			}
		case ebiten.KeyBackspace:
			if len(a.Buf) > 0 && a.Cursor > 0 {
				a.Buf = slices.Delete(a.Buf, a.Cursor-1, a.Cursor)
			}
			a.Cursor = max(0, a.Cursor-1)
		}
	}

	if a.Frames > 30 { // debounce previous input when the Asker is opened.
		var chars []rune
		chars = ebiten.AppendInputChars(chars)
		if len(chars) > 0 {
			a.Buf = slices.Insert(a.Buf, a.Cursor, chars...)
		}
		a.Cursor += len(chars)
	}
	a.Midget.Update()
	return MidgetOK
}

func (a Asker) Draw(s *Surface) {
	a.Midget.Draw(s)
	// Simulate a cursor with a pipe character to simplify rendering.
	ebitenutil.DebugPrintAt(s,
		fmt.Sprintf("%s>%s|%s", a.Prompt,
			string(a.Buf[0:a.Cursor]),
			string(a.Buf[a.Cursor:]),
		),
		a.Bounds.Min.X, a.Bounds.Min.Y)
}

func Bounds(x, y, w, h int) Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

func (m *Midget) Ask(x, y, w, h int, prompt, def string, on func(res string) bool) *Asker {
	ask := Ask(Bounds(x, y, w, h), prompt, def, on)
	m.Add(ask)
	return ask
}

func (m *Midget) YesNo(x, y, w, h int, prompt, def string, on func(res bool)) *Asker {
	wrap := func(sres string) bool {
		on(sres == def)
		return true
	}
	ask := Ask(Bounds(x, y, w, h), prompt, def, wrap)
	m.Add(ask)
	return ask
}

func Accept(sres string) bool { return true }

func (m *Midget) Error(x, y, w, h int, err error) *Asker {
	if err == nil {
		return nil
	}
	msg := err.Error()
	ask := Ask(Bounds(x, y, max(w, len(msg)*8), h), msg, "", Accept)
	ask.Style = ask.Style.ForError()
	m.Add(ask)
	return ask
}

func (m *Midget) AskString(x, y, w, h int, prompt string, str *string) *Asker {
	on := func(sres string) bool {
		*str = sres
		return true
	}
	return m.Ask(x, y, w, h, prompt, *str, on)
}

func (m *Midget) AskInt(x, y, w, h int, prompt string, i *int) *Asker {
	on := func(sres string) bool {
		res, err := strconv.Atoi(sres)
		if err == nil {
			*i = res
			return true
		} else {
			m.Error(x+20, y+20, w, h, err)
			return false
		}
	}
	return m.Ask(x, y, w, h, prompt, strconv.Itoa(*i), on)
}

func (m *Midget) AskByte(x, y, w, h int, prompt string, i *byte) *Asker {
	on := func(sres string) bool {
		res, err := strconv.Atoi(sres)
		if err == nil {
			*i = byte(res)
			return true
		} else {
			m.Error(x+20, y+20, w, h, err)
			return false
		}
	}
	return m.Ask(x, y, w, h, prompt, strconv.Itoa(int(*i)), on)
}

func (m *Midget) AskFlag(x, y, w, h int, prompt string, i *Flag) *Asker {
	on := func(sres string) bool {
		res, err := strconv.Atoi(sres)
		if err == nil {
			*i = Flag(res)
			return true
		} else {
			m.Error(x+20, y+20, w, h, err)
			return false
		}
	}
	return m.Ask(x, y, w, h, prompt, strconv.Itoa(int(*i)), on)
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
	t.Cursor = tile
	err := t.Midget.Update()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.Selected = tile
		t.On(tile.X, tile.Y)
		return MidgetOK
	}
	return err
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

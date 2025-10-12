package masite

import "image"
import "strconv"
import "strings"
import "slices"
import "fmt"
import "errors"
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/vector"
import "github.com/hajimehoshi/ebiten/v2/inpututil"
import "github.com/hajimehoshi/ebiten/v2/ebitenutil"

// caption is the caption of a midget
type Caption struct {
	Text   string
	Bounds Rectangle
	Close  Rectangle
}

// Midget is a small widget.
// It simply implements the ebiten.Game interface
// Midgets propagage events using update bubbling.
// A midget should return Terminate from its Update function
// to request to close itself.
// A midget should return MidgetOK from its Update function
// to request to stop other widgets from updating.
// For example, when a keypress or mouyse click was handled, etc.
type Midget struct {
	Kids   []Game
	Bounds Rectangle
	Style
	Done    bool
	Drag    bool
	Lock    bool
	From    Point
	Frames  int
	Caption Caption
}

func MakeMidget(bounds Rectangle) Midget {
	return Midget{Bounds: bounds, Style: DefaultStyle()}
}

func AbsoluteMouse() Point {
	return image.Pt(ebiten.CursorPosition())
}

func (m *Midget) RelativeMouse() Point {
	rm := image.Pt(ebiten.CursorPosition()).Sub(m.Bounds.Min)
	if m.Caption.Text != "" {
		rm = rm.Sub(image.Pt(0, m.Caption.Bounds.Dy()))
	}
	return rm
}

func (m *Midget) IsMouseIn() bool {
	return image.Pt(ebiten.CursorPosition()).In(m.Bounds)
}

func (m *Midget) IsMouseInCaption() bool {
	return image.Pt(ebiten.CursorPosition()).In(m.Caption.Bounds)
}

func (m *Midget) IsMouseInClose() bool {
	return image.Pt(ebiten.CursorPosition()).In(m.Caption.Close)
}

func (m *Midget) Update() error {
	m.Frames++
	err := m.UpdateKids()
	if err != nil {
		return err
	}
	err = m.UpdateCaption()
	if err != nil {
		return err
	}
	return m.UpdateDrag()
}

const CaptionHeight = 16
const CaptionCloseMargin = 2
const CaptionCloseSize = 14

func (m *Midget) SetCaption(text string) {
	m.Caption.Text = text
	m.Caption.Bounds = m.Bounds
	m.Caption.Bounds.Max.Y = m.Caption.Bounds.Min.Y + CaptionHeight
	m.Caption.Close = m.Caption.Bounds
	m.Caption.Close.Min.X = m.Caption.Close.Max.X - CaptionCloseSize
	m.Caption.Close = m.Caption.Close.Inset(CaptionCloseMargin)
}

func (m *Midget) Draw(s *Surface) {
	bounds := m.Bounds
	m.Style.DrawBox(s, bounds)

	if m.Caption.Text != "" {
		m.Style.DrawBox(s, m.Caption.Bounds)
		m.Style.DrawBox(s, m.Caption.Close)
		m.Style.DrawClose(s, m.Caption.Close.Inset(CaptionCloseMargin))

		ebitenutil.DebugPrintAt(s, m.Caption.Text, m.Caption.Bounds.Min.X, m.Caption.Bounds.Min.Y)
	}
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
			// drop the midget
			return MidgetOK // stop event handling for other widgets
		} else if errors.Is(err, MidgetOK) {
			return MidgetOK // stop event handling for other widgets
		} else if err != nil {
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
			if delta.Eq(Point{}) {
				return nil
			}
			m.Bounds = m.Bounds.Add(delta)
			m.Caption.Bounds = m.Caption.Bounds.Add(delta)
			m.Caption.Close = m.Caption.Close.Add(delta)
			m.From = now
			return MidgetOK
		}
	}
	return nil
}

// UpdateCaption allows the Midget's caption to be interacted with.
func (m *Midget) UpdateCaption() error {
	if !m.IsMouseInClose() {
		return nil
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return Termination
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
	Close  RGBA
	Stroke int
}

func DefaultStyle() Style {
	s := Style{}
	s.Cursor = RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55}
	s.Border = RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
	s.Shadow = RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xaa}
	s.Filled = RGBA{R: 0x00, G: 0x00, B: 0x55, A: 0xaa}
	s.Close = RGBA{R: 0xff, G: 0xaa, B: 0xaa, A: 0xaa}
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

func DrawX(surface *Surface, r Rectangle, thick int, col RGBA) {
	DrawLine(surface, r, thick, col)
	r.Min.X, r.Max.X = r.Max.X, r.Min.X
	DrawLine(surface, r, thick, col)
}

func (s Style) DrawRect(surface *Surface, r Rectangle) {
	DrawRect(surface, r, int(s.Stroke), s.Border)
}

func (s Style) DrawCursor(surface *Surface, r Rectangle) {
	DrawRect(surface, r, int(s.Stroke), s.Cursor)
}

func (s Style) DrawClose(surface *Surface, r Rectangle) {
	DrawX(surface, r, int(s.Stroke), s.Close)
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
	head, text, _ := strings.Cut(prompt, "\n")

	a := &Asker{Midget: MakeMidget(bounds),
		Prompt: text, On: on, Buf: []rune(def),
	}
	a.Cursor = len(a.Buf)
	a.SetCaption(head)
	return a
}

func (a *Asker) Update() error {
	var keys []Key
	keys = inpututil.AppendJustPressedKeys(keys)
	handled := len(keys) > 0
	for _, key := range keys {
		switch key {
		case ebiten.KeyEnter:
			done := a.On(string(a.Buf))
			if done {
				return Termination
			}
		case ebiten.KeyC:
			if inpututil.KeyPressDuration(ebiten.KeyControlLeft) > 0 ||
				inpututil.KeyPressDuration(ebiten.KeyControlRight) > 0 {
				WriteClipboardRunes(a.Buf)
				a.Midget.Update()
				return MidgetOK
			}
		case ebiten.KeyV:
			if inpututil.KeyPressDuration(ebiten.KeyControlLeft) > 0 ||
				inpututil.KeyPressDuration(ebiten.KeyControlRight) > 0 {
				clip := ReadClipboardRunes()
				if len(clip) > 0 {
					a.Buf = append(a.Buf, clip...)
					a.Cursor += len(clip)
					a.Midget.Update()
					return MidgetOK
				}
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
		case ebiten.KeyF1, ebiten.KeyF2, ebiten.KeyF3, ebiten.KeyF4, ebiten.KeyF5,
			ebiten.KeyF6, ebiten.KeyF7, ebiten.KeyF8, ebiten.KeyF9, ebiten.KeyF10,
			ebiten.KeyF11, ebiten.KeyF12, ebiten.KeyPause:
			handled = false
		default:
		}
	}

	if a.Frames > 30 { // debounce previous input when the Asker is opened.
		var chars []rune
		chars = ebiten.AppendInputChars(chars)
		if len(chars) > 0 {
			a.Buf = slices.Insert(a.Buf, a.Cursor, chars...)
			a.Cursor += len(chars)
			handled = true
		}
	}
	err := a.Midget.Update()
	if err != nil {
		return err
	}
	if handled {
		return MidgetOK
	} else {
		return nil
	}
}

func (a Asker) Draw(s *Surface) {
	a.Midget.Draw(s)
	// Simulate a cursor with a pipe character to simplify rendering.

	bounds := a.Bounds
	if a.Caption.Text != "" {
		bounds.Min.Y = a.Caption.Bounds.Max.Y
	}

	ebitenutil.DebugPrintAt(s,
		fmt.Sprintf("%s>%s|%s", a.Prompt,
			string(a.Buf[0:a.Cursor]),
			string(a.Buf[a.Cursor:]),
		),
		bounds.Min.X, bounds.Min.Y,
	)
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

func (m *Midget) AskText(x, y, w, h int, prompt string, t TextEncoding) *Asker {
	on := func(sres string) bool {
		err := t.UnmarshalText([]byte(sres))
		if err == nil {
			return true
		} else {
			m.Error(x+20, y+20, w, h, err)
			return false
		}
	}
	enc, _ := t.MarshalText()
	return m.Ask(x, y, w, h, prompt, string(enc), on)
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

func (t *Tiler) SetCaption(text string) {
	t.Midget.SetCaption(text)
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

	err := t.Midget.Update()
	if err != nil {
		return err
	}

	if !t.IsMouseIn() {
		return nil
	}
	mouse := t.RelativeMouse()
	tile := image.Pt(mouse.X/t.Tw, mouse.Y/t.Th)
	if !t.IsMouseInCaption() {
		t.Cursor = tile
	}
	if !t.IsMouseInCaption() && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.Selected = tile
		t.On(tile.X, tile.Y)
		return MidgetOK
	}
	return nil
}

func (t Tiler) Draw(s *Surface) {
	t.Midget.Draw(s)

	bounds := t.Bounds
	if t.Caption.Text != "" {
		bounds.Min.Y = t.Caption.Bounds.Max.Y
	}

	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		float64(bounds.Min.X),
		float64(bounds.Min.Y),
	)

	if t.Surface != nil {
		s.DrawImage(t.Surface, &opts)
	}
	t.Style.DrawCursor(s, Bounds(t.Cursor.X*t.Tw, t.Cursor.Y*t.Th, t.Tw, t.Th).Add(bounds.Min))
	t.Style.DrawCursor(s, Bounds(t.Selected.X*t.Tw, t.Selected.Y*t.Th, t.Tw, t.Th).Add(bounds.Min))
}

func (m *Midget) Tile(x, y int, surface *Surface, on func(x, y int)) *Tiler {
	w, h := surface.Size()
	Tile := Tile(Bounds(x, y, w, h), surface, on)
	m.Add(Tile)
	return Tile
}

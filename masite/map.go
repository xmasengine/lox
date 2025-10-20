package masite

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Format is just the lowercase extension including the '.' prefix.
type Format string

const (
	JSONFormat   Format = ".json"
	MasiteFormat Format = ".xml"
	BasicFormat  Format = ".bas"
)

func (f Format) Unmarshal(buf []byte, ptr any) error {
	switch f {
	case ".json", ".js", ".masite":
		return json.Unmarshal(buf, ptr)
	case ".xml", ".mas", ".maxite":
		return xml.Unmarshal(buf, ptr)
	default:
		return errors.New("format not supported: " + string(f))
	}
}

func (f Format) Marshal(ptr any) ([]byte, error) {
	switch f {
	case ".json", ".js", ".masite":
		return json.MarshalIndent(ptr, "", "    ")
	case ".xml", ".mas", ".maxite":
		return xml.MarshalIndent(ptr, "", "    ")
	case ".bas":
		return MarshalBasic(ptr)
	default:
		return nil, errors.New("format not supported: " + string(f))
	}
}

type Flag byte

const (
	FlagExtended       Flag = 1 << 0
	FlagHorizontalFlip Flag = 1 << 1
	FlagVerticalFlip   Flag = 1 << 2
	FlagSpritePalette  Flag = 1 << 3
	FlagOnTop          Flag = 1 << 4
	FlagSolid          Flag = 1 << 5
	FlagHarm           Flag = 1 << 6
	FlagBless          Flag = 1 << 7
)

func (f Flag) Is(s Flag) bool {
	return (f & s) == s
}

func (f Flag) Render(screen *Surface, bounds Rectangle) {

	if f.Is(FlagOnTop) {
		color := RGBA{R: 0, G: 127, B: 127, A: 128}
		DrawRect(screen, bounds, 2, color)
	}
	if f.Is(FlagSolid) {
		color := RGBA{R: 127, G: 127, B: 127, A: 240}
		DrawRect(screen, bounds.Inset(2), 2, color)
	}
	if f.Is(FlagBless) {
		color := RGBA{R: 255, G: 255, B: 0, A: 128}
		DrawRect(screen, bounds.Inset(4), 2, color)
	}

	if f.Is(FlagHarm) {
		color := RGBA{R: 255, G: 0, B: 0, A: 128}
		DrawRect(screen, bounds.Inset(6), 2, color)
	}
}

func (f Flag) Buffer() *bytes.Buffer {
	b := bytes.Buffer{}
	if f.Is(FlagExtended) {
		b.WriteRune('E')
	}
	if f.Is(FlagHorizontalFlip) {
		b.WriteRune('H')
	}
	if f.Is(FlagVerticalFlip) {
		b.WriteRune('V')
	}
	if f.Is(FlagSpritePalette) {
		b.WriteRune('s')
	}
	if f.Is(FlagOnTop) {
		b.WriteRune('T')
	}
	if f.Is(FlagSolid) {
		b.WriteRune('S')
	}
	if f.Is(FlagBless) {
		b.WriteRune('B')
	}
	if f.Is(FlagHarm) {
		b.WriteRune('h')
	}
	return &b
}

func (f Flag) String() string {
	return f.Buffer().String()
}

func (f Flag) MarshalText() (text []byte, err error) {
	return f.Buffer().Bytes(), nil
}

func (f *Flag) UnmarshalText(text []byte) error {
	if bytes.ContainsAny(text, "0123456789") {
		i, err := strconv.Atoi(string(text))
		if err != nil {
			return err
		}
		*f = Flag(i)
		return nil
	}
	v := Flag(0)
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case 'E':
			v |= FlagExtended
		case 'H':
			v |= FlagHorizontalFlip
		case 'V':
			v |= FlagVerticalFlip
		case 's':
			v |= FlagSpritePalette
		case 'T':
			v |= FlagOnTop
		case 'S':
			v |= FlagSolid
		case 'B':
			v |= FlagBless
		case 'h':
			v |= FlagHarm
		default:
			return fmt.Errorf("Unknown character in Flag: %c", text[i])
		}
	}
	*f = v
	return nil
}

type Cell struct {
	Index byte `json:"index" xml:"index,attr"`
	Flag  Flag `json:"flag" xml:"flag,attr"`
}

type Row struct {
	Cells []Cell `json:"cells" xml:"cells"`
}

type Direction int

const (
	North Direction = iota
	East
	South
	West
	In
	Out
	Up
	Down
	LastDirection
)

func (e Direction) String() string {
	switch e {
	case North:
		return "north"
	case East:
		return "east"
	case South:
		return "south"
	case West:
		return "west"
	case In:
		return "in"
	case Out:
		return "out"
	case Up:
		return "up"
	case Down:
		return "down"
	case LastDirection:
		return "last"
	default:
		return "unknown"
	}
}

func (e Direction) MarshalText() (text []byte, err error) {
	var str = e.String()
	if str == "unknown" {
		return nil, fmt.Errorf("error: Direction not known")
	}
	return []byte(str), nil
}

func (e *Direction) UnmarshalText(text []byte) error {
	v := Direction(0)
	str := string(text)
	switch str {
	case "north":
		v = North
	case "east":
		v = East
	case "south":
		v = South
	case "west":
		v = West
	case "in":
		v = In
	case "out":
		v = Out
	case "up":
		v = Up
	case "down":
		v = Down
	case "last":
		v = LastDirection
	default:
		fmt.Errorf("error: Direction not known")
	}
	*e = v
	return nil
}

type Exit struct {
	Name   string `json:"name" xml:"name,attr,omitempty"`     // Name of the linked tile map.
	Number int    `json:"number" xml:"number,attr,omitempty"` // Linked Map number in basic.
}

type Exits struct {
	North Exit `json:"north,omitempty" xml:"north,omitempty"`
	East  Exit `json:"east,omitempty" xml:"east,omitempty"`
	South Exit `json:"south,omitempty" xml:"south,omitempty"`
	West  Exit `json:"west,omitempty" xml:"west,omitempty"`
	In    Exit `json:"in,omitempty" xml:"in,omitempty"`
	Out   Exit `json:"out,omitempty" xml:"out,omitempty"`
	Up    Exit `json:"up,omitempty" xml:"up,omitempty"`
	Down  Exit `json:"down,omitempty" xml:"down,omitempty"`
}

// Kin is the type of a Presence
type Kin int

const (
	EmptyKin Kin = iota
	HumanKin
	ItemKin
	DoorKin
	StairKin
	ChestKin
	ClueKin
	SpiritKin
	FoeKin
	LastKin
)

func (e Kin) String() string {
	switch e {
	case EmptyKin:
		return "empty"
	case HumanKin:
		return "human"
	case ItemKin:
		return "item"
	case DoorKin:
		return "door"
	case StairKin:
		return "stair"
	case ChestKin:
		return "chest"
	case ClueKin:
		return "clue"
	case SpiritKin:
		return "spirit"
	case FoeKin:
		return "foe"
	case LastKin:
		return "last"
	default:
		return "unknown"
	}
}

func (e Kin) MarshalText() (text []byte, err error) {
	var str = e.String()
	if str == "unknown" {
		return nil, fmt.Errorf("error: Kin not known")
	}
	return []byte(str), nil
}

func (e *Kin) UnmarshalText(text []byte) error {
	v := Kin(0)
	str := string(text)
	switch str {
	case "empty":
		v = EmptyKin
	case "human":
		v = HumanKin
	case "item":
		v = ItemKin
	case "door":
		v = DoorKin
	case "stair":
		v = StairKin
	case "chest":
		v = ChestKin
	case "clue":
		v = ClueKin
	case "spirit":
		v = SpiritKin
	case "foe":
		v = FoeKin
	case "last":
		v = LastKin
	default:
		fmt.Errorf("error: Kin not known")
	}
	*e = v
	return nil
}

const MaxPresence = 4

// A presence is anything that is present on the tile map apart
// from the player.
type Presence struct {
	Kin    Kin    `json:"kin" xml:"kin,attr"` // Kin is the kind of presence.
	X      int    `json:"x" xml:"x,attr"`     // X where the presence starts out.
	Y      int    `json:"y" xml:"y,attr"`     // Y where the presence starts out.
	Width  int    `json:"width" xml:"width,attr"`
	Height int    `json:"height" xml:"height,attr"`
	Offset int    `json:"offset" xml:"offset,attr"`
	Frames int    `json:"frames" xml:"frames,attr,omitempty"`
	Item   int    `json:"item" xml:"item,attr,omitempty"`
	Money  int    `json:"money" xml:"money,attr,omitempty"`
	Talk   string `json:"talk" xml:"talk,omitempty"`
	Basic  string `json:"basic" xml:"basic,omitempty"`
}

// Sprites associated with this map.
// We allow only one sprite sheet per map,
// the player has a separate sprite sheet.
type Sprites struct {
	From    string   `json:"from" xml:"from,attr"`               // From where to load the sprites or tiles.
	Number  int      `json:"number" xml:"number,attr,omitempty"` // Sheet number in basic.
	Surface *Surface `json:"-" xml:"-"`                          // Ebiten Surface for display.
}

type Map struct {
	XMLName   xml.Name   `json:"-" xml:"map"`
	Width     int        `json:"width" xml:"width,attr"`
	Height    int        `json:"height" xml:"height,attr"`
	Tw        int        `json:"tw" xml:"tw,attr"`
	Th        int        `json:"th" xml:"th,attr"`
	Offset    int        `json:"offset" xml:"offset,attr"`
	From      string     `json:"from" xml:"from,attr"`               // From where to load the images tiles.
	Prefix    string     `json:"prefix" xml:"prefix,attr"`           // Prefix in basic.
	Number    int        `json:"number" xml:"number,attr,omitempty"` // Map number in basic.
	Border    int        `json:"border" xml:"border,attr,omitempty"` // Border color.
	Music     int        `json:"music" xml:"music,attr,omitempty"`   // Music index.
	Sprites   Sprites    `json:"sprites" xml:"sprites"`              // Sprites.
	Exits     Exits      `json:"exits" xml:"exits"`                  // Exits.
	Presences []Presence `json:"presences" xml:"presences"`          // Presences.
	Rows      []Row      `json:"rows" xml:"rows"`                    // Rows.
	Surface   *Surface   `json:"-" xml:"-"`                          // Ebiten Surface for display.
	Flags     bool       `json:"-" xml:"-"`                          // If true flags fill be drawn.
}

func FormatFor(name string) Format {
	return Format(strings.ToLower(filepath.Ext(name)))
}

const TW = 8
const TH = 8
const PRE = "map"

func NewMap(w, h int, from string) (*Map, error) {
	res := &Map{Width: w, Height: h, Th: TH, Tw: TW, Prefix: PRE}
	err := res.LoadSurface(from)
	if err != nil {
		return nil, err
	}

	for y := 0; y < h; y++ {
		cells := make([]Cell, w)
		row := Row{Cells: cells}
		res.Rows = append(res.Rows, row)
	}
	return res, nil
}

func LoadMap(from string) (*Map, error) {
	f, err := os.Open(from)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadMapFromFile(f)
}

func LoadMapFromFile(f *os.File) (*Map, error) {
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	res := &Map{}
	err = FormatFor(f.Name()).Unmarshal(buf, res)
	if err != nil {
		return nil, err
	}
	// resize in case of non-coresponence
	res.Resize(res.Width, res.Height)

	err = res.LoadSurface(res.From)
	if err != nil {
		return nil, err
	}

	if res.Sprites.From != "" {
		err = res.Sprites.LoadSurface(res.Sprites.From)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (m *Map) Resize(w, h int) {
	if h < 1 || w < 1 {
		return
	}

	rlen := len(m.Rows)
	if h > rlen {
		m.Rows = append(m.Rows, make([]Row, h-rlen)...)
	} else if h < len(m.Rows) {
		m.Rows = m.Rows[0:h]
	}
	for y := 0; y < len(m.Rows); y++ {
		row := m.Rows[y]
		clen := len(row.Cells)
		if w > clen {
			row.Cells = append(row.Cells, make([]Cell, w-clen)...)
		} else if w < len(row.Cells) {
			row.Cells = row.Cells[0:w]
		}
		m.Rows[y] = row
	}
	m.Width = w
	m.Height = h
}

func (m *Map) Wrap(by int) {
	offx := by
	if by >= m.Width {
		offx = by % m.Width
	}
	if by < 0 {
		offx = m.Width + (by % m.Width)
	}

	for y, row := range m.Rows {
		pre, post := row.Cells[:offx], row.Cells[offx:]
		row.Cells = append(post, pre...)
		m.Rows[y] = row
	}
}

func (m *Map) Roll(by int) {
	offy := by
	if by >= m.Height {
		offy = by % m.Height
	}
	if by < 0 {
		offy = m.Height + (by % m.Height)
	}

	pre, post := m.Rows[:offy], m.Rows[offy:]
	m.Rows = append(post, pre...)
}

func (m *Map) LoadSurface(name string) error {
	img, err := LoadSurface(FromName(name))
	if err != nil {
		return errors.Join(errors.New("Cannot load image:"+name), err)
	}
	m.From = name
	m.Surface = img
	return nil
}

func (s *Sprites) LoadSurface(name string) error {
	img, err := LoadSurface(FromName(name))
	if err != nil {
		return errors.Join(errors.New("Cannot load image:"+name), err)
	}
	s.From = name
	s.Surface = img
	return nil
}

func (m *Map) ToTile(at Point, camera Rectangle) Point {
	off := at.Sub(camera.Min)
	return image.Pt(off.X/m.Tw, off.Y/m.Th)
}

// Puts the cell in the map.
// If the map is in flag mode, only the cell flag will be set.
func (m *Map) Put(atTile Point, cell Cell) {
	if !m.Inside(atTile) {
		return
	}
	m.Rows[atTile.Y].Cells[atTile.X] = cell
}

// Puts the cell flag in the map.
func (m *Map) PutFlag(atTile Point, flag Flag) {
	if !m.Inside(atTile) {
		return
	}
	m.Rows[atTile.Y].Cells[atTile.X].Flag = flag
}

// Puts the cell index in the map without changing the flags.
func (m *Map) PutIndex(atTile Point, idx byte) {
	if !m.Inside(atTile) {
		return
	}
	m.Rows[atTile.Y].Cells[atTile.X].Index = idx
}

const DefaultPresenceWidth = 8
const DefaultPresenceHeight = 16

// Puts the presnece into the map.
// If the map is in flag mode, only the cell flag will be set.
func (m *Map) PutPresence(atTile Point, presence Presence) {
	if len(m.Presences) >= MaxPresence {
		return
	}
	presence.X = atTile.X * m.Tw
	presence.Y = atTile.Y * m.Th
	if presence.Width < 1 {
		presence.Width = DefaultPresenceWidth
	}
	if presence.Height < 1 {
		presence.Height = DefaultPresenceHeight
	}
	m.Presences = append(m.Presences, presence)
}

func (m *Map) Inside(atTile Point) bool {
	if atTile.X < 0 || atTile.X >= m.Width {
		return false
	}
	if atTile.Y < 0 || atTile.Y >= m.Height {
		return false
	}
	return true
}

func (m *Map) Get(atTile Point) (cell Cell) {
	if !m.Inside(atTile) {
		return Cell{}
	}
	return m.Rows[atTile.Y].Cells[atTile.X]
}

func (m *Map) Save(to string) error {
	f, err := os.Create(to)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.SaveToFile(f)
}

func (m *Map) SaveToFile(f *os.File) error {
	buf, err := FormatFor(f.Name()).Marshal(m)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	return err
}

func (m *Map) Export(to string, form Format) error {
	f, err := os.Create(to)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.ExportToFile(f, form)
}

func (m *Map) ExportToFile(f *os.File, form Format) error {
	buf, err := form.Marshal(m)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	return err
}

var blockColor = RGBA{R: 66, B: 66, G: 66, A: 0xaa}

// var presenceColor = RGBA{R: 00, B: 66, G: 66, A: 0xaa}
var presenceColor = RGBA{R: 240, B: 66, G: 66, A: 0xaa}

const FeetHeight = 8

func (m *Map) RenderPresences(screen *Surface, camera Rectangle) {

	starty := camera.Min.Y / m.Th
	if starty < 0 {
		starty = 0
	}

	for _, presence := range m.Presences {

		atx := presence.X - camera.Min.X
		aty := presence.Y - camera.Min.Y

		if m.Sprites.Surface == nil || m.Flags {
			to := Bounds(atx, aty, m.Tw, m.Th)
			FillRect(screen, to, presenceColor)
			// draw colored rectangle if sprites are not available
			// or if flags mode is set
		}

		if m.Sprites.Surface == nil {
			continue
		}

		aty = aty - presence.Height + FeetHeight
		// "shift up so the "feet" stand on the position of the presence.
		ab := m.Sprites.Surface.Bounds()
		tilew := ab.Dx() / m.Tw
		id := presence.Offset
		idx := id % tilew
		idy := id / tilew
		fx := idx * m.Tw
		fy := idy * m.Th
		from := image.Rect(fx, fy, fx+presence.Width, fy+presence.Height)
		sub := m.Sprites.Surface.SubImage(from).(*Surface)
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(atx), float64(aty))
		if sub != nil {
			screen.DrawImage(sub, &opts)
		}
	}
}

func (m *Map) Render(screen *Surface, camera Rectangle) {
	ab := m.Surface.Bounds()

	starty := camera.Min.Y / m.Th
	if starty < 0 {
		starty = 0
	}
	endy := min(camera.Max.Y/m.Th, len(m.Rows)-1)

	// This draws the whole layer. Only draw visible part using a camera.
	for ty := starty; ty < endy; ty++ {
		row := m.Rows[ty]

		startx := max(camera.Min.X/m.Tw, 0)
		endx := min(camera.Max.X/m.Tw, len(row.Cells)-1)
		for tx := startx; tx < endx; tx++ {
			cell := row.Cells[tx]
			id := int(cell.Index)
			if cell.Flag&FlagExtended != 0 {
				id += 255
			}
			tilew := ab.Dx() / m.Tw
			idx := id % tilew
			idy := id / tilew
			fx := idx * m.Tw
			fy := idy * m.Th

			from := image.Rect(fx, fy, fx+m.Tw, fy+m.Th)
			sub := m.Surface.SubImage(from).(*Surface)
			opts := ebiten.DrawImageOptions{}
			if cell.Flag&FlagHorizontalFlip != 0 {
				opts.GeoM.Scale(-1, 1)
				opts.GeoM.Translate(float64(m.Tw), 0)
			}
			if cell.Flag&FlagVerticalFlip != 0 {
				opts.GeoM.Scale(1, -1)
				opts.GeoM.Translate(0, float64(m.Th))
			}

			atx := int(tx)*m.Tw - camera.Min.X
			aty := int(ty)*m.Th - camera.Min.Y

			opts.GeoM.Translate(float64(atx), float64(aty))

			if sub != nil {
				screen.DrawImage(sub, &opts)
			}

			to := Bounds(atx, aty, m.Tw, m.Th)
			if m.Flags {
				cell.Flag.Render(screen, to)
			}
		}
	}
	m.RenderPresences(screen, camera)
}

func (m *Map) FloodFill(atTile Point, cell Cell) {
	now := m.Get(atTile)
	if now.Index == cell.Index && now.Flag == cell.Flag {
		return // already ok
	}
	if !m.Inside(atTile) {
		return
	}

	m.Put(atTile, cell)
	// the floodfill is recursive but the maps are small so
	// it should not cause problems.
	for dx := -1; dx <= 1; dx++ {
		at2 := atTile
		at2.X += dx
		now2 := m.Get(at2)
		if now2.Index == now.Index {
			m.FloodFill(at2, cell)
		}
	}
	for dy := -1; dy <= 1; dy++ {
		at2 := atTile
		at2.Y += dy
		now2 := m.Get(at2)
		if now2.Index == now.Index {
			m.FloodFill(at2, cell)
		}
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
)

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Format is just the lowercase extension including the '.' prefix.
type Format string

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
	FlagExtended       Flag = 1
	FlagHorizontalFlip Flag = 2
	FlagVerticalFlip   Flag = 4
	FlagSpritePalette  Flag = 8
	FlagOnTop          Flag = 16
	FlagSolid          Flag = 32
	FlagBless          Flag = 64
	FlagHarm           Flag = 128
)

type Cell struct {
	Index byte `json:"index" xml:"index,attr"`
	Flag  Flag `json:"flag" xml:"flag,attr"`
}

type Row struct {
	Cells []Cell `json:"cells" xml:"cells"`
}

type Map struct {
	Width  int    `json:"width" xml:"width,attr"`
	Height int    `json:"height" xml:"height,attr"`
	Tw     int    `json:"tw" xml:"tw,attr"`
	Th     int    `json:"th" xml:"th,attr"`
	Offset int    `json:"offset" xml:"offset,attr"`
	From   string `json:"from" xml:"from,attr"`     // From where to load the images tiles.
	Prefix string `json:"prefix" xml:"prefix,attr"` // Prefix in basic
	Rows   []Row  `json:"rows" xml:"rows"`          // Rows.

	Surface *Surface `json:"-" xml:"-"` // Ebiten Surface for display.
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
	buf, err := os.ReadFile(from)
	if err != nil {
		println(from, err.Error())
		return nil, err
	}
	res := &Map{}
	err = FormatFor(from).Unmarshal(buf, res)
	if err != nil {
		return nil, err
	}
	err = res.LoadSurface(res.From)
	if err != nil {
		return nil, err
	}
	return res, nil
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

func (m *Map) ToTile(at Point, camera Rectangle) Point {
	off := at.Sub(camera.Min)
	return image.Pt(off.X/m.Tw, off.Y/m.Th)
}

func (m *Map) Put(atTile Point, cell Cell) {
	if atTile.X < 0 || atTile.X >= m.Width {
		return
	}
	if atTile.Y < 0 || atTile.Y >= m.Height {
		return
	}
	m.Rows[atTile.Y].Cells[atTile.X] = cell
}

func (m *Map) Get(atTile Point) (cell Cell) {
	if atTile.X < 0 || atTile.X >= m.Width {
		return Cell{}
	}
	if atTile.Y < 0 || atTile.Y >= m.Height {
		return Cell{}
	}
	return m.Rows[atTile.Y].Cells[atTile.X]
}

func (m *Map) Save(to string) error {
	buf, err := FormatFor(to).Marshal(m)
	if err != nil {
		return err
	}
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(buf)
	return err
}

var blockColor = RGBA{R: 66, B: 66, G: 66, A: 0xaa}

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

			to := Bounds(atx, atx, m.Tw, m.Th)

			if cell.Flag&FlagSolid != 0 {
				DrawRect(screen, to, 3, blockColor)
			}
		}
	}
}

func col2b(col color.Color) byte {
	cr, cb, cg, _ := col.RGBA()

	b := byte(cr >> 14)
	g := byte(cb >> 14)
	r := byte(cg >> 14)
	return r<<4 | g<<2 | b
}

func PaletteToBasic(out io.Writer, bitmap image.PalettedImage, pre string, poff int) error {
	cm := bitmap.ColorModel()
	palette, ok := cm.(color.Palette)
	if !ok {
		return fmt.Errorf("Cannot get palette")
	}
	if len(palette) > 16 {
		return fmt.Errorf("Too many pallet entries, can only have 16: %s", len(palette))
	}
	fmt.Fprintf(out, "' Palette subroutine, call be called with GOSUB %s_palette\n", pre)
	fmt.Fprintf(out, "%s_palette: PROCEDURE\n", pre)
	for pidx, entry := range palette {
		pr, pg, pb, pa := entry.RGBA()
		fmt.Fprintf(out, "' palette entry %d: %0x,%0x,%0x,%0x\n", pidx, pr, pg, pb, pa)
		fmt.Fprintf(out, "\tPALETTE %d,$%02x\n", poff+pidx, col2b(entry))
	}
	fmt.Fprintf(out, "END\n\n")
	return nil
}

func ImageToBasic(out io.Writer, bitmap image.PalettedImage, pre string, cw, ch, poff int) error {
	fmt.Fprintf(out, "' Generated with res2bas\n\n")
	PaletteToBasic(out, bitmap, pre, poff)
	fmt.Fprintf(out, "' Bitmap output: %s\n", pre)
	fmt.Fprintf(out, "%s_bitmap:\n", pre)

	bw, bh := bitmap.Bounds().Dx(), bitmap.Bounds().Dy()

	for cy := 0; cy < bh; cy += ch {
		for cx := 0; cx < bw; cx += cw {
			if !IsBitmapEmpty(bitmap, cw, ch, cx, cy) {
				BitmapToBasic(out, bitmap, pre, cw, ch, cx, cy)
			}
		}
	}
	return nil
}

func IsBitmapEmpty(bitmap image.PalettedImage, cw, ch, cx, cy int) bool {
	for y := cy; y < cy+ch; y++ {
		for x := cx; x < cx+cw; x++ {
			idx := bitmap.ColorIndexAt(x, y)
			if idx > 16 {
				errExit(fmt.Errorf("Color out of range at (%d, %d): %s", x, y, idx))
			}
			if idx != 0 {
				return false
			}
		}
	}
	return true
}

func BitmapToBasic(out io.Writer, bitmap image.PalettedImage, pre string, cw, ch, cx, cy int) error {
	fmt.Fprintf(out, "' sprite at (%d, %d)\"\n", cx, cy)
	for y := cy; y < cy+ch; y++ {
		fmt.Fprintf(out, "\tBITMAP \"")
		for x := cx; x < cx+cw; x++ {
			idx := bitmap.ColorIndexAt(x, y)
			if idx > 15 {
				errExit(fmt.Errorf("Color out of range at (%d, %d): %s", x, y, idx))
			}
			if idx == 0 {
				fmt.Fprintf(out, ".")
			} else {
				fmt.Fprintf(out, "%x", idx)
			}
		}
		fmt.Fprintf(out, "\"\n")
	}
	return nil
}

func (m *Map) Basic(out io.Writer) error {
	fmt.Fprintf(out, "' Generated with res2bas\n\n")
	fmt.Fprintf(out, "' Screen for tile map %s, offset: %d\n", m.Prefix, m.Offset)
	fmt.Fprintf(out, "%s_map: \n", m.Prefix)
	for y := 0; y < 24; y++ {
		fmt.Fprintf(out, "DATA BYTE ")
		for x := 0; x < 32; x++ {
			cell := m.Get(image.Pt(x, y))
			fmt.Fprintf(out, "$%02x,$%02x", cell.Index, cell.Flag)
		}
		fmt.Fprintf(out, "\n")
	}
	if m.From != "" {
		pali, err := LoadPaletted(FromName(m.From))
		if err != nil {
			return err
		}
		err = ImageToBasic(out, pali, m.Prefix, m.Tw, m.Th, m.Offset)
		if err != nil {
			return err
		}
	}
	return nil
}

type Basicer interface {
	Basic(out io.Writer) error
}

func MarshalBasic(ptr any) ([]byte, error) {
	basicer, ok := ptr.(Basicer)
	if !ok {
		return nil, errors.New("Can only mashal a Basicer to basic")
	}
	writer := &bytes.Buffer{}
	if err := basicer.Basic(writer); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

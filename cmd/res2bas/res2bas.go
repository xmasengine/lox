// res2bas converts resources such as images to basic.
// Currently it targets CVBasic and the Master System.
//
// The default mode is to convert sprites from a
// paletted PNG or GIF file to basic BITMAP statements,
// preserving the palette as PALETTE statements.
// The paletted images mmust have 16 or less colors,
// and color 0 is the transparent color.
// The bitmaps are exported in 8x16 cells, and empty sprites will be skipped.
//
// In tile mode this tool convert sprites from a
// paletted PNG or GIF file to basic BITMAP statements
// preserving the palette as PALETTE statements.
// The paletted images mmust have 16 or less colors,
// and color 0 is the transparent color.
// The bitmaps are exported in 8x16 cells, and empty times will be kept.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"os"
	"path/filepath"
)

import (
	tiled "github.com/lafriks/go-tiled"
)

func errExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exit(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func warn(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "warning:"+msg+"\n", args...)
}

func main() {
	var err error

	var img string
	var bas string
	var pre string
	var mod string

	flag.StringVar(&img, "i", "", "input TMX, PNG or GIF file name or STDIN by default")
	flag.StringVar(&bas, "o", "", "output bas file name or STDOUT by default")
	flag.StringVar(&pre, "p", "sprite", "label prefix in basic output")
	flag.StringVar(&mod, "m", "sprite", "mode, one of sprite,tile,tmx,map,defpal")
	flag.Parse()

	in := os.Stdin
	out := os.Stdout

	if img != "" {
		in, err = os.Open(img)
		errExit(err)
		defer in.Close()
	}

	if bas != "" {
		out, err = os.Create(bas)
		errExit(err)
		defer out.Close()
	}

	if mod == "defpal" {
		defpal(out)
		return
	}

	if mod == "tmx" {
		dn := filepath.Dir(img)
		tilemap, err := tiled.LoadReader(dn, in)
		errExit(err)
		tmx(out, tilemap, pre, dn)
		return
	}

	decoded, _, err := image.Decode(in)
	errExit(err)
	bitmap, ok := decoded.(image.PalettedImage)
	if !ok {
		errExit(fmt.Errorf("Not a paletted image"))
	}

	poff := 0
	ph := 8
	if mod == "sprite" {
		poff = 16
		ph = 16
	}

	generate(out, bitmap, pre, 8, ph, poff)
}

func col2b(col color.Color) byte {
	cr, cb, cg, _ := col.RGBA()

	b := byte(cr >> 14)
	g := byte(cb >> 14)
	r := byte(cg >> 14)
	return r<<4 | g<<2 | b
}

var pal = []byte{
	0x00, 0x00, 0x0c, 0x2e, 0x20, 0x30, 0x02, 0x3c, 0x17, 0x2B, 0x0f, 0x2f, 0x08, 0x33, 0x2a, 0x3f}

var names = []string{
	"Transparent", "Black", "Green", "Lime", "Navy", "Blue", "Brown", "Cyan",
	"Red", "Scarlet", "Khaki", "Yellow", "Leaf", "Magenta", "Gray", "White",
}

func defpal(out *os.File) {
}

func tmx(out *os.File, tm *tiled.Map, pre, dn string) {
	fmt.Fprintf(out, "' Generated with res2bas\n\n")
	fmt.Fprintf(out, "' Screen for tile map %s\n", pre)
	fmt.Fprintf(out, "%s_map: \n", pre)

	if len(tm.Layers) != 1 {
		exit("Tile map must have exactly 1 layer")
	}
	if len(tm.Tilesets) != 1 {
		exit("Tile map must have exactly 1 tile set")
	}
	layer := tm.Layers[0]
	set := tm.Tilesets[0]
	offset := layer.Properties.GetInt("offset")
	warn("offset: %d", offset)
	for y := 0; y < 24; y++ {
		fmt.Fprintf(out, "DATA BYTE ")
		for x := 0; x < 32; x++ {
			tile := layer.Tiles[y*tm.Width+x]
			var idx byte
			var flag byte
			if tile != nil && !tile.Nil {
				set := tile.Tileset
				if set == nil {
					warn("tile has no tile set: (%d, %d): %d", x, y, tile.ID)
				}

				id := uint32(offset) + tile.ID - set.FirstGID
				idx = byte(id)
				if id > 255 {
					idx = byte(id - 255)
					flag |= 1
				}
				if id > 511 || id < 0 {
					warn("tile out of range: (%d, %d): %d", x, y, id)
				}
				if tile.HorizontalFlip {

					flag |= 2
				}
				if tile.VerticalFlip {
					flag |= 4
				}
				if tile.DiagonalFlip {
					flag |= 6
				}
				if flag != 0 {
					warn("flip: (%d, %d): %d,%d,%d: %d", x, y, id, idx, flag, tile.ID)
				}
			} else {
				warn("nil tile: (%d, %d)", x, y)
			}
			if x > 0 {
				fmt.Fprintf(out, ",")
			}
			fmt.Fprintf(out, "$%02x,$%02x", idx, flag)
		}
		fmt.Fprintf(out, "\n")
	}
	if set.Image != nil && set.Image.Source != "" {
		tname := filepath.Join(dn, set.Image.Source)
		warn("tile set source: %s", tname)
		tim, err := os.Open(tname)
		errExit(err)
		defer tim.Close()
		tdec, _, err := image.Decode(tim)
		errExit(err)
		tbmp, ok := tdec.(image.PalettedImage)
		if !ok {
			exit("Not a paletted image: %s", tname)
		}
		generate(out, tbmp, pre, 8, 8, 0)

	} else {
		warn("tile set has no image")
	}
}

func genpal(out *os.File, bitmap image.PalettedImage, pre string, poff int) {
	cm := bitmap.ColorModel()
	palette, ok := cm.(color.Palette)
	if !ok {
		errExit(fmt.Errorf("Cannot get palette"))
	}
	if len(palette) > 16 {
		errExit(fmt.Errorf("Too many pallet entries, can only have 16: %s", len(palette)))
	}
	fmt.Fprintf(out, "' Palette subroutine, call be called with GOSUB %s_palette\n", pre)
	fmt.Fprintf(out, "%s_palette: PROCEDURE\n", pre)
	for pidx, entry := range palette {
		pr, pg, pb, pa := entry.RGBA()
		fmt.Fprintf(out, "' palette entry %d: %0x,%0x,%0x,%0x\n", pidx, pr, pg, pb, pa)
		fmt.Fprintf(out, "\tPALETTE %d,$%02x\n", poff+pidx, col2b(entry))
	}
	fmt.Fprintf(out, "END\n\n")
}

func generate(out *os.File, bitmap image.PalettedImage, pre string, cw, ch, poff int) {
	fmt.Fprintf(out, "' Generated with res2bas\n\n")
	genpal(out, bitmap, pre, poff)
	fmt.Fprintf(out, "' Bitmap output: %s\n", pre)
	fmt.Fprintf(out, "%s_bitmap:\n", pre)

	bw, bh := bitmap.Bounds().Dx(), bitmap.Bounds().Dy()

	for cy := 0; cy < bh; cy += ch {
		for cx := 0; cx < bw; cx += cw {
			if !isCellEmpty(bitmap, cw, ch, cx, cy) {
				cell(out, bitmap, pre, cw, ch, cx, cy)
			}
		}
	}
}

func isCellEmpty(bitmap image.PalettedImage, cw, ch, cx, cy int) bool {
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

func cell(out *os.File, bitmap image.PalettedImage, pre string, cw, ch, cx, cy int) {
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
}

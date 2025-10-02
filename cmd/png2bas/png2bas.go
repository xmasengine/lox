// png2bas converts paletted PNG or GIF files with sprites to basic
// BITMAP, preserving the palette as PALETTE statements.
// Currently this tool is for use with cvbasic for master system only.
// The bitmaps are exported in 8x16 cells for sprites, and empty
// sprites will be skipped.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"os"
)

func errExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	var err error

	var img string
	var bas string
	var pre string

	flag.StringVar(&img, "i", "", "input png or GIF file name or STDIN by default")
	flag.StringVar(&bas, "o", "", "output bas file name or STDOUT by default")
	flag.StringVar(&pre, "p", "sprite", "label prefix in basic output")
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

	decoded, _, err := image.Decode(in)
	errExit(err)
	bitmap, ok := decoded.(image.PalettedImage)
	if !ok {
		errExit(fmt.Errorf("Not a paletted image"))
	}
	generate(out, bitmap, pre, 8, 16)
}

func col2b(col color.Color) byte {
	cr, cb, cg, _ := col.RGBA()

	b := byte(cr >> 14)
	g := byte(cb >> 14)
	r := byte(cg >> 14)
	return r<<4 | g<<2 | b
}

func generate(out *os.File, bitmap image.PalettedImage, pre string, cw, ch int) {
	cm := bitmap.ColorModel()
	palette, ok := cm.(color.Palette)
	if !ok {
		errExit(fmt.Errorf("Cannot get palette"))
	}
	if len(palette) > 16 {
		errExit(fmt.Errorf("Too many pallet entries, can only have 16: %s", len(palette)))
	}
	fmt.Fprintf(out, "' Generated with png2bas\n\n")
	fmt.Fprintf(out, "' Palette subroutine, call be called with GOSUB %s_palette\n", pre)
	fmt.Fprintf(out, "%s_palette: PROCEDURE\n", pre)
	for pidx, entry := range palette {
		pr, pg, pb, pa := entry.RGBA()
		fmt.Fprintf(out, "' palette entry %d: %0x,%0x,%0x,%0x\n", pidx, pr, pg, pb, pa)
		fmt.Fprintf(out, "\tPALETTE %d,$%0x\n", 16+pidx, col2b(entry))
	}
	fmt.Fprintf(out, "END\n")
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
			if idx > 16 {
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

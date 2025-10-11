package masite

import "io"
import "image"
import "image/color"
import "fmt"
import "errors"
import "bytes"

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

package masite

import "github.com/hajimehoshi/ebiten/v2"
import "image"
import "image/color"
import "errors"
import "encoding"
import "os"
import "fmt"

// A few type aliases and helpers for convenience
type (
	Color           = color.Color
	RGBA            = color.RGBA
	Image           = image.Image
	Surface         = ebiten.Image
	Game            = ebiten.Game
	Rectangle       = image.Rectangle
	Point           = image.Point
	Key             = ebiten.Key
	TextMarshaler   = encoding.TextMarshaler
	TextUnmarshaler = encoding.TextUnmarshaler
)

type TextEncoding interface {
	TextMarshaler
	TextUnmarshaler
}

var (
	Termination = ebiten.Termination
	MidgetOK    = errors.New("OK")
)

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

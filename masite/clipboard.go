package masite

import (
	"golang.design/x/clipboard"
)

import (
	"bytes"
	"context"
	"unicode/utf8"
)

var clipboardAvailable = false

type ClipboardFormat = clipboard.Format

const (
	ClipboardText  = clipboard.FmtText
	ClipboardImage = clipboard.FmtImage
)

func ReadClipboard(form ClipboardFormat) []byte {
	if !clipboardAvailable {
		return nil
	}
	return clipboard.Read(form)
}

func WriteClipboard(form ClipboardFormat, data []byte) <-chan struct{} {
	if !clipboardAvailable {
		return nil
	}
	return clipboard.Write(form, data)
}

func WriteClipboardRunes(runes []rune) {
	buf := bytes.Buffer{}
	for i := 0; i < len(runes); i++ {
		buf.WriteRune(runes[i])
	}
	WriteClipboard(ClipboardText, buf.Bytes())
}

func ReadClipboardRunes() []rune {
	buf := ReadClipboard(ClipboardText)
	if len(buf) < 1 {
		return nil
	}
	res := []rune{}
	for i := 0; i < len(buf); {
		r, size := utf8.DecodeRune(buf[i:])
		if r == utf8.RuneError {
			return res
		}
		res = append(res, r)
		i += size
	}
	return res
}

func WatchClipboard(ctx context.Context, form ClipboardFormat) <-chan []byte {
	if !clipboardAvailable {
		return nil
	}
	return clipboard.Watch(ctx, form)
}

func init() {
	err := clipboard.Init()
	if err != nil {
		println("clipboard not available", "err", err.Error())
	}
	clipboardAvailable = err == nil
}

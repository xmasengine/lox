// Package fir is for interfacing fith Furnace the open source chiptune tracker.
// For simplicity I only support the fir text export format.
package fir

import "io"
import "errors"
import "strings"

import "strconv"
import "rsc.io/markdown"

func Read(in io.Reader) (*Song, error) {
	buf, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	parser := markdown.Parser{}
	doc := parser.Parse(string(buf))
	return ReadDoc(doc)
}

func TextPlainText(text *markdown.Text) string {
	return text.Inline[0].(*markdown.Plain).Text
}

func ListStrings(block markdown.Block) []string {
	res := []string{}
	list := block.(*markdown.List)
	for _, sub := range list.Items {
		block := sub.(*markdown.Item).Blocks[0]
		var txt *markdown.Text
		if para, ok := block.(*markdown.Paragraph); ok {
			txt = para.Text
		} else {
			txt = block.(*markdown.Text)
		}
		res = append(res, TextPlainText(txt))
	}
	return res
}

func ListToMap(strs []string, sep string) map[string]string {
	res := map[string]string{}
	for _, str := range strs {
		key, value, _ := strings.Cut(str, sep)
		key = strings.Trim(key, "  \t")
		value = strings.Trim(value, "  \t")
		res[key] = value
	}
	return res
}

const GeneratorBlock = 1
const InfoBlock = 3
const ChipsBlock = 5

func FindHeading(blocks []markdown.Block, name string) int {
	for i := 0; i < len(blocks); i++ {
		if hd, ok := blocks[i].(*markdown.Heading); ok {
			if TextPlainText(hd.Text) == name {
				return i
			}
		}
	}
	return -1
}

func ReadDoc(doc *markdown.Document) (*Song, error) {
	song := &Song{}
	if len(doc.Blocks) < 10 {
		return nil, errors.New("Too short for a Furnace song text export")
	}
	// doc.Blocks[]
	song.Generator = TextPlainText(doc.Blocks[GeneratorBlock].(*markdown.Paragraph).Text)

	// Song info
	infoMap := ListToMap(ListStrings(doc.Blocks[InfoBlock]), ":")
	song.Info.Name = infoMap["name"]
	song.Info.Author = infoMap["author"]
	song.Info.System = infoMap["system"]
	song.Info.Tuning, _ = strconv.Atoi(infoMap["tuning"])
	song.Info.Instruments, _ = strconv.Atoi(infoMap["instruments"])
	song.Info.Wavetables, _ = strconv.Atoi(infoMap["wavetables"])
	song.Info.Samples, _ = strconv.Atoi(infoMap["samples"])

	// Chip Info
	mchips := doc.Blocks[ChipsBlock].(*markdown.List)
	for _, mchip := range mchips.Items {
		ichip := mchip.(*markdown.Item)
		var chip Chip
		chip.Name = TextPlainText(ichip.Blocks[0].(*markdown.Text))
		chipMap := ListToMap(ListStrings(ichip.Blocks[1]), ":")
		chip.ID, _ = strconv.Atoi(chipMap["id"])
		chip.Volume, _ = strconv.Atoi(chipMap["volume"])
		chip.Panning, _ = strconv.Atoi(chipMap["panning"])
		chip.FrontRear, _ = strconv.Atoi(chipMap["front/rear"])
		song.Chips = append(song.Chips, chip)
	}

	subIndex := FindHeading(doc.Blocks, "Subsongs")
	if subIndex < 0 {
		return song, errors.New("subsong not found")
	}

	var subSong Subsong
	subSongIndex := subIndex + 1
	subSong.ID, _ = strconv.Atoi((TextPlainText(doc.Blocks[subSongIndex].(*markdown.Heading).Text)))
	subSongIndex++
	subSongMap := ListToMap(ListStrings(doc.Blocks[subSongIndex]), ":")
	subSong.TickRate, _ = strconv.Atoi(subSongMap["tick rate"])
	subSong.TimeBase, _ = strconv.Atoi(subSongMap["time base"])
	subSong.PatternLength, _ = strconv.Atoi(subSongMap["pattern length"])

	speeds := strings.Split(subSongMap["speeds"], " ")
	for _, speed := range speeds {
		ispeed, _ := strconv.Atoi(speed)
		subSong.Speeds = append(subSong.Speeds, ispeed)
	}
	tempo, tempoDiv, _ := strings.Cut(subSongMap["virtual tempo"], "/")
	subSong.Tempo, _ = strconv.Atoi(tempo)
	subSong.TempoDiv, _ = strconv.Atoi(tempoDiv)
	subSongIndex += 2

	orderm := doc.Blocks[subSongIndex].(*markdown.CodeBlock)
	for _, line := range orderm.Text {
		var order Order
		pre, body, _ := strings.Cut(line, " | ")
		order.ID, _ = strconv.Atoi(pre)
		channels := strings.Split(body, " ")
		for _, channel := range channels {
			ichannel, _ := strconv.Atoi(channel)
			order.Channels = append(order.Channels, ichannel)
		}
		subSong.Orders = append(subSong.Orders, order)
	}
	subSongIndex += 2

	patterns := doc.Blocks[subSongIndex].(*markdown.Paragraph)
	var pattern *Pattern
	for _, tline := range patterns.Text.Inline {
		if plain, ok := tline.(*markdown.Plain); ok {
			line := plain.Text
			if strings.HasPrefix(line, "----- ORDER ") {
				if pattern != nil {
					subSong.Patterns = append(subSong.Patterns, *pattern)
				} else {
					pattern = &Pattern{}
					pats := strings.TrimPrefix(line, "----- ORDER ")
					pattern.Order, _ = strconv.Atoi(pats)
				}
			} else {
				row := Row{}
				parts := strings.Split(line, "|")
				row.Tick, _ = strconv.Atoi(strings.Trim(parts[0], " \t"))
				for c := 1; c < len(parts); c++ {
					fields := strings.Split(parts[c], " ")
					channel := Channel{}
					channel.NoteName = fields[0]
					if fields[1] == ".." {
						channel.Channel = -1
					} else {
						channel.Channel, _ = strconv.Atoi(fields[1])
					}
					row.Channels = append(row.Channels, channel)
				}
				pattern.Rows = append(pattern.Rows, row)
			}
		}
	}
	if pattern != nil {
		subSong.Patterns = append(subSong.Patterns, *pattern)
	}

	song.Subsongs = append(song.Subsongs, subSong)

	return song, nil
}

type Information struct {
	Name        string ``
	Author      string
	System      string
	Tuning      int
	Instruments int
	Wavetables  int
	Samples     int
}

type Chip struct {
	Name      string
	ID        int
	Volume    int
	Panning   int
	FrontRear int
}

type Macro struct {
	Name   string
	Speed  int
	Levels []int
}

type Instrument struct {
	Type   int
	Macros []Macro
}

type Wavetable struct {
}

type Sample struct {
}

type Song struct {
	Generator   string
	Info        Information
	Chips       []Chip
	Instruments []Instrument
	Wavetables  []Wavetable
	Samples     []Sample
	Subsongs    []Subsong
}

type Subsong struct {
	ID            int
	TickRate      int
	Speeds        []int
	Tempo         int
	TempoDiv      int
	TimeBase      int
	PatternLength int
	Orders        []Order
	Patterns      []Pattern
}

type Order struct {
	ID       int
	Channels []int
}

type Pattern struct {
	Order int
	Rows  []Row
}

type Note int

const (
	Off Note = iota
	A0
	A0S
	B0
	C0
	C0S
	D0
	D0S
	E0
	F0
	F0S
	G0
	G0S
	A1
	A1S
	B1
	C1
	C1S
	D1
	D1S
	E1
	F1
	F1S
	G1
	G1S
	A2
	A2S
	B2
	C2
	C2S
	D2
	D2S
	E2
	F2
	F2S
	G2
	G2S
	A3
	A3S
	B3
	C3
	C3S
	D3
	D3S
	E3
	F3
	F3S
	G3
	G3S
	A4
	A4S
	B4
	C4
	C4S
	D4
	D4S
	E4
	F4
	F4S
	G4
	G4S
	A5
	A5S
	B5
	C5
	C5S
	D5
	D5S
	E5
	F5
	F5S
	G5
	G5S
	A6
	A6S
	B6
	C6
	C6S
	D6
	D6S
	E6
	F6
	F6S
	G6
	G6S
)

type Channel struct {
	NoteName string
	Channel  int
}

type Row struct {
	Tick     int
	Channels []Channel
}

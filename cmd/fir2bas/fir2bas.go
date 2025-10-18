package main

import "strings"
import "os"
import "io"
import "flag"
import "fmt"

import "github.com/xmasengine/lox/fir"

// import "strings"

var input string
var output string

func main() {
	var err error

	flag.StringVar(&input, "i", "-", "ABC input file")
	flag.StringVar(&output, "o", "-", "BASIC input file")
	flag.Parse()

	in := os.Stdin
	out := os.Stdin
	if input != "-" {
		in, err = os.Open(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		defer in.Close()
	}
	if output != "-" {
		out, err = os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		defer out.Close()
	}

	song, err := fir.Read(in)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fir2bas(out, song)

}

const Channels = 4
const TicksPerSecond = 50

var instruments = []string{"W", "X", "Y", "Z"}

func fir2bas(out io.Writer, song *fir.Song) {
	playing := [Channels]bool{}
	voices := [Channels]int{}

	label := strings.ToLower(strings.ReplaceAll(song.Info.Name, " ", "_"))
	fmt.Fprintf(out, "' %s: %s\n", song.Info.Name, song.Info.Author)
	sub := song.Subsongs[0]
	// The ticks per note is imply (the first) speed
	// TickRate is beats per minute
	ticksPerNote := sub.Speeds[0]
	fmt.Fprintf(out, "music_%s: DATA BYTE %d\n", label, ticksPerNote)
	// TODO: should order patterns.
	for _, pattern := range sub.Patterns {
		for _, row := range pattern.Rows {
			parts := [4]string{"-", "-", "-", "-"}
			for cn, channel := range row.Channels {
				if channel.NoteName == "OFF" {
					playing[cn] = false
				} else if channel.NoteName == "..." {
					if playing[cn] {
						parts[cn] = "S"
					}
				} else {
					instrument := channel.Channel
					istr := ""
					if instrument >= 0 || instrument < len(instruments) {
						if voices[cn] != instrument {
							voices[cn] = instrument
							istr = instruments[instrument]
						}
					}
					parts[cn] = fmt.Sprintf("%c%c%s", channel.NoteName[0], channel.NoteName[2], istr)
					if channel.NoteName[1] == '#' {
						parts[cn] = fmt.Sprintf("%c%c#%s", channel.NoteName[0], channel.NoteName[2], istr)
					}
					playing[cn] = true
				}
			}
			fmt.Fprintf(out, "    MUSIC %s\n", strings.Join(parts[:], ","))
		}
	}
	fmt.Fprintf(out, "    MUSIC REPEAT\n")
	fmt.Fprintf(out, "    MUSIC STOP\n")

}

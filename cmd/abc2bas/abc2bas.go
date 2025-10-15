package main

import "io"
import "strconv"
import "os"
import "flag"
import "fmt"
import "strings"

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
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		defer in.Close()
	}
	if output != "-" {
		out, err = os.Create(output)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		defer out.Close()
	}
	err = abc2bas(in, out)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

type Track struct {
	Instrument byte
	Notes      []string
}

type Song struct {
	Title  string
	Tempo  byte
	Tracks [3]Track
	Drums  Track
}

func (s *Song) Note(track int, note string) string {
	s.Tracks[track].Notes = append(s.Tracks[track].Notes, note)
	return note
}

func (s *Song) Notes(track int, note string, length int) string {
	for i := 0; i < length; i++ {
		s.Note(track, note)
	}
	return note
}

func abc2bas(in io.Reader, out io.Writer) error {
	buf, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	abc := string(buf)

	var song Song
	for line := range strings.SplitSeq(abc, "\n") {
		line, _, _ = strings.Cut(line, "%")
		line = strings.Trim(line, " \t")
		if len(line) < 1 {
			continue
		}
		var notes string
		track := 0
		if len(line) > 3 && line[0:3] == "[V:" {
			n, _ := strconv.Atoi(line[3:4])
			if n > 0 {
				track = n - 1
			}
			if track > 2 {
				track = 2
			}
			notes, _ = strings.CutSuffix(line, "[|]")
			notes = strings.Trim(notes, " \t")
		} else {
			prefix, suffix, _ := strings.Cut(line, ": ")
			switch prefix[0] {
			case 'K':
				switch suffix {
				case "1/1":
					song.Tempo = 100
				case "1/2":
					song.Tempo = 50
				case "1/4":
					song.Tempo = 25
				case "1/8":
					song.Tempo = 12
				case "1/16":
					song.Tempo = 6
				default:
					song.Tempo = 25
				}
			case 'T':
				song.Title = strings.ReplaceAll(strings.ToLower(strings.Trim(suffix, " \t")), " ", "_")
			default:
				// ignore
			}
		}
		last := "-"
		acc := ""
		acci := 0
		for i, note := range notes {
			switch note {
			case 'A':
				last = song.Note(track, "A4"+acc)
			case 'B':
				last = song.Note(track, "B4"+acc)
			case 'C':
				last = song.Note(track, "C4"+acc)
			case 'D':
				last = song.Note(track, "D4"+acc)
			case 'E':
				last = song.Note(track, "E4"+acc)
			case 'F':
				last = song.Note(track, "F4"+acc)
			case 'G':
				last = song.Note(track, "G4"+acc)
			case 'a':
				last = song.Note(track, "A5"+acc)
			case 'b':
				last = song.Note(track, "B5"+acc)
			case 'c':
				last = song.Note(track, "C5"+acc)
			case 'd':
				last = song.Note(track, "D5"+acc)
			case 'e':
				last = song.Note(track, "E5"+acc)
			case 'f':
				last = song.Note(track, "F5"+acc)
			case 'g':
				last = song.Note(track, "G5"+acc)
			case ',':
				last = song.Note(track, "-")
			case '1':
				song.Notes(track, last, 1)
			case '2':
				song.Notes(track, "S", 1)
			case '3':
				song.Notes(track, "S", 2)
			case '4':
				song.Notes(track, "S", 3)
			case '5':
				song.Notes(track, "S", 4)
			case '6':
				song.Notes(track, "S", 5)
			case '7':
				song.Notes(track, "S", 6)
			case '8':
				song.Notes(track, "S", 7)
			case '9':
				song.Notes(track, "S", 7)
			case '_':
				acc = "#"
				acci = i
			case '-':
				acc = ""
			default:
			}
			if i > acci+2 {
				acc = ""
				acci = 0
			}
		}
	}

	fmt.Fprintf(out, "music_%s: DATA BYTE %d\n", song.Title, song.Tempo)

	last := max(len(song.Tracks[0].Notes), len(song.Tracks[1].Notes), len(song.Tracks[2].Notes))

	for i := 0; i < last; i++ {
		n1, n2, n3 := "-", "-", "-"
		if i < len(song.Tracks[0].Notes) {
			n1 = song.Tracks[0].Notes[i]
		}
		if i < len(song.Tracks[1].Notes) {
			n2 = song.Tracks[1].Notes[i]
		}
		if i < len(song.Tracks[2].Notes) {
			n3 = song.Tracks[2].Notes[i]
		}
		fmt.Fprintf(out, "\tMUSIC %s,%s,%s\n", n1, n2, n3)
	}
	fmt.Fprintln(out, "\tMUSIC REPEAT")
	fmt.Fprintln(out, "\tMUSIC STOP")

	return nil
}

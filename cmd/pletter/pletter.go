package main

import "flag"
import "os"
import "fmt"
import "github.com/xmas/lox/pletter"

var input string
var output string
var saveLength bool

func errExit(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(2)
	}
}

func main() {
	var err error

	flag.StringVar(&input, "i", "-", "input file to compress")
	flag.StringVar(&output, "o", "-", "output file to compress to")
	flag.BoolVar(&saveLength, "l", false, "save length in file header")
	flag.Parse()
	compressor := pletter.New(saveLength)
	in := os.Stdin
	out := os.Stdin
	if input != "" && input != "-" {
		in, err = os.Open(input)
		errExit(err)
		defer in.Close()
	}

	if output != "" && output != "-" {
		out, err = os.Create(output)
		errExit(err)
		defer out.Close()
	}

	err = compressor.Load(in)
	errExit(err)
	_, err = compressor.Save(out)
	errExit(err)
}

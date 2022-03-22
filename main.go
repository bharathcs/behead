package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	reportingDuration = time.Duration(10) * time.Second
	usageString       = `behead - take all contents after skipping a number of lines.

behead [-n count] [-f INPUT-FILE] [-o OUTPUT-FILE]

This utility will display all content after the first count lines from the specified input file, or of the standard
input if no files are specified. The utility will write to the output file, or standard out if no file is specified.
If the utility detects a specified output file, it will report progress to the standard output.

-n count
    Skip the first count lines of the input

-f INPUT-FILE
    Input will be read from this file, or standard in if not specified.

-o OUTPUT-FILE
    Output will be written to this file, or to standard out if not specified.
`
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usageString)
		flag.PrintDefaults()
	}

	numberOfLinesToSkip := flag.Int64("n", 0, "number of lines to skip")
	filepath := flag.String("f", "", "file to cut")
	outputFilepath := flag.String("o", "", "file to output to")
	flag.Parse()

	var (
		r   io.Reader
		wr  io.Writer
		err error
	)

	r, err = getReader(*filepath)
	exitIfError(err)
	wr, err = getWriter(*outputFilepath)
	exitIfError(err)

	if err = behead(r, wr, *numberOfLinesToSkip); err != nil {
		exitIfError(err)
	}
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

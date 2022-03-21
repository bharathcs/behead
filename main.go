package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const reportingDuration = time.Duration(10) * time.Second

func main() {
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

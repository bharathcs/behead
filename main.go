package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

const reportingDuration = time.Duration(10) * time.Second

func main() {
	numberOfLinesToSkip := flag.Int64("n", 0, "line number to numberOfLinesToSkip from")
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

func behead(r io.Reader, wr io.Writer, numberOfLinesToSkip int64) error {
	if numberOfLinesToSkip < 0 {
		return fmt.Errorf("non negative numberOfLinesToSkip required (CLI flag: -n %d)", numberOfLinesToSkip)
	}

	sc := bufio.NewScanner(r)

	lineCount := int64(0)
	done := make(chan interface{}, 1)
	if wr != os.Stdout {
		go report(&lineCount, reportingDuration, done, os.Stdout)
	}

	for sc.Scan() {
		if current := atomic.LoadInt64(&lineCount); current < numberOfLinesToSkip {
			atomic.AddInt64(&lineCount, 1)
			continue
		}

		_, err := wr.Write([]byte(sc.Text() + "\n"))
		if err != nil {
			return fmt.Errorf("unable to write to output anymore: %w", err)
		}
		atomic.AddInt64(&lineCount, 1)
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("did not end reading cleanly: %w", err)
	}

	return nil
}

func report(lineCountAddr *int64, reportingDuration time.Duration, stop <-chan interface{}, wr io.Writer) {
	var (
		start              time.Time
		now                time.Time
		lastTime           time.Time
		lastLineCount      int64
		lineCount          int64
		timeElapsed        time.Duration
		overallTimeElapsed time.Duration
		rate               float64
		overallRate        float64
	)

	start = now

	select {
	case _ = <-stop:
		printLine(wr, overallTimeElapsed, rate, overallRate, lineCount, "\n")
		break
	case now = <-time.NewTicker(reportingDuration).C:
		now = time.Now()
		lineCount = atomic.LoadInt64(lineCountAddr)

		rate = float64(lineCount-lastLineCount) / timeElapsed.Seconds()
		timeElapsed = now.Sub(lastTime)

		overallRate = float64(lineCount) / now.Sub(start).Seconds()
		overallTimeElapsed = now.Sub(start)

		printLine(wr, overallTimeElapsed, rate, overallRate, lineCount, "\r")
		lastLineCount = lineCount
		lastTime = now
	}
}

func printLine(wr io.Writer, overallTimeElapsed time.Duration, rate float64, overallRate float64, lineCount int64, suffix string) {
	fmt.Fprintf(
		wr,
		"%v have elapsed. %0.2f lines/sec (last %v), %0.2f lines/sec (overall), %d total rows     %s",
		time.Duration(overallTimeElapsed-(overallTimeElapsed%time.Second)),
		rate, reportingDuration, overallRate, lineCount, suffix,
	)
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getReader(filepath string) (r io.Reader, err error) {
	if len(filepath) == 0 {
		return os.Stdin, nil
	} else if f, err := os.Open(filepath); err != nil {
		return nil, fmt.Errorf("input file cannot be opened (CLI flag: -f %s): %w", filepath, err)
	} else {
		return f, nil
	}
}

func getWriter(filepath string) (r io.Writer, err error) {
	if len(filepath) == 0 {
		return os.Stdout, nil
	} else if f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755); err != nil {
		return nil, fmt.Errorf("output file cannot be opened (CLI flag: -o %s): %w", filepath, err)
	} else {
		return f, nil
	}
}

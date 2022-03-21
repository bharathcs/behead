package main

import (
	"fmt"
	"io"
	"os"
)

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

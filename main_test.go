package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func Test_main(t *testing.T) {
	runMain := func(linesToSkip int64, inputFile, outputFile string) (inFile, outFile string, err error) {
		//program, _ := os.Executable()
		cmdScript := fmt.Sprintf("%s -n %d < %s > %s", "go run .", linesToSkip, inputFile, outputFile)
		t.Log(cmdScript)
		cmd := exec.Command(
			"/bin/sh", "-c",
			cmdScript,
		)

		if _, err = cmd.CombinedOutput(); err != nil {
			return "", "", err
		}

		r, err := os.Open(inputFile)
		in, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read input %v", err)
			return "", "", err
		}

		r, err = os.Open(outputFile)
		out, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read output %v", err)
			return "", "", err
		}

		return string(in), string(out), err
	}

	test := func(inputFileContents string, numLinesToSkip int, t *testing.T) {
		outputFileContents := ""
		sc := bufio.NewScanner(strings.NewReader(inputFileContents))
		i := 0
		for sc.Scan() {
			if i < numLinesToSkip {
				i++
				continue
			}
			outputFileContents += sc.Text() + "\n"
			i++

		}

		tmp := path.Join(os.TempDir(), "main")
		os.Mkdir(tmp, 0750)

		inputFilepath := path.Join(tmp, "input")
		outputFilepath := path.Join(tmp, "output")
		if iFile, err := os.OpenFile(inputFilepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755); err == nil {
			iFile.Truncate(0)
			iFile.WriteString(inputFileContents)
			iFile.Close()
		} else {
			t.Fatal("can't create input file", err)
			return
		}

		if oFile, err := os.OpenFile(outputFilepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755); err == nil {
			oFile.Truncate(0)
			oFile.Close()
		} else {
			t.Fatal("can't create output file", err)
			return
		}

		inputAfter, outputAfter, err := runMain(int64(numLinesToSkip), inputFilepath, outputFilepath)
		if err != nil {
			t.Fatalf("Got err %v", err)
			return
		}
		if inputFileContents != inputAfter {
			t.Fatalf("wanted input %q, got %q", inputFileContents, inputAfter)
		}
		if outputFileContents != outputAfter {
			t.Fatalf("wanted output %q got %q", outputFileContents, outputAfter)
		}
	}

	var inputFileContents string
	for i := 0; i < 100; i++ {
		inputFileContents += fmt.Sprint(i) + "\n"
	}

	tests := []struct {
		name              string
		inputFileContents string
		linesToSkip       int
	}{
		{
			inputFileContents: inputFileContents,
			linesToSkip:       0,
		},
		{
			inputFileContents: inputFileContents,
			linesToSkip:       5,
		},
		{
			inputFileContents: inputFileContents,
			linesToSkip:       50,
		},
		{
			inputFileContents: inputFileContents,
			linesToSkip:       80,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(tt.inputFileContents, tt.linesToSkip, t)
		})
	}
}

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
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

func Test_behead_happyPath(t *testing.T) {
	type args struct {
		r                   io.Reader
		numberOfLinesToSkip int64
	}
	tests := []struct {
		name   string
		args   args
		wantWr string
	}{
		{
			name: "no input",
			args: args{
				r:                   strings.NewReader(""),
				numberOfLinesToSkip: 100,
			},
			wantWr: "",
		},
		{
			name: "overshoot lines to skip",
			args: args{
				r:                   strings.NewReader(""),
				numberOfLinesToSkip: 10,
			},
			wantWr: "",
		},
		{
			name: "simple test",
			args: args{
				r:                   strings.NewReader("foo\nbar\nbaz\n"),
				numberOfLinesToSkip: 1,
			},
			wantWr: "bar\nbaz\n",
		},
		{
			name: "will add terminating newline",
			args: args{
				r:                   strings.NewReader("foo\nbar\nbaz"),
				numberOfLinesToSkip: 2,
			},
			wantWr: "baz\n",
		},
		{
			name: "will add terminating newline",
			args: args{
				r:                   strings.NewReader("foo\nbar\nbaz"),
				numberOfLinesToSkip: 1,
			},
			wantWr: "bar\nbaz\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := &bytes.Buffer{}
			err := behead(tt.args.r, wr, tt.args.numberOfLinesToSkip)
			if err != nil {
				t.Errorf("behead() error = %v, wantErr %v", err, false)
				return
			}
			if gotWr := wr.String(); gotWr != tt.wantWr {
				t.Errorf("behead() gotWr = %v, want %v", unquote(gotWr, t), unquote(tt.wantWr, t))
			}
		})
	}
}

type badReader struct{}
type badWriter struct{}

func (b badWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("oops")
}

func (b badReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("oops")
}

func Test_behead_errors(t *testing.T) {
	type args struct {
		r                   io.Reader
		w                   io.Writer
		numberOfLinesToSkip int64
	}
	tests := []struct {
		name            string
		args            args
		errStringPrefix string
	}{
		{
			name: "bad reader",
			args: args{
				r:                   badReader{},
				w:                   &bytes.Buffer{},
				numberOfLinesToSkip: 100,
			},
			errStringPrefix: "did not end reading cleanly:",
		},
		{
			name: "bad writer",
			args: args{
				r:                   strings.NewReader("foo\n"),
				w:                   badWriter{},
				numberOfLinesToSkip: 0,
			},
			errStringPrefix: "unable to write to output anymore:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := behead(tt.args.r, tt.args.w, tt.args.numberOfLinesToSkip)
			if err == nil {
				t.Errorf("behead() error = %v, wantErr %q", err, tt.errStringPrefix)
				return
			}
			if !strings.Contains(err.Error(), tt.errStringPrefix) {
				t.Errorf("behead() gotWr = %v, want %v", err, tt.errStringPrefix)
				return
			}
		})
	}
}

func Test_getReader(t *testing.T) {

	tempDir := path.Join(os.TempDir(), "getReader")
	os.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)

	emptyFilepath := path.Join(tempDir, "emptyFile")
	normalFilepath := path.Join(tempDir, "normalFile")
	nonExistentFilepath := path.Join(tempDir, "nonExistentFile")

	normalFileContents := ""
	for i := 0; i < 100; i++ {
		normalFileContents += fmt.Sprint(i) + "\n"
	}

	if _, err := os.Create(emptyFilepath); err != nil {
		t.Fatal("unable to run test: temp files could not be created", err)
		return
	}

	if f, err := os.OpenFile(normalFilepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755); err == nil {
		_, err = f.WriteString(normalFileContents)
	} else if err != nil && f.Close() != nil {
		t.Fatal("unable to run test: temp files could not be created", err)
		return
	}

	tests := []struct {
		name     string
		filepath string
		wantR    string
		wantErr  bool
	}{
		{
			name:     "empty file",
			filepath: emptyFilepath,
			wantR:    "",
			wantErr:  false,
		},
		{
			name:     "non empty file",
			filepath: normalFilepath,
			wantR:    normalFileContents,
			wantErr:  false,
		},
		{
			name:     "non existent file",
			filepath: nonExistentFilepath,
			wantR:    "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := getReader(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if contents, _ := ioutil.ReadAll(gotR); string(contents) != tt.wantR {
					t.Errorf("getReader() gotR = %q, want %q", unquote(string(contents), t), tt.wantR)
				}
			}
		})
	}

	t.Run("empty filepath", func(t *testing.T) {
		gotR, err := getReader("")
		if err != nil {
			t.Errorf("getReader() err = %v", err)
		}
		if gotR != os.Stdin {
			t.Errorf("getReader() gotR = %v, want %v", gotR, os.Stdin)
		}
	})
}

func Test_getWriter(t *testing.T) {

	tempDir := path.Join(os.TempDir(), "getWriter")
	os.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)

	nonExistentFilepath := path.Join(tempDir, "notAFileYet")
	normalFilepath := path.Join(tempDir, "normalFile")

	normalFileContents := ""
	for i := 0; i < 100; i++ {
		normalFileContents += fmt.Sprint(i) + "\n"
	}

	if _, err := os.Stat(nonExistentFilepath); os.IsExist(err) {
		t.Fatal("unable to run test: temp files exists where it shouldn't", err)
	}
	if f, err := os.OpenFile(normalFilepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755); err == nil {
		_, err = f.WriteString(normalFileContents)
	} else if err != nil {
		t.Fatal("unable to run test: temp files could not be created", err)
	}

	newContent := "foo\n"

	t.Run("test non existent file", func(t *testing.T) {
		gotW, err := getWriter(nonExistentFilepath)
		gotW.Write([]byte(newContent))

		var res []byte

		if err != nil {
			t.Errorf("getWriter() error = %v, wantErr %v", err, false)
			return
		}
		if file, err := os.Open(nonExistentFilepath); err != nil {
			t.Errorf("getWriter() did not create an openable file: %v", err)
		} else if res, err = ioutil.ReadAll(file); err != nil {
			t.Errorf("getWriter() created a file that ioutil gave an error for: %v", err)
		}

		if string(res) != ""+newContent {
			t.Errorf("getWriter()'s wr saved %q instead of %q", unquote(string(res), t), unquote(""+newContent, t))
		}
	})

	t.Run("test existent file", func(t *testing.T) {
		gotW, err := getWriter(normalFilepath)
		gotW.Write([]byte(newContent))

		var res []byte

		if err != nil {
			t.Errorf("getWriter() error = %v, wantErr %v", err, false)
			return
		}
		if file, err := os.Open(normalFilepath); err != nil {
			t.Errorf("getWriter() did not create an openable file: %v", err)
		} else if res, err = ioutil.ReadAll(file); err != nil {
			t.Errorf("getWriter() created a file that ioutil gave an error for: %v", err)
		}

		if string(res) != normalFileContents+newContent {
			t.Errorf("getWriter()'s wr saved %q instead of %q", unquote(string(res), t), unquote(normalFileContents+newContent, t))
		}
	})

	t.Run("empty filepath", func(t *testing.T) {
		gotW, err := getWriter("")
		if err != nil {
			t.Errorf("getWriter() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(gotW, os.Stdout) {
			t.Errorf("getWriter() gotR = %v, want %v", gotW, os.Stdout)
		}
	})
}

func Test_report(t *testing.T) {
	lineCount := int64(94)

	name := "all in one"
	lineCountAddr := &lineCount
	customReportingDuration := time.Duration(10) * time.Millisecond
	stop := make(chan interface{}, 1)

	r, w, _ := os.Pipe()

	t.Run(name, func(t *testing.T) {
		go report(lineCountAddr, customReportingDuration, stop, w)

		time.Sleep(time.Duration(104) * time.Millisecond)
		close(stop)
		time.Sleep(time.Duration(10) * time.Millisecond)
		w.Close()

		res, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatalf("failed to read from mocked Stdin %v", err)
			return
		}

		resEscaped := unquote(string(res), t)

		if n := strings.Count(string(res), "\r"); n > 1 {
			t.Errorf("report() wanted returns, instead got %q", resEscaped)
			return
		}
	})
}

func unquote(res string, t *testing.T) string {
	resEscaped, err := strconv.Unquote(`"` + string(res) + `"`)
	if err != nil {
		t.Fatalf("failed to escape from from mocked Stdin %v", err)
	}
	return resEscaped
}

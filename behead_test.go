package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

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

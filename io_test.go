package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

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

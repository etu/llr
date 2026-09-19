package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

// TestStreamLinesMatchesPrintLines checks that streaming produces the same
// output as the batch path for the same input.
func TestStreamLinesMatchesPrintLines(t *testing.T) {
	inputs := map[string]string{
		"empty":                "",
		"single line":          "hello\n",
		"no trailing newline":  "hello\nworld",
		"blank lines":          "a\n\nb\n\n",
		"only newline":         "\n",
		"long lines":           "this line is far too long to fit\nshort\n",
		"tabs":                 "a\tb\tc\n",
		"ansi":                 "\x1b[31mred text that is long\x1b[0m\n",
		"wide runes":           "日本語のテキスト\n",
		"carriage returns":     "a\r\nb\r\n",
		"line longer than buf": strings.Repeat("x", 100000) + "\nend\n",
	}

	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			var want bytes.Buffer
			if err := printLines(&want, 10, strings.Split(input, "\n")); err != nil {
				t.Fatal(err)
			}

			var got bytes.Buffer
			if err := streamLines(strings.NewReader(input), &got, 10); err != nil {
				t.Fatal(err)
			}

			if got.String() != want.String() {
				t.Errorf("streamLines = %q, want %q", got.String(), want.String())
			}
		})
	}
}

// TestStreamLinesEmitsBeforeEOF is the reason streaming exists: with an
// input that stays open (like `journalctl -f`), each line must come out as
// soon as it is written rather than when the input is closed.
func TestStreamLinesEmitsBeforeEOF(t *testing.T) {
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	done := make(chan error, 1)
	go func() {
		err := streamLines(inR, outW, 5)
		outW.Close()
		done <- err
	}()

	out := bufio.NewReader(outR)
	readLine := func() string {
		t.Helper()
		type result struct {
			line string
			err  error
		}
		ch := make(chan result, 1)
		go func() {
			line, err := out.ReadString('\n')
			ch <- result{line, err}
		}()
		select {
		case r := <-ch:
			if r.err != nil {
				t.Fatalf("reading output: %v", r.err)
			}
			return r.line
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for output while input is still open")
			return ""
		}
	}

	for _, tc := range []struct{ in, want string }{
		{"hello world\n", "hello\n"},
		{"second line\n", "secon\n"},
		{"ok\n", "ok\n"},
	} {
		if _, err := inW.Write([]byte(tc.in)); err != nil {
			t.Fatal(err)
		}
		if got := readLine(); got != tc.want {
			t.Errorf("after writing %q got %q, want %q", tc.in, got, tc.want)
		}
	}

	inW.Close()
	if err := <-done; err != nil {
		t.Errorf("streamLines returned %v", err)
	}
}

type errReader struct {
	data string
	err  error
	done bool
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, r.err
	}
	r.done = true
	return copy(p, r.data), nil
}

// TestStreamLinesReadError checks that lines read before a failure are still
// printed and that the error is reported.
func TestStreamLinesReadError(t *testing.T) {
	wantErr := io.ErrUnexpectedEOF
	var got bytes.Buffer

	err := streamLines(&errReader{data: "one\npartial", err: wantErr}, &got, 80)

	if err != wantErr {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
	if got.String() != "one\npartial\n" {
		t.Errorf("output = %q, want %q", got.String(), "one\npartial\n")
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestStreamLinesWriteError(t *testing.T) {
	err := streamLines(strings.NewReader("hello\n"), failWriter{}, 80)
	if err != io.ErrClosedPipe {
		t.Errorf("err = %v, want %v", err, io.ErrClosedPipe)
	}
}

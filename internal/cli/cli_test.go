package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/catatsuy/bento/internal/cli"
)

func TestRun_versionFlg(t *testing.T) {
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cl := NewCLI(outStream, errStream, inputStream, nil)

	args := strings.Split("bento -version", " ")
	status := cl.Run(args)

	if status != ExitCodeOK {
		t.Errorf("ExitStatus=%d, want %d", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("bento version %s", Version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("Output=%q, want %q", errStream.String(), expected)
	}
}

func TestRun_helpSuccess(t *testing.T) {
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cl := NewCLI(outStream, errStream, inputStream, nil)

	args := strings.Split("bento -help", " ")
	status := cl.Run(args)

	if status != ExitCodeOK {
		t.Errorf("ExitStatus=%d, want %d", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("bento version %s", Version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("Output=%q, want %q", errStream.String(), expected)
	}
}

func TestRun_hSuccess(t *testing.T) {
	outStream, errStream, inputStream := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	cl := NewCLI(outStream, errStream, inputStream, nil)

	args := strings.Split("bento -h", " ")
	status := cl.Run(args)

	if status != ExitCodeOK {
		t.Errorf("ExitStatus=%d, want %d", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("bento version %s", Version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("Output=%q, want %q", errStream.String(), expected)
	}
}

func TestCLI_translateFile(t *testing.T) {
	ctx := context.TODO()

	tmpFile, err := os.Open("testdata/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	mockTranslator := &MockTranslator{
		TranslateTextFunc: func(ctx context.Context, prompt, text, model string) (string, error) {
			return "これはテストです。\nドットなしの別の行", nil
		},
	}

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cl := NewCLI(outStream, errStream, tmpFile, mockTranslator)

	err = cl.MultiRequest(ctx, "", "test", 1000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "これはテストです。\nドットなしの別の行\n"
	if outStream.String() != expected {
		t.Errorf("expected %q, but got %q", expected, outStream.String())
	}
}

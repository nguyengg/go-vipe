package vipe

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/shlex"
	"github.com/mattn/go-tty"
)

// From writes src io.Reader to a temporary file, opens the editor to that file, passing along any optional args, then
// return that file for reading with read index at start of file.
//
// Caller is responsible for closing and deleting the returned file once done.
//
// Environment variable EDITOR and VISUAL may be used to select the editor which defaults to vi. If the editor exits
// with a non-zero status code, the returned error wraps an exec.ExitError.
func From(src io.Reader, args ...string) (*os.File, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("vipe: create temp error: %w", err)
	}

	if _, err = io.Copy(f, src); err != nil {
		return nil, fmt.Errorf("vipe: write temp error: %w", err)
	}

	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("vipe: close temp error: %w", err)
	}

	name := f.Name()
	if err = FromFile(name, args...); err != nil {
		_ = os.Remove(name)
		return nil, err
	}

	f, err = os.Open(name)
	if err != nil {
		_ = os.Remove(name)
		return nil, fmt.Errorf("vipe: reopen temp error: %w", err)
	}

	return f, nil
}

// FromFile is a variant of From that uses an existing file instead.
//
// Call this if you want complete control over the file, or if you've a file ready for editing. If the editor exits
// with a non-zero status code, the returned error wraps an exec.ExitError.
func FromFile(name string, args ...string) (err error) {
	editor := []string{"vi"}
	args = append(args, name)

	// I don't like /usr/bin/editor so skipping it.

	if v, ok := os.LookupEnv("EDITOR"); ok {
		if editor, err = shlex.Split(v); err != nil {
			return
		}
	}
	if v, ok := os.LookupEnv("VISUAL"); ok {
		if editor, err = shlex.Split(v); err != nil {
			return
		}
	}

	t, err := tty.Open()
	if err != nil {
		return
	}
	defer func() {
		_ = t.Close()
	}()

	cmd := exec.Command(editor[0], append(editor[1:], args...)...)
	cmd.Stdin = t.Input()
	cmd.Stdout = t.Output()
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("vipe: %s exited nonzero, aborting: %w", strings.Join(editor, " "), err)
		}
		return fmt.Errorf("vipe: %s exec error: %w", strings.Join(editor, " "), err)
	}

	return nil
}

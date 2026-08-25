package xt //nolint:testpackage

import (
	"bytes"
	"os"
	"testing"
)

func TestFileModeUnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    os.FileMode
		wantErr bool
	}{
		{name: "unquoted", input: "644", want: 0o644},
		{name: "leading zero", input: "0644", want: 0o644},
		{name: "double quoted", input: `"0755"`, want: 0o755},
		{name: "single quoted", input: "'0755'", want: 0o755},
		{name: "padded", input: "  0644  ", want: 0o644},
		{name: "invalid letters", input: "abc", wantErr: true},
		{name: "invalid octal", input: "999", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var mode FileMode

			err := mode.UnmarshalText([]byte(test.input))
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if mode.Mode() != test.want {
				t.Fatalf("mode = %o, want %o", mode.Mode(), test.want)
			}
		})
	}
}

func TestFileModeMarshalAndString(t *testing.T) {
	t.Parallel()

	mode := FileMode(0o644)

	got, err := mode.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, []byte("0644")) {
		t.Fatalf("MarshalText = %q", got)
	}

	if mode.String() != "0644" {
		t.Fatalf("String = %q", mode.String())
	}
}

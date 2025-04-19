package xt

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FileMode is used to unmarshal a unix file mode from the config file.
type FileMode os.FileMode //nolint:recvcheck // it works this way...

// UnmarshalText turns a unix file mode, wrapped in quotes or not, into a usable os.FileMode.
func (f *FileMode) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(strings.Trim(string(text), `"'`))

	fm, err := strconv.ParseUint(str, 8, 32)
	if err != nil {
		return fmt.Errorf("file_mode (%s) is invalid: %w", str, err)
	}

	*f = FileMode(os.FileMode(fm))

	return nil
}

// MarshalText satisfies an encoder.TextMarshaler interface.
func (f FileMode) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// String creates a unix-octal version of a file mode.
func (f FileMode) String() string {
	return fmt.Sprintf("%04o", f)
}

// Mode returns the compatible os.FileMode.
func (f FileMode) Mode() os.FileMode {
	return os.FileMode(f)
}

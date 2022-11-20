package editor

import (
	"bytes"
)

type CodeEditor func(*bytes.Buffer) (*bytes.Buffer, error)

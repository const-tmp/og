package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/fs"
	"os"
	"testing"
)

func TestName34(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write([]byte("start"))
	for i := 0; i < 10; i++ {
		read, err := io.ReadAll(b)
		require.NoError(t, err)
		//n, err := b.Read(p)
		//require.NoError(t, err)
		t.Log("read", string(read))
		b.Write([]byte(fmt.Sprint(i)))
	}
	t.Log(b.String())
	b.Write([]byte("sdfdsf"))
	t.Log(b.String())

}

func TestDfskjdnf(t *testing.T) {
	f, err := os.Open(".gitigdnore")
	if err != nil {
		t.Logf("%v\t%T", err, err)
	} else {
		f.Close()
	}
	t.Log(errors.Is(err, fs.ErrNotExist))
}

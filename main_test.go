package main

import (
	"github.com/nullc4t/gensta/pkg/inspector"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		t.Log(path, info, err)
		return nil
	})
	require.NoError(t, err)
	err = filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
		t.Log(path, d, err)
		return nil
	})
	files, err := filepath.Glob("go.mod")
	require.NoError(t, err)
	for i, file := range files {
		t.Log(i, file)
	}
	info, err := os.ReadDir(wd)
	require.NoError(t, err)
	for i, entry := range info {
		t.Log(i, entry.Name(), entry.Type(), entry.IsDir())
		fi, err := entry.Info()
		require.NoError(t, err)
		t.Log(fi)
	}
}

func TestPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Log(wd)
	t.Log(filepath.Join(wd, ".."))

	path, err := inspector.SearchFileUp("go.mod", wd, 3)
	require.NoError(t, err)
	t.Log(path)

	path, err = inspector.SearchFileDown("go.mod")
	require.NoError(t, err)
	t.Log(path)

	mod, err := inspector.GetModuleNameFromGoMod(path)
	require.NoError(t, err)
	t.Log(mod)
}

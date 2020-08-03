package gitgetter

import (
	"errors"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func path(p string) string {
	return filepath.FromSlash(p)
}

func writeFile(fs billy.Filesystem, filename string, data []byte) error {
	f, err := fs.OpenFile(path(filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func TestReadFile(t *testing.T) {
	treeReader := FileTreeReader{fs: memfs.New()}
	content := []byte("test")
	err := writeFile(treeReader.fs, "/test.txt", content)
	require.NoError(t, err)

	b, err := treeReader.ReadFile("test.txt")
	require.NoError(t, err)
	require.Equal(t, b, content)
}

type file struct {
	path  string
	isDir bool
}

func TestWalk(t *testing.T) {
	treeReader := FileTreeReader{fs: memfs.New()}
	err := writeFile(treeReader.fs, "/file.txt", []byte{})
	require.NoError(t, err)
	err = writeFile(treeReader.fs, "/dir/file.txt", []byte{})
	require.NoError(t, err)

	visited := make([]file, 0, 4)
	err = treeReader.Walk(func(filepath string, isDir bool, err error) error {
		visited = append(visited, file{filepath, isDir})
		return nil
	})
	require.NoError(t, err)

	require.Len(t, visited, 4)
	require.Contains(t, visited, file{path("/"), true})
	require.Contains(t, visited, file{path("/dir"), true})
	require.Contains(t, visited, file{path("/file.txt"), false})
	require.Contains(t, visited, file{path("/dir/file.txt"), false})
}

func TestWalkSkipDir(t *testing.T) {
	treeReader := FileTreeReader{fs: memfs.New()}
	err := writeFile(treeReader.fs, "/file.txt", []byte{})
	require.NoError(t, err)

	err = treeReader.Walk(func(fpath string, isDir bool, err error) error {
		if isDir {
			return filepath.SkipDir
		}
		require.Fail(t, "directory not skipped", "visited the file %q in the skipped directory", fpath)
		return nil
	})
	require.NoError(t, err)
}

func TestWalkError(t *testing.T) {
	treeReader := FileTreeReader{fs: memfs.New()}
	err := writeFile(treeReader.fs, "/file.txt", []byte{})
	require.NoError(t, err)

	err = treeReader.Walk(func(fpath string, isDir bool, err error) error {
		if fpath != path("/") {
			return errors.New("test walk")
		}
		return nil
	})
	require.EqualError(t, err, "test walk")
}

func TestGetter(t *testing.T) {
	var options *git.CloneOptions
	clone = func(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error) {
		options = o

		err := writeFile(worktree, "/a.txt", []byte{})
		require.NoError(t, err)
		err = writeFile(worktree, "/subdir/b.txt", []byte{})
		require.NoError(t, err)
		return nil, nil
	}

	t.Run("Head ref by default", func(t *testing.T) {
		_, err := Get("http://test.com/")
		require.NoError(t, err)
		require.Equal(t, plumbing.HEAD, options.ReferenceName)
	})

	t.Run("Custom reference", func(t *testing.T) {
		_, err := Get("http://test.com#refs/tags/1.0")
		require.NoError(t, err)
		require.Equal(t, "refs/tags/1.0", string(options.ReferenceName))
	})

	t.Run("Subdirectory", func(t *testing.T) {
		r, err := Get("http://test.com//subdir")
		require.NoError(t, err)

		visited := make([]string, 0, 2)
		err = r.Walk(func(fpath string, isDir bool, err error) error {
			visited = append(visited, fpath)
			return nil
		})
		require.NoError(t, err)

		require.Equal(t, []string{path("/"), path("/b.txt")}, visited)
	})
}

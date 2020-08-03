package gitgetter

import (
	"bytes"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"path/filepath"
)

var clone = git.Clone

// Get clones git repository into in-memory filesystem and returns
// FileTreeReader with repository source files. An URL can be any
// valid URL supported by official git client. In case a host is
// a known one (github.com, bitbucket.org, gitlab.com, gitea.com)
// then specifying protocol is optional.
//
// Git references can be specified with a hash fragment at the end
// of the URL: `host.com/repository#refs/tags/1.0.0`
//
// Subdirectories can be specified with a double-slash within path
// part of the URL. If a host is a known one, then directory will
// be resolved automatically after single slash, considering path
// consist of user and repository name only.
//
// Examples:
// ```
// http://unknownhost.com/user/repo//subdirectory
// github.com/user/repo/subdirectory
// ```
func Get(repourl string) (*FileTreeReader, error) {
	r, err := parseUrl(repourl)
	if err != nil {
		return nil, err
	}
	ref := plumbing.HEAD
	if r.ref != "" {
		ref = plumbing.ReferenceName(r.ref)
	}
	fs := memfs.New()
	_, err = clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:               r.raw,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
		ReferenceName:     ref,
		Depth:             1,
		Tags:              git.NoTags,
	})
	if err != nil {
		return nil, err
	}
	if r.subdir != "" {
		fs = chroot.New(fs, r.subdir)
	}
	return &FileTreeReader{fs}, nil
}

// WalkFunc is the type of the function called for each file or directory
// visited by Walk. The path argument contains the argument to Walk as a
// prefix; that is, if Walk is called with "dir", which is a directory
// containing the file "a", the walk function will be called with argument
// "dir/a".
//
// If there was a problem walking to the file or directory named by path, the
// incoming error will describe the problem and the function can decide how
// to handle that error (and Walk will not descend into that directory). If an
// error is returned, processing stops. The sole exception is when the function
// returns the special value filepath.SkipDir. If the function returns SkipDir
// when invoked on a directory, Walk skips the directory's contents entirely.
// If the function returns SkipDir when invoked on a non-directory file,
// Walk skips the remaining files in the containing directory.
type walkFunc = func(filepath string, isDir bool, err error) error

// FileTreeReader provides an API to read directory and it's files
// from Billy's filesystem.
type FileTreeReader struct {
	fs billy.Filesystem
}

// Walk walks the file tree, calling walkFn for each file or directory
// in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func (r FileTreeReader) Walk(walkFn walkFunc) error {
	if err := r.walk(string(filepath.Separator), true, walkFn); err != filepath.SkipDir {
		return err
	}
	return nil
}

func (r FileTreeReader) walk(path string, isDir bool, walkFn walkFunc) error {
	err := walkFn(path, isDir, nil)
	if err != nil {
		return err
	}
	if !isDir {
		return nil
	}

	infos, err := r.fs.ReadDir(path)
	if err != nil {
		return walkFn(path, isDir, err)
	}

	for _, info := range infos {
		filename := r.fs.Join(path, info.Name())
		err = r.walk(filename, info.IsDir(), walkFn)
		if err != nil {
			if !info.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}

// ReadFile reads the file from file tree named by filename and returns
// the contents. A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from
// Read as an error to be reported.
func (r FileTreeReader) ReadFile(filename string) ([]byte, error) {
	f, err := r.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var capacity int64 = bytes.MinRead

	if fi, err := r.fs.Stat(filename); err == nil {
		// As initial capacity for readAll, use Size + a little extra in case Size
		// is zero, and to avoid another allocation after Read has filled the
		// buffer. The readAll call will read into its allocated internal buffer
		// cheaply. If the size was wrong, we'll either waste some space off the end
		// or reallocate as needed, but in the overwhelmingly common case we'll get
		// it just right.
		if size := fi.Size() + bytes.MinRead; size > capacity {
			capacity = size
		}
	}
	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(f)
	return buf.Bytes(), err
}

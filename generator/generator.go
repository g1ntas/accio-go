package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const templateExt = ".accio"

const (
	fileRegular fileType = iota
	fileDir
)

type fileType uint

type OptionFn func(*Runner)

// OnExistsFn handles files that already exist at target path.
// Return true to overwrite file or false to skip it.
type OnExistsFn func(path string) bool

// OnErrorFn is called if error occurred when processing file.
// Return true to skip the file and continue process, or false
// to terminate Runner and return the error.
type OnErrorFn func(err error) bool

// OnSuccessFn is called on each successfully generated file.
// First argument holds path of the source file, and second
// argument - path of the generated file.
type OnSuccessFn func(src, dst string)

type blueprint = struct {
	Body     string
	Filename string
	Skip     bool
}

type BlueprintParser interface {
	Parse(b []byte) (*blueprint, error)
}

// FileTreeReader is an abstraction over any system-agnostic
// file tree. In the case of generator, it provides full structure,
// that should be scanned, read and generated at the filepath relative
// to the working directory.
type FileTreeReader interface {
	// ReadFile reads the file from file tree named by filename and returns
	// the contents.
	ReadFile(filename string) ([]byte, error)

	// Walk walks the file tree, calling walkFn for each file or directory
	// in the tree, including root. All errors that arise visiting files
	// and directories are filtered by walkFn.
	Walk(walkFn func(filepath string, isDir bool, err error) error) error
}

type Filesystem interface {
	WriteFile(name string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

type RunError struct {
	Err  error
	Path string
}

func (e *RunError) Error() string {
	return fmt.Sprintf("generating %s: %s", e.Path, e.Err.Error())
}

func (e *RunError) Unwrap() error {
	return e.Err
}

func IgnoreFile(p string) OptionFn {
	p = normalizePath(p)
	return func(r *Runner) {
		r.ignore[p] = fileRegular
	}
}

func IgnoreDir(p string) OptionFn {
	p = normalizePath(p)
	return func(r *Runner) {
		r.ignore[p] = fileDir
	}
}

func OnFileExists(fn OnExistsFn) OptionFn {
	return func(r *Runner) {
		r.onExists = fn
	}
}

func OnError(fn OnErrorFn) OptionFn {
	return func(r *Runner) {
		r.onError = fn
	}
}

func OnSuccess(fn OnSuccessFn) OptionFn {
	return func(r *Runner) {
		r.onSuccess = fn
	}
}

type Runner struct {
	fs        Filesystem
	mp        BlueprintParser
	writeDir  string // absolute path to the directory to write generated files
	onExists  OnExistsFn
	onError   OnErrorFn
	onSuccess OnSuccessFn
	// ignore defines files to ignore during run, where key is a filepath within generator's structure
	ignore map[string]fileType
}

func NewRunner(fs Filesystem, mp BlueprintParser, dir string, options ...OptionFn) *Runner {
	r := &Runner{
		fs:       fs,
		mp:       mp,
		writeDir: dir,
		ignore:   make(map[string]fileType),
		onExists: func(_ string) bool {
			return false
		},
		onError: func(_ error) bool {
			return false
		},
		onSuccess: func(_, _ string) {},
	}
	for _, option := range options {
		option(r)
	}
	return r
}

// Run generates all the files from FileTreeReader by walking over
// each file, reading it and writing it at relative path in working
// directory. If file ends with extension `.accio`, then it's parsed with
// BlueprintParser, which returns file's content and additional metadata,
// like custom filepath, and whether file should be skipped.
func (r *Runner) Run(ftr FileTreeReader) error {
	return ftr.Walk(func(fpath string, isDir bool, err error) error {
		if err != nil {
			return r.handleError(err, fpath)
		}
		fpath = normalizePath(fpath)
		// skip specified files and directories
		if ignoredFile, ok := r.ignore[fpath]; ok {
			switch {
			case isDir && ignoredFile == fileDir:
				return filepath.SkipDir
			case !isDir && ignoredFile == fileRegular:
				return nil
			}
		}
		// do nothing with directories
		if isDir {
			return nil
		}
		body, err := ftr.ReadFile(fpath)
		if err != nil {
			return r.handleError(err, fpath)
		}
		target := filepath.Join(r.writeDir, fpath)
		if hasTemplateExtension(target) {
			target = target[:len(target)-len(templateExt)] // remove ext
			tpl, err := r.mp.Parse(body)
			switch {
			case err != nil:
				return r.handleError(err, fpath)
			case tpl.Skip:
				return nil
			case tpl.Filename != "":
				basename := filepath.Base(target)
				target = joinWithinRoot(r.writeDir, tpl.Filename)
				stat, err := r.fs.Stat(target)
				// if path is directory, then attach filename of source file
				if err == nil && stat.IsDir() {
					target = filepath.Join(target, basename)
				}
			}
			body = []byte(tpl.Body)
		}
		// if file exists, call callback to decide if it should be skipped
		_, err = r.fs.Stat(target)
		if err == nil && !r.onExists(target) {
			return nil
		}
		if err != nil && !os.IsNotExist(err) {
			return r.handleError(err, fpath)
		}
		err = r.fs.MkdirAll(filepath.Dir(target), 0755)
		if err != nil {
			return r.handleError(err, fpath)
		}
		err = r.fs.WriteFile(target, body, 0775)
		if err != nil {
			return r.handleError(err, fpath)
		}
		r.onSuccess(fpath, target)
		return nil
	})
}

func (r *Runner) handleError(err error, path string) error {
	err = &RunError{err, path}
	if r.onError(err) {
		return nil
	}
	return err
}

func hasTemplateExtension(path string) bool {
	return len(path) > len(templateExt) && path[len(path)-len(templateExt):] == templateExt
}

// joinWithinRoot joins two paths ensuring that one (relative) path ends up
// inside the other (root) path. If relative path evaluates to be outside root
// directory, then it's treated as there's no parent directory and root is final.
func joinWithinRoot(root, relpath string) string {
	sep := string(filepath.Separator)
	parts := strings.Split(filepath.Clean(relpath), sep)
	for _, part := range parts {
		if part != ".." {
			break
		}
		parts = parts[1:]
	}
	return filepath.Join(root, strings.Join(parts, sep))
}

// normalizePath cleans up the path and normalizes it so it can
// be compared with other paths referring to the same file but
// containing different path format.
func normalizePath(p string) string {
	p = filepath.Clean(p)
	if len(p) > 0 && p[0] == filepath.Separator {
		return p[1:]
	}
	return p
}

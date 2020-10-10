package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const templateExt = ".accio"

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

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
}

type NopLogger struct{}

func (l NopLogger) Debug(_ ...interface{}) {
}

func (l NopLogger) Info(_ ...interface{}) {
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

func WithLogger(l Logger) OptionFn {
	return func(r *Runner) {
		r.log = l
	}
}

func SkipErrors(r *Runner) {
	r.skipErrors = true
}

func IgnorePath(p string) OptionFn {
	p = normalizePath(p)
	return func(r *Runner) {
		r.ignore[p] = struct{}{}
	}
}

func OnFileExists(fn OnExistsFn) OptionFn {
	return func(r *Runner) {
		r.onExists = fn
	}
}

type Runner struct {
	fs         Filesystem
	bluepr     BlueprintParser
	log        Logger
	writeDir   string // absolute path to the directory to write generated files
	skipErrors bool
	onExists   OnExistsFn
	// ignore defines files to ignore during run, where key is a filepath within generator's structure
	ignore map[string]struct{}
}

func NewRunner(fs Filesystem, bp BlueprintParser, dir string, options ...OptionFn) *Runner {
	r := &Runner{
		fs:       fs,
		bluepr:   bp,
		log:      NopLogger{},
		writeDir: dir,
		ignore:   make(map[string]struct{}),
		onExists: func(_ string) bool {
			return false
		},
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
		r.log.Debug("visiting ", fpath)
		// skip specified files and directories
		if _, ok := r.ignore[fpath]; ok {
			if isDir {
				r.log.Debug("skip directory")
				return filepath.SkipDir
			}
			r.log.Debug("skip file")
			return nil
		}
		// do nothing with directories
		if isDir {
			r.log.Debug("is a directory, do nothing")
			return nil
		}
		body, err := ftr.ReadFile(fpath)
		if err != nil {
			return r.handleError(err, fpath)
		}
		target := filepath.Join(r.writeDir, fpath)
		r.log.Debug("file will be written at ", target)
		if hasTemplateExtension(target) {
			r.log.Debug("file is a blueprint, parsing...")
			target = target[:len(target)-len(templateExt)] // remove ext
			tpl, err := r.bluepr.Parse(body)
			switch {
			case err != nil:
				return r.handleError(err, fpath)
			case tpl.Skip:
				r.log.Debug("blueprint: skipping file...")
				return nil
			case tpl.Filename != "":
				basename := filepath.Base(target)
				target = joinWithinRoot(r.writeDir, tpl.Filename)
				stat, err := r.fs.Stat(target)
				// if path is directory, then attach filename of source file
				if err == nil && stat.IsDir() {
					target = filepath.Join(target, basename)
				}
				r.log.Debug("blueprint: file's write destination changed to ", target)
			}
			body = []byte(tpl.Body)
		}
		// if file exists, call callback to decide if it should be skipped
		_, err = r.fs.Stat(target)
		if err == nil && !r.onExists(target) {
			r.log.Debug("file already exists, skipping...")
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
		r.log.Debug("file created at", fpath)
		return nil
	})
}

func (r *Runner) handleError(err error, path string) error {
	err = &RunError{err, path}
	if r.skipErrors {
		r.log.Info("ERROR: ", err.Error(), ". Skipping...")
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

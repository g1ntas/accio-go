package generator

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const templateExt = ".accio"

type Template interface {
	Body() []byte
	Filename() string
	Skip() bool
}

type TemplateEngine interface {
	Parse(b []byte, data map[string]interface{}) (Template, error)
}

type Runner struct {
	prompter Prompter
	fs ReaderWalkerWriter
	tplEngine TemplateEngine
}

func NewRunner(p Prompter, fs ReaderWalkerWriter, tpl TemplateEngine) *Runner {
	return &Runner{p, fs, tpl}
}

func (r *Runner) Run(gen *Generator, writeDir string, forceOverwrite bool) error {
	manifestPath := path.Join(gen.Dest, manifestFilename)
	data, err := gen.prompt(r.prompter)
	if err != nil {
		return err
	}
	return r.fs.Walk(gen.Dest, func(abspath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ignore non-files and manifest
		if !info.Mode().IsRegular() || abspath == manifestPath {
			return nil
		}
		b, err := r.fs.ReadFile(abspath)
		if err != nil {
			return err
		}
		relpath, err := filepath.Rel(gen.Dest, abspath)
		if err != nil {
			return err
		}
		if r.hasTemplateExtension(abspath) {
			tpl, err := r.tplEngine.Parse(b, data)
			if err != nil {
				return err
			}
			if f := tpl.Filename(); f != "" {
				relpath = f
			} else {
				// remove template extension
				relpath = relpath[:len(relpath)-len(templateExt)]
			}
			if tpl.Skip() {
				return nil
			}
			b = tpl.Body()
		}
		err = r.fs.WriteFile(joinWithinRoot(writeDir, relpath), b, info.Mode())
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *Runner) hasTemplateExtension(path string) bool {
	l := len(path)
	return path[l-len(templateExt):l] == templateExt
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
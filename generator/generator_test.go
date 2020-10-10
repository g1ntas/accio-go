package generator

import (
	"encoding/json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// blueprintBlueprintMock represents BlueprintParser implementation.
type blueprintParserMock struct{}

var _ BlueprintParser = (*blueprintParserMock)(nil)

// Parse decodes json into Blueprint
func (e *blueprintParserMock) Parse(b []byte) (*blueprint, error) {
	tpl := &blueprint{}
	err := json.Unmarshal(b, tpl)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

// fileTreeReaderMock implements FileTreeReader for testing.
type fileTreeReaderMock struct {
	fs afero.Fs
}

var _ FileTreeReader = (*fileTreeReaderMock)(nil)

func (r *fileTreeReaderMock) Walk(walkFn func(filepath string, isDir bool, err error) error) error {
	return afero.Walk(r.fs, "", func(path string, info os.FileInfo, err error) error {
		return walkFn(path, info.IsDir(), err)
	})
}

func (r *fileTreeReaderMock) ReadFile(fpath string) ([]byte, error) {
	return afero.ReadFile(r.fs, fpath)
}

// fsOpFn performs operation on the given filesystem.
type fsOpFn func(afero.Fs) error

// assertFn performs test assertion on the given filesystem.
type assertFn func(*testing.T, afero.Fs)

var overwriteExisting = OnFileExists(func(p string) bool {
	return true
})

var noOutput = []assertFn{doesntExist("/output")}

// file creates a file with a specified content at target path.
func file(filename string, content string) fsOpFn {
	return func(fs afero.Fs) error {
		return afero.WriteFile(fs, filename, []byte(content), 0775)
	}
}

// dir creates a directory at target path.
func dir(dirname string) fsOpFn {
	return func(fs afero.Fs) error {
		return fs.Mkdir(dirname, 0755)
	}
}

// doesntExist asserts that target path doesn't exist.
func doesntExist(filename string) assertFn {
	return func(t *testing.T, fs afero.Fs) {
		exists, err := afero.Exists(fs, filename)
		require.NoError(t, err)
		require.Falsef(t, exists, "file %s exists", filename)
	}
}

// fileExists asserts if target file exists and contains given content.
func fileExists(filename string, content string) assertFn {
	return func(t *testing.T, fs afero.Fs) {
		b, err := afero.ReadFile(fs, filename)
		if os.IsNotExist(err) {
			require.FailNowf(t, "expected file or directory", "file %s doesn't exist", filename)
		}
		require.NoError(t, err)
		require.Equal(t, content, string(b))
	}
}

var runnerTests = []struct {
	name    string
	input   []fsOpFn   // sequence of FS operations to prepare initial generator directory structure
	output  []assertFn // sequence of test assertions to perform after runner completes
	options []OptionFn
}{
	{
		"no files",
		[]fsOpFn{},
		noOutput,
		[]OptionFn{},
	},
	{
		"ignore directories",
		[]fsOpFn{dir("/generator/a")},
		noOutput,
		[]OptionFn{},
	},
	{
		"write static file",
		[]fsOpFn{file("/generator/a.txt", "test")},
		[]assertFn{fileExists("/output/a.txt", "test")},
		[]OptionFn{},
	},
	{
		"write multiple files",
		[]fsOpFn{file("/generator/a.txt", "file1"), file("/generator/b.txt", "file2")},
		[]assertFn{fileExists("/output/a.txt", "file1"), fileExists("/output/b.txt", "file2")},
		[]OptionFn{},
	},
	{
		"write nested files",
		[]fsOpFn{dir("/generator/abc"), file("/generator/abc/test.txt", "file")},
		[]assertFn{fileExists("/output/abc/test.txt", "file")},
		[]OptionFn{},
	},
	{
		"write blueprint file",
		[]fsOpFn{file("/generator/test.txt.accio", `{"body": "test"}`)},
		[]assertFn{fileExists("/output/test.txt", "test")},
		[]OptionFn{},
	},
	{
		"blueprint | skip file",
		[]fsOpFn{file("/generator/test.txt.accio", `{"skip": true}`)},
		noOutput,
		[]OptionFn{},
	},
	{
		"blueprint | custom filename",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "custom.txt", "body": "---"}`)},
		[]assertFn{fileExists("/output/custom.txt", "---")},
		[]OptionFn{},
	},
	{
		"blueprint | nested custom filename",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "dir/custom.txt"}`)},
		[]assertFn{fileExists("/output/dir/custom.txt", "")},
		[]OptionFn{},
	},
	{
		"blueprint | append static name if filename is directory",
		[]fsOpFn{dir("/output/abc"), file("/generator/test.txt.accio", `{"filename": "abc"}`)},
		[]assertFn{fileExists("/output/abc/test.txt", "")},
		[]OptionFn{},
	},
	{
		"blueprint | don't write outside root",
		[]fsOpFn{file("/generator/test.txt.accio", `{"filename": "../../../custom.txt"}`)},
		[]assertFn{fileExists("/output/custom.txt", "")},
		[]OptionFn{},
	},
	{
		"blueprint | write to root",
		[]fsOpFn{file("/generator/abc/test.txt.accio", `{"filename": "custom.txt"}`)},
		[]assertFn{fileExists("/output/custom.txt", "")},
		[]OptionFn{},
	},
	{
		"overwrite if file exists",
		[]fsOpFn{file("/generator/test.txt", "new"), file("/output/test.txt", "old")},
		[]assertFn{fileExists("/output/test.txt", "new")},
		[]OptionFn{overwriteExisting},
	},
	{
		"skip if file exists by default",
		[]fsOpFn{file("/generator/test.txt", "new"), file("/output/test.txt", "old")},
		[]assertFn{fileExists("/output/test.txt", "old")},
		[]OptionFn{},
	},
	{
		"ignore file",
		[]fsOpFn{file("/generator/ignore.txt", ""), file("/generator/file.txt", "")},
		[]assertFn{doesntExist("/output/ignore.txt"), fileExists("/output/file.txt", "")},
		[]OptionFn{IgnorePath("ignore.txt")},
	},
	{
		"ignore directory",
		[]fsOpFn{file("/generator/ignore/a.txt", ""), file("/generator/ignore/b.txt", "")},
		[]assertFn{doesntExist("/output/ignore/a.txt"), doesntExist("/output/ignore/b.txt")},
		[]OptionFn{IgnorePath("ignore")},
	},
}

func TestRunner(t *testing.T) {
	for _, test := range runnerTests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			err := fs.Mkdir("/generator", 0755)
			require.NoError(t, err)

			// build initial FS structure
			for _, fsOperation := range test.input {
				err := fsOperation(fs)
				require.NoError(t, err)
			}

			runner := NewRunner(
				fs,
				&blueprintParserMock{},
				"/output",
				test.options...,
			)

			fileTree := &fileTreeReaderMock{fs: afero.NewBasePathFs(fs, "/generator")}
			err = runner.Run(fileTree)
			require.NoError(t, err)

			// assert final FS structure
			for _, assertion := range test.output {
				assertion(t, fs)
			}
		})
	}
}

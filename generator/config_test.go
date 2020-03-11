package generator

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	noError  = true
	hasError = false
)

var emptyGen = Generator{}

// conf is alias type for config to improve readability
type conf = map[string]interface{}

// strOfLen generates string of length n
func strOfLen(n int) string {
	return strings.Repeat("a", n)
}

// string creates human-readable representation of Generator
func (g *Generator) string() string {
	// shorten help if too long
	var help string
	if len(g.Help) > 10 {
		help = fmt.Sprintf("%.10s...", g.Help)
	} else {
		help = g.Help
	}
	// stringify prompts in format [var]:[type]
	prompts := make([]string, len(g.Prompts))
	i := 0
	for k, p := range g.Prompts {
		prompts[i] = k + ":" + p.kind()
		i++
	}
	return fmt.Sprintf("%q %v", help, prompts)
}

var configTests = []struct {
	name  string
	input map[string]interface{}
	gen   Generator
	ok    bool
}{
	// help
	{
		"help",
		conf{"help": "abc"},
		Generator{Help: "abc", Prompts: PromptMap{}},
		noError,
	},

	// prompts
	{
		"Prompt empty type",
		conf{"prompts": conf{"var": conf{"type": "", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt invalid type",
		conf{"prompts": conf{"var": conf{"type": "invalid", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt empty message",
		conf{"prompts": conf{"var": conf{"type": "input", "message": ""}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt message longer than 128 characters",
		conf{"prompts": conf{"var": conf{"type": "input", "message": strOfLen(129)}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt var name longer than 64 characters",
		conf{"prompts": conf{strOfLen(65): conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt with valid var name",
		conf{"prompts": conf{"_Var_1": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"_Var_1": &input{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt with var name starting with digit",
		conf{"prompts": conf{"0var": conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt with var name containing invalid characters",
		conf{"prompts": conf{"test-var": conf{"type": "input", "message": "test"}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt type input",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &input{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt help",
		conf{"prompts": conf{"var": conf{"type": "input", "message": "test", "help": "abc"}}},
		Generator{Prompts: PromptMap{"var": &input{Base{Msg: "test", HelpText: "abc"}}}},
		noError,
	},
	{
		"Prompt type integer",
		conf{"name": "a", "prompts": conf{"var": conf{"type": "integer", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &integer{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type confirm",
		conf{"prompts": conf{"var": conf{"type": "confirm", "message": "test"}}},
		Generator{Prompts: PromptMap{"var": &confirm{Base{Msg: "test"}}}},
		noError,
	},
	{
		"Prompt type choice",
		conf{"prompts": conf{"var": conf{
			"type":    "choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Prompts: PromptMap{"var": &choice{
			Base{Msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
	{
		"Prompt 'choice' without options",
		conf{"prompts": conf{"var": conf{
			"type":    "choice",
			"message": "test",
		}}},
		emptyGen,
		hasError,
	},
	{
		"Prompt type multi choice",
		conf{"prompts": conf{"var": conf{
			"type":    "multi-choice",
			"options": []string{"a", "b"},
			"message": "test",
		}}},
		Generator{Prompts: PromptMap{"var": &multiChoice{
			Base{Msg: "test"},
			[]string{"a", "b"},
		},
		}},
		noError,
	},
	{
		"Prompt 'multi-choice' without options",
		conf{"prompts": conf{"var": conf{
			"type":    "multi-choice",
			"message": "test",
		}}},
		emptyGen,
		hasError,
	},
}

type mockReader struct {
	content []byte
}

func (r *mockReader) ReadFile(_ string) ([]byte, error) {
	return r.content, nil
}

func TestConfigReading(t *testing.T) {
	for _, test := range configTests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := toml.NewEncoder(buf).Encode(test.input)
			require.NoError(t, err)

			gen := &Generator{Prompts: make(PromptMap)}
			err = gen.ReadConfig(&mockReader{buf.Bytes()})
			switch {
			case !test.ok:
				assert.Error(t, err)
			case test.ok:
				assert.NoError(t, err)
			default:
				assert.Equal(t, &test.gen, gen)
			}
		})
	}
}

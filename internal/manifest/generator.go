package manifest

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"sort"
	"unicode"
)

type Generator struct {
	Help    string    `toml:"help"`
	Ignore  []string  `toml:"ignore"`
	Prompts PromptMap `toml:"prompts"`
}

func NewGenerator() *Generator {
	return &Generator{
		Prompts: make(PromptMap),
	}
}

func (g *Generator) PromptAll(prompter Prompter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	// sort prompts by keys, so they always appear in the same order
	keys, i := make([]string, len(g.Prompts)), 0
	for k := range g.Prompts {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		val, err := g.Prompts[k].Prompt(prompter)
		if err != nil {
			return map[string]interface{}{}, err
		}
		data[k] = val
	}
	return data, nil
}

func ReadToml(b []byte) (*Generator, error) {
	g := NewGenerator()
	err := toml.Unmarshal(b, &g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (m PromptMap) UnmarshalTOML(data interface{}) error {
	prompts := data.(map[string]interface{})
	for key, def := range prompts {
		if err := validatePromptKey(key); err != nil {
			return err
		}
		mapping := def.(map[string]interface{})
		typ, ok := mapping["type"].(string)
		if !ok {
			return fmt.Errorf("missing type for prompt %q", key)
		}
		base := Base{}
		var err error
		if base.Msg, err = parsePromptMessage(mapping, key); err != nil {
			return err
		}
		base.HelpText, _ = mapping["help"].(string)
		switch typ {
		case promptInput:
			m[key] = &input{base}
		case promptInteger:
			m[key] = &integer{base}
		case promptConfirm:
			m[key] = &confirm{base}
		case promptChoice, promptMultiChoice:
			opts, err := parsePromptOptions(mapping, key)
			if err != nil {
				return err
			}
			if typ == promptChoice {
				m[key] = &choice{base, opts}
			} else {
				m[key] = &multiChoice{base, opts}
			}
		default:
			return fmt.Errorf("unknown type %q in prompt %q", typ, key)
		}
	}
	return nil
}

func parsePromptOptions(conf map[string]interface{}, key string) ([]string, error) {
	opts, ok := conf["options"].([]interface{})
	if !ok || len(opts) == 0 {
		return []string{}, fmt.Errorf("options for prompt %q were not specified or are invalid", key)
	}
	options := make([]string, len(opts))
	for i, v := range opts {
		options[i], ok = v.(string)
		if !ok {
			return []string{}, fmt.Errorf("encountered non-string element while parsing options for prompt %q, make sure that all options are of type string", key)
		}
	}
	return options, nil
}

func validatePromptKey(k string) error {
	if len(k) > 64 {
		return fmt.Errorf("prompt key %q is too long, it must be not longer than 64 characters", k)
	}
	if isDigit(rune(k[0])) {
		return fmt.Errorf("prompt key %q should start with a letter or underscore, but got digit instead", k)
	}
	for _, r := range k {
		if !isLetter(r) && !isDigit(r) && r != '_' {
			return fmt.Errorf("prompt key %q contains invalid character %q", k, r)
		}
	}
	return nil
}

func parsePromptMessage(conf map[string]interface{}, key string) (string, error) {
	s, _ := conf["message"].(string)
	if len(s) == 0 {
		return "", fmt.Errorf("required setting 'message' is missing for prompt %q", key)
	}
	if len(s) > 128 {
		return "", fmt.Errorf("prompt message for %q is too long, it must be not longer than 128 characters", key)
	}
	return s, nil
}

// isLetter checks whether r is an ASCII valid letter ([a-zA-Z]).
func isLetter(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsLetter(r)
}

// isDigit checks whether r is an ASCII valid numeric digit ([0-9]).
func isDigit(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsDigit(r)
}

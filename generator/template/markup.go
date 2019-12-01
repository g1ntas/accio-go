package template

import (
	"errors"
	"github.com/g1ntas/accio/markup"
	"log"
)

const (
	tagFilename = "filename"
	tagSkip     = "skipif"
	tagTemplate = "template"
	tagPartial  = "partial"
	tagVariable = "variable"
	attrName    = "name"
)

var errSkipTag = errors.New("skip tag")

type Markup struct {
	Filename string
	Skip     string
	Body     string
	Vars     vars
	Partials map[string]string
}

// vars holds variables in their original order
type vars [][2]string

func (v *vars) find(key string) int {
	for i, val := range *v {
		if val[0] == key {
			return i
		}
	}
	return -1
}

func newMarkup() *Markup {
	return &Markup{
		Vars:     make(vars, 0, 5),
		Partials: make(map[string]string),
	}
}

func (m *Markup) setPartial(name, val string) {
	if _, ok := m.Partials[name]; ok {
		log.Printf("Partial definition for %q already exists. Overwriting...", name)
	}
	m.Partials[name] = val
}

func (m *Markup) setVar(name, val string) {
	if i := m.Vars.find(name); i != -1 {
		log.Printf("Variable definition for %q already exists. Overwriting...", name)
		m.Vars[i] = [2]string{name, val}
		return
	}
	m.Vars = append(m.Vars, [2]string{name, val})
}

func parse(p *markup.Parser) *Markup {
	m := newMarkup()
	for _, tag := range p.Tags {
		switch tag.Name {
		case tagFilename:
			m.Filename = parseScriptBody(tag)
		case tagSkip:
			m.Skip = parseScriptBody(tag)
		case tagTemplate:
			m.Body = parseBody(tag)
		case tagPartial, tagVariable:
			name, err := parseTagNameAttr(tag)
			if err == errSkipTag {
				continue
			}
			if tag.Body == nil {
				log.Printf("Tag %q with name %q is missing body. Skipping...\n", tag.Name, name)
				continue
			}
			switch tag.Name {
			case tagPartial:
				m.setPartial(name, parseBody(tag))
			case tagVariable:
				m.setVar(name, parseScriptBody(tag))
			}
		}
	}
	return m
}

func parseTagNameAttr(tag *markup.TagNode) (string, error) {
	var name string
	for _, attr := range tag.Attributes {
		if attr.Name == attrName {
			name = attr.Value
		}
	}
	if len(name) == 0 {
		log.Printf("Tag %q is missing %q attribute. Skipping...", tag.Name, attrName)
		return "", errSkipTag
	}
	return name, nil
}

func parseBody(tag *markup.TagNode) string {
	if tag.Body == nil {
		return ""
	}
	return tag.Body.Content
}

func parseScriptBody(tag *markup.TagNode) string {
	if tag.Body == nil {
		return ""
	}
	if tag.Body.Inline {
		return wrapInlineScript(tag.Body.Content)
	}
	return tag.Body.Content
}

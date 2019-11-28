package template // todo: rename to Model (ModelParser)

import (
	"github.com/cbroglie/mustache"
	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/markup"
)

// todo: rename to parser
type Parser struct {
	data   map[string]interface{}
	markup *Markup
	tpl    generator.Template
}

func NewParser(d map[string]interface{}) *Parser {
	return &Parser{data: d}
}

func (p *Parser) Parse(b []byte) (*generator.Template, error) {
	mp, err := markup.Parse(string(b), "", "")
	if err != nil {
		return nil, err
	}
	p.markup = parse(mp)

	// todo: compile scripts
	data := p.copyData()
	for k, val := range p.markup.Vars {
		s, err := execute(val, p.data)
		if err != nil {
			return nil, err
		}
		data[k] = s
	}
	// todo: merge variables with data
	// todo: render templates and partials
	err = p.renderTemplate(data)
	return &p.tpl, nil
}

func (p *Parser) renderTemplate(data map[string]interface{}) error {
	provider := &mustache.StaticProvider{p.markup.Partials}
	content, err := mustache.RenderPartials(p.markup.Body, provider, data)
	if err != nil {
		return err
	}
	p.tpl.Body = content
	return nil
}

func (p *Parser) copyData() map[string]interface{} {
	d := make(map[string]interface{})
	for k, v := range p.data {
		d[k] = v
	}
	return d
}

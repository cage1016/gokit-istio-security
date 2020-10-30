package postgres

import (
	"bytes"
	"fmt"
	"text/template"
)

var engine Engine

type Engine interface {
	init()
	Execute(name string, model interface{}) (string, error)
	ExecuteString(data string, model interface{}) (string, error)
	AssetDir(name string) ([]string, error)
}

type DefaultEngine struct {
	t *template.Template
}

func NewEngine() Engine {
	if engine == nil {
		engine = &DefaultEngine{}
		engine.init()
	}
	return engine
}

func (e *DefaultEngine) init() {
	e.t = template.New("default")
}

func (e *DefaultEngine) Execute(name string, model interface{}) (string, error) {
	d, err := Asset(fmt.Sprintf("sqls/%s.tmpl", name))
	if err != nil {
		return "", err
	}
	tmp, err := e.t.Parse(string(d))
	if err != nil {
		return "", err
	}
	ret := bytes.NewBufferString("")
	err = tmp.Execute(ret, model)
	return ret.String(), err
}

func (e *DefaultEngine) ExecuteString(data string, model interface{}) (string, error) {
	tmp, err := e.t.Parse(data)
	if err != nil {
		return "", err
	}
	ret := bytes.NewBufferString("")
	err = tmp.Execute(ret, model)
	return ret.String(), err
}

func (e *DefaultEngine) AssetDir(name string) ([]string, error) {
	return AssetDir(name)
}

package templates

import (
	"strings"
	"text/template"
)

func NewRoot() (*template.Template, error) {
	t := template.New("gensta")
	t = t.Funcs(template.FuncMap{
		"argNames":        argNames,
		"args":            argsSting,
		"funcArgs":        renderArgs,
		"join":            strings.Join,
		"appendFormatter": appendFormatter,
		"lower1":          lower1,
		"receiver":        receiver,
		"dict":            dict,
	})
	var err error
	t, err = t.Parse(NamedArgs)
	if err != nil {
		return nil, err
	}
	t, err = t.Parse(Method)
	if err != nil {
		return nil, err
	}
	return t, nil
}

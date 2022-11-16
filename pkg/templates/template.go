package templates

import "text/template"

type (
	Dot        any
	DataGetter func() Dot
)

func NewCRUD() (*template.Template, error) {
	tmlp, err := template.New("crud").Funcs(FuncMap).Parse(CRUDTemplate)
	if err != nil {
		return nil, err
	}
	return tmlp, nil
}

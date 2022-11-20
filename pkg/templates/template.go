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

func NewRepo() (*template.Template, error) {
	tmlp, err := template.New("repo").Funcs(FuncMap).Parse(CRUDRepoTemplate)
	if err != nil {
		return nil, err
	}
	return tmlp, nil
}

func NewGeneralRepo() (*template.Template, error) {
	tmlp, err := template.New("general_repo").Funcs(FuncMap).Parse(GeneralRepo)
	if err != nil {
		return nil, err
	}
	return tmlp, nil
}

package templates

var GeneralRepo = `package {{ .PackageName }}

import (
	"gorm.io/gorm"
)

// Repo is general interface for your database.
// You can add new methods here, 
// file will not be regenerated if exists
type Repo interface {
	{{ range .Repos }}
	{{- .Method }}() {{ .Package }}.{{ .Type }}
	{{ end }}
}

type repo struct {
	db            *gorm.DB
	{{ range .Repos }}
	{{- .Package }} {{ .Package }}.{{ .Type }}
	{{ end }}
}

func New(db *gorm.DB) Repo {
	return repo{
		db:            db,
		{{ range .Repos }}
		{{- .Package }}:       {{ .Package }}.New(db),
		{{ end }}
	}
}

{{ range .Repos }}
func (r repo) {{ .Method }}() {{ .Package }}.{{ .Type }} {
	return r.{{ .Package }}
}
{{ end }}`

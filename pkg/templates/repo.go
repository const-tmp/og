package templates

var CRUDRepoTemplate = `package {{ .Package }}

import (
	"context"
	"gorm.io/gorm"
)

type (
	// Repo is an interface generated for your usage.
	// You can add new methods to it, it will not be regenerated if exists.
	Repo interface {
		CRUD
	}

	// repo is an implementation.
	repo struct {
		db   *gorm.DB
		crud CRUD
	}
)

func New(db *gorm.DB, omit ...string) Repo {
	return repo{db, NewCRUD(db, omit...)}
}

func (r repo) Create(ctx context.Context, v {{ .Type }}, omit ...string) (*{{ .Type }}, error) {
	return r.crud.Create(ctx, v, omit...)
}

func (r repo) GetOrCreate(ctx context.Context, v {{ .Type }}, omit ...string) (*{{ .Type }}, error) {
	return r.crud.GetOrCreate(ctx, v, omit...)
}

func (r repo) GetByID(ctx context.Context, v {{ .Type }}) (*{{ .Type }}, error) {
	return r.crud.GetByID(ctx, v)
}

func (r repo) Query(ctx context.Context, v {{ .Type }}, omit ...string) ([]*{{ .Type }}, error) {
	return r.crud.Query(ctx, v, omit...)
}

func (r repo) QueryOne(ctx context.Context, v {{ .Type }}, omit ...string) (*{{ .Type }}, error) {
	return r.crud.QueryOne(ctx, v, omit...)
}

func (r repo) UpdateField(ctx context.Context, v {{ .Type }}, column string, value any) error {
	return r.crud.UpdateField(ctx, v, column, value)
}

func (r repo) Update(ctx context.Context, v {{ .Type }}, omit ...string) (err error) {
	return r.crud.Update(ctx, v, omit...)
}

func (r repo) UpdateMap(ctx context.Context, v {{ .Type }}, m map[string]any) error {
	return r.crud.UpdateMap(ctx, v, m)
}

func (r repo) Delete(ctx context.Context, v {{ .Type }}) error {
	return r.crud.Delete(ctx, v)
}`

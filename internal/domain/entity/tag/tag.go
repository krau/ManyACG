package tag

import (
	"slices"

	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Tag struct {
	ID    objectuuid.ObjectUUID
	Name  string
	Alias []string
}

func NewTag(id objectuuid.ObjectUUID, name string, alias []string) *Tag {
	return &Tag{
		ID:    id,
		Name:  name,
		Alias: alias,
	}
}

func (t *Tag) AddAlias(alias string) {
	if slices.Contains(t.Alias, alias) {
		return
	}
	t.Alias = append(t.Alias, alias)
}

func (t *Tag) RemoveAlias(alias string) {
	t.Alias = slices.DeleteFunc(t.Alias, func(a string) bool {
		return a == alias
	})
}

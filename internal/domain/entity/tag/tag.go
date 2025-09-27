package tag

import (
	"slices"

	"github.com/krau/ManyACG/pkg/objectuuid"
)

type TagID = objectuuid.ObjectUUID

type Tag struct {
	ID    TagID
	Name  string
	Alias []string
}

func NewTag(id TagID, name string, alias []string) *Tag {
	return &Tag{
		ID:    id,
		Name:  name,
		Alias: alias,
	}
}

func (t *Tag) AddAlias(alias string) {
	for _, a := range t.Alias {
		if a == alias {
			return
		}
	}
	t.Alias = append(t.Alias, alias)
}

func (t *Tag) RemoveAlias(alias string) {
	t.Alias = slices.DeleteFunc(t.Alias, func(a string) bool {
		return a == alias
	})
}

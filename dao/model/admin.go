package model

import "ManyACG-Bot/types"

type AdminModel struct {
	UserID      int64              `bson:"user_id"`
	Permissions []types.Permission `bson:"permissions"`
	GrantBy     int64              `bson:"grant_by"`
	SuperAdmin  bool               `bson:"super_admin"`
}

func (a *AdminModel) HasPermission(p types.Permission) bool {
	for _, permission := range a.Permissions {
		if permission == p {
			return true
		}
	}
	return false
}

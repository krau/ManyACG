package types

type AdminModel struct {
	UserID      int64        `bson:"user_id" json:"user_id"`
	Permissions []Permission `bson:"permissions" json:"permissions"`
	GrantBy     int64        `bson:"grant_by" json:"grant_by"`
	SuperAdmin  bool         `bson:"super_admin" json:"super_admin"`
}

func (a *AdminModel) HasPermission(p Permission) bool {
	for _, permission := range a.Permissions {
		if permission == p {
			return true
		}
	}
	return false
}

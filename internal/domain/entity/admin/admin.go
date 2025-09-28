package admin

import (
	"slices"
	"sync"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Admin struct {
	ID          objectuuid.ObjectUUID
	TelegramID  int64
	permissions []shared.Permission
	mu          sync.RWMutex
}

func (a *Admin) HasPermission(permission shared.Permission) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if slices.Contains(a.permissions, shared.PermissionSudo) {
		return true
	}
	return slices.Contains(a.permissions, permission)
}

func NewAdmin(id objectuuid.ObjectUUID, telegramID int64, permissions []shared.Permission) *Admin {
	return &Admin{
		ID:          id,
		TelegramID:  telegramID,
		permissions: permissions,
	}
}

func (a *Admin) Promote(permission shared.Permission) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.HasPermission(permission) {
		a.permissions = append(a.permissions, permission)
	}
}

func (a *Admin) Demote(permission shared.Permission) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.HasPermission(permission) {
		a.permissions = slices.DeleteFunc(a.permissions, func(p shared.Permission) bool {
			return p == permission
		})
	}
}

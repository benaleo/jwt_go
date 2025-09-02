package graph

import (
	"jwt_go/graph/model"
)

// MARK : Mapper User with Role And Role Permission
func toGraphQLUser(u model.UserDB) *model.User {
	user := model.User{
		ID:        int32(u.ID),
		Name:      u.Name,
		Username:  u.Username,
		RoleID:    int32(u.RoleID),
		Role:      toGraphQLRole(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	// Optional string fields -> pointers if non-empty
	if u.Email != "" {
		user.Email = &u.Email
	}
	if u.Address != "" {
		user.Address = &u.Address
	}
	if u.Phone != "" {
		user.Phone = &u.Phone
	}

	// Set deleted_at if it exists
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		user.DeletedAt = &t
	}

	// Set avatar if it exists
	if u.Avatar != "" {
		user.Avatar = &u.Avatar
	}

	return &user
}

// MARK : Mapper Role
func toGraphQLRole(r model.RoleDB) *model.Role {
	// Basic role mapper with empty permissions slice
	role := model.Role{
		ID:        int32(r.ID),
		Name:      r.Name,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Permissions: []*model.Permission{},
	}

	// Set deleted_at if it exists
	if r.DeletedAt.Valid {
		t := r.DeletedAt.Time
		role.DeletedAt = &t
	}

	return &role
}

// toGraphQLRoleWithPermissions maps a role and its permissions
func toGraphQLRoleWithPermissions(r model.RoleDB, perms []model.PermissionDB) *model.Role {
	role := toGraphQLRole(r)
	if len(perms) == 0 {
		role.Permissions = []*model.Permission{}
		return role
	}
	list := make([]*model.Permission, 0, len(perms))
	for _, p := range perms {
		list = append(list, toGraphQLPermission(p))
	}
	role.Permissions = list
	return role
}

// MARK : Mapper Permission
func toGraphQLPermission(p model.PermissionDB) *model.Permission {
	permission := model.Permission{
		ID:   int32(p.ID),
		Name: p.Name,
	}

	return &permission
}

// MARK : Mapper Role Permission
func toGraphQLRolePermission(rp model.RolePermissionDB) *model.RolePermission {
	rolePermission := model.RolePermission{
		RoleID:       int32(rp.RoleID),
		PermissionID: int32(rp.PermissionID),
	}

	return &rolePermission
}

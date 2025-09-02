package config

import (
	"fmt"
	"strings"

	"jwt_go/graph/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed seeds initial permissions, SUPERADMIN role, and an admin user.
// Idempotent: safe to run multiple times.
func Seed(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	// 1) Ensure permissions exist
	permNames := []string{
		"user.view",
		"user.create",
		"user.edit",
		"user.delete",
		"role.view",
		"role.create",
		"role.edit",
		"role.delete",
	}

	var permissions []model.PermissionDB
	for _, name := range permNames {
		var p model.PermissionDB
		err := db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&p).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				p = model.PermissionDB{Name: name}
				if err := db.Create(&p).Error; err != nil {
					return fmt.Errorf("seed: create permission %s: %w", name, err)
				}
			} else {
				return fmt.Errorf("seed: find permission %s: %w", name, err)
			}
		}
		permissions = append(permissions, p)
	}

	// 2) Ensure SUPERADMIN role exists
	var role model.RoleDB
	if err := db.Where("LOWER(name) = ?", "superadmin").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			role = model.RoleDB{Name: "SUPERADMIN", IsActive: true}
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("seed: create role SUPERADMIN: %w", err)
			}
		} else {
			return fmt.Errorf("seed: find role SUPERADMIN: %w", err)
		}
	}

	// 2b) Ensure USER role exists with no permissions
	var userRole model.RoleDB
	if err := db.Where("LOWER(name) = ?", "user").First(&userRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			userRole = model.RoleDB{Name: "USER", IsActive: true}
			if err := db.Create(&userRole).Error; err != nil {
				return fmt.Errorf("seed: create role USER: %w", err)
			}
		} else {
			return fmt.Errorf("seed: find role USER: %w", err)
		}
	}
	// Ensure USER role has no permissions (idempotent)
	if err := db.Where("role_id = ?", userRole.ID).Delete(&model.RolePermissionDB{}).Error; err != nil {
		return fmt.Errorf("seed: clear USER role permissions: %w", err)
	}

	// 3) Ensure role_permissions include ALL permissions for SUPERADMIN
	// Collect existing permission IDs for this role
	var existingRPs []model.RolePermissionDB
	if err := db.Where("role_id = ?", role.ID).Find(&existingRPs).Error; err != nil {
		return fmt.Errorf("seed: find role_permissions: %w", err)
	}
	existing := map[int]struct{}{}
	for _, rp := range existingRPs {
		existing[rp.PermissionID] = struct{}{}
	}

	var toInsert []model.RolePermissionDB
	for _, p := range permissions {
		if _, ok := existing[p.ID]; !ok {
			toInsert = append(toInsert, model.RolePermissionDB{RoleID: role.ID, PermissionID: p.ID})
		}
	}
	if len(toInsert) > 0 {
		if err := db.Create(&toInsert).Error; err != nil {
			return fmt.Errorf("seed: create role_permissions: %w", err)
		}
	}

	// 4) Ensure admin user exists with SUPERADMIN role
	var admin model.UserDB
	if err := db.Where("LOWER(username) = ?", "admin").First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hash, herr := bcrypt.GenerateFromPassword([]byte("rahasia123"), bcrypt.DefaultCost)
			if herr != nil {
				return fmt.Errorf("seed: hash admin password: %w", herr)
			}
			admin = model.UserDB{
				Name:     "SUPERADMIN",
				Username: "admin",
				Email:    "",
				Password: string(hash),
				RoleID:   role.ID,
				IsActive: true,
			}
			if err := db.Create(&admin).Error; err != nil {
				return fmt.Errorf("seed: create admin user: %w", err)
			}
		} else {
			return fmt.Errorf("seed: find admin user: %w", err)
		}
	} else {
		// Ensure admin has SUPERADMIN role if somehow missing
		if admin.RoleID != role.ID {
			admin.RoleID = role.ID
			if err := db.Save(&admin).Error; err != nil {
				return fmt.Errorf("seed: update admin role: %w", err)
			}
		}
	}

	return nil
}

package model

import (
	"time"

	"gorm.io/gorm"
)

// MARK : INTERFACE
type AuditableDB interface {
	IsAuditableDB()
	GetIsActiveDB() bool
	GetDeletedAtDB() *time.Time
	GetCreatedAtDB() time.Time
	GetUpdatedAtDB() time.Time
}

// MARK : USER
type UserDB struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"-" gorm:"not null"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	RoleID   int    `json:"role_id"`
	Role     RoleDB `gorm:"foreignKey:RoleID;references:ID"`

	IsActive  bool           `json:"is_active"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserDB) TableName() string {
	return "users"
}

// MARK : ROLE
type RoleDB struct {
	ID        int            `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	IsActive  bool           `json:"is_active"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (RoleDB) TableName() string {
	return "roles"
}

// MARK : PERMISSION
type PermissionDB struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

func (PermissionDB) TableName() string {
	return "permissions"
}

// MARK : ROLE PERMISSION
type RolePermissionDB struct {
	RoleID       int `gorm:"primaryKey" json:"role_id"`
	PermissionID int `gorm:"primaryKey" json:"permission_id"`
}

func (RolePermissionDB) TableName() string {
	return "role_permissions"
}

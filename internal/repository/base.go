package repository

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseRepository struct {
}

type BaseUUIDModel struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *BaseUUIDModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}

type MapUpdate = map[string]interface{}

func getModelID(model interface{}) string {
	val := reflect.ValueOf(model)

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Look for fields with `gorm:"primaryKey"` tag first
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		gormTag := field.Tag.Get("gorm")
		if strings.Contains(gormTag, "primaryKey") {
			fieldValue := val.Field(i)
			return fmt.Sprintf("%v", fieldValue.Interface())
		}
	}

	// Fallback to "ID" field if no explicit primary key
	idField := val.FieldByName("ID")
	if idField.IsValid() {
		return fmt.Sprintf("%v", idField.Interface())
	}

	return ""
}

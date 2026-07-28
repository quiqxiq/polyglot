package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB is a custom GORM type that stores a map[string]string as a PostgreSQL
// JSONB value. It implements driver.Valuer and sql.Scanner so GORM can use it
// with jsonb columns without pulling in gorm.io/datatypes.
type JSONB map[string]string

// Value marshals the map to JSON bytes for storage.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

// Scan unmarshals JSON bytes (or a string) from the database into the map.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("jsonb: unsupported scan type %T", value)
	}

	if len(bytes) == 0 {
		*j = JSONB{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

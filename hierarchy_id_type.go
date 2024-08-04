package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// HierarchyId is a structure to represent database hierarchy ids.
type HierarchyId struct {
	// Path of the hierarchy (e.g "/1/2/3/4/")
	Data HierarchyIdData
}

// GormDataTypeInterface to specify the nema of data type.
func (HierarchyId) GormDataType() string {
	return "hierarchyid"
}

// GormDBDataTypeInterface defines the data type to apply in the database.
func (HierarchyId) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	if db.Dialector.Name() != "sqlserver" {
		panic("hierarchyid is only supported on SQL Server")
	}

	return "hierarchyid"
}

// Get all parents of a hierarchyid.
func (j *HierarchyId) GetParents() []HierarchyId {
	p := []HierarchyId{}
	pd := GetParents(j.Data)

	for _, d := range pd {
		p = append(p, HierarchyId{Data: d})
	}

	return p
}

// When marshaling to JSON, we want the field formatted as a string.
func (j HierarchyId) MarshalJSON() ([]byte, error) {
	return json.Marshal(ToString(j.Data))
}

// When unmarshaling from JSON, we want to parse the string into the field.
func (j *HierarchyId) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	str := ""

	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	j.Data, err = FromString(str)
	if err != nil {
		return err
	}

	return nil
}

// Value implements the driver.Valuer interface.
//
// Used to provide a value to the SQL server for storage.
func (j HierarchyId) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}

	data, err := Encode(j.Data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Scan implements the sql.Scanner interface.
//
// Used to read the value provided by the SQL server.
func (j *HierarchyId) Scan(src any) error {
	if src == nil {
		j.Data = nil
		return nil
	}

	switch src := src.(type) {
	case []byte:
		var err error
		j.Data, err = Decode(src)
		if err != nil {
			return err
		}
	default:
		return errors.New("incompatible type to scan")
	}

	return nil
}

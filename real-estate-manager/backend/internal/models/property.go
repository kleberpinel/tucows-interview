package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// NullString wraps sql.NullString with proper JSON marshaling
type NullString struct {
	sql.NullString
}

// MarshalJSON implements json.Marshaler interface
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

// NullInt32 wraps sql.NullInt32 with proper JSON marshaling
type NullInt32 struct {
	sql.NullInt32
}

// MarshalJSON implements json.Marshaler interface
func (ni NullInt32) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int32)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ni *NullInt32) UnmarshalJSON(data []byte) error {
	var i *int32
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Valid = true
		ni.Int32 = *i
	} else {
		ni.Valid = false
	}
	return nil
}

// FlexibleString can unmarshal both string and number JSON values as strings
type FlexibleString string

// UnmarshalJSON implements json.Unmarshaler interface for FlexibleString
func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexibleString(s)
		return nil
	}
	
	// If that fails, try as number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*fs = FlexibleString(n.String())
		return nil
	}
	
	return errors.New("cannot unmarshal into FlexibleString")
}

// String returns the string value
func (fs FlexibleString) String() string {
	return string(fs)
}

type Property struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Location    string     `json:"location" db:"location"`
	Price       float64    `json:"price" db:"price"`
	Description NullString `json:"description" db:"description"`
	Photos      PhotoList  `json:"photos" db:"photos"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	
	// SimplyRETS specific fields
	ExternalID    NullString `json:"external_id,omitempty" db:"external_id"`
	MLSNumber     NullString `json:"mls_number,omitempty" db:"mls_number"`
	PropertyType  NullString `json:"property_type,omitempty" db:"property_type"`
	Bedrooms      NullInt32  `json:"bedrooms,omitempty" db:"bedrooms"`
	Bathrooms     NullInt32  `json:"bathrooms,omitempty" db:"bathrooms"`
	SquareFeet    NullInt32  `json:"square_feet,omitempty" db:"square_feet"`
	LotSize       NullString `json:"lot_size,omitempty" db:"lot_size"`
	YearBuilt     NullInt32  `json:"year_built,omitempty" db:"year_built"`
}

// Photo represents a property photo
type Photo struct {
	URL      string `json:"url"`
	LocalURL string `json:"local_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// PhotoList is a slice of photos that implements SQL driver interfaces
type PhotoList []Photo

// Value implements the driver.Valuer interface for database storage
func (p PhotoList) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for database retrieval
func (p *PhotoList) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into PhotoList")
	}
	
	return json.Unmarshal(bytes, p)
}

// SimplyRETS API Response structures
type SimplyRETSProperty struct {
	ListingID    string                     `json:"listingId"`
	MLSNumber    FlexibleString             `json:"mlsId"`
	Address      SimplyRETSAddress          `json:"address"`
	ListPrice    float64                    `json:"listPrice"`
	Property     SimplyRETSPropertyDetails  `json:"property"`
	Photos       []string                   `json:"photos"`
	Remarks      string                     `json:"remarks"`
}

type SimplyRETSAddress struct {
	Full         string         `json:"full"`
	Unit         string         `json:"unit"`
	StreetNumber FlexibleString `json:"streetNumber"`
	StreetName   string         `json:"streetName"`
	City         string         `json:"city"`
	State        string         `json:"state"`
	PostalCode   string         `json:"postalCode"`
}

type SimplyRETSPropertyDetails struct {
	PropertyType string `json:"type"`
	Style        string `json:"style"`
	YearBuilt    int    `json:"yearBuilt"`
	Stories      int    `json:"stories"`
	Area         int    `json:"area"`
	LotSize      string `json:"lotSize"`
	Bedrooms     int    `json:"bedrooms"`
	Bathrooms    int    `json:"bathrooms"`
}

// ProcessingStatus represents the status of property processing
type ProcessingStatus struct {
	ID              int       `json:"id"`
	Status          string    `json:"status"` // "running", "completed", "failed"
	TotalProperties int       `json:"total_properties"`
	ProcessedCount  int       `json:"processed_count"`
	FailedCount     int       `json:"failed_count"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}
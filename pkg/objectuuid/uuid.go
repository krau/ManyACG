package objectuuid

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ObjectUUID will transform MongoDB's ObjectID (12 bytes) into UUID format (16 bytes).
// The last 4 bytes are always zero.
type ObjectUUID uuid.UUID

var (
	Nil ObjectUUID
)

func New() ObjectUUID {
	oid := newObjectID()
	var cu ObjectUUID
	copy(cu[:12], oid[:])
	return cu
}

// FromObjectID converts MongoDB's ObjectID to ObjectUUID
func FromObjectID(oid ObjectID) ObjectUUID {
	var cu ObjectUUID
	copy(cu[:12], oid[:])
	return cu
}

func FromObjectIDHex(oidHex string) (ObjectUUID, error) {
	oid, err := ObjectIDFromHex(oidHex)
	if err != nil {
		return Nil, err
	}
	return FromObjectID(oid), nil
}

func (cu ObjectUUID) IsZero() bool {
	return cu == Nil
}

// ToObjectID converts ObjectUUID back to MongoDB's ObjectID
func (cu ObjectUUID) ToObjectID() ObjectID {
	var oid ObjectID
	copy(oid[:], cu[:12])
	return oid
}

// Hex returns objectID hex string
func (cu ObjectUUID) Hex() string {
	u, err := uuid.FromBytes(cu[:])
	if err != nil {
		return ""
	}
	return hex.EncodeToString(u[:12])
}

// Scan implements sql.Scanner (for reading from DB)
func (c *ObjectUUID) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		parsed, err := uuid.FromBytes(v)
		if err != nil {
			return err
		}
		*c = ObjectUUID(parsed)
		return nil
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*c = ObjectUUID(parsed)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into ObjectUUID", value)
	}
}

// Value implements driver.Valuer (for writing into DB)
func (c ObjectUUID) Value() (driver.Value, error) {
	return uuid.UUID(c).String(), nil
}

func (c ObjectUUID) String() string {
	return c.Hex()
}

// MarshalText returns the ObjectUUID as text (hex of ObjectID part).
func (cu ObjectUUID) MarshalText() ([]byte, error) {
	return []byte(cu.Hex()), nil
}

// UnmarshalText parses ObjectUUID from hex text.
func (cu *ObjectUUID) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*cu = Nil
		return nil
	}
	oid, err := ObjectIDFromHex(string(b))
	if err != nil {
		return err
	}
	*cu = FromObjectID(oid)
	return nil
}

// MarshalJSON outputs ObjectUUID as a JSON string (ObjectID hex).
func (cu ObjectUUID) MarshalJSON() ([]byte, error) {
	return MarshalJSON(cu.Hex())
}

// UnmarshalJSON parses JSON into ObjectUUID.
// Supports:
//   - null → leave unchanged
//   - "507f1f77bcf86cd799439011" (24 hex chars)
//   - {"$oid": "..."} (Mongo Extended JSON)
//   - "" → Nil
func (cu *ObjectUUID) UnmarshalJSON(b []byte) error {
	// null -> keep unchanged (same as official ObjectID)
	if string(b) == "null" {
		return nil
	}

	var err error
	var str string

	if err = UnmarshalJSON(b, &str); err == nil {
		if len(str) == 0 {
			*cu = Nil
			return nil
		}
		if len(str) != 24 {
			return fmt.Errorf("cannot unmarshal into ObjectUUID, length must be 24 but is %d", len(str))
		}
		oid, err := ObjectIDFromHex(str)
		if err != nil {
			return err
		}
		*cu = FromObjectID(oid)
		return nil
	}

	var m map[string]any
	if err = UnmarshalJSON(b, &m); err != nil {
		return err
	}
	oidVal, ok := m["$oid"]
	if !ok {
		return errors.New("not an extended JSON ObjectID")
	}
	oidStr, ok := oidVal.(string)
	if !ok {
		return errors.New("not an extended JSON ObjectID")
	}
	if oidStr == "" {
		*cu = Nil
		return nil
	}
	oid, err := ObjectIDFromHex(oidStr)
	if err != nil {
		return err
	}
	*cu = FromObjectID(oid)
	return nil
}

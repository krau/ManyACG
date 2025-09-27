package objectuuid

import (
	"database/sql/driver"
	"encoding/hex"
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

// ToObjectID converts ObjectUUID back to MongoDB's ObjectID
func (cu ObjectUUID) ToObjectID() ObjectID {
	var oid ObjectID
	copy(oid[:], cu[:12])
	return oid
}

// String returns objectID hex string
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

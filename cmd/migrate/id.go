package migrate

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoUUID will transform MongoDB's ObjectID (12 bytes) into UUID format (16 bytes).
// The last 4 bytes are always zero.
type MongoUUID uuid.UUID

func NewMongoUUID() MongoUUID {
	oid := primitive.NewObjectID()
	var cu MongoUUID
	copy(cu[:12], oid[:])
	return cu
}

// FromObjectID converts MongoDB's ObjectID to MongoUUID
func FromObjectID(oid primitive.ObjectID) MongoUUID {
	var cu MongoUUID
	copy(cu[:12], oid[:])
	return cu
}

// ToObjectID converts MongoUUID back to MongoDB's ObjectID
func (cu MongoUUID) ToObjectID() primitive.ObjectID {
	var oid primitive.ObjectID
	copy(oid[:], cu[:12])
	return oid
}

// String returns objectID hex string
func (cu MongoUUID) Hex() string {
	u, err := uuid.FromBytes(cu[:])
	if err != nil {
		return ""
	}
	return hex.EncodeToString(u[:12])
}

// Scan implements sql.Scanner (for reading from DB)
func (c *MongoUUID) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		parsed, err := uuid.FromBytes(v)
		if err != nil {
			return err
		}
		*c = MongoUUID(parsed)
		return nil
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*c = MongoUUID(parsed)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into MongoUUID", value)
	}
}

// Value implements driver.Valuer (for writing into DB)
func (c MongoUUID) Value() (driver.Value, error) {
	return uuid.UUID(c).String(), nil
}

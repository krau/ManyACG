package objectuuid

import "testing"

func TestObjectUUID(t *testing.T) {
	oid := newObjectID()
	cu := FromObjectID(oid)
	if cu.ToObjectID() != oid {
		t.Errorf("ToObjectID() = %v; want %v", cu.ToObjectID(), oid)
	}

	hex := cu.Hex()
	if hex != oid.Hex() {
		t.Errorf("Hex() = %s; want %s", hex, oid.Hex())
	}

	newCu := New()
	if newCu.ToObjectID() == oid {
		t.Errorf("New() = %v; want different from %v", newCu, oid)
	}
}

func TestObjectUUIDs(t *testing.T) {
	oid1 := FromObjectID(newObjectID())
	oid2 := FromObjectID(newObjectID())
	oid3 := FromObjectID(newObjectID())

	var list ObjectUUIDs
	if !list.IsEmpty() {
		t.Errorf("IsEmpty() = false; want true")
	}
	if list.Len() != 0 {
		t.Errorf("Len() = %d; want 0", list.Len())
	}

	list.Add(oid1, oid2)
	if list.IsEmpty() {
		t.Errorf("IsEmpty() = true; want false")
	}
	if list.Len() != 2 {
		t.Errorf("Len() = %d; want 2", list.Len())
	}

	list.Add(oid1) // adding duplicate
	if list.Len() != 2 {
		t.Errorf("Len() after adding duplicate = %d; want 2", list.Len())
	}

	list.Remove(oid1)
	if list.Len() != 1 {
		t.Errorf("Len() after removing = %d; want 1", list.Len())
	}
	list.Remove(oid3) // removing non-existing
	if list.Len() != 1 {
		t.Errorf("Len() after removing non-existing = %d; want 1", list.Len())
	}

	list.Remove(oid2)
	if !list.IsEmpty() {
		t.Errorf("IsEmpty() after removing all = false; want true")
	}
	if list.Len() != 0 {
		t.Errorf("Len() after removing all = %d; want 0", list.Len())
	}
}

func TestObjectUUID_ScanAndValue(t *testing.T) {
	oid := FromObjectID(newObjectID())

	val, err := oid.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	var scanned ObjectUUID
	err = scanned.Scan(val)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if scanned != oid {
		t.Errorf("Scanned value = %v; want %v", scanned, oid)
	}
}

func TestObjectUUID_ScanInvalid(t *testing.T) {
	var cu ObjectUUID
	err := cu.Scan(12345) // invalid type
	if err == nil {
		t.Fatalf("Scan() with invalid type did not return error")
	}
}

func BenchmarkNewObjectUUID(b *testing.B) {
	for b.Loop() {
		_ = New()
	}
}

func BenchmarkObjectUUID_Hex(b *testing.B) {
	oid := FromObjectID(newObjectID())
	b.ResetTimer()
	for b.Loop() {
		_ = oid.Hex()
	}
}

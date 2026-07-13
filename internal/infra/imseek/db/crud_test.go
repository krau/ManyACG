package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestDeleteImage(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)

	// Add 3 images
	for i := range 3 {
		tx, _ := d.Begin(ctx)
		id, _ := d.AddImage(ctx, tx, []byte{byte(i + 1), 0, 0, 0}, "/img", "", "", "")
		d.AddVector(ctx, tx, id, make([]byte, 32))
		d.AddVectorStats(ctx, tx, id, 1)
		tx.Commit()
	}

	// Delete image 2
	if _, err := d.DeleteImage(ctx, 2); err != nil {
		t.Fatalf("DeleteImage(2): %v", err)
	}

	// Image 2 should be gone
	var path string
	err := d.sql.QueryRowContext(ctx, "SELECT path FROM image WHERE id = 2").Scan(&path)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}

	// Vector and stats for image 2 should be gone
	var count int
	d.sql.QueryRowContext(ctx, "SELECT COUNT(*) FROM vector WHERE id = 2").Scan(&count)
	if count != 0 {
		t.Fatalf("vector count for deleted image = %d want 0", count)
	}
	d.sql.QueryRowContext(ctx, "SELECT COUNT(*) FROM vector_stats WHERE id = 2").Scan(&count)
	if count != 0 {
		t.Fatalf("vector_stats count for deleted image = %d want 0", count)
	}

	// Images 1 and 3 should still exist
	imgCount, _, err := d.GetCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if imgCount != 2 {
		t.Fatalf("image count after delete = %d want 2", imgCount)
	}

	// Delete non-existent image returns ErrNoRows
	_, err = d.DeleteImage(ctx, 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("DeleteImage(999): expected ErrNoRows, got %v", err)
	}

	// Orphan vector id that belonged to deleted image 2 must not remap to image 3
	_, err = d.GetImageIDByVectorID(ctx, 2)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("orphan vid 2: expected ErrNoRows, got id/err %v", err)
	}
	// Remaining images still map correctly
	imgID, err := d.GetImageIDByVectorID(ctx, 1)
	if err != nil || imgID != 1 {
		t.Fatalf("vid 1 -> %d err=%v want 1", imgID, err)
	}
	imgID, err = d.GetImageIDByVectorID(ctx, 3)
	if err != nil || imgID != 3 {
		t.Fatalf("vid 3 -> %d err=%v want 3", imgID, err)
	}
}

func TestGetImage(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)

	tx, _ := d.Begin(ctx)
	id, _ := d.AddImage(ctx, tx, []byte{0xAB, 0xCD}, "/test.jpg", "", "", "")
	d.AddVector(ctx, tx, id, make([]byte, 32))
	d.AddVectorStats(ctx, tx, id, 5)
	tx.Commit()

	img, err := d.GetImage(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if img.ID != id {
		t.Errorf("ID = %d want %d", img.ID, id)
	}
	if img.Path != "/test.jpg" {
		t.Errorf("Path = %q want /test.jpg", img.Path)
	}

	_, err = d.GetImage(ctx, 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

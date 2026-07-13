package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (d *DB) AddImage(ctx context.Context, tx *sql.Tx, hash []byte, path, artworkID, phash, dhash string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx,
		`INSERT INTO image (hash, path, artwork_id, phash, dhash) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		hash, path, artworkID, phash, dhash).Scan(&id)
	return id, err
}

func (d *DB) CheckImageHash(ctx context.Context, hash []byte) (int64, bool, error) {
	var id int64
	err := d.sql.QueryRowContext(ctx, `SELECT id FROM image WHERE hash = ?`, hash).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (d *DB) FindByExactPerceptualHash(ctx context.Context, phash, dhash, excludePath string) (path string, ok bool, err error) {
	if phash == "" || dhash == "" {
		return "", false, nil
	}
	err = d.sql.QueryRowContext(ctx,
		`SELECT path FROM image WHERE phash = ? AND dhash = ? AND path <> ? LIMIT 1`,
		phash, dhash, excludePath).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

// GetImagePath returns the path of an image (picture id hex in ManyACG).
func (d *DB) GetImagePath(ctx context.Context, id int64) (string, error) {
	var path string
	err := d.sql.QueryRowContext(ctx, `SELECT path FROM image WHERE id = ?`, id).Scan(&path)
	return path, err
}

func (d *DB) AddVector(ctx context.Context, tx *sql.Tx, id int64, vector []byte) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO vector (id, vector) VALUES (?, ?)`, id, vector)
	return err
}

func (d *DB) AddVectorStats(ctx context.Context, tx *sql.Tx, id int64, vectorCount int64) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO vector_stats (id, vector_count, total_vector_count)
		 VALUES (?, ?, COALESCE((SELECT MAX(total_vector_count) FROM vector_stats), 0) + ?)`,
		id, vectorCount, vectorCount)
	return err
}

func (d *DB) GetImageIDByVectorID(ctx context.Context, vectorID int64) (int64, error) {
	var id int64
	err := d.sql.QueryRowContext(ctx,
		`SELECT id FROM vector_stats
		 WHERE total_vector_count - vector_count < ? AND total_vector_count >= ?
		 LIMIT 1`,
		vectorID, vectorID).Scan(&id)
	return id, err
}

type VectorIDRange struct {
	Lo int64 // first vector id (inclusive)
	Hi int64 // last vector id (inclusive); Hi-Lo+1 == vector_count
}

func (d *DB) GetVectorIDRange(ctx context.Context, imageID int64) (VectorIDRange, error) {
	var count, total int64
	err := d.sql.QueryRowContext(ctx,
		`SELECT vector_count, total_vector_count FROM vector_stats WHERE id = ?`,
		imageID).Scan(&count, &total)
	if err != nil {
		return VectorIDRange{}, err
	}
	return VectorIDRange{Lo: total - count + 1, Hi: total}, nil
}

func (d *DB) CountImageUnindexed(ctx context.Context) (int64, error) {
	var n int64
	err := d.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM vector_stats WHERE indexed = 0`).Scan(&n)
	return n, err
}

type VectorRow struct {
	ID               int64
	Vector           []byte
	TotalVectorCount int64
}

func (d *DB) GetVectorsUnindexed(ctx context.Context, limit, offset int64) ([]VectorRow, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT v.id, v.vector, s.total_vector_count
		 FROM vector v JOIN vector_stats s ON v.id = s.id
		 WHERE s.indexed = 0 ORDER BY v.id ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVectorRows(rows)
}

func (d *DB) GetVectors(ctx context.Context, limit, offset int64) ([]VectorRow, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT v.id, v.vector, s.total_vector_count
		 FROM vector v JOIN vector_stats s ON v.id = s.id
		 ORDER BY v.id ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVectorRows(rows)
}

func scanVectorRows(rows *sql.Rows) ([]VectorRow, error) {
	var out []VectorRow
	for rows.Next() {
		var r VectorRow
		if err := rows.Scan(&r.ID, &r.Vector, &r.TotalVectorCount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DB) SetIndexedBatch(ctx context.Context, ids []int64) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `UPDATE vector_stats SET indexed = 1 WHERE id = ?`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) GetCount(ctx context.Context) (int64, int64, error) {
	var images int64
	if err := d.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM image`).Scan(&images); err != nil {
		return 0, 0, err
	}
	var total sql.NullInt64
	if err := d.sql.QueryRowContext(ctx,
		`SELECT MAX(total_vector_count) FROM vector_stats`).Scan(&total); err != nil {
		return 0, 0, err
	}
	if !total.Valid {
		return images, 0, nil
	}
	return images, total.Int64, nil
}

func (d *DB) CountVectors(ctx context.Context) (int64, error) {
	var n int64
	err := d.sql.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(LENGTH(vector)/?), 0) FROM vector`,
		32).Scan(&n)
	return n, err
}

type VectorBound struct {
	ImageID int64
	Total   int64 // inclusive upper bound of this image's vector ids
	Count   int64 // number of descriptors; range is (Total-Count, Total]
}

func (d *DB) GetAllVectorBounds(ctx context.Context) ([]VectorBound, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT id, total_vector_count, vector_count FROM vector_stats ORDER BY total_vector_count ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VectorBound
	for rows.Next() {
		var b VectorBound
		if err := rows.Scan(&b.ImageID, &b.Total, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (d *DB) GuessCodeSize(ctx context.Context) (int, error) {
	var blobLen, vecCount int64
	err := d.sql.QueryRowContext(ctx,
		`SELECT length(v.vector), s.vector_count
		 FROM vector v JOIN vector_stats s ON v.id = s.id
		 WHERE s.vector_count > 0 LIMIT 1`).Scan(&blobLen, &vecCount)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if vecCount == 0 {
		return 0, nil
	}
	return int(blobLen / vecCount), nil
}

func (d *DB) Begin(ctx context.Context) (*sql.Tx, error) {
	return d.sql.BeginTx(ctx, nil)
}

func (d *DB) DeleteImage(ctx context.Context, id int64) (VectorIDRange, error) {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return VectorIDRange{}, fmt.Errorf("delete image %d: begin tx: %w", id, err)
	}
	defer tx.Rollback()

	var rng VectorIDRange
	var count, total int64
	err = tx.QueryRowContext(ctx,
		`SELECT vector_count, total_vector_count FROM vector_stats WHERE id = ?`, id).
		Scan(&count, &total)
	if err == nil {
		rng = VectorIDRange{Lo: total - count + 1, Hi: total}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return VectorIDRange{}, fmt.Errorf("delete image %d: read stats: %w", id, err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM vector WHERE id = ?`, id); err != nil {
		return VectorIDRange{}, fmt.Errorf("delete image %d: delete vector: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM vector_stats WHERE id = ?`, id); err != nil {
		return VectorIDRange{}, fmt.Errorf("delete image %d: delete stats: %w", id, err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM image WHERE id = ?`, id)
	if err != nil {
		return VectorIDRange{}, fmt.Errorf("delete image %d: delete image: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return VectorIDRange{}, fmt.Errorf("delete image %d: rows affected: %w", id, err)
	}
	if n == 0 {
		return VectorIDRange{}, sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return VectorIDRange{}, err
	}
	return rng, nil
}

type ImageInfo struct {
	ID        int64
	Path      string
	ArtworkID string
}

func (d *DB) GetImage(ctx context.Context, id int64) (*ImageInfo, error) {
	var img ImageInfo
	err := d.sql.QueryRowContext(ctx,
		`SELECT id, path, artwork_id FROM image WHERE id = ?`, id).Scan(&img.ID, &img.Path, &img.ArtworkID)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

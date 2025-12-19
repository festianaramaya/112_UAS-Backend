package repository

import (
	"context"
	"database/sql"
	"uas/app/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AchievementRepository struct {
	DB *sqlx.DB
}

func NewAchievementRepository(db *sqlx.DB) *AchievementRepository {
	return &AchievementRepository{DB: db}
}

/* ================= CREATE ================= */

func (r *AchievementRepository) Create(ref *model.AchievementReference) error {
	query := `
		INSERT INTO achievement_references
		(student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return r.DB.QueryRowx(
		query,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
	).Scan(&ref.ID, &ref.CreatedAt)
}


func (r *AchievementRepository) GetByStudentID(
	ctx context.Context,
	studentID uuid.UUID,
) ([]model.AchievementReference, error) {

	var results []model.AchievementReference

	query := `
		SELECT
			id,
			student_id,
			mongo_achievement_id,
			status,
			created_at,
			updated_at
		FROM achievement_references
		WHERE student_id = $1
		ORDER BY created_at DESC
	`

	err := r.DB.SelectContext(ctx, &results, query, studentID)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *AchievementRepository) GetAll(
	ctx context.Context,
) ([]model.AchievementReference, error) {

	var refs []model.AchievementReference

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
	`

	err := r.DB.SelectContext(ctx, &refs, query)
	return refs, err
}

/* ================= UPDATE ================= */

func (r *AchievementRepository) UpdateStatus(
    ctx context.Context,
    id uuid.UUID,
    status string,
    verifiedBy sql.NullString,
    rejectionNote sql.NullString,
) error {
    // Gunakan transaksi agar kedua update sukses atau gagal bersamaan
    tx, err := r.DB.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Update status di tabel utama
    queryUpdate := `
        UPDATE achievement_references
        SET status = $2, verified_by = $3, rejection_note = $4, updated_at = NOW()
        WHERE id = $1`
    
    if _, err := tx.ExecContext(ctx, queryUpdate, id, status, verifiedBy, rejectionNote); err != nil {
        return err
    }

    // 2. CATAT KE RIWAYAT (Agar tidak null saat di-GET)
    queryHistory := `
        INSERT INTO achievement_status_histories (achievement_id, status, note, updated_at)
        VALUES ($1, $2, $3, NOW())`
    
    note := ""
    if rejectionNote.Valid {
        note = rejectionNote.String
    }

    if _, err := tx.ExecContext(ctx, queryHistory, id, status, note); err != nil {
        return err
    }

    return tx.Commit()
}

func (r *AchievementRepository) UpdateTimestamp(
	ctx context.Context,
	id uuid.UUID,
) error {

	query := `
		UPDATE achievement_references
		SET updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

/* ================= DELETE ================= */

func (r *AchievementRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	query := `
		DELETE FROM achievement_references
		WHERE id = $1
	`

	res, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *AchievementRepository) GetAdviseeAchievements(
	ctx context.Context,
	lecturerID uuid.UUID,
) ([]model.AchievementFull, error) {

	var achievements []model.AchievementFull

	query := `
		SELECT *
		FROM achievements
		WHERE lecturer_id = $1
	`

	err := r.DB.SelectContext(ctx, &achievements, query, lecturerID)
	if err != nil {
		return nil, err
	}

	return achievements, nil
}

func (r *AchievementRepository) GetStatistics(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'verified') AS verified,
			COUNT(*) FILTER (WHERE status = 'rejected') AS rejected,
			COUNT(*) FILTER (WHERE status = 'submitted') AS submitted
		FROM achievement_references
	`

	var stats struct {
		Total     int `db:"total"`
		Verified  int `db:"verified"`
		Rejected  int `db:"rejected"`
		Submitted int `db:"submitted"`
	}

	err := r.DB.GetContext(ctx, &stats, query)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"total":     stats.Total,
		"verified":  stats.Verified,
		"rejected":  stats.Rejected,
		"submitted": stats.Submitted,
	}, nil
}

func (r *AchievementRepository) GetStudentReport(
	ctx context.Context,
	studentID string,
) (map[string]int, error) {

	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'verified') AS verified,
			COUNT(*) FILTER (WHERE status = 'submitted') AS submitted,
			COUNT(*) FILTER (WHERE status = 'rejected') AS rejected
		FROM achievement_references
		WHERE student_id = $1
	`

	var result struct {
		Total     int `db:"total"`
		Verified  int `db:"verified"`
		Submitted int `db:"submitted"`
		Rejected  int `db:"rejected"`
	}

	err := r.DB.GetContext(ctx, &result, query, studentID)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"total":     result.Total,
		"verified":  result.Verified,
		"submitted": result.Submitted,
		"rejected":  result.Rejected,
	}, nil
}

func (r *AchievementRepository) GetStatusHistory(
	ctx context.Context,
	id uuid.UUID,
) ([]model.AchievementHistory, error) {

	query := `
		SELECT
			status,
			note,
			updated_at
		FROM achievement_status_histories
		WHERE achievement_id = $1
		ORDER BY updated_at ASC
	`

	var history []model.AchievementHistory

	err := r.DB.SelectContext(ctx, &history, query, id)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (r *AchievementRepository) GetByLecturerID(
	ctx context.Context,
	lecturerID uuid.UUID,
) ([]model.AchievementReference, error) {

	var results []model.AchievementReference

	// Menggunakan JOIN untuk menghubungkan tabel prestasi dengan tabel mahasiswa
	// agar kita bisa memfilter berdasarkan advisor_id (dosen pembimbing)
	query := `
		SELECT 
			ar.id, 
			ar.student_id, 
			ar.mongo_achievement_id, 
			ar.status, 
			ar.created_at, 
			ar.updated_at
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		WHERE s.advisor_id = $1
		ORDER BY ar.created_at DESC
	`

	err := r.DB.SelectContext(ctx, &results, query, lecturerID)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *AchievementRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AchievementReference, error) {
    var result model.AchievementReference
    query := `SELECT * FROM achievement_references WHERE id = $1`
    
    err := r.DB.GetContext(ctx, &result, query, id)
    if err != nil {
        return nil, err
    }
    return &result, nil
}
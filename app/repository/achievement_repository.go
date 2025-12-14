package repository

import (
	"context"
	"database/sql"
	"uas/app/model"

	"github.com/jmoiron/sqlx"
)

type AchievementRepository struct {
	DB *sqlx.DB
}

func NewAchievementRepository(db *sqlx.DB) *AchievementRepository {
    return &AchievementRepository{DB: db}
}

// CREATE
func (r *AchievementRepository) Create(ctx context.Context, ref *model.AchievementReference) (string, error) {
    // ... (Implementasi INSERT/RETURNING id) ...
    query := `
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, NOW(), NOW()
		) RETURNING id
	`
	var pgID string
	err := r.DB.QueryRowContext(ctx, query, ref.StudentID, ref.MongoAchievementID, ref.Status).Scan(&pgID)
	if err != nil {
		return "", err
	}
	return pgID, nil
}

// GET DETAIL
func (r *AchievementRepository) GetByID(ctx context.Context, id string) (model.AchievementReference, error) {
    // ... (Implementasi SELECT BY ID) ...
    var ref model.AchievementReference
    query := `
        SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
        FROM achievement_references
        WHERE id = $1
    `
    err := r.DB.GetContext(ctx, &ref, query, id)
    return ref, err
}

// GET ALL (METHOD HILANG YANG DIPERLUKAN OLEH SERVICE)
func (r *AchievementRepository) GetAll(ctx context.Context) ([]model.AchievementReference, error) {
    var refs []model.AchievementReference
    query := `
        SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
        FROM achievement_references
    `
    err := r.DB.SelectContext(ctx, &refs, query)
    return refs, err
}

// UPDATE STATUS (SUDAH ADA)
func (r *AchievementRepository) UpdateStatus(ctx context.Context, id string, status string, verifiedBy sql.NullString, rejectionNote sql.NullString) error {
    // ... (Implementasi UPDATE STATUS) ...
    query := `
		UPDATE achievement_references
		SET status = $1, verified_by = $2, rejection_note = $3, updated_at = NOW(),
			verified_at = CASE WHEN $1 IN ('verified', 'rejected') THEN NOW() ELSE verified_at END
		WHERE id = $4
	`
	res, err := r.DB.ExecContext(ctx, query, status, verifiedBy, rejectionNote, id)
	if err != nil { return err }
	if rows, _ := res.RowsAffected(); rows == 0 { return sql.ErrNoRows }
	return nil
}

// UPDATE TIMESTAMP (METHOD HILANG YANG DIPERLUKAN OLEH SERVICE)
func (r *AchievementRepository) UpdateTimestamp(ctx context.Context, id string) error {
    query := `
        UPDATE achievement_references
        SET updated_at = NOW()
        WHERE id = $1
    `
    res, err := r.DB.ExecContext(ctx, query, id)
    if err != nil { return err }
    if rows, _ := res.RowsAffected(); rows == 0 { return sql.ErrNoRows }
    return nil
}

// DELETE (METHOD HILANG YANG DIPERLUKAN OLEH SERVICE)
func (r *AchievementRepository) Delete(ctx context.Context, id string) error {
    query := `
        DELETE FROM achievement_references
        WHERE id = $1
    `
    res, err := r.DB.ExecContext(ctx, query, id)
    if err != nil { return err }
    if rows, _ := res.RowsAffected(); rows == 0 { return sql.ErrNoRows }
    return nil
}
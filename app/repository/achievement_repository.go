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


/* ================= GET ================= */

func (r *AchievementRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.AchievementReference, error) {

	var ref model.AchievementReference

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`

	err := r.DB.GetContext(ctx, &ref, query, id)
	return ref, err
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

	query := `
		UPDATE achievement_references
		SET status = $2,
		    verified_by = $3,
		    rejection_note = $4,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		id,
		status,
		verifiedBy,
		rejectionNote,
	)

	return err
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

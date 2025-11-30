package repository

import (
	"database/sql"
	"uas/app/model"
)

type AchievementRepository struct {
	DB *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{DB: db}
}

// CREATE (status default = "draft")
func (r *AchievementRepository) Create(a *model.AchievementReference) error {
	// FIX: Hapus ID dari list kolom INSERT. Tambahkan RETURNING untuk mengambil ID dan timestamps.
	query := `
	INSERT INTO achievement_references
	(student_id, mongo_achievement_id, status)
	VALUES ($1, $2, $3)
    RETURNING id, created_at, updated_at
	`

    // FIX: Hapus a.ID dari parameter Exec. Gunakan QueryRow untuk menangkap hasil.
	row := r.DB.QueryRow(
		query,
		a.StudentID,
		a.MongoAchievementID,
		a.Status,
	)
    // FIX: Tangkap nilai yang dikembalikan
    return row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

// GET BY ID
func (r *AchievementRepository) GetByID(id string) (*model.AchievementReference, error) {
	query := `
	SELECT id, student_id, mongo_achievement_id, status,
	        submitted_at, verified_at, verified_by, rejection_note,
	        created_at, updated_at
	FROM achievement_references
	WHERE id = $1
	`
    // Note: Nama tabel harusnya 'achievement_references' (plural) sesuai DDL.
    // Query yang Anda berikan menggunakan 'achievement_reference' (singular), diasumsikan 'achievement_references'.

	row := r.DB.QueryRow(query, id)

	var a model.AchievementReference

	err := row.Scan(
		&a.ID,
		&a.StudentID,
		&a.MongoAchievementID,
		&a.Status,
		&a.SubmittedAt,
		&a.VerifiedAt,
		&a.VerifiedBy,
		&a.RejectionNote,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}

// UPDATE STATUS
func (r *AchievementRepository) UpdateStatus(
	id string,
	status string,
	verifiedBy *string,
	rejectionNote *string,
) error {
    // Note: Logika CASE di SQL yang Anda berikan sudah kompleks dan cenderung benar.
    // Diasumsikan *string* yang dikirim dari Go dapat diinterpretasikan dengan benar oleh SQL.

	query := `
	UPDATE achievement_references
	SET status = $1,
		verified_by = $2,
		rejection_note = $3,
		verified_at = CASE WHEN $1='verified' THEN NOW() ELSE NULL END,
		submitted_at = CASE WHEN $1='submitted' THEN NOW() ELSE submitted_at END,
		updated_at = NOW()
	WHERE id = $4
	`

	_, err := r.DB.Exec(query, status, verifiedBy, rejectionNote, id)
	return err
}
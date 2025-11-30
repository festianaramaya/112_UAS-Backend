package repository

import (
	"database/sql"
	"uas/app/model"
)

type LecturerRepository struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{DB: db}
}

// CREATE lecturer
func (r *LecturerRepository) Create(lec *model.Lecturer) error {
	// FIX: Hapus ID dari kolom INSERT. Tambahkan RETURNING.
	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
        RETURNING id, created_at
	`

	row := r.DB.QueryRow(query,
		lec.UserID,
		lec.LecturerID,
		lec.Department,
	)
    // FIX: Tangkap ID dan created_at yang dihasilkan DB
	return row.Scan(&lec.ID, &lec.CreatedAt)
}

func (r *LecturerRepository) GetAll() ([]model.Lecturer, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturers []model.Lecturer
	for rows.Next() {
		var lec model.Lecturer
		if err := rows.Scan(
			&lec.ID,
			&lec.UserID,
			&lec.LecturerID,
			&lec.Department,
			&lec.CreatedAt,
		); err != nil {
			return nil, err
		}
		lecturers = append(lecturers, lec)
	}

	return lecturers, nil
}

// GET lecturer by ID
func (r *LecturerRepository) GetLecturerByID(id string) (*model.Lecturer, error) {
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE id = $1
	`
    // ... (Logika GetLecturerByID sudah benar)
	row := r.DB.QueryRow(query, id)

	var lec model.Lecturer

	err := row.Scan(
		&lec.ID,
		&lec.UserID,
		&lec.LecturerID,
		&lec.Department,
		&lec.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &lec, nil
}
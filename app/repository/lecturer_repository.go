package repository

import (
	"uas/app/model"

	"github.com/jmoiron/sqlx"
)

type LecturerRepository struct {
	DB *sqlx.DB
}

func NewLecturerRepository(db *sqlx.DB) *LecturerRepository {
	return &LecturerRepository{DB: db}
}

// CREATE lecturer
func (r *LecturerRepository) Create(lec *model.Lecturer) error {
	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	row := r.DB.QueryRow(
		query,
		lec.UserID,
		lec.LecturerID,
		lec.Department,
	)

	return row.Scan(&lec.ID, &lec.CreatedAt)
}

// GET ALL lecturers
func (r *LecturerRepository) GetAll() ([]model.Lecturer, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
	`)
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

	row := r.DB.QueryRow(query, id)

	var lec model.Lecturer
	if err := row.Scan(
		&lec.ID,
		&lec.UserID,
		&lec.LecturerID,
		&lec.Department,
		&lec.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &lec, nil
}

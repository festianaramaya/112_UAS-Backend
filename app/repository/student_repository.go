package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
)

type StudentRepository struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{DB: db}
}

// GetStudentByID - Mengambil detail mahasiswa berdasarkan ID UUID
func (r *StudentRepository) GetStudentByID(ctx context.Context, id string) (*model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at, updated_at
		FROM students
		WHERE id = $1
	`
	row := r.DB.QueryRowContext(ctx, query, id)

	var s model.Student
    // Penangkapan error dari Scan
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.CreatedAt,
        &s.UpdatedAt, 
	)

	// Pengecekan error setelah Scan
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Create - Membuat data mahasiswa baru
func (r *StudentRepository) Create(ctx context.Context, s *model.Student) error {
	// Query tidak menyertakan 'id' karena DB yang generate. Menggunakan RETURNING untuk mendapatkan ID baru.
	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
	`

	row := r.DB.QueryRowContext(ctx, query,
		s.UserID,
		s.StudentID,
		s.ProgramStudy,
		s.AcademicYear,
		s.AdvisorID,
	)
    // Tangkap ID dan timestamps yang dihasilkan DB
	return row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

// GetByUserID - Mengambil data mahasiswa berdasarkan UserID (UUID)
func (r *StudentRepository) GetByUserID(ctx context.Context, userID string) (*model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at, updated_at
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`
	row := r.DB.QueryRowContext(ctx, query, userID)
	var s model.Student
	err := row.Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetAll - Mengambil daftar semua mahasiswa
func (r *StudentRepository) GetAll(ctx context.Context) ([]model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at, updated_at
		FROM students
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student

	for rows.Next() {
		var s model.Student
        // Menggunakan := untuk mendeklarasikan 'err' baru di scope loop adalah umum, 
        // namun untuk menghindari error 'declared and not used', pastikan err di scan dicek.
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err 
		}

		students = append(students, s)
	}
    
    // Pengecekan error dari iterasi rows (rows.Err())
    if err := rows.Err(); err != nil {
        return nil, err
    }

	return students, nil
}
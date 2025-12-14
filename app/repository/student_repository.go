package repository

import (
	"context"
	"database/sql"
	"uas/app/model"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type StudentRepository struct {
	DB *sqlx.DB
}

// GANTI: Constructor menerima *sqlx.DB
func NewStudentRepository(db *sqlx.DB) *StudentRepository {
	return &StudentRepository{DB: db}
}

// ----------------------------------------------------------------------------------
// UTILITY SCANNER: Fungsi untuk men-scan baris SELECT ke model.Student (tanpa updated_at)
// ----------------------------------------------------------------------------------

func scanStudent(row *sql.Row, s *model.Student) error {
	// Variabel penampung untuk field yang bisa NULL
	var nullAdvisorID sql.NullString
	
	// Scan 7 kolom
	err := row.Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, 
		&nullAdvisorID, // Scan ke sql.NullString (untuk advisor_id)
		&s.CreatedAt, 
		// Hapus &nullUpdatedAt
	)

	if err != nil {
		return err
	}
	
	// Konversi nilai Nullable kembali ke struct Student
	s.AdvisorID = nullAdvisorID
	// Hapus s.UpdatedAt

	return nil
}

func scanStudents(rows *sql.Rows, students *[]model.Student) error {
	for rows.Next() {
		var s model.Student
		var nullAdvisorID sql.NullString
		// Hapus var nullUpdatedAt

		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, 
			&nullAdvisorID, // Scan ke sql.NullString
			&s.CreatedAt, 
			// Hapus &nullUpdatedAt
		); err != nil {
			return err
		}

		// Konversi Nullable ke struct Student
		s.AdvisorID = nullAdvisorID
		// Hapus s.UpdatedAt
		
		*students = append(*students, s)
	}

	return rows.Err()
}


// ----------------------------------------------------------------------------------
// Implementasi Methods
// ----------------------------------------------------------------------------------

// GetStudentByID - Mengambil detail mahasiswa berdasarkan ID UUID
func (r *StudentRepository) GetStudentByID(ctx context.Context, id string) (*model.Student, error) {
	// FIX: Hapus updated_at dari SELECT
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE id = $1
	`
	row := r.DB.QueryRowContext(ctx, query, id)

	var s model.Student
	if err := scanStudent(row, &s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &s, nil
}

// Create - Membuat data mahasiswa baru
func (r *StudentRepository) Create(ctx context.Context, s *model.Student) error {
	// FIX: Hapus updated_at dari RETURNING
	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	// Siapkan nilai AdvisorID untuk INSERT. Jika NULLABLE, gunakan Valuenya.
	advisorIDValue := s.AdvisorID.String
	if !s.AdvisorID.Valid {
		advisorIDValue = "" 
	}


	row := r.DB.QueryRowContext(ctx, query,
		s.UserID,
		s.StudentID,
		s.ProgramStudy,
		s.AcademicYear,
		advisorIDValue, 
	)
	
	// FIX: Hapus &s.UpdatedAt dari Scan
	return row.Scan(&s.ID, &s.CreatedAt)
}

// GetByUserID - Mengambil data mahasiswa berdasarkan UserID (UUID)
func (r *StudentRepository) GetByUserID(ctx context.Context, userID string) (*model.Student, error) {
	// FIX: Hapus updated_at dari SELECT
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`
	row := r.DB.QueryRowContext(ctx, query, userID)
	var s model.Student
	
	if err := scanStudent(row, &s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &s, nil
}

// GetAll - Mengambil daftar semua mahasiswa
func (r *StudentRepository) GetAll(ctx context.Context) ([]model.Student, error) {
	// FIX: Hapus updated_at dari SELECT
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student
	if err := scanStudents(rows, &students); err != nil {
		return nil, err
	}
	
	return students, nil
}

// UpdateAdvisorID memperbarui kolom advisor_id di tabel students
func (r *StudentRepository) UpdateAdvisorID(ctx context.Context, studentID string, advisorID sql.NullString) error {
    query := `
        UPDATE students
        SET advisor_id = $1
        WHERE id = $2
    `
    // Gunakan Exec untuk menjalankan perintah non-query
    result, err := r.DB.ExecContext(ctx, query, advisorID, studentID)
    if err != nil {
        return err
    }

    // Cek apakah ada baris yang terpengaruh (pastikan studentID ditemukan)
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return fmt.Errorf("student with ID %s not found", studentID) // Tambahkan error handling jika ID tidak ditemukan
    }

    return nil
}
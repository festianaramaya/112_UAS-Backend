// app/model/student.go (PERBAIKAN FINAL)

package model

import (
    "time"
    "database/sql" // <--- Import wajib
)

type Student struct {
    ID           string    `json:"id" db:"id"`
    UserID       string    `json:"user_id" db:"user_id"`
    StudentID    string    `json:"student_id" db:"student_id"`
    ProgramStudy string    `json:"program_study" db:"program_study"`
    AcademicYear string    `json:"academic_year" db:"academic_year"`
    
    AdvisorID    sql.NullString `json:"advisor_id" db:"advisor_id"` // <--- NULLABLE STRING
    
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
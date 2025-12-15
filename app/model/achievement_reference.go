package model

import (
	"database/sql"
	"time"
)

// AchievementReference adalah struktur untuk tabel di PostgreSQL (Metadata/Reference)
type AchievementReference struct {
    ID                 string         `db:"id" json:"id"`
    StudentID          string         `db:"student_id" json:"studentId"`
    MongoAchievementID string         `db:"mongo_achievement_id" json:"mongoAchievementId"`
    Status             string         `db:"status" json:"status"` // ENUM: draft, submitted, verified, rejected
    SubmittedAt        sql.NullTime   `db:"submitted_at" json:"submittedAt"`
    VerifiedAt         sql.NullTime   `db:"verified_at" json:"verifiedAt"`
    VerifiedBy         sql.NullString `db:"verified_by" json:"verifiedBy"` // user_id Dosen Wali
    RejectionNote      sql.NullString `db:"rejection_note" json:"rejectionNote"`
    CreatedAt          time.Time      `db:"created_at" json:"createdAt"`
    UpdatedAt          time.Time      `db:"updated_at" json:"updatedAt"`
}

// AchievementFull adalah gabungan dari PG Reference dan Mongo Detail
// Digunakan untuk response GET Detail
type AchievementFull struct {
    AchievementReference
    MongoDetails MongoAchievement `json:"details"`
}

// AchievementCreateUpdate digunakan untuk Request Body POST/PUT
type AchievementCreateUpdate struct {
    StudentID       string `json:"student_id"` // Hanya untuk POST
    AchievementType string `json:"achievement_type"`
    Title           string `json:"title"`
    Description     string `json:"description"`
    Points          float64 `json:"points"`
    Tags            []string `json:"tags"`
    Details         map[string]interface{} `json:"details"`
    Attachments     []Attachment `json:"attachments"`
}

type AchievementCreateResponse struct {
	ID                 string    `json:"id"`
	MongoAchievementID string    `json:"mongo_achievement_id"`
	StudentID          string    `json:"student_id"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

type AchievementCreateRequest struct {
	StudentID   string `json:"student_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
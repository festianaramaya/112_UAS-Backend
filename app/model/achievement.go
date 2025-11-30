package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AchievementAttachment merepresentasikan dokumen pendukung
type AchievementAttachment struct {
	FileName  string    `json:"fileName" bson:"fileName"`
	FileUrl   string    `json:"fileUrl" bson:"fileUrl"`
	FileType  string    `json:"fileType" bson:"fileType"`
	UploadedAt time.Time `json:"uploadedAt" bson:"uploadedAt"`
}

// AchievementDetails menampung field dinamis berdasarkan tipe prestasi
type AchievementDetails struct {
	// Competition fields
	CompetitionName    string `json:"competitionName,omitempty" bson:"competitionName,omitempty"`
	CompetitionLevel   string `json:"competitionLevel,omitempty" bson:"competitionLevel,omitempty"`
	Rank               *int   `json:"rank,omitempty" bson:"rank,omitempty"`
	MedalType          string `json:"medalType,omitempty" bson:"medalType,omitempty"`

	// Publication fields
	PublicationType    string   `json:"publicationType,omitempty" bson:"publicationType,omitempty"`
	PublicationTitle   string `json:"publicationTitle,omitempty" bson:"publicationTitle,omitempty"`
	Authors            []string `json:"authors,omitempty" bson:"authors,omitempty"`
	Publisher          string   `json:"publisher,omitempty" bson:"publisher,omitempty"`
	Issn               string   `json:"issn,omitempty" bson:"issn,omitempty"`

	// Organization fields
	OrganizationName   string `json:"organizationName,omitempty" bson:"organizationName,omitempty"`
	Position           string `json:"position,omitempty" bson:"position,omitempty"`
	PeriodStart        *time.Time `json:"periodStart,omitempty" bson:"periodStart,omitempty"` // period.start
	PeriodEnd          *time.Time `json:"periodEnd,omitempty" bson:"periodEnd,omitempty"`   // period.end
	
	// Certification fields
	CertificationName  string `json:"certificationName,omitempty" bson:"certificationName,omitempty"`
	IssuedBy           string `json:"issuedBy,omitempty" bson:"issuedBy,omitempty"`
	CertificationNumber string `json:"certificationNumber,omitempty" bson:"certificationNumber,omitempty"`
	ValidUntil         *time.Time `json:"validUntil,omitempty" bson:"validUntil,omitempty"`

	// General common fields
	EventDate          *time.Time `json:"eventDate,omitempty" bson:"eventDate,omitempty"`
	Location           string `json:"location,omitempty" bson:"location,omitempty"`
	Organizer          string `json:"organizer,omitempty" bson:"organizer,omitempty"`
	Score              *float64 `json:"score,omitempty" bson:"score,omitempty"`
}

// Achievement adalah model utama untuk MongoDB Collection
type Achievement struct {
	ID                primitive.ObjectID   `json:"id" bson:"_id,omitempty"` // MongoDB ID
	StudentID         string               `json:"studentId" bson:"studentId"` // Ref ke PostgreSQL (UUID)
	AchievementType   string               `json:"achievementType" bson:"achievementType"` // e.g. 'competition'
	Title             string               `json:"title" bson:"title"`
	Description       string               `json:"description" bson:"description"`
	Details           AchievementDetails   `json:"details" bson:"details"` // Field dinamis
	CustomFields      map[string]interface{} `json:"customFields,omitempty" bson:"customFields,omitempty"`
	Attachments       []AchievementAttachment `json:"attachments" bson:"attachments"`
	Tags              []string             `json:"tags" bson:"tags"`
	Points            float64              `json:"points" bson:"points"`
	CreatedAt         time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time            `json:"updatedAt" bson:"updatedAt"`
}
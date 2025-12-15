package model

import (
	"time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoAchievement adalah struktur untuk Collection di MongoDB (Detail)
type MongoAchievement struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    StudentID      string             `bson:"student_id" json:"student_id"`
    AchievementType string             `bson:"achievement_type" json:"achievementType"`
    Title          string             `bson:"title" json:"title"`
    Description    string             `bson:"description" json:"description"`
    Points         float64            `bson:"points" json:"points"`
    Tags           []string           `bson:"tags" json:"tags"`
    Details        map[string]interface{} `bson:"details" json:"details"`
    Attachments    []Attachment       `bson:"attachments" json:"attachments"`
    CreatedAt      time.Time          `bson:"created_at" json:"createdAt"`
    UpdatedAt      time.Time          `bson:"updated_at" json:"updatedAt"`
}

type Attachment struct {
    Filename string `bson:"filename" json:"filename"`
    Url      string `bson:"url" json:"url"`
}
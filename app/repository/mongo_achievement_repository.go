package repository

import (
	"context"
	"fmt"
	"uas/app/model"
    
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// MongoAchievementRepository mengelola operasi database MongoDB
type MongoAchievementRepository struct {
	Collection *mongo.Collection
}

// Create menyimpan detail prestasi baru ke MongoDB
func (r *MongoAchievementRepository) Create(ctx context.Context, achievement *model.MongoAchievement) (primitive.ObjectID, error) {
    achievement.CreatedAt = time.Now()
    achievement.UpdatedAt = time.Now()
    res, err := r.Collection.InsertOne(ctx, achievement)
    if err != nil {
        return primitive.NilObjectID, err
    }
    return res.InsertedID.(primitive.ObjectID), nil
}

// GetByID mengambil detail prestasi berdasarkan ID Mongo
func (r *MongoAchievementRepository) GetByID(ctx context.Context, id primitive.ObjectID) (model.MongoAchievement, error) {
	var achievement model.MongoAchievement
	filter := bson.M{"_id": id}
	err := r.Collection.FindOne(ctx, filter).Decode(&achievement)
	return achievement, err
}

// Delete menghapus detail prestasi dari MongoDB (digunakan untuk Rollback)
func (r *MongoAchievementRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    filter := bson.M{"_id": id}
    res, err := r.Collection.DeleteOne(ctx, filter)
    if err != nil {
        return err
    }
    if res.DeletedCount == 0 {
        return fmt.Errorf("no achievement found with ID %s to delete", id.Hex())
    }
    return nil
}

// achievementData harus berupa map[string]interface{} atau bson.M
func (r *MongoAchievementRepository) Update(ctx context.Context, id primitive.ObjectID, achievementData map[string]interface{}) error {
    updateFields := bson.M{"$set": achievementData}
    updateFields["$set"].(bson.M)["updated_at"] = time.Now() // Set updated_at

    filter := bson.M{"_id": id}
    
    res, err := r.Collection.UpdateOne(ctx, filter, updateFields)
    if err != nil {
        return err
    }
    if res.ModifiedCount == 0 {
        return fmt.Errorf("no achievement found with ID %s to update", id.Hex())
    }
    return nil
}

// NewMongoAchievementRepository adalah constructor untuk MongoAchievementRepository
func NewMongoAchievementRepository(collection *mongo.Collection) *MongoAchievementRepository {
    return &MongoAchievementRepository{Collection: collection}
}
package repository

import (
	"context"
	"fmt"
	"time"
	"uas/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAchievementRepository struct {
	Collection *mongo.Collection
}

func NewMongoAchievementRepository(col *mongo.Collection) *MongoAchievementRepository {
	return &MongoAchievementRepository{Collection: col}
}

func (r *MongoAchievementRepository) Create(
	ctx context.Context,
	achievement *model.MongoAchievement,
) (primitive.ObjectID, error) {

	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	res, err := r.Collection.InsertOne(ctx, achievement)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *MongoAchievementRepository) GetByID(
	ctx context.Context,
	id primitive.ObjectID,
) (model.MongoAchievement, error) {

	var a model.MongoAchievement
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	return a, err
}

func (r *MongoAchievementRepository) Update(
	ctx context.Context,
	id primitive.ObjectID,
	data map[string]interface{},
) error {

	data["updated_at"] = time.Now()
	res, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": data},
	)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("no data updated")
	}
	return nil
}

func (r *MongoAchievementRepository) Delete(
	ctx context.Context,
	id primitive.ObjectID,
) error {

	res, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *MongoAchievementRepository) AddAttachment(
	ctx context.Context,
	id primitive.ObjectID,
	attachment model.Attachment,
) error {

	update := bson.M{
		"$push": bson.M{
			"attachments": attachment,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	return err
}

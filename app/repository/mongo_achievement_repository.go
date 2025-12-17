package repository

import (
	"context"
	"fmt"
	"time"
	"uas/app/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAchievementRepository struct {
	Collection *mongo.Collection
}

func (r *MongoAchievementRepository) Create(
	ctx context.Context,
	data *model.MongoAchievement,
) (primitive.ObjectID, error) {

	data.ID = primitive.NewObjectID()

	_, err := r.Collection.InsertOne(ctx, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return data.ID, nil
}


func NewMongoAchievementRepository(col *mongo.Collection) *MongoAchievementRepository {
	return &MongoAchievementRepository{
		Collection: col,
	}
}

func (r *MongoAchievementRepository) GetAdviseeAchievements(
	ctx context.Context,
	lecturerID uuid.UUID,
) ([]model.AchievementFull, error) {

	filter := bson.M{
		"lecturer_id": lecturerID.String(), // ⬅️ Mongo simpan UUID sebagai string
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []model.AchievementFull
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
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

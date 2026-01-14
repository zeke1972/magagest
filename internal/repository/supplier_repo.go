// internal/repository/promotion_repo.go

package repository

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ricambi-manager/internal/domain"
)

type PromotionRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewPromotionRepository(db *mongo.Database) *PromotionRepository {
	return &PromotionRepository{
		collection: db.Collection("promotions"),
		db:         db,
	}
}

func (r *PromotionRepository) Create(ctx context.Context, promotion *domain.Promotion) error {
	if promotion.ID.IsZero() {
		promotion.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, promotion)
	return err
}

func (r *PromotionRepository) Update(ctx context.Context, promotion *domain.Promotion) error {
	promotion.UpdatedAt = time.Now()

	filter := bson.M{"_id": promotion.ID}
	update := bson.M{"$set": promotion}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPromotionNotFound
	}

	return nil
}

func (r *PromotionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrPromotionNotFound
	}

	return nil
}

func (r *PromotionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Promotion, error) {
	var promotion domain.Promotion
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&promotion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPromotionNotFound
		}
		return nil, err
	}

	return &promotion, nil
}

func (r *PromotionRepository) FindByCode(ctx context.Context, code string) (*domain.Promotion, error) {
	var promotion domain.Promotion
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}

	err := r.collection.FindOne(ctx, filter).Decode(&promotion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPromotionNotFound
		}
		return nil, err
	}

	return &promotion, nil
}

func (r *PromotionRepository) FindActive(ctx context.Context, date time.Time) ([]*domain.Promotion, error) {
	filter := bson.M{
		"is_active":  true,
		"valid_from": bson.M{"$lte": date},
		"$or": []bson.M{
			{"valid_to": bson.M{"$gte": date}},
			{"valid_to": time.Time{}},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var promotions []*domain.Promotion
	if err = cursor.All(ctx, &promotions); err != nil {
		return nil, err
	}

	return promotions, nil
}

func (r *PromotionRepository) FindExpired(ctx context.Context, date time.Time) ([]*domain.Promotion, error) {
	filter := bson.M{
		"valid_to": bson.M{"$lt": date, "$ne": time.Time{}},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var promotions []*domain.Promotion
	if err = cursor.All(ctx, &promotions); err != nil {
		return nil, err
	}

	return promotions, nil
}

func (r *PromotionRepository) FindExpiringSoon(ctx context.Context, days int) ([]*domain.Promotion, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"is_active": true,
		"valid_to":  bson.M{"$gte": now, "$lte": futureDate},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var promotions []*domain.Promotion
	if err = cursor.All(ctx, &promotions); err != nil {
		return nil, err
	}

	return promotions, nil
}

func (r *PromotionRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Promotion, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "valid_from", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var promotions []*domain.Promotion
	if err = cursor.All(ctx, &promotions); err != nil {
		return nil, err
	}

	return promotions, nil
}

func (r *PromotionRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "code", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "valid_from", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "valid_to", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "priority", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// internal/repository/kit_repo.go

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

type KitRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewKitRepository(db *mongo.Database) *KitRepository {
	return &KitRepository{
		collection: db.Collection("kits"),
		db:         db,
	}
}

func (r *KitRepository) Create(ctx context.Context, kit *domain.Kit) error {
	if kit.ID.IsZero() {
		kit.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, kit)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return mongo.ErrClientDisconnected
		}
		return err
	}

	return nil
}

func (r *KitRepository) Update(ctx context.Context, kit *domain.Kit) error {
	kit.UpdatedAt = time.Now()

	filter := bson.M{"_id": kit.ID}
	update := bson.M{"$set": kit}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrKitNotFound
	}

	return nil
}

func (r *KitRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrKitNotFound
	}

	return nil
}

func (r *KitRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Kit, error) {
	var kit domain.Kit
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&kit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrKitNotFound
		}
		return nil, err
	}

	return &kit, nil
}

func (r *KitRepository) FindByCode(ctx context.Context, code string) (*domain.Kit, error) {
	var kit domain.Kit
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}

	err := r.collection.FindOne(ctx, filter).Decode(&kit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrKitNotFound
		}
		return nil, err
	}

	return &kit, nil
}

func (r *KitRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Kit, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"code": bson.M{"$regex": query, "$options": "i"}},
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}

	return kits, nil
}

func (r *KitRepository) FindByCategory(ctx context.Context, category string) ([]*domain.Kit, error) {
	filter := bson.M{
		"category":  category,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}

	return kits, nil
}

func (r *KitRepository) FindContainingArticle(ctx context.Context, articleID primitive.ObjectID) ([]*domain.Kit, error) {
	filter := bson.M{
		"components.article_id": articleID,
		"is_active":             true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}

	return kits, nil
}

func (r *KitRepository) FindPopular(ctx context.Context, limit int) ([]*domain.Kit, error) {
	filter := bson.M{"is_active": true}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "sales_count", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}

	return kits, nil
}

func (r *KitRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Kit, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	filter := bson.M{"is_active": true}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kits []*domain.Kit
	if err = cursor.All(ctx, &kits); err != nil {
		return nil, err
	}

	return kits, nil
}

func (r *KitRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true})
}

func (r *KitRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "components.article_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "sales_count", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *KitRepository) Exists(ctx context.Context, code string) (bool, error) {
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

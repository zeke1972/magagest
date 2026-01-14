// internal/repository/budget_repo.go

package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ricambi-manager/internal/domain"
)

type BudgetRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewBudgetRepository(db *mongo.Database) *BudgetRepository {
	return &BudgetRepository{
		collection: db.Collection("budgets"),
		db:         db,
	}
}

func (r *BudgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	if budget.ID.IsZero() {
		budget.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, budget)
	return err
}

func (r *BudgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	budget.UpdatedAt = time.Now()

	filter := bson.M{"_id": budget.ID}
	update := bson.M{"$set": budget}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrBudgetNotFound
	}

	return nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrBudgetNotFound
	}

	return nil
}

func (r *BudgetRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Budget, error) {
	var budget domain.Budget
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&budget)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}

	return &budget, nil
}

func (r *BudgetRepository) FindByEntity(ctx context.Context, entityID primitive.ObjectID, budgetType domain.BudgetType) ([]*domain.Budget, error) {
	filter := bson.M{
		"entity_id": entityID,
		"type":      budgetType,
	}

	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}, {Key: "quarter", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) FindByEntityAndPeriod(ctx context.Context, entityID primitive.ObjectID, budgetType domain.BudgetType, year, quarter int) (*domain.Budget, error) {
	var budget domain.Budget
	filter := bson.M{
		"entity_id": entityID,
		"type":      budgetType,
		"year":      year,
		"quarter":   quarter,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&budget)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}

	return &budget, nil
}

func (r *BudgetRepository) FindActive(ctx context.Context, date time.Time) ([]*domain.Budget, error) {
	filter := bson.M{
		"start_date": bson.M{"$lte": date},
		"end_date":   bson.M{"$gte": date},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) FindByType(ctx context.Context, budgetType domain.BudgetType) ([]*domain.Budget, error) {
	filter := bson.M{
		"type": budgetType,
	}

	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) FindExpired(ctx context.Context, date time.Time) ([]*domain.Budget, error) {
	filter := bson.M{
		"end_date": bson.M{"$lt": date},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Budget, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "year", Value: -1}, {Key: "quarter", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "entity_id", Value: 1}, {Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "year", Value: 1}, {Key: "quarter", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "start_date", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "end_date", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

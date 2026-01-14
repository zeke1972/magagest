// internal/repository/operator_repo.go

package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ricambi-manager/internal/domain"
)

type OperatorRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewOperatorRepository(db *mongo.Database) *OperatorRepository {
	return &OperatorRepository{
		collection: db.Collection("operators"),
		db:         db,
	}
}

func (r *OperatorRepository) Create(ctx context.Context, operator *domain.Operator) error {
	if operator.ID.IsZero() {
		operator.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, operator)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("operator with this username already exists")
		}
		return err
	}

	return nil
}

func (r *OperatorRepository) Update(ctx context.Context, operator *domain.Operator) error {
	operator.UpdatedAt = time.Now()

	filter := bson.M{"_id": operator.ID}
	update := bson.M{"$set": operator}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Operator, error) {
	var operator domain.Operator
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&operator)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrOperatorNotFound
		}
		return nil, err
	}

	return &operator, nil
}

func (r *OperatorRepository) FindByUsername(ctx context.Context, username string) (*domain.Operator, error) {
	var operator domain.Operator
	filter := bson.M{"username": strings.ToLower(strings.TrimSpace(username))}

	err := r.collection.FindOne(ctx, filter).Decode(&operator)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrOperatorNotFound
		}
		return nil, err
	}

	return &operator, nil
}

func (r *OperatorRepository) FindByEmail(ctx context.Context, email string) (*domain.Operator, error) {
	var operator domain.Operator
	filter := bson.M{"email": strings.ToLower(strings.TrimSpace(email))}

	err := r.collection.FindOne(ctx, filter).Decode(&operator)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrOperatorNotFound
		}
		return nil, err
	}

	return &operator, nil
}

func (r *OperatorRepository) FindBySessionToken(ctx context.Context, token string) (*domain.Operator, error) {
	var operator domain.Operator
	filter := bson.M{
		"session_token":  token,
		"session_expiry": bson.M{"$gt": time.Now()},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&operator)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrOperatorNotFound
		}
		return nil, err
	}

	return &operator, nil
}

func (r *OperatorRepository) FindByProfile(ctx context.Context, profile domain.ProfileType) ([]*domain.Operator, error) {
	filter := bson.M{
		"profile":   profile,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operators []*domain.Operator
	if err = cursor.All(ctx, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *OperatorRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Operator, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operators []*domain.Operator
	if err = cursor.All(ctx, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *OperatorRepository) FindActive(ctx context.Context) ([]*domain.Operator, error) {
	filter := bson.M{"is_active": true}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operators []*domain.Operator
	if err = cursor.All(ctx, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *OperatorRepository) FindLocked(ctx context.Context) ([]*domain.Operator, error) {
	filter := bson.M{"is_locked": true}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operators []*domain.Operator
	if err = cursor.All(ctx, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *OperatorRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *OperatorRepository) CountActive(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true})
}

func (r *OperatorRepository) UpdateLastLogin(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"last_login": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) UpdatePassword(ctx context.Context, operatorID primitive.ObjectID, passwordHash string) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"password_hash":        passwordHash,
			"last_password_change": time.Now(),
			"updated_at":           time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) IncrementFailedAttempts(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$inc": bson.M{"failed_attempts": 1},
		"$set": bson.M{
			"last_failed_attempt": time.Now(),
			"updated_at":          time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) ResetFailedAttempts(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"failed_attempts": 0,
			"updated_at":      time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) Lock(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"is_locked":  true,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) Unlock(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"is_locked":       false,
			"failed_attempts": 0,
			"updated_at":      time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) UpdateSession(ctx context.Context, operatorID primitive.ObjectID, token string, expiry time.Time) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"session_token":  token,
			"session_expiry": expiry,
			"last_login":     time.Now(),
			"updated_at":     time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) ClearSession(ctx context.Context, operatorID primitive.ObjectID) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$set": bson.M{
			"session_token":  "",
			"session_expiry": time.Time{},
			"updated_at":     time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) AddAuditEntry(ctx context.Context, operatorID primitive.ObjectID, entry domain.AuditEntry) error {
	filter := bson.M{"_id": operatorID}
	update := bson.M{
		"$push": bson.M{"audit_log": entry},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrOperatorNotFound
	}

	return nil
}

func (r *OperatorRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Operator, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"full_name": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operators []*domain.Operator
	if err = cursor.All(ctx, &operators); err != nil {
		return nil, err
	}

	return operators, nil
}

func (r *OperatorRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "profile", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_locked", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "session_token", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "session_expiry", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *OperatorRepository) Exists(ctx context.Context, username string) (bool, error) {
	filter := bson.M{"username": strings.ToLower(strings.TrimSpace(username))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *OperatorRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"email": strings.ToLower(strings.TrimSpace(email))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *OperatorRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	filter := bson.M{
		"session_expiry": bson.M{"$lt": time.Now(), "$ne": time.Time{}},
	}
	update := bson.M{
		"$set": bson.M{
			"session_token":  "",
			"session_expiry": time.Time{},
			"updated_at":     time.Now(),
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

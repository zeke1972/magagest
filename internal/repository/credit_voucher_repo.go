// internal/repository/credit_voucher_repo.go

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

type CreditVoucherRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewCreditVoucherRepository(db *mongo.Database) *CreditVoucherRepository {
	return &CreditVoucherRepository{
		collection: db.Collection("credit_vouchers"),
		db:         db,
	}
}

func (r *CreditVoucherRepository) Create(ctx context.Context, voucher *domain.CreditVoucher) error {
	if voucher.ID.IsZero() {
		voucher.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, voucher)
	return err
}

func (r *CreditVoucherRepository) Update(ctx context.Context, voucher *domain.CreditVoucher) error {
	voucher.UpdatedAt = time.Now()

	filter := bson.M{"_id": voucher.ID}
	update := bson.M{"$set": voucher}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrVoucherNotFound
	}

	return nil
}

func (r *CreditVoucherRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.CreditVoucher, error) {
	var voucher domain.CreditVoucher
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&voucher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, err
	}

	return &voucher, nil
}

func (r *CreditVoucherRepository) FindByCode(ctx context.Context, code string) (*domain.CreditVoucher, error) {
	var voucher domain.CreditVoucher
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}

	err := r.collection.FindOne(ctx, filter).Decode(&voucher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrVoucherNotFound
		}
		return nil, err
	}

	return &voucher, nil
}

func (r *CreditVoucherRepository) FindByCustomer(ctx context.Context, customerID primitive.ObjectID) ([]*domain.CreditVoucher, error) {
	filter := bson.M{"customer_id": customerID}

	opts := options.Find().SetSort(bson.D{{Key: "issued_date", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vouchers []*domain.CreditVoucher
	if err = cursor.All(ctx, &vouchers); err != nil {
		return nil, err
	}

	return vouchers, nil
}

func (r *CreditVoucherRepository) FindActiveByCustomer(ctx context.Context, customerID primitive.ObjectID) ([]*domain.CreditVoucher, error) {
	now := time.Now()
	filter := bson.M{
		"customer_id":      customerID,
		"remaining_amount": bson.M{"$gt": 0},
		"status":           bson.M{"$in": []domain.VoucherStatus{domain.VoucherStatusIssued, domain.VoucherStatusPartiallyUsed}},
		"$or": []bson.M{
			{"expiry_date": bson.M{"$gte": now}},
			{"expiry_date": time.Time{}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vouchers []*domain.CreditVoucher
	if err = cursor.All(ctx, &vouchers); err != nil {
		return nil, err
	}

	return vouchers, nil
}

func (r *CreditVoucherRepository) FindExpired(ctx context.Context, date time.Time) ([]*domain.CreditVoucher, error) {
	filter := bson.M{
		"expiry_date":      bson.M{"$lt": date, "$ne": time.Time{}},
		"status":           bson.M{"$ne": domain.VoucherStatusExpired},
		"remaining_amount": bson.M{"$gt": 0},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vouchers []*domain.CreditVoucher
	if err = cursor.All(ctx, &vouchers); err != nil {
		return nil, err
	}

	return vouchers, nil
}

func (r *CreditVoucherRepository) FindExpiringSoon(ctx context.Context, days int) ([]*domain.CreditVoucher, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"expiry_date":      bson.M{"$gte": now, "$lte": futureDate},
		"remaining_amount": bson.M{"$gt": 0},
		"status":           bson.M{"$in": []domain.VoucherStatus{domain.VoucherStatusIssued, domain.VoucherStatusPartiallyUsed}},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vouchers []*domain.CreditVoucher
	if err = cursor.All(ctx, &vouchers); err != nil {
		return nil, err
	}

	return vouchers, nil
}

func (r *CreditVoucherRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "code", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "customer_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "expiry_date", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// internal/repository/customer_repo.go

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

type CustomerRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewCustomerRepository(db *mongo.Database) *CustomerRepository {
	return &CustomerRepository{
		collection: db.Collection("customers"),
		db:         db,
	}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	if customer.ID.IsZero() {
		customer.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, customer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("customer with this code already exists")
		}
		return err
	}

	return nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	customer.UpdatedAt = time.Now()

	filter := bson.M{"_id": customer.ID}
	update := bson.M{"$set": customer}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Customer, error) {
	var customer domain.Customer
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}

	return &customer, nil
}

func (r *CustomerRepository) FindByCode(ctx context.Context, code string) (*domain.Customer, error) {
	var customer domain.Customer
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}

	err := r.collection.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}

	return &customer, nil
}

func (r *CustomerRepository) FindByVATNumber(ctx context.Context, vatNumber string) (*domain.Customer, error) {
	var customer domain.Customer
	filter := bson.M{"vat_number": strings.ToUpper(strings.TrimSpace(vatNumber))}

	err := r.collection.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, err
	}

	return &customer, nil
}

func (r *CustomerRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Customer, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"code": bson.M{"$regex": query, "$options": "i"}},
			{"company_name": bson.M{"$regex": query, "$options": "i"}},
			{"vat_number": bson.M{"$regex": query, "$options": "i"}},
			{"contact_info.email": bson.M{"$regex": query, "$options": "i"}},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "company_name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindByCategory(ctx context.Context, category domain.CustomerCategory, limit int) ([]*domain.Customer, error) {
	filter := bson.M{
		"category":  category,
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "company_name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindByCreditClass(ctx context.Context, creditClass domain.CreditClass, limit int) ([]*domain.Customer, error) {
	filter := bson.M{
		"credit_info.credit_class": creditClass,
		"is_active":                true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "company_name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindWithFidoWarning(ctx context.Context, warningThreshold float64) ([]*domain.Customer, error) {
	filter := bson.M{
		"$expr": bson.M{
			"$gte": []interface{}{
				bson.M{"$multiply": []interface{}{
					bson.M{"$divide": []interface{}{"$credit_info.current_exposure", "$credit_info.fido_limit"}},
					100,
				}},
				warningThreshold,
			},
		},
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindWithBlockedSales(ctx context.Context) ([]*domain.Customer, error) {
	filter := bson.M{
		"credit_info.block_sales": true,
		"is_active":               true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindWithOverduePayments(ctx context.Context) ([]*domain.Customer, error) {
	filter := bson.M{
		"credit_info.overdue_amount": bson.M{"$gt": 0},
		"is_active":                  true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "credit_info.overdue_amount", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Customer, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "company_name", Value: 1}})

	filter := bson.M{"is_active": true}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true})
}

func (r *CustomerRepository) UpdateExposure(ctx context.Context, customerID primitive.ObjectID, unpaidInvoices, openOrders float64) error {
	filter := bson.M{"_id": customerID}
	update := bson.M{
		"$set": bson.M{
			"credit_info.unpaid_invoices":   unpaidInvoices,
			"credit_info.open_orders":       openOrders,
			"credit_info.current_exposure":  unpaidInvoices + openOrders,
			"credit_info.last_credit_check": time.Now(),
			"updated_at":                    time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) BlockSales(ctx context.Context, customerID primitive.ObjectID, reason string) error {
	filter := bson.M{"_id": customerID}
	update := bson.M{
		"$set": bson.M{
			"credit_info.block_sales":  true,
			"credit_info.block_reason": reason,
			"updated_at":               time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) UnblockSales(ctx context.Context, customerID primitive.ObjectID) error {
	filter := bson.M{"_id": customerID}
	update := bson.M{
		"$set": bson.M{
			"credit_info.block_sales":  false,
			"credit_info.block_reason": "",
			"updated_at":               time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) AddDiscountRule(ctx context.Context, customerID primitive.ObjectID, rule domain.DiscountRule) error {
	filter := bson.M{"_id": customerID}
	update := bson.M{
		"$push": bson.M{"discount_grid": rule},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) RemoveDiscountRule(ctx context.Context, customerID, ruleID primitive.ObjectID) error {
	filter := bson.M{"_id": customerID}
	update := bson.M{
		"$pull": bson.M{"discount_grid": bson.M{"id": ruleID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Customer, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *CustomerRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "vat_number", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "company_name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "credit_info.credit_class", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "credit_info.block_sales", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "contact_info.email", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *CustomerRepository) Exists(ctx context.Context, code string) (bool, error) {
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CustomerRepository) VATExists(ctx context.Context, vatNumber string) (bool, error) {
	filter := bson.M{"vat_number": strings.ToUpper(strings.TrimSpace(vatNumber))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CustomerRepository) GetTopCustomers(ctx context.Context, limit int) ([]*domain.Customer, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "credit_info.current_exposure", Value: -1}})

	filter := bson.M{"is_active": true}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err = cursor.All(ctx, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

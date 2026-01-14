// internal/repository/article_repo.go

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

type ArticleRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewArticleRepository(db *mongo.Database) *ArticleRepository {
	return &ArticleRepository{
		collection: db.Collection("articles"),
		db:         db,
	}
}

func (r *ArticleRepository) Create(ctx context.Context, article *domain.Article) error {
	if article.ID.IsZero() {
		article.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, article)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("article with this code already exists")
		}
		return err
	}

	return nil
}

func (r *ArticleRepository) Update(ctx context.Context, article *domain.Article) error {
	article.UpdatedAt = time.Now()

	filter := bson.M{"_id": article.ID}
	update := bson.M{"$set": article}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrArticleNotFound
	}

	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrArticleNotFound
	}

	return nil
}

func (r *ArticleRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Article, error) {
	var article domain.Article
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

func (r *ArticleRepository) FindByCode(ctx context.Context, code string) (*domain.Article, error) {
	var article domain.Article
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}

	err := r.collection.FindOne(ctx, filter).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

func (r *ArticleRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.Article, error) {
	var article domain.Article
	filter := bson.M{"barcodes": barcode}

	err := r.collection.FindOne(ctx, filter).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

func (r *ArticleRepository) SearchByCode(ctx context.Context, query string, limit int) ([]*domain.Article, error) {
	query = strings.ToUpper(strings.TrimSpace(query))

	filter := bson.M{
		"$or": []bson.M{
			{"code": bson.M{"$regex": query, "$options": "i"}},
			{"code": bson.M{"$regex": "^" + query}},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) SearchByDescription(ctx context.Context, query string, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"extended_desc": bson.M{"$regex": query, "$options": "i"}},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "description", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) SearchFuzzy(ctx context.Context, query string, limit int) ([]*domain.Article, error) {
	query = strings.TrimSpace(query)

	filter := bson.M{
		"$or": []bson.M{
			{"code": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"extended_desc": bson.M{"$regex": query, "$options": "i"}},
			{"brand": bson.M{"$regex": query, "$options": "i"}},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindByApplicability(ctx context.Context, make, model string, year int, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"applicability": bson.M{
			"$elemMatch": bson.M{
				"make":      bson.M{"$regex": "^" + make + "$", "$options": "i"},
				"model":     bson.M{"$regex": "^" + model + "$", "$options": "i"},
				"year_from": bson.M{"$lte": year},
				"$or": []bson.M{
					{"year_to": bson.M{"$gte": year}},
					{"year_to": 0},
				},
			},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindByFamily(ctx context.Context, family string, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"family":    family,
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindByPrecodice(ctx context.Context, precodice string, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"precodice": precodice,
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindByClassification(ctx context.Context, classification string, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"classification": classification,
		"is_active":      true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindLowStock(ctx context.Context, limit int) ([]*domain.Article, error) {
	filter := bson.M{
		"$expr": bson.M{
			"$lte": []interface{}{"$stock.available", "$stock.reorder_point"},
		},
		"is_active": true,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "stock.available", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindWithExpiredNetPrices(ctx context.Context, date time.Time) ([]*domain.Article, error) {
	filter := bson.M{
		"pricing.net_prices": bson.M{
			"$elemMatch": bson.M{
				"valid_to": bson.M{"$lte": date, "$ne": time.Time{}},
			},
		},
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindReplacementChain(ctx context.Context, code string) ([]*domain.Article, error) {
	var articles []*domain.Article
	visited := make(map[string]bool)

	current := code
	for {
		if visited[current] {
			break
		}
		visited[current] = true

		article, err := r.FindByCode(ctx, current)
		if err != nil {
			break
		}

		articles = append(articles, article)

		if article.ReplacedBy == "" {
			break
		}

		current = article.ReplacedBy
	}

	return articles, nil
}

func (r *ArticleRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Article, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) FindAll(ctx context.Context, skip, limit int) ([]*domain.Article, error) {
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "code", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*domain.Article
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true})
}

func (r *ArticleRepository) CountByFilter(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *ArticleRepository) UpdateStock(ctx context.Context, articleID primitive.ObjectID, quantity, reserved float64) error {
	filter := bson.M{"_id": articleID}
	update := bson.M{
		"$set": bson.M{
			"stock.quantity":          quantity,
			"stock.reserved":          reserved,
			"stock.available":         quantity - reserved,
			"stock.last_movement_date": time.Now(),
			"updated_at":              time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrArticleNotFound
	}

	return nil
}

func (r *ArticleRepository) BulkUpdatePrices(ctx context.Context, updates map[primitive.ObjectID]float64) error {
	var models []mongo.WriteModel

	for id, price := range updates {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"pricing.list_price": price,
					"updated_at":         time.Now(),
				},
			})
		models = append(models, model)
	}

	if len(models) == 0 {
		return nil
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := r.collection.BulkWrite(ctx, models, opts)
	return err
}

func (r *ArticleRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "barcodes", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "description", Value: "text"}, {Key: "extended_desc", Value: "text"}},
		},
		{
			Keys: bson.D{{Key: "family", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "precodice", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "classification", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "applicability.make", Value: 1}, {Key: "applicability.model", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "stock.available", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *ArticleRepository) Exists(ctx context.Context, code string) (bool, error) {
	filter := bson.M{"code": strings.ToUpper(strings.TrimSpace(code))}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ArticleRepository) BarcodeExists(ctx context.Context, barcode string) (bool, error) {
	filter := bson.M{"barcodes": barcode}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

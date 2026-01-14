// internal/usecase/manage_stock.go

package usecase

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"ricambi-manager/internal/domain"
	"ricambi-manager/internal/repository"
)

type ManageStockUseCase struct {
	articleRepo *repository.ArticleRepository
	kitRepo     *repository.KitRepository
}

func NewManageStockUseCase(
	articleRepo *repository.ArticleRepository,
	kitRepo *repository.KitRepository,
) *ManageStockUseCase {
	return &ManageStockUseCase{
		articleRepo: articleRepo,
		kitRepo:     kitRepo,
	}
}

type StockMovement struct {
	ArticleID    primitive.ObjectID
	ArticleCode  string
	Quantity     float64
	MovementType string
	Reason       string
	OperatorID   string
}

func (uc *ManageStockUseCase) AddStock(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity float64,
	operator *domain.Operator,
) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	if err := article.AddStock(quantity); err != nil {
		return err
	}

	operator.AddAuditEntry(
		"add_stock",
		"warehouse",
		articleID.Hex(),
		fmt.Sprintf("Added %.2f units to %s", quantity, article.Code),
		"",
	)

	return uc.articleRepo.Update(ctx, article)
}

func (uc *ManageStockUseCase) RemoveStock(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity float64,
	operator *domain.Operator,
) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	if err := article.RemoveStock(quantity); err != nil {
		return err
	}

	operator.AddAuditEntry(
		"remove_stock",
		"warehouse",
		articleID.Hex(),
		fmt.Sprintf("Removed %.2f units from %s", quantity, article.Code),
		"",
	)

	return uc.articleRepo.Update(ctx, article)
}

func (uc *ManageStockUseCase) ReserveStock(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity float64,
) error {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	if err := article.ReserveStock(quantity); err != nil {
		return err
	}

	return uc.articleRepo.Update(ctx, article)
}

func (uc *ManageStockUseCase) ReleaseReservedStock(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity float64,
) error {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	if err := article.ReleaseReservedStock(quantity); err != nil {
		return err
	}

	return uc.articleRepo.Update(ctx, article)
}

func (uc *ManageStockUseCase) GetLowStockArticles(ctx context.Context, limit int) ([]*domain.Article, error) {
	return uc.articleRepo.FindLowStock(ctx, limit)
}

func (uc *ManageStockUseCase) CheckStockAvailability(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity float64,
) (bool, error) {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return false, err
	}

	return article.Stock.Available >= quantity, nil
}

func (uc *ManageStockUseCase) CheckKitAvailability(
	ctx context.Context,
	kitID primitive.ObjectID,
	quantity float64,
) (bool, []string, error) {
	kit, err := uc.kitRepo.FindByID(ctx, kitID)
	if err != nil {
		return false, nil, err
	}

	articleIDs := make([]primitive.ObjectID, len(kit.Components))
	for i, comp := range kit.Components {
		articleIDs[i] = comp.ArticleID
	}

	articles, err := uc.articleRepo.FindByIDs(ctx, articleIDs)
	if err != nil {
		return false, nil, err
	}

	articleMap := make(map[primitive.ObjectID]*domain.Article)
	for _, article := range articles {
		articleMap[article.ID] = article
	}

	// CanFulfill ritorna (bool, []string) NON (bool, []string, error)
	canFulfill, missingArticles := kit.CanFulfill(quantity, articleMap)
	return canFulfill, missingArticles, nil
}

func (uc *ManageStockUseCase) ReserveKitComponents(
	ctx context.Context,
	kitID primitive.ObjectID,
	quantity float64,
) error {
	kit, err := uc.kitRepo.FindByID(ctx, kitID)
	if err != nil {
		return err
	}

	articleIDs := make([]primitive.ObjectID, len(kit.Components))
	for i, comp := range kit.Components {
		articleIDs[i] = comp.ArticleID
	}

	articles, err := uc.articleRepo.FindByIDs(ctx, articleIDs)
	if err != nil {
		return err
	}

	articleMap := make(map[primitive.ObjectID]*domain.Article)
	for _, article := range articles {
		articleMap[article.ID] = article
	}

	if err := kit.ReserveComponents(quantity, articleMap); err != nil {
		return err
	}

	for _, article := range articles {
		if err := uc.articleRepo.Update(ctx, article); err != nil {
			return err
		}
	}

	return nil
}

func (uc *ManageStockUseCase) UpdateArticleStock(
	ctx context.Context,
	articleID primitive.ObjectID,
	quantity, reserved float64,
) error {
	return uc.articleRepo.UpdateStock(ctx, articleID, quantity, reserved)
}

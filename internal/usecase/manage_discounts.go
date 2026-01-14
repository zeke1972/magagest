// internal/usecase/manage_discounts.go

package usecase

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"ricambi-manager/internal/domain"
	"ricambi-manager/internal/repository"
)

type ManageDiscountsUseCase struct {
	customerRepo  *repository.CustomerRepository
	articleRepo   *repository.ArticleRepository
	promotionRepo *repository.PromotionRepository
}

func NewManageDiscountsUseCase(
	customerRepo *repository.CustomerRepository,
	articleRepo *repository.ArticleRepository,
	promotionRepo *repository.PromotionRepository,
) *ManageDiscountsUseCase {
	return &ManageDiscountsUseCase{
		customerRepo:  customerRepo,
		articleRepo:   articleRepo,
		promotionRepo: promotionRepo,
	}
}

type DiscountCalculation struct {
	BasePrice         float64
	CustomerDiscount  float64
	PromotionDiscount float64
	NetPrice          float64
	FinalPrice        float64
	TotalDiscount     float64
	DiscountPercent   float64
	AppliedRule       *domain.DiscountRule
	AppliedPromotion  *domain.Promotion
}

func (uc *ManageDiscountsUseCase) CalculateFinalPrice(
	ctx context.Context,
	customer *domain.Customer,
	article *domain.Article,
	quantity float64,
) (*DiscountCalculation, error) {
	calc := &DiscountCalculation{
		BasePrice: article.Pricing.ListPrice,
	}

	netPrice := article.GetNetPrice(customer.ID)
	if netPrice != nil {
		calc.NetPrice = netPrice.Price
		calc.BasePrice = netPrice.Price
	}

	activePromotions, err := uc.promotionRepo.FindActive(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	var bestPromotion *domain.Promotion
	var bestPromotionDiscount float64

	for _, promo := range activePromotions {
		if !promo.IsApplicableToArticle(article) {
			continue
		}
		if !promo.IsApplicableToCustomer(customer) {
			continue
		}

		canUse, _ := promo.CanBeUsed(customer.ID.Hex(), quantity, calc.BasePrice*quantity)
		if !canUse {
			continue
		}

		discount := promo.CalculateDiscount(calc.BasePrice, quantity)
		if discount > bestPromotionDiscount {
			bestPromotionDiscount = discount
			bestPromotion = promo
		}
	}

	discountRule := customer.GetApplicableDiscount(article, quantity)
	if discountRule != nil {
		calc.AppliedRule = discountRule
		calc.CustomerDiscount = calc.BasePrice * (discountRule.DiscountPercent / 100)
	}

	if bestPromotion != nil {
		calc.AppliedPromotion = bestPromotion
		calc.PromotionDiscount = bestPromotionDiscount / quantity
	}

	if calc.CustomerDiscount > 0 && calc.PromotionDiscount > 0 {
		if calc.PromotionDiscount > calc.CustomerDiscount {
			calc.CustomerDiscount = 0
			calc.AppliedRule = nil
		} else {
			calc.PromotionDiscount = 0
			calc.AppliedPromotion = nil
		}
	}

	calc.FinalPrice = calc.BasePrice - calc.CustomerDiscount - calc.PromotionDiscount
	calc.TotalDiscount = calc.CustomerDiscount + calc.PromotionDiscount

	if calc.BasePrice > 0 {
		calc.DiscountPercent = (calc.TotalDiscount / calc.BasePrice) * 100
	}

	return calc, nil
}

func (uc *ManageDiscountsUseCase) AddCustomerDiscountRule(
	ctx context.Context,
	customerID primitive.ObjectID,
	rule domain.DiscountRule,
) error {
	customer, err := uc.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return err
	}

	if err := customer.AddDiscountRule(rule); err != nil {
		return err
	}

	return uc.customerRepo.Update(ctx, customer)
}

func (uc *ManageDiscountsUseCase) RemoveCustomerDiscountRule(
	ctx context.Context,
	customerID, ruleID primitive.ObjectID,
) error {
	customer, err := uc.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return err
	}

	customer.RemoveDiscountRule(ruleID)

	return uc.customerRepo.Update(ctx, customer)
}

func (uc *ManageDiscountsUseCase) AddNetPriceToArticle(
	ctx context.Context,
	articleID, customerID primitive.ObjectID,
	netPrice domain.NetPrice,
) error {
	article, err := uc.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	netPrice.CustomerID = customerID
	article.AddNetPrice(netPrice)

	return uc.articleRepo.Update(ctx, article)
}

func (uc *ManageDiscountsUseCase) GetExpiredNetPrices(ctx context.Context) ([]*domain.Article, error) {
	return uc.articleRepo.FindWithExpiredNetPrices(ctx, time.Now())
}

func (uc *ManageDiscountsUseCase) ValidateDiscount(
	ctx context.Context,
	operator *domain.Operator,
	discountPercent float64,
) error {
	if discountPercent < 0 || discountPercent > 100 {
		return errors.New("invalid discount percentage")
	}

	if discountPercent > 20 && operator.Profile != domain.ProfileAdmin {
		return errors.New("discount exceeds authorized limit")
	}

	return nil
}

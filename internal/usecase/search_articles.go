// internal/usecase/search_articles.go

package usecase

import (
	"context"
	"strings"

	"ricambi-manager/internal/domain"
	"ricambi-manager/internal/repository"
)

type SearchArticlesUseCase struct {
	articleRepo *repository.ArticleRepository
}

func NewSearchArticlesUseCase(articleRepo *repository.ArticleRepository) *SearchArticlesUseCase {
	return &SearchArticlesUseCase{
		articleRepo: articleRepo,
	}
}

type SearchResult struct {
	Articles     []*domain.Article
	TotalResults int
	SearchType   string
	Query        string
	Highlights   map[string][]string
}

func (uc *SearchArticlesUseCase) SearchByCode(ctx context.Context, code string, limit int) (*SearchResult, error) {
	articles, err := uc.articleRepo.SearchByCode(ctx, code, limit)
	if err != nil {
		return nil, err
	}

	result := &SearchResult{
		Articles:     articles,
		TotalResults: len(articles),
		SearchType:   "code",
		Query:        code,
		Highlights:   make(map[string][]string),
	}

	for _, article := range articles {
		result.Highlights[article.Code] = []string{article.Code}
	}

	return result, nil
}

func (uc *SearchArticlesUseCase) SearchByDescription(ctx context.Context, description string, limit int) (*SearchResult, error) {
	articles, err := uc.articleRepo.SearchByDescription(ctx, description, limit)
	if err != nil {
		return nil, err
	}

	result := &SearchResult{
		Articles:     articles,
		TotalResults: len(articles),
		SearchType:   "description",
		Query:        description,
		Highlights:   make(map[string][]string),
	}

	return result, nil
}

func (uc *SearchArticlesUseCase) SearchByBarcode(ctx context.Context, barcode string) (*domain.Article, error) {
	article, err := uc.articleRepo.FindByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}

	if article.ReplacedBy != "" {
		replacementArticle, err := uc.articleRepo.FindByCode(ctx, article.ReplacedBy)
		if err == nil {
			return replacementArticle, nil
		}
	}

	return article, nil
}

func (uc *SearchArticlesUseCase) SearchByApplicability(ctx context.Context, vehicleMake, model string, year int, limit int) (*SearchResult, error) {
	articles, err := uc.articleRepo.FindByApplicability(ctx, vehicleMake, model, year, limit)
	if err != nil {
		return nil, err
	}

	result := &SearchResult{
		Articles:     articles,
		TotalResults: len(articles),
		SearchType:   "applicability",
		Query:        vehicleMake + " " + model,
		Highlights:   make(map[string][]string),
	}

	return result, nil
}

func (uc *SearchArticlesUseCase) FuzzySearch(ctx context.Context, query string, limit int) (*SearchResult, error) {
	articles, err := uc.articleRepo.SearchFuzzy(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	scoredArticles := uc.rankArticles(articles, query)

	result := &SearchResult{
		Articles:     scoredArticles,
		TotalResults: len(scoredArticles),
		SearchType:   "fuzzy",
		Query:        query,
		Highlights:   make(map[string][]string),
	}

	return result, nil
}

func (uc *SearchArticlesUseCase) rankArticles(articles []*domain.Article, query string) []*domain.Article {
	query = strings.ToLower(query)

	type scoredArticle struct {
		article *domain.Article
		score   int
	}

	scored := make([]scoredArticle, len(articles))

	for i, article := range articles {
		score := 0

		if strings.Contains(strings.ToLower(article.Code), query) {
			score += 100
		}
		if strings.HasPrefix(strings.ToLower(article.Code), query) {
			score += 50
		}
		if strings.Contains(strings.ToLower(article.Description), query) {
			score += 30
		}
		if strings.Contains(strings.ToLower(article.Brand), query) {
			score += 20
		}

		scored[i] = scoredArticle{article: article, score: score}
	}

	// Bubble sort per ordinare per score decrescente
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	result := make([]*domain.Article, len(scored))
	for i, s := range scored {
		result[i] = s.article
	}

	return result
}

func (uc *SearchArticlesUseCase) SearchWithReplacement(ctx context.Context, code string) (*domain.Article, error) {
	article, err := uc.articleRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if article.ReplacedBy != "" {
		replacementArticle, err := uc.articleRepo.FindByCode(ctx, article.ReplacedBy)
		if err == nil {
			return replacementArticle, nil
		}
	}

	return article, nil
}

func (uc *SearchArticlesUseCase) GetReplacementChain(ctx context.Context, code string) ([]*domain.Article, error) {
	return uc.articleRepo.FindReplacementChain(ctx, code)
}

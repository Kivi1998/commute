package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/pkg/amap"
	"github.com/haojia/commute/internal/pkg/doubao"
	"github.com/haojia/commute/internal/repository"
)

const (
	aiCacheTTL       = 24 * time.Hour
	poiEnrichConcurr = 5
)

type AIService struct {
	repo   *repository.AIRepo
	doubao *doubao.Client
	amap   *amap.Client
}

func NewAIService(repo *repository.AIRepo, doubaoClient *doubao.Client, amapClient *amap.Client) *AIService {
	return &AIService{repo: repo, doubao: doubaoClient, amap: amapClient}
}

// RecommendCompanies 主入口：缓存 → 豆包 → POI 二次校验 → 写缓存
//
// ExcludeNames 不参与 cache_key，但出现时强制绕开缓存（每次"再换一批"都是新鲜数据）
func (s *AIService) RecommendCompanies(ctx context.Context, userID int64, in model.AIRecommendInput) (*model.AIRecommendResult, error) {
	hasExcludes := len(in.ExcludeNames) > 0
	cacheKey := repository.BuildCacheKey(userID, in)

	// 1. 查缓存（有 exclude 则跳过缓存，保证能换新）
	if !in.ForceRefresh && !hasExcludes {
		if cached, err := s.repo.FindCache(ctx, cacheKey); err == nil {
			return buildResultFromCache(cached)
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}

	// 2. 调豆包
	if !s.doubao.Configured() {
		return nil, doubao.ErrNotConfigured
	}
	raw, err := s.doubao.RecommendCompanies(ctx, doubao.RecommendCompaniesInput{
		City:            in.City,
		Position:        in.Position,
		ExperienceYears: in.ExperienceYears,
		CompanyTypes:    in.CompanyTypes,
		Count:           in.Count,
		ExcludeNames:    in.ExcludeNames,
	})
	if err != nil {
		return nil, err
	}

	// 3. 对每条做 POI 二次校验（并发 5）
	companies := make([]model.AIRecommendedCompany, len(raw.Companies))
	var wg sync.WaitGroup
	sem := make(chan struct{}, poiEnrichConcurr)

	for i, c := range raw.Companies {
		companies[i] = model.AIRecommendedCompany{
			Name:        c.Name,
			Category:    c.Category,
			Industry:    c.Industry,
			AddressHint: c.AddressHint,
			Reason:      c.Reason,
		}
		wg.Add(1)
		go func(idx int, src doubao.RecommendedCompany) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			enrichWithPOI(ctx, s.amap, &companies[idx], src, in.City)
		}(i, c)
	}
	wg.Wait()

	// 4. 写缓存（exclude 场景不写缓存，因为是"再换一批"的临时结果）
	if !hasExcludes {
		if cached, err := s.repo.InsertCache(ctx, userID, in, cacheKey, companies,
			raw.Usage.PromptTokens, raw.Usage.CompletionTokens, aiCacheTTL); err == nil {
			return &model.AIRecommendResult{
				FromCache:   false,
				CachedAt:    &cached.RequestedAt,
				ExpiresAt:   &cached.ExpiresAt,
				Companies:   companies,
				TokenInput:  raw.Usage.PromptTokens,
				TokenOutput: raw.Usage.CompletionTokens,
			}, nil
		} else {
			return nil, err
		}
	}
	return &model.AIRecommendResult{
		FromCache:   false,
		Companies:   companies,
		TokenInput:  raw.Usage.PromptTokens,
		TokenOutput: raw.Usage.CompletionTokens,
	}, nil
}

// enrichWithPOI 调用高德 POI 搜索匹配精确坐标
func enrichWithPOI(ctx context.Context, client *amap.Client, dst *model.AIRecommendedCompany, src doubao.RecommendedCompany, region string) {
	// 第一轮：公司名精确搜索
	items, err := client.POISearch(ctx, src.Name, region, 3)
	if err != nil || len(items) == 0 {
		// 第二轮：公司名 + 地址提示
		if src.AddressHint != "" {
			items, _ = client.POISearch(ctx, src.Name+" "+src.AddressHint, region, 3)
		}
	}
	if len(items) == 0 {
		dst.LocationConfident = false
		return
	}

	p := items[0]
	dst.ResolvedAddress = &p.Address
	dst.ResolvedLongitude = &p.Longitude
	dst.ResolvedLatitude = &p.Latitude
	if p.Province != "" {
		dst.ResolvedProvince = &p.Province
	}
	if p.City != "" {
		dst.ResolvedCity = &p.City
	}
	if p.District != "" {
		dst.ResolvedDistrict = &p.District
	}
	dst.LocationConfident = p.Longitude != 0 && p.Latitude != 0
}

func buildResultFromCache(c *repository.CachedRecommendation) (*model.AIRecommendResult, error) {
	var wrapper struct {
		Companies []model.AIRecommendedCompany `json:"companies"`
	}
	if err := json.Unmarshal(c.RawResponse, &wrapper); err != nil {
		return nil, err
	}
	return &model.AIRecommendResult{
		FromCache:   true,
		CachedAt:    &c.RequestedAt,
		ExpiresAt:   &c.ExpiresAt,
		Companies:   wrapper.Companies,
		TokenInput:  c.TokenInput,
		TokenOutput: c.TokenOutput,
	}, nil
}

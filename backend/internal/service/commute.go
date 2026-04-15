package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/pkg/amap"
	"github.com/haojia/commute/internal/pkg/citydict"
	"github.com/haojia/commute/internal/repository"
)

const (
	commuteCacheTTL       = 7 * 24 * time.Hour
	commuteMaxConcurrency = 5
	companySoftLimit      = 20
)

type CommuteService struct {
	repo        *repository.CommuteRepo
	addressRepo *repository.AddressRepo
	companyRepo *repository.CompanyRepo
	amap        *amap.Client
}

func NewCommuteService(
	repo *repository.CommuteRepo,
	addressRepo *repository.AddressRepo,
	companyRepo *repository.CompanyRepo,
	amap *amap.Client,
) *CommuteService {
	return &CommuteService{
		repo: repo, addressRepo: addressRepo, companyRepo: companyRepo, amap: amap,
	}
}

// Calculate 主入口：批量计算并返回响应
func (s *CommuteService) Calculate(ctx context.Context, userID int64, in model.CommuteCalculateInput) (*model.CommuteCalculateResponse, error) {
	home, err := s.addressRepo.Get(ctx, userID, in.HomeID)
	if err != nil {
		return nil, fmt.Errorf("home: %w", err)
	}

	companies := make([]*model.Company, 0, len(in.CompanyIDs))
	for _, cid := range in.CompanyIDs {
		c, err := s.companyRepo.Get(ctx, userID, cid)
		if err != nil {
			return nil, fmt.Errorf("company %d: %w", cid, err)
		}
		companies = append(companies, c)
	}

	buffer := in.BufferMinutes
	if buffer == 0 {
		buffer = 5
	}
	weekday := in.Weekday
	if weekday == 0 {
		weekday = 1
	}

	// 可选保存查询会话
	var queryID *int64
	if in.SaveQuery {
		id, err := s.repo.CreateQuery(ctx, userID, in)
		if err != nil {
			return nil, fmt.Errorf("save query: %w", err)
		}
		queryID = &id
	}

	// 构造所有 (company × direction × mode) 任务
	type task struct {
		company       *model.Company
		direction     string
		transportMode string
		departTime    string // HH:MM
	}
	tasks := make([]task, 0)
	for _, c := range companies {
		for _, mode := range in.TransportModes {
			tasks = append(tasks, task{c, "to_work", mode, in.Morning.Time})
			tasks = append(tasks, task{c, "to_home", mode, in.Evening.Time})
		}
	}

	// 结果聚合
	type output struct {
		companyID int64
		item      *model.CommuteResultItem
		err       *model.CommuteCalcError
		cacheHit  bool
	}
	outCh := make(chan output, len(tasks))

	sem := make(chan struct{}, commuteMaxConcurrency)
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 起终点经纬度（direction 决定方向）
			var originLng, originLat, destLng, destLat float64
			if t.direction == "to_work" {
				originLng, originLat = home.Longitude, home.Latitude
				destLng, destLat = t.company.Longitude, t.company.Latitude
			} else {
				originLng, originLat = t.company.Longitude, t.company.Latitude
				destLng, destLat = home.Longitude, home.Latitude
			}

			item, hit, err := s.computeOne(ctx, userID, queryID, home, t.company,
				t.direction, t.transportMode, t.departTime,
				weekday, buffer,
				originLng, originLat, destLng, destLat,
				in.ForceRefresh,
			)
			if err != nil {
				outCh <- output{
					companyID: t.company.ID,
					err: &model.CommuteCalcError{
						Direction: t.direction, TransportMode: t.transportMode, Message: err.Error(),
					},
				}
				return
			}
			outCh <- output{companyID: t.company.ID, item: item, cacheHit: hit}
		}(t)
	}

	wg.Wait()
	close(outCh)

	// 按公司聚合
	byCompany := map[int64]*model.CompanyCommute{}
	for _, c := range companies {
		byCompany[c.ID] = &model.CompanyCommute{
			CompanyID: c.ID, CompanyName: c.Name,
			CompanyLongitude: c.Longitude, CompanyLatitude: c.Latitude,
			Items: []model.CommuteResultItem{}, Errors: []model.CommuteCalcError{},
		}
	}
	summary := model.CommuteSummary{TotalCompanies: len(companies)}
	for o := range outCh {
		summary.TotalCalculations++
		cc := byCompany[o.companyID]
		if o.err != nil {
			summary.Failures++
			cc.Errors = append(cc.Errors, *o.err)
			continue
		}
		if o.cacheHit {
			summary.CacheHits++
		}
		cc.Items = append(cc.Items, *o.item)
	}

	results := make([]model.CompanyCommute, 0, len(companies))
	for _, c := range companies {
		results = append(results, *byCompany[c.ID])
	}

	return &model.CommuteCalculateResponse{
		QueryID: queryID, Home: home, Weekday: weekday, BufferMinutes: buffer,
		Results: results, Summary: summary,
	}, nil
}

// computeOne 单条计算。先查缓存，未命中调高德，写入 result。
func (s *CommuteService) computeOne(
	ctx context.Context,
	userID int64, queryID *int64,
	home *model.HomeAddress, company *model.Company,
	direction, mode, departTimeHM string,
	weekday, buffer int,
	originLng, originLat, destLng, destLat float64,
	forceRefresh bool,
) (*model.CommuteResultItem, bool, error) {
	key := repository.CommuteResultKey{
		HomeID: home.ID, CompanyID: company.ID,
		TransportMode: mode, Direction: direction,
		DepartTime: departTimeHM, Weekday: weekday,
	}

	if !forceRefresh {
		if cached, err := s.repo.FindCachedResult(ctx, key); err == nil {
			return toItem(cached, true), true, nil
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, false, err
		}
	}

	// 构造下周一（或指定 weekday）的 depart time
	departureTime, err := nextWeekdayTime(time.Now(), weekday, departTimeHM)
	if err != nil {
		return nil, false, fmt.Errorf("parse time: %w", err)
	}

	// 公交需要 citycode：优先家所在城市，回退公司所在城市
	cityCode := ""
	if mode == "transit" {
		if home.City != nil && *home.City != "" {
			cityCode = citydict.Lookup(*home.City)
		}
		if cityCode == "" && company.City != nil && *company.City != "" {
			cityCode = citydict.Lookup(*company.City)
		}
	}

	dir, err := s.amap.DirectionByMode(ctx, mode, amap.DirectionOptions{
		OriginLng: originLng, OriginLat: originLat,
		DestLng: destLng, DestLat: destLat,
		DepartureTime: departureTime,
		CityCode:      cityCode,
	})
	if err != nil {
		return nil, false, err
	}

	durRawMin := (dir.DurationSec + 59) / 60
	durMin := durRawMin + buffer
	distKm := float64(dir.DistanceMeter) / 1000.0

	// 计算到达时间
	arriveAt := departureTime.Add(time.Duration(durMin) * time.Minute)

	inserted, err := s.repo.InsertResult(ctx, repository.CommuteResultInsert{
		UserID: userID, QueryID: queryID,
		HomeID: home.ID, CompanyID: company.ID,
		Direction: direction, TransportMode: mode,
		DepartTime: departTimeHM, ArriveTime: arriveAt.Format("15:04"),
		Weekday:        weekday,
		DurationMin:    durMin,
		DurationMinRaw: durRawMin,
		DistanceKM:     round2(distKm),
		CostYuan:       dir.CostYuan,
		TransferCount:  dir.TransferCount,
		Polyline:       dir.Polyline,
		RouteDetail:    dir.RouteDetail,
		ExpiresAt:      time.Now().Add(commuteCacheTTL),
	})
	if err != nil {
		return nil, false, fmt.Errorf("insert result: %w", err)
	}

	return toItem(inserted, false), false, nil
}

func toItem(r *model.CommuteResult, fromCache bool) *model.CommuteResultItem {
	_ = json.RawMessage(nil)
	return &model.CommuteResultItem{
		Direction:      r.Direction,
		TransportMode:  r.TransportMode,
		DepartTime:     r.DepartTime,
		ArriveTime:     r.ArriveTime,
		DurationMin:    r.DurationMin,
		DurationMinRaw: r.DurationMinRaw,
		DistanceKM:     r.DistanceKM,
		CostYuan:       r.CostYuan,
		TransferCount:  r.TransferCount,
		Polyline:       r.Polyline,
		FromCache:      fromCache,
		ResultID:       r.ID,
	}
}

// GetResultDetail 详情接口
func (s *CommuteService) GetResultDetail(ctx context.Context, userID, id int64) (*model.CommuteResult, error) {
	return s.repo.GetResult(ctx, userID, id)
}

func (s *CommuteService) ListQueryResults(ctx context.Context, userID, queryID int64) ([]model.CommuteResult, error) {
	return s.repo.ListResultsByQuery(ctx, userID, queryID)
}

func (s *CommuteService) ListQueries(ctx context.Context, userID int64, limit int) ([]model.CommuteQueryListItem, error) {
	return s.repo.ListQueries(ctx, userID, limit)
}

func (s *CommuteService) DeleteQuery(ctx context.Context, userID, id int64) error {
	return s.repo.DeleteQuery(ctx, userID, id)
}

// GetQueryDetail 查询会话详情（用于恢复配置）
func (s *CommuteService) GetQueryDetail(ctx context.Context, userID, id int64) (*model.CommuteQuery, error) {
	return s.repo.GetQuery(ctx, userID, id)
}

// nextWeekdayTime 返回"下一个指定星期几的 HH:MM"。用于高德 departure_time。
// weekday: ISO 1=Mon..7=Sun
func nextWeekdayTime(now time.Time, weekday int, hm string) (time.Time, error) {
	t, err := time.Parse("15:04", hm)
	if err != nil {
		return time.Time{}, err
	}
	// Go 的 Weekday 是 0=Sun..6=Sat；ISO 是 1=Mon..7=Sun
	goTarget := time.Weekday(weekday % 7)
	days := (int(goTarget) - int(now.Weekday()) + 7) % 7
	if days == 0 {
		// 若今天是目标日且当前时间已过目标时间，用下周；否则今天。
		todayAt := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if now.After(todayAt) {
			days = 7
		}
	}
	target := now.AddDate(0, 0, days)
	return time.Date(target.Year(), target.Month(), target.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100.0
}

package amap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// DirectionResult 统一归一化后的路径规划结果
type DirectionResult struct {
	DurationSec    int             `json:"duration_sec"`
	DistanceMeter  int             `json:"distance_meter"`
	CostYuan       *float64        `json:"cost_yuan,omitempty"`
	TransferCount  *int            `json:"transfer_count,omitempty"`
	RouteDetail    json.RawMessage `json:"route_detail"`
}

// DirectionOptions 路径规划参数
type DirectionOptions struct {
	OriginLng      float64
	OriginLat      float64
	DestLng        float64
	DestLat        float64
	DepartureTime  time.Time // 用于驾车路况估算，公交 date/time
	CityCode       string    // 公交必需（city1=city2=同城默认）
}

// --- 驾车 ---

type drivingResp struct {
	Route struct {
		Paths []struct {
			Distance string `json:"distance"`
			Cost     struct {
				Duration     string `json:"duration"`
				Tolls        string `json:"tolls"`
				TollDistance string `json:"toll_distance"`
				TaxiFee      string `json:"taxi_fee"`
				TrafficLights string `json:"traffic_lights"`
			} `json:"cost"`
		} `json:"paths"`
	} `json:"route"`
}

func (c *Client) Driving(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("strategy", "32")
	params.Set("show_fields", "cost")
	if !opt.DepartureTime.IsZero() {
		params.Set("departure_time", strconv.FormatInt(opt.DepartureTime.Unix(), 10))
	}

	var r drivingResp
	if err := c.doGet(ctx, "/v5/direction/driving", params, &r); err != nil {
		return nil, err
	}
	if len(r.Route.Paths) == 0 {
		return nil, fmt.Errorf("amap driving: no paths")
	}
	p := r.Route.Paths[0]
	dur, _ := strconv.Atoi(p.Cost.Duration)
	dist, _ := strconv.Atoi(p.Distance)
	taxi, _ := strconv.ParseFloat(p.Cost.TaxiFee, 64)

	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		CostYuan:      &taxi,
		RouteDetail:   raw,
	}, nil
}

// --- 公交 ---

type transitResp struct {
	Route struct {
		Transits []struct {
			Distance        string `json:"distance"`
			WalkingDistance string `json:"walking_distance"`
			Cost            struct {
				Duration   string `json:"duration"`
				TransitFee string `json:"transit_fee"`
			} `json:"cost"`
			Segments []struct {
				Walking json.RawMessage `json:"walking"`
				Bus     *struct {
					Buslines []struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"buslines"`
				} `json:"bus"`
			} `json:"segments"`
		} `json:"transits"`
	} `json:"route"`
}

func (c *Client) Transit(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	if opt.CityCode == "" {
		return nil, fmt.Errorf("amap transit: city code required")
	}
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("city1", opt.CityCode)
	params.Set("city2", opt.CityCode)
	params.Set("strategy", "0")
	params.Set("show_fields", "cost,navi")
	params.Set("AlternativeRoute", "1")
	if !opt.DepartureTime.IsZero() {
		params.Set("date", opt.DepartureTime.Format("2006-01-02"))
		params.Set("time", opt.DepartureTime.Format("15:04"))
	}

	var r transitResp
	if err := c.doGet(ctx, "/v5/direction/transit/integrated", params, &r); err != nil {
		return nil, err
	}
	if len(r.Route.Transits) == 0 {
		return nil, fmt.Errorf("amap transit: no routes")
	}
	t := r.Route.Transits[0]
	dur, _ := strconv.Atoi(t.Cost.Duration)
	dist, _ := strconv.Atoi(t.Distance)
	fee, _ := strconv.ParseFloat(t.Cost.TransitFee, 64)

	// 换乘次数 = 有 buslines 的 segment 数量
	transfers := 0
	for _, seg := range t.Segments {
		if seg.Bus != nil && len(seg.Bus.Buslines) > 0 {
			transfers += len(seg.Bus.Buslines)
		}
	}
	// 换乘次数减 1 作为"换乘次数"更合理（第一次乘车不算换乘）
	if transfers > 0 {
		transfers -= 1
	}

	raw, _ := json.Marshal(t)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		CostYuan:      &fee,
		TransferCount: &transfers,
		RouteDetail:   raw,
	}, nil
}

// --- 骑行 ---

type bicyclingResp struct {
	Data struct {
		Paths []struct {
			Distance int `json:"distance"`
			Duration int `json:"duration"`
		} `json:"paths"`
	} `json:"data"`
}

func (c *Client) Bicycling(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))

	var r bicyclingResp
	if err := c.doGet(ctx, "/v5/direction/bicycling", params, &r); err != nil {
		return nil, err
	}
	if len(r.Data.Paths) == 0 {
		return nil, fmt.Errorf("amap bicycling: no paths")
	}
	p := r.Data.Paths[0]
	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   p.Duration,
		DistanceMeter: p.Distance,
		RouteDetail:   raw,
	}, nil
}

// --- 步行 ---

type walkingResp struct {
	Route struct {
		Paths []struct {
			Distance string `json:"distance"`
			Cost     struct {
				Duration string `json:"duration"`
			} `json:"cost"`
		} `json:"paths"`
	} `json:"route"`
}

func (c *Client) Walking(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("show_fields", "cost")

	var r walkingResp
	if err := c.doGet(ctx, "/v5/direction/walking", params, &r); err != nil {
		return nil, err
	}
	if len(r.Route.Paths) == 0 {
		return nil, fmt.Errorf("amap walking: no paths")
	}
	p := r.Route.Paths[0]
	dur, _ := strconv.Atoi(p.Cost.Duration)
	dist, _ := strconv.Atoi(p.Distance)
	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		RouteDetail:   raw,
	}, nil
}

// DirectionByMode 根据出行方式分派
func (c *Client) DirectionByMode(ctx context.Context, mode string, opt DirectionOptions) (*DirectionResult, error) {
	switch mode {
	case "driving":
		return c.Driving(ctx, opt)
	case "transit":
		return c.Transit(ctx, opt)
	case "cycling":
		return c.Bicycling(ctx, opt)
	case "walking":
		return c.Walking(ctx, opt)
	default:
		return nil, fmt.Errorf("amap: unsupported mode %q", mode)
	}
}

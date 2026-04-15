package amap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DirectionResult 统一归一化后的路径规划结果
type DirectionResult struct {
	DurationSec   int             `json:"duration_sec"`
	DistanceMeter int             `json:"distance_meter"`
	CostYuan      *float64        `json:"cost_yuan,omitempty"`
	TransferCount *int            `json:"transfer_count,omitempty"`
	Polyline      string          `json:"polyline"` // 拼接后的路线点串 "lng,lat;lng,lat;..."
	RouteDetail   json.RawMessage `json:"route_detail"`
}

// DirectionOptions 路径规划参数
type DirectionOptions struct {
	OriginLng     float64
	OriginLat     float64
	DestLng       float64
	DestLat       float64
	DepartureTime time.Time // 用于驾车路况估算，公交 date/time
	CityCode      string    // 公交必需（city1=city2=同城默认）
}

// joinPolylines 把多段 "lng,lat;lng,lat" 合并为一条（相邻段去重首点）
func joinPolylines(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	if len(nonEmpty) == 1 {
		return nonEmpty[0]
	}
	// 简单拼接，前端绘制时可容忍极小重复点
	return strings.Join(nonEmpty, ";")
}

// --- 驾车 ---

type drivingStep struct {
	Instruction string `json:"instruction"`
	Polyline    string `json:"polyline"`
}

type drivingResp struct {
	Route struct {
		Paths []struct {
			Distance string `json:"distance"`
			Cost     struct {
				Duration      string `json:"duration"`
				Tolls         string `json:"tolls"`
				TollDistance  string `json:"toll_distance"`
				TaxiFee       string `json:"taxi_fee"`
				TrafficLights string `json:"traffic_lights"`
			} `json:"cost"`
			Steps []drivingStep `json:"steps"`
		} `json:"paths"`
	} `json:"route"`
}

func (c *Client) Driving(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("strategy", "32")
	params.Set("show_fields", "cost,polyline")
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

	polys := make([]string, 0, len(p.Steps))
	for _, s := range p.Steps {
		polys = append(polys, s.Polyline)
	}

	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		CostYuan:      &taxi,
		Polyline:      joinPolylines(polys...),
		RouteDetail:   raw,
	}, nil
}

// --- 公交 ---

type transitBusline struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Polyline json.RawMessage `json:"polyline"`
}

// transitWalkingStep polyline 字段高德可能返回字符串或 { polyline: "..." } 对象，
// 用 json.RawMessage 接住再兜底解析。
type transitWalkingStep struct {
	Polyline json.RawMessage `json:"polyline"`
}

type transitSegment struct {
	Walking *struct {
		Polyline json.RawMessage     `json:"polyline"`
		Steps    []transitWalkingStep `json:"steps"`
	} `json:"walking"`
	Bus *struct {
		Buslines []transitBusline `json:"buslines"`
	} `json:"bus"`
}

// flexString 尝试把 json.RawMessage 解析成 polyline 字符串：
// 1) 直接是 "lng,lat;..." 字符串 → 取出
// 2) 是 { "polyline": "..." } 对象 → 取 polyline 字段
func flexString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj struct {
		Polyline string `json:"polyline"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj.Polyline
	}
	return ""
}

type transitResp struct {
	Route struct {
		Transits []struct {
			Distance        string `json:"distance"`
			WalkingDistance string `json:"walking_distance"`
			Cost            struct {
				Duration   string `json:"duration"`
				TransitFee string `json:"transit_fee"`
			} `json:"cost"`
			Segments []transitSegment `json:"segments"`
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
	params.Set("show_fields", "cost,navi,polyline")
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

	transfers := 0
	polys := make([]string, 0, len(t.Segments)*2)
	for _, seg := range t.Segments {
		if seg.Walking != nil {
			wk := flexString(seg.Walking.Polyline)
			if wk != "" {
				polys = append(polys, wk)
			} else {
				for _, ws := range seg.Walking.Steps {
					if s := flexString(ws.Polyline); s != "" {
						polys = append(polys, s)
					}
				}
			}
		}
		if seg.Bus != nil && len(seg.Bus.Buslines) > 0 {
			transfers += len(seg.Bus.Buslines)
			for _, bl := range seg.Bus.Buslines {
				if s := flexString(bl.Polyline); s != "" {
					polys = append(polys, s)
				}
			}
		}
	}
	if transfers > 0 {
		transfers -= 1
	}

	raw, _ := json.Marshal(t)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		CostYuan:      &fee,
		TransferCount: &transfers,
		Polyline:      joinPolylines(polys...),
		RouteDetail:   raw,
	}, nil
}

// --- 骑行 ---

type bicyclingStep struct {
	Polyline string `json:"polyline"`
}

type bicyclingResp struct {
	Data struct {
		Paths []struct {
			Distance int             `json:"distance"`
			Duration int             `json:"duration"`
			Steps    []bicyclingStep `json:"steps"`
		} `json:"paths"`
	} `json:"data"`
}

func (c *Client) Bicycling(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("show_fields", "polyline")

	var r bicyclingResp
	if err := c.doGet(ctx, "/v5/direction/bicycling", params, &r); err != nil {
		return nil, err
	}
	if len(r.Data.Paths) == 0 {
		return nil, fmt.Errorf("amap bicycling: no paths")
	}
	p := r.Data.Paths[0]

	polys := make([]string, 0, len(p.Steps))
	for _, s := range p.Steps {
		polys = append(polys, s.Polyline)
	}

	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   p.Duration,
		DistanceMeter: p.Distance,
		Polyline:      joinPolylines(polys...),
		RouteDetail:   raw,
	}, nil
}

// --- 步行 ---

type walkingStep struct {
	Polyline string `json:"polyline"`
}

type walkingResp struct {
	Route struct {
		Paths []struct {
			Distance string `json:"distance"`
			Cost     struct {
				Duration string `json:"duration"`
			} `json:"cost"`
			Steps []walkingStep `json:"steps"`
		} `json:"paths"`
	} `json:"route"`
}

func (c *Client) Walking(ctx context.Context, opt DirectionOptions) (*DirectionResult, error) {
	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", opt.OriginLng, opt.OriginLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", opt.DestLng, opt.DestLat))
	params.Set("show_fields", "cost,polyline")

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

	polys := make([]string, 0, len(p.Steps))
	for _, s := range p.Steps {
		polys = append(polys, s.Polyline)
	}

	raw, _ := json.Marshal(p)
	return &DirectionResult{
		DurationSec:   dur,
		DistanceMeter: dist,
		Polyline:      joinPolylines(polys...),
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

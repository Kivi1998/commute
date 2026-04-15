package amap

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// --- 地理编码 ---

type GeocodeItem struct {
	FormattedAddress string  `json:"formatted_address"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	District         string  `json:"district"`
	Longitude        float64 `json:"longitude"`
	Latitude         float64 `json:"latitude"`
	Level            string  `json:"level"`
}

type geocodeResp struct {
	Geocodes []struct {
		FormattedAddress string `json:"formatted_address"`
		Province         string `json:"province"`
		City             any    `json:"city"` // 有时返回 []
		District         string `json:"district"`
		Location         string `json:"location"`
		Level            string `json:"level"`
	} `json:"geocodes"`
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (c *Client) Geocode(ctx context.Context, address, city string) ([]GeocodeItem, error) {
	params := url.Values{}
	params.Set("address", address)
	if city != "" {
		params.Set("city", city)
	}

	var r geocodeResp
	if err := c.doGet(ctx, "/v3/geocode/geo", params, &r); err != nil {
		return nil, err
	}
	out := make([]GeocodeItem, 0, len(r.Geocodes))
	for _, g := range r.Geocodes {
		parts := strings.Split(g.Location, ",")
		if len(parts) != 2 {
			continue
		}
		lng, _ := strconv.ParseFloat(parts[0], 64)
		lat, _ := strconv.ParseFloat(parts[1], 64)
		out = append(out, GeocodeItem{
			FormattedAddress: g.FormattedAddress,
			Province:         g.Province,
			City:             asString(g.City),
			District:         g.District,
			Longitude:        lng,
			Latitude:         lat,
			Level:            g.Level,
		})
	}
	return out, nil
}

// --- 逆地理编码 ---

type RegeocodeResult struct {
	FormattedAddress string `json:"formatted_address"`
	Province         string `json:"province"`
	City             string `json:"city"`
	District         string `json:"district"`
	ADCode           string `json:"adcode"`
}

type regeocodeResp struct {
	Regeocode struct {
		FormattedAddress string `json:"formatted_address"`
		AddressComponent struct {
			Province string `json:"province"`
			City     any    `json:"city"`
			District string `json:"district"`
			ADCode   string `json:"adcode"`
		} `json:"addressComponent"`
	} `json:"regeocode"`
}

func (c *Client) Regeocode(ctx context.Context, lng, lat float64) (*RegeocodeResult, error) {
	params := url.Values{}
	params.Set("location", strconv.FormatFloat(lng, 'f', 6, 64)+","+strconv.FormatFloat(lat, 'f', 6, 64))

	var r regeocodeResp
	if err := c.doGet(ctx, "/v3/geocode/regeo", params, &r); err != nil {
		return nil, err
	}
	return &RegeocodeResult{
		FormattedAddress: r.Regeocode.FormattedAddress,
		Province:         r.Regeocode.AddressComponent.Province,
		City:             asString(r.Regeocode.AddressComponent.City),
		District:         r.Regeocode.AddressComponent.District,
		ADCode:           r.Regeocode.AddressComponent.ADCode,
	}, nil
}

// --- POI 搜索 ---

type POIItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Type      string  `json:"type"`
	Province  string  `json:"pname"`
	City      string  `json:"cityname"`
	District  string  `json:"adname"`
}

type poiResp struct {
	POIs []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Address  any    `json:"address"`
		Location string `json:"location"`
		Type     string `json:"type"`
		PName    any    `json:"pname"`
		CityName any    `json:"cityname"`
		ADName   any    `json:"adname"`
	} `json:"pois"`
}

func (c *Client) POISearch(ctx context.Context, keyword, region string, pageSize int) ([]POIItem, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	params := url.Values{}
	params.Set("keywords", keyword)
	if region != "" {
		params.Set("region", region)
	}
	params.Set("page_size", strconv.Itoa(pageSize))
	params.Set("page_num", "1")

	var r poiResp
	if err := c.doGet(ctx, "/v5/place/text", params, &r); err != nil {
		return nil, err
	}
	out := make([]POIItem, 0, len(r.POIs))
	for _, p := range r.POIs {
		parts := strings.Split(p.Location, ",")
		if len(parts) != 2 {
			continue
		}
		lng, _ := strconv.ParseFloat(parts[0], 64)
		lat, _ := strconv.ParseFloat(parts[1], 64)
		out = append(out, POIItem{
			ID: p.ID, Name: p.Name,
			Address:   asString(p.Address),
			Longitude: lng, Latitude: lat,
			Type:     p.Type,
			Province: asString(p.PName),
			City:     asString(p.CityName),
			District: asString(p.ADName),
		})
	}
	return out, nil
}

package model

type EnumItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

type Enums struct {
	CompanyType    []EnumItem `json:"company_type"`
	CompanyStatus  []EnumItem `json:"company_status"`
	CompanySource  []EnumItem `json:"company_source"`
	TransportMode  []EnumItem `json:"transport_mode"`
	TimeStrategy   []EnumItem `json:"time_strategy"`
	CommuteDirection []EnumItem `json:"commute_direction"`
}

var AllEnums = Enums{
	CompanyType: []EnumItem{
		{Value: "big_tech", Label: "大厂"},
		{Value: "mid_tech", Label: "中厂"},
		{Value: "startup", Label: "创业公司"},
		{Value: "foreign", Label: "外企"},
		{Value: "other", Label: "其他"},
	},
	CompanyStatus: []EnumItem{
		{Value: "watching", Label: "关注"},
		{Value: "applied", Label: "已投递"},
		{Value: "interviewing", Label: "面试中"},
		{Value: "offered", Label: "已 offer"},
		{Value: "rejected", Label: "已拒/被拒"},
		{Value: "archived", Label: "归档"},
	},
	CompanySource: []EnumItem{
		{Value: "ai_recommend", Label: "AI 推荐"},
		{Value: "manual", Label: "手动添加"},
	},
	TransportMode: []EnumItem{
		{Value: "transit", Label: "公交/地铁", Icon: "🚇"},
		{Value: "driving", Label: "驾车", Icon: "🚗"},
		{Value: "cycling", Label: "骑行", Icon: "🚴"},
		{Value: "walking", Label: "步行", Icon: "🚶"},
	},
	TimeStrategy: []EnumItem{
		{Value: "depart_at", Label: "指定出发时间"},
		{Value: "arrive_by", Label: "指定到达时间"},
	},
	CommuteDirection: []EnumItem{
		{Value: "to_work", Label: "去上班"},
		{Value: "to_home", Label: "回家"},
	},
}

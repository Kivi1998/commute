package doubao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// RecommendCompaniesInput 推荐参数
type RecommendCompaniesInput struct {
	City             string
	Position         string
	ExperienceYears  *int
	CompanyTypes     []string // big_tech / mid_tech / startup / foreign / other
	Count            int      // 建议 15-30
}

// RecommendedCompany AI 返回的单个公司
type RecommendedCompany struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Industry    string `json:"industry"`
	AddressHint string `json:"address_hint"`
	Reason      string `json:"reason"`
}

// RecommendCompaniesResult 推荐结果
type RecommendCompaniesResult struct {
	Companies []RecommendedCompany `json:"companies"`
	Usage     Usage                // 方便上层追踪 token
}

const systemPrompt = `你是一位专业的互联网行业分析师，熟悉国内各城市的科技公司分布。

请根据用户提供的城市和求职岗位，推荐本地真实存在的公司列表。要求：
1. 公司必须真实存在，且在目标城市有办公地点。
2. 均衡覆盖大厂、中厂、创业公司三个层级。
3. 每家公司给出真实办公地点的简要描述（区/街道级别，如"北京海淀区中关村"）。
4. 推荐理由需结合公司业务与目标岗位的契合度，控制在 50 字以内。
5. 严格输出 JSON 格式，不要包含任何解释性文字、Markdown 标记或代码块。
6. 禁止虚构不存在的公司名称。`

var categoryLabelCN = map[string]string{
	"big_tech": "大厂",
	"mid_tech": "中厂",
	"startup":  "创业公司",
	"foreign":  "外企",
	"other":    "其他",
}

func buildUserPrompt(in RecommendCompaniesInput) string {
	count := in.Count
	if count <= 0 {
		count = 20
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("城市：%s\n", in.City))
	b.WriteString(fmt.Sprintf("求职岗位：%s\n", in.Position))
	if in.ExperienceYears != nil {
		b.WriteString(fmt.Sprintf("工作经验：%d 年\n", *in.ExperienceYears))
	}
	if len(in.CompanyTypes) > 0 {
		labels := make([]string, 0, len(in.CompanyTypes))
		for _, t := range in.CompanyTypes {
			if l, ok := categoryLabelCN[t]; ok {
				labels = append(labels, l)
			}
		}
		if len(labels) > 0 {
			b.WriteString(fmt.Sprintf("偏好公司类型：%s\n", strings.Join(labels, "、")))
		}
	}
	b.WriteString(fmt.Sprintf(`
请推荐 %d 家符合条件的公司。严格按以下 JSON 格式返回（必须是纯 JSON，不要任何其他文字）：

{
  "companies": [
    {
      "name": "公司名称",
      "category": "big_tech | mid_tech | startup | foreign | other",
      "industry": "所属行业",
      "address_hint": "办公地点的区/街道级别描述",
      "reason": "推荐理由（50字内）"
    }
  ]
}`, count))
	return b.String()
}

// RecommendCompanies 调用豆包生成公司推荐
func (c *Client) RecommendCompanies(ctx context.Context, in RecommendCompaniesInput) (*RecommendCompaniesResult, error) {
	temp := 0.7
	topP := 0.9
	maxTokens := 4096

	resp, err := c.ChatCompletion(ctx, ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildUserPrompt(in)},
		},
		Temperature: &temp,
		TopP:        &topP,
		MaxTokens:   &maxTokens,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("doubao: empty choices")
	}

	content := resp.Choices[0].Message.Content

	// 兼容偶尔包裹 ```json ... ``` 的情况
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed struct {
		Companies []RecommendedCompany `json:"companies"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("doubao decode companies: %w; content=%s", err, truncate(content, 200))
	}

	// 过滤无效分类，归并到 other
	for i := range parsed.Companies {
		cat := parsed.Companies[i].Category
		if _, ok := categoryLabelCN[cat]; !ok {
			parsed.Companies[i].Category = "other"
		}
	}

	return &RecommendCompaniesResult{
		Companies: parsed.Companies,
		Usage:     resp.Usage,
	}, nil
}

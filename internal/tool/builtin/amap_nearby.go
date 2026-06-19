package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(nearbySearchTool{}) }

type nearbySearchTool struct{}

func (nearbySearchTool) Name() string { return "nearby_search" }

func (nearbySearchTool) Description() string {
	return "搜索指定坐标附近的POI（兴趣点），如餐厅、酒店、地铁站等。location 格式为 '经度,纬度'。"
}

func (nearbySearchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "location":{"type":"string","description":"中心点坐标，格式 '经度,纬度'，如 '116.397428,39.90923'"},
  "keywords":{"type":"string","description":"搜索关键词，如'餐厅'、'酒店'"},
  "types":{"type":"string","description":"POI类型编码，如 '060000'（餐饮）"},
  "radius":{"type":"integer","description":"搜索半径（米），默认1000，最大50000","default":1000},
  "offset":{"type":"integer","description":"每页记录数，默认20","default":20},
  "page":{"type":"integer","description":"页码，默认1","default":1}
},
"required":["location"]
}`)
}

func (nearbySearchTool) ReadOnly() bool { return true }

func (nearbySearchTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Location string `json:"location"`
		Keywords string `json:"keywords"`
		Types    string `json:"types"`
		Radius   int    `json:"radius"`
		Offset   int    `json:"offset"`
		Page     int    `json:"page"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Location == "" {
		return "", fmt.Errorf("location is required")
	}
	if p.Radius <= 0 {
		p.Radius = 1000
	}
	if p.Offset <= 0 {
		p.Offset = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}

	client := apiclient.NewAmapClient()
	radius := fmt.Sprintf("%d", p.Radius)
	offset := fmt.Sprintf("%d", p.Offset)
	page := fmt.Sprintf("%d", p.Page)
	result, err := client.NearbySearch(p.Location, p.Keywords, p.Types, radius, offset, page, "base")
	if err != nil {
		return "", fmt.Errorf("nearby search failed: %w", err)
	}

	// Parse POI response
	pois := parsePOIs(result)
	b, _ := json.MarshalIndent(map[string]any{
		"success": true,
		"count":   result["count"],
		"pois":    pois,
	}, "", "  ")
	return string(b), nil
}

// parsePOIs extracts POI entries from AMAP response.
func parsePOIs(data map[string]any) []map[string]any {
	rawPois, ok := data["pois"].([]any)
	if !ok {
		return nil
	}
	var pois []map[string]any
	for _, raw := range rawPois {
		if poi, ok := raw.(map[string]any); ok {
			pois = append(pois, map[string]any{
				"id":       poi["id"],
				"name":     poi["name"],
				"location": poi["location"],
				"address":  poi["address"],
				"cityname": poi["cityname"],
				"adname":   poi["adname"],
				"type":     poi["type"],
			})
		}
	}
	return pois
}

func init() { tool.RegisterBuiltin(fuzzySearchTool{}) }

type fuzzySearchTool struct{}

func (fuzzySearchTool) Name() string { return "fuzzy_search_poi" }

func (fuzzySearchTool) Description() string {
	return "根据关键词模糊搜索POI（兴趣点），支持指定城市。"
}

func (fuzzySearchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "keywords":{"type":"string","description":"搜索关键词，如'北京烤鸭'"},
  "city":{"type":"string","description":"城市名称或citycode，如'北京'或'010'"},
  "offset":{"type":"integer","description":"每页记录数，默认20","default":20},
  "page":{"type":"integer","description":"页码，默认1","default":1}
},
"required":["keywords"]
}`)
}

func (fuzzySearchTool) ReadOnly() bool { return true }

func (fuzzySearchTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Keywords string `json:"keywords"`
		City     string `json:"city"`
		Offset   int    `json:"offset"`
		Page     int    `json:"page"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Keywords == "" {
		return "", fmt.Errorf("keywords is required")
	}
	if p.Offset <= 0 {
		p.Offset = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}

	client := apiclient.NewAmapClient()
	result, err := client.FuzzySearch(p.Keywords, p.City, fmt.Sprintf("%d", p.Offset), fmt.Sprintf("%d", p.Page))
	if err != nil {
		return "", fmt.Errorf("fuzzy search failed: %w", err)
	}

	pois := parsePOIs(result)
	b, _ := json.MarshalIndent(map[string]any{
		"success": true,
		"count":   result["count"],
		"pois":    pois,
	}, "", "  ")
	return string(b), nil
}

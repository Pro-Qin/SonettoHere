package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(transitRouteTool{}) }

type transitRouteTool struct{}

func (transitRouteTool) Name() string { return "get_transit_route" }

func (transitRouteTool) Description() string {
	return "查询两个坐标点之间的公交路线规划。origin 和 destination 格式为 '经度,纬度'。"
}

func (transitRouteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "origin":{"type":"string","description":"起点坐标，格式 '经度,纬度'"},
  "destination":{"type":"string","description":"终点坐标，格式 '经度,纬度'"},
  "city":{"type":"string","description":"城市名称或citycode，如'北京'"},
  "strategy":{"type":"integer","description":"公交换乘策略: 0=最快捷,1=最经济,2=最少换乘,3=最少步行,5=不乘地铁","default":0},
  "date":{"type":"string","description":"出发日期 YYYY-MM-DD"},
  "time":{"type":"string","description":"出发时间 HH:MM"}
},
"required":["origin","destination","city"]
}`)
}

func (transitRouteTool) ReadOnly() bool { return true }

func (transitRouteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		City        string `json:"city"`
		Strategy    int    `json:"strategy"`
		Date        string `json:"date"`
		Time        string `json:"time"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Origin == "" || p.Destination == "" || p.City == "" {
		return "", fmt.Errorf("origin, destination, and city are required")
	}

	client := apiclient.NewAmapClient()
	result, err := client.TransitRoute(p.Origin, p.Destination, p.City, strconv.Itoa(p.Strategy), p.Date, p.Time)
	if err != nil {
		return "", fmt.Errorf("transit route failed: %w", err)
	}

	routes := parseTransitRoutes(result)
	b, _ := json.MarshalIndent(map[string]any{
		"success": true,
		"count":   result["count"],
		"routes":  routes,
	}, "", "  ")
	return string(b), nil
}

// parseTransitRoutes parses AMAP transit route response.
func parseTransitRoutes(data map[string]any) []map[string]any {
	route, ok := data["route"].(map[string]any)
	if !ok {
		return nil
	}
	transits, ok := route["transits"].([]any)
	if !ok {
		return nil
	}

	var routes []map[string]any
	for _, t := range transits {
		transit, ok := t.(map[string]any)
		if !ok {
			continue
		}
		cost, _ := transit["cost"].(string)
		duration, _ := transit["duration"].(string)
		walkingDist, _ := transit["walking_distance"].(string)
		if walkingDist == "" {
			if wd, ok2 := transit["walking_distance"].(float64); ok2 {
				walkingDist = strconv.Itoa(int(wd))
			}
		}

		costF, _ := strconv.ParseFloat(cost, 64)
		durI, _ := strconv.Atoi(duration)
		walkI, _ := strconv.Atoi(walkingDist)

		routes = append(routes, map[string]any{
			"cost":            costF,
			"duration":        durI,
			"walking_distance": walkI,
		})
	}
	return routes
}

func init() { tool.RegisterBuiltin(cyclingRouteTool{}) }

type cyclingRouteTool struct{}

func (cyclingRouteTool) Name() string { return "get_cycling_route" }

func (cyclingRouteTool) Description() string {
	return "查询两个坐标点之间的骑行路线规划。origin 和 destination 格式为 '经度,纬度'。"
}

func (cyclingRouteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "origin":{"type":"string","description":"起点坐标，格式 '经度,纬度'"},
  "destination":{"type":"string","description":"终点坐标，格式 '经度,纬度'"}
},
"required":["origin","destination"]
}`)
}

func (cyclingRouteTool) ReadOnly() bool { return true }

func (cyclingRouteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Origin == "" || p.Destination == "" {
		return "", fmt.Errorf("origin and destination are required")
	}

	client := apiclient.NewAmapClient()
	result, err := client.CyclingRoute(p.Origin, p.Destination)
	if err != nil {
		return "", fmt.Errorf("cycling route failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

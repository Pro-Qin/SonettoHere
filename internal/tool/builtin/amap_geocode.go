package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(geocodeTool{}) }

type geocodeTool struct{}

func (geocodeTool) Name() string { return "geocode_address" }

func (geocodeTool) Description() string {
	return "将详细地址转换为经纬度坐标（GCJ-02 火星坐标系）。返回坐标可直接用于 nearby_search / get_transit_route / get_cycling_route。"
}

func (geocodeTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "address":{"type":"string","description":"详细地址字符串，如'北京市海淀区中关村大街'"}
},
"required":["address"]
}`)
}

func (geocodeTool) ReadOnly() bool { return true }

func (geocodeTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Address == "" {
		return "", fmt.Errorf("address is required")
	}

	client := apiclient.NewAmapClient()
	result, err := client.Geocode(p.Address)
	if err != nil {
		return "", fmt.Errorf("geocode failed: %w", err)
	}

	// Parse AMAP geo response
	status, _ := result["status"].(string)
	if status == "1" {
		if geocodes, ok := result["geocodes"].([]any); ok && len(geocodes) > 0 {
			if first, ok := geocodes[0].(map[string]any); ok {
				loc, _ := first["location"].(string)
				b, _ := json.MarshalIndent(map[string]any{
					"success":  true,
					"address":  p.Address,
					"location": loc,
				}, "", "  ")
				return string(b), nil
			}
		}
	}
	return fmt.Sprintf(`{"success":false,"error":"地理编码失败，未找到'%s'的坐标"}`, p.Address), nil
}

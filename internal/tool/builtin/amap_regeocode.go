package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(regeocodeTool{}) }

type regeocodeTool struct{}

func (regeocodeTool) Name() string { return "regeocode" }

func (regeocodeTool) Description() string {
	return "将经纬度坐标（GCJ-02 火星坐标系）转换为详细地址信息。"
}

func (regeocodeTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "location":{"type":"string","description":"坐标，格式 '经度,纬度'，如 '116.397428,39.90923'"}
},
"required":["location"]
}`)
}

func (regeocodeTool) ReadOnly() bool { return true }

func (regeocodeTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Location == "" {
		return "", fmt.Errorf("location is required")
	}

	client := apiclient.NewAmapClient()
	result, err := client.ReGeocode(p.Location)
	if err != nil {
		return "", fmt.Errorf("regeocode failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

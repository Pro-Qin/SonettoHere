package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(weatherTool{}) }

type weatherTool struct{}

func (weatherTool) Name() string { return "get_current_weather" }

func (weatherTool) Description() string {
	return "获取指定城市的天气信息。支持实时天气、多天预报、逐小时预报、分钟级降水、生活指数。city 和 adcode 二选一即可。"
}

func (weatherTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "city":{"type":"string","description":"城市名称，如'北京'、'Shanghai'"},
  "adcode":{"type":"string","description":"行政区划代码，如'110000'"},
  "extended":{"type":"boolean","description":"返回体感温度/能见度/气压/紫外线/AQI等扩展信息"},
  "forecast":{"type":"boolean","description":"返回最多7天预报"},
  "hourly":{"type":"boolean","description":"返回24小时逐小时预报"},
  "minutely":{"type":"boolean","description":"返回分钟级降水预报（仅国内）"},
  "indices":{"type":"boolean","description":"返回18项生活指数"},
  "lang":{"type":"string","description":"语言：zh/en","default":"zh"}
},
"anyOf":[
  {"required":["city"]},
  {"required":["adcode"]}
]
}`)
}

func (weatherTool) ReadOnly() bool { return true }

func (weatherTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		City     string `json:"city"`
		Adcode   string `json:"adcode"`
		Extended bool   `json:"extended"`
		Forecast bool   `json:"forecast"`
		Hourly   bool   `json:"hourly"`
		Minutely bool   `json:"minutely"`
		Indices  bool   `json:"indices"`
		Lang     string `json:"lang"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.City == "" && p.Adcode == "" {
		return "", fmt.Errorf("city or adcode is required")
	}
	if p.Lang == "" {
		p.Lang = "zh"
	}

	client := apiclient.NewUapisClient()
	result, err := client.GetWeather(p.City, p.Adcode, p.Extended, p.Forecast, p.Hourly, p.Minutely, p.Indices, p.Lang)
	if err != nil {
		return "", fmt.Errorf("weather query failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

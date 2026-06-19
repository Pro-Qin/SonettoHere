package builtin

import (
	"context"
	"encoding/json"
	"os"

	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(healthCheckTool{}) }

type healthCheckTool struct{}

func (healthCheckTool) Name() string { return "health_check" }

func (healthCheckTool) Description() string {
	return "检查 Reasonix 系统运行状态和各组件健康状况。返回 tools 注册状态、API key 配置状态以及版本信息。"
}

func (healthCheckTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "component":{"type":"string","description":"可选：检查特定组件 'tools'/'apikeys'/'all'，不传则返回概要"}
}
}`)
}

func (healthCheckTool) ReadOnly() bool { return true }

func (healthCheckTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Component string `json:"component"`
	}
	json.Unmarshal(args, &p)

	builtins := tool.Builtins()
	toolNames := make([]string, len(builtins))
	for i, t := range builtins {
		toolNames[i] = t.Name()
	}

	apiKeys := map[string]struct {
		Set  bool   `json:"set"`
		Hint string `json:"hint"`
	}{
		"UAPIS_API_KEY": {
			Set:  os.Getenv("UAPIS_API_KEY") != "",
			Hint: "天气/节假日/娱乐工具",
		},
		"AMAP_API_KEY": {
			Set:  os.Getenv("AMAP_API_KEY") != "",
			Hint: "地图工具 (https://console.amap.com)",
		},
		"TODOIST_API_TOKEN": {
			Set:  os.Getenv("TODOIST_API_TOKEN") != "",
			Hint: "Todoist 任务管理",
		},
	}

	result := map[string]any{
		"status":      "ok",
		"tools_count": len(toolNames),
		"tools":       toolNames,
		"api_keys":    apiKeys,
	}
	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

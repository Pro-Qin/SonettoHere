package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(currentTimeTool{}) }

type currentTimeTool struct{}

func (currentTimeTool) Name() string { return "current_time" }

func (currentTimeTool) Description() string {
	return "获取当前系统时间。"
}

func (currentTimeTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "timezone":{"type":"string","description":"时区，如 'Asia/Shanghai'、'America/New_York'，默认 Local","default":"Local"}
}
}`)
}

func (currentTimeTool) ReadOnly() bool { return true }

func (currentTimeTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Timezone string `json:"timezone"`
	}
	json.Unmarshal(args, &p)

	loc := time.Local
	if p.Timezone != "" && p.Timezone != "Local" {
		var err error
		loc, err = time.LoadLocation(p.Timezone)
		if err != nil {
			return "", fmt.Errorf("invalid timezone %q: %w", p.Timezone, err)
		}
	}

	now := time.Now().In(loc)
	b, _ := json.MarshalIndent(map[string]any{
		"success":  true,
		"time":     now.Format("2006-01-02 15:04:05"),
		"timezone": loc.String(),
		"date":     now.Format("2006-01-02"),
		"weekday":  now.Weekday().String(),
		"unix":     now.Unix(),
	}, "", "  ")
	return string(b), nil
}

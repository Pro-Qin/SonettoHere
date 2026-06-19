package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(holidayTool{}) }

type holidayTool struct{}

func (holidayTool) Name() string { return "holiday_calendar" }

func (holidayTool) Description() string {
	return "查询指定日期/月份/年份的万年历与节假日信息。支持农历、节气、法定节假日。date/month/year 三选一。"
}

func (holidayTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "date":{"type":"string","description":"按天查询，格式 YYYY-MM-DD（与 month/year 三选一）"},
  "month":{"type":"string","description":"按月查询，格式 YYYY-MM（与 date/year 三选一）"},
  "year":{"type":"string","description":"按年查询，格式 YYYY（与 date/month 三选一）"},
  "timezone":{"type":"string","description":"时区","default":"Asia/Shanghai"},
  "holiday_type":{"type":"string","description":"节日类型: all/legal/legal_rest/legal_workday/solar/lunar/term","default":"all"},
  "include_nearby":{"type":"boolean","description":"是否返回前后最近节日（仅 date 模式）"},
  "nearby_limit":{"type":"integer","description":"最近节日数量限制","default":7}
},
"anyOf":[
  {"required":["date"]},
  {"required":["month"]},
  {"required":["year"]}
]
}`)
}

func (holidayTool) ReadOnly() bool { return true }

func (holidayTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Date          string `json:"date"`
		Month         string `json:"month"`
		Year          string `json:"year"`
		Timezone      string `json:"timezone"`
		HolidayType   string `json:"holiday_type"`
		IncludeNearby bool   `json:"include_nearby"`
		NearbyLimit   int    `json:"nearby_limit"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Date == "" && p.Month == "" && p.Year == "" {
		return "", fmt.Errorf("date, month, or year is required")
	}
	if p.Timezone == "" {
		p.Timezone = "Asia/Shanghai"
	}
	if p.HolidayType == "" {
		p.HolidayType = "all"
	}
	if p.NearbyLimit <= 0 {
		p.NearbyLimit = 7
	}

	client := apiclient.NewUapisClient()
	result, err := client.GetHolidayCalendar(p.Date, p.Month, p.Year, p.Timezone, p.HolidayType, p.IncludeNearby, p.NearbyLimit)
	if err != nil {
		return "", fmt.Errorf("holiday query failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(tarotTool{}) }

type tarotTool struct{}

func (tarotTool) Name() string { return "tarot_reading" }

func (tarotTool) Description() string {
	return "塔罗牌占卜，随机抽取指定数量的塔罗牌并解读。每次最多5张。"
}

func (tarotTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "count":{"type":"integer","description":"抽取牌数（1-5），默认3","default":3}
}
}`)
}

func (tarotTool) ReadOnly() bool { return true }

func (tarotTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Count int `json:"count"`
	}
	json.Unmarshal(args, &p)
	if p.Count <= 0 || p.Count > 5 {
		p.Count = 3
	}

	client := apiclient.NewUapisClient()
	result, err := client.GetTarot(p.Count)
	if err != nil {
		return "", fmt.Errorf("tarot failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

func init() { tool.RegisterBuiltin(answerBookTool{}) }

type answerBookTool struct{}

func (answerBookTool) Name() string { return "answer_book" }

func (answerBookTool) Description() string {
	return "答案之书 —— 给出一个随机回答来帮助你做决定或寻找灵感。"
}

func (answerBookTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "question":{"type":"string","description":"你心中的问题"}
}
}`)
}

func (answerBookTool) ReadOnly() bool { return true }

func (answerBookTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Question string `json:"question"`
	}
	json.Unmarshal(args, &p)

	client := apiclient.NewUapisClient()
	result, err := client.GetAnswerBook(p.Question)
	if err != nil {
		return "", fmt.Errorf("answer book failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "data": result}, "", "  ")
	return string(b), nil
}

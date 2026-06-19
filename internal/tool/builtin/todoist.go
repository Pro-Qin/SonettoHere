package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"reasonix/internal/apiclient"
	"reasonix/internal/tool"
)

func init() { tool.RegisterBuiltin(todoistAddTool{}) }

type todoistAddTool struct{}

func (todoistAddTool) Name() string { return "todoist_add" }

func (todoistAddTool) Description() string {
	return "向 Todoist 添加新任务。"
}

func (todoistAddTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "content":{"type":"string","description":"任务名称/内容，例如'完成项目报告'"},
  "due_string":{"type":"string","description":"截止日期。支持自然语言（明天下午3点、下周五）或 YYYY-MM-DD"},
  "priority":{"type":"integer","description":"1=低, 2=中, 3=高, 4=紧急","default":1},
  "project_id":{"type":"string","description":"所属项目ID。不填默认 Inbox。先通过 todoist_list_projects 获取项目ID"}
},
"required":["content"]
}`)
}

func (todoistAddTool) ReadOnly() bool { return false }

func (todoistAddTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Content    string `json:"content"`
		DueString  string `json:"due_string"`
		Priority   int    `json:"priority"`
		ProjectID  string `json:"project_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.Content == "" {
		return "", fmt.Errorf("content is required")
	}
	if p.Priority < 1 || p.Priority > 4 {
		p.Priority = 1
	}

	client := apiclient.NewTodoistClient()
	task, err := client.AddTask(p.Content, p.ProjectID, p.DueString, p.Priority)
	if err != nil {
		return "", fmt.Errorf("add task failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "task": task}, "", "  ")
	return string(b), nil
}

func init() { tool.RegisterBuiltin(todoistListTool{}) }

type todoistListTool struct{}

func (todoistListTool) Name() string { return "todoist_list_tasks" }

func (todoistListTool) Description() string {
	return "列出 Todoist 任务。可按项目ID或过滤器筛选。"
}

func (todoistListTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "project_id":{"type":"string","description":"项目ID，只返回该项目下的任务"},
  "filter":{"type":"string","description":"过滤器，如 'today'、'p1'、'view all'"}
}
}`)
}

func (todoistListTool) ReadOnly() bool { return true }

func (todoistListTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		Filter    string `json:"filter"`
	}
	json.Unmarshal(args, &p)

	client := apiclient.NewTodoistClient()
	tasks, err := client.GetTasks(p.ProjectID, p.Filter)
	if err != nil {
		return "", fmt.Errorf("list tasks failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "tasks": tasks, "count": len(tasks)}, "", "  ")
	return string(b), nil
}

func init() { tool.RegisterBuiltin(todoistCloseTool{}) }

type todoistCloseTool struct{}

func (todoistCloseTool) Name() string { return "todoist_complete_task" }

func (todoistCloseTool) Description() string {
	return "将 Todoist 任务标记为已完成。"
}

func (todoistCloseTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "task_id":{"type":"string","description":"任务ID"}
},
"required":["task_id"]
}`)
}

func (todoistCloseTool) ReadOnly() bool { return false }

func (todoistCloseTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.TaskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	client := apiclient.NewTodoistClient()
	if err := client.CloseTask(p.TaskID); err != nil {
		return "", fmt.Errorf("complete task failed: %w", err)
	}

	return `{"success":true,"message":"任务已标记为完成"}`, nil
}

func init() { tool.RegisterBuiltin(todoistDeleteTool{}) }

type todoistDeleteTool struct{}

func (todoistDeleteTool) Name() string { return "todoist_delete_task" }

func (todoistDeleteTool) Description() string {
	return "删除一个 Todoist 任务。"
}

func (todoistDeleteTool) Schema() json.RawMessage {
	return json.RawMessage(`{
"type":"object",
"properties":{
  "task_id":{"type":"string","description":"任务ID"}
},
"required":["task_id"]
}`)
}

func (todoistDeleteTool) ReadOnly() bool { return false }

func (todoistDeleteTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.TaskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	client := apiclient.NewTodoistClient()
	if err := client.DeleteTask(p.TaskID); err != nil {
		return "", fmt.Errorf("delete task failed: %w", err)
	}

	return `{"success":true,"message":"任务已删除"}`, nil
}

func init() { tool.RegisterBuiltin(todoistListProjectsTool{}) }

type todoistListProjectsTool struct{}

func (todoistListProjectsTool) Name() string { return "todoist_list_projects" }

func (todoistListProjectsTool) Description() string {
	return "列出 Todoist 中的所有项目。在添加任务前应先调用此工具确认项目存在。"
}

func (todoistListProjectsTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

func (todoistListProjectsTool) ReadOnly() bool { return true }

func (todoistListProjectsTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	client := apiclient.NewTodoistClient()
	projects, err := client.GetProjects()
	if err != nil {
		return "", fmt.Errorf("list projects failed: %w", err)
	}

	b, _ := json.MarshalIndent(map[string]any{"success": true, "projects": projects, "count": len(projects)}, "", "  ")
	return string(b), nil
}

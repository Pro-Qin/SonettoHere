# SonettoHere v2.0

基于 LangChain + LangGraph 的 ReAct AI Agent，支持 CLI 和 QQ Bot 双入口。

## 能力概览

SonettoHere 内置 **30 个 Skill**，覆盖 10 个领域：

| 领域 | 技能 | 说明 |
|------|------|------|
| **Todo** | 添加/列出/完成/取消/删除/更新/查询任务、列出项目 | Todoist 任务管理 |
| **地图** | 周边搜索、地址编码、公交路线、骑行路线、模糊地址 | 高德地图 API |
| **网络** | 天气查询、智能搜索、网页抓取、节假日日历 | UAPI + 通用 HTTP |
| **文件** | 文件读写/删改/目录操作、PDF 阅读、Word 阅读 | 本地文件系统操作 |
| **开发** | 语法检查、代码质量分析、单元测试运行、调试器 | Python/JS/TS 代码工具 |
| **系统** | 当前时间、Python 脚本执行 | 系统级工具 |
| **任务追踪** | 多步骤任务进度管理 | 内存中任务状态跟踪 |
| **交互** | 向用户提问 | 非交互环境自动降级 |
| **娱乐** | 答案之书、塔罗牌 | UAPI + 内置 78 张塔罗牌库 |

> 每个复杂领域的 Skill 采用**两步调用**模式：LLM 先通过 `get_doc=true` 读取同目录的 `SKILL.md` 获取领域知识，再带真实参数执行。简单 Skill 可跳过文档一步到位。

## 环境要求

- Python >= 3.11
- DeepSeek API Key（或其他 OpenAI 兼容的 API）

## 安装

```bash
# 克隆项目
cd SonettoHere

# 创建虚拟环境
python -m venv .venv
source .venv/bin/activate   # Linux/macOS
# .venv\Scripts\activate    # Windows

# 安装依赖
pip install -e .
```

## 配置

### 1. 创建 .env 文件

```bash
cp .env.example .env
```

### 2. 填写 API Key

```env
# 必填 — LLM 后端
DEEPSEEK_API_KEY=sk-your-key-here
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1

# 可选 — 按需填写（不影响无需对应 Skill 的基础对话）
TODOIST_API_TOKEN=your-todoist-token    # Todo 技能
UAPIS_API_KEY=your-uapi-key             # 天气/搜索/答案之书
AMAP_API_KEY=your-amap-key             # 地图技能
QQ_APPID=your-appid                     # QQ Bot
QQ_TOKEN=your-token                     # QQ Bot
```

> DeepSeek API 兼容 OpenAI 格式，也可以用其他兼容服务替换（修改 `DEEPSEEK_BASE_URL` 和模型名即可）。

## 使用

### CLI 模式（默认）

```bash
python main.py
# 或
python main.py cli
```

交互命令：
- 直接输入问题即可对话
- `/clear` — 清空当前对话
- `/exit` — 退出

### QQ Bot 模式

```bash
python main.py qqbot
```

> QQ Bot 适配器目前为骨架，待 botpy SDK 集成。

### 编程调用

```python
import asyncio
from clients.cli import SonettoCLI

async def main():
    cli = SonettoCLI()
    await cli.run()

asyncio.run(main())
```

## 项目结构

```
SonettoHere/
├── main.py                  # 入口选择器 (cli / qqbot)
├── pyproject.toml           # 项目元数据与依赖
├── .env.example             # 环境变量模板
│
├── config/
│   ├── settings.py          # Pydantic Settings（从 .env 加载）
│   └── personas/
│       ├── AGENTS.md        # Agent 行为规则
│       ├── SOUL.md          # 人设（Sonetto 大姐姐）
│       └── MEMORY.md        # 固定记忆
│
├── agent/
│   ├── graph.py             # LangGraph StateGraph 构建
│   ├── state.py             # AgentState TypedDict
│   └── prompts.py           # 系统提示词组装（启动缓存）
│
├── skills/                  # ★ 30 个 Skill，每个文件夹含 SKILL.md
│   ├── base.py              # SkillBase 基类 + SharedAPIClient
│   ├── __init__.py          # get_all_skills() 集中注册
│   ├── todo/                # Todoist（8 个 Skill）
│   ├── map/                 # 高德地图（5 个 Skill）
│   ├── network/             # 网络服务（4 个 Skill）
│   ├── system/              # 系统工具（2 个 Skill）
│   ├── files/               # 文件操作（3 个 Skill）
│   ├── development/         # 开发工具（4 个 Skill）
│   ├── task/                # 任务追踪（1 个 Skill）
│   ├── interaction/         # 用户交互（1 个 Skill）
│   └── entertainment/       # 娱乐（2 个 Skill）
│
├── memory/
│   ├── short_term.py        # 短期记忆（token 自动裁剪）
│   ├── long_term.py         # 长期记忆（关键词评分检索）
│   ├── extractor.py         # 对话回合摘要提取
│   └── preference.py        # 用户偏好画像
│
├── callbacks/
│   └── printer.py           # 彩色终端输出
│
└── clients/
    ├── cli.py               # 异步 CLI 入口
    └── qqbot.py             # QQ Bot 适配器（骨架）
```

## 架构

```
用户输入 → [Agent: LLM + bind_tools(skills)] → [Skills: 执行] → 循环
              ↑                                    ↓
              │                    [final answer / get_doc=true 返回文档]
              └── LLM 读完文档后继续循环，下一步带真实参数调用
```

- **LLM 后端**：DeepSeek Chat（OpenAI 兼容 tool calling）
- **Agent 框架**：LangGraph `create_react_agent` + MemorySaver 状态持久化
- **记忆系统**：短期记忆（token 阈值自动裁剪）+ 长期记忆（关键词+置信度+时间衰减评分检索）
- **Skill 模式**：每个 Skill = `SKILL.md`（领域知识）+ `skill_*.py`（执行代码），通过 `get_doc=true` 按需加载文档

## License

MIT

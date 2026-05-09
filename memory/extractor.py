"""回合摘要提取 — LLM 驱动的对话关键信息抽取。"""

import datetime
import json
import os
from typing import Any

EXTRACTION_PROMPT = """你是一个专业的文本分析师，能够根据Agent对话记录提取出关键信息。

你的输出格式为json，键名如下（缺失的字段请返回空字符串或空数组）：
{
    "summary": "提取后的摘要",
    "keywords": ["关键词1", "关键词2", "关键词3"],
    "correction": "（可选）如果用户在对话中修正了Agent的错误，说明修正信息",
    "mistake": "（可选）如果Agent在调用工具或理解问题时出现错误，说明错误信息",
    "user_preferences": [{"key":"称呼","value":"Miso","confidence":0.9,"evidence":"用户明确说"}],
    "user_habits": [{"habit":"倾向简短回答","confidence":0.7,"evidence":"多次要求简洁"}],
    "error_patterns": [{"type":"tool_selection","tool":"nearby_search","trigger":"找附近实体店时","detail":"错误地使用了模糊搜索"}],
    "anti_patterns": [{"type":"tool_selection","tool":"nearby_search","instruction":"涉及附近搜索优先使用nearby而非fuzzy"}],
    "confidence": 0.0-1.0
}"""


def extract_from_messages(
    messages: list[dict[str, str]],
    llm,
    max_retries: int = 2,
) -> dict[str, Any]:
    """使用 LLM 从对话消息中提取关键摘要。"""
    prompt = [
        {"role": "system", "content": EXTRACTION_PROMPT},
        {"role": "user", "content": f"请根据以下对话记录提取出关键信息：\n\n{messages}"},
    ]

    for _ in range(max_retries + 1):
        try:
            raw = llm.generate(prompt)
            parsed = json.loads(raw)
            return _normalize(parsed)
        except (json.JSONDecodeError, Exception):
            continue

    fallback = ""
    if messages:
        fallback = " | ".join(
            f"{m.get('role', '?')}: {str(m.get('content', ''))[:80]}"
            for m in messages[-6:]
        )
    return _normalize({"summary": fallback or "提取失败", "keywords": [], "confidence": 0.2})


def _normalize(record: dict[str, Any]) -> dict[str, Any]:
    return {
        "summary": record.get("summary", ""),
        "keywords": record.get("keywords", []),
        "correction": record.get("correction", ""),
        "mistake": record.get("mistake", ""),
        "user_preferences": record.get("user_preferences", []),
        "user_habits": record.get("user_habits", []),
        "error_patterns": record.get("error_patterns", []),
        "anti_patterns": record.get("anti_patterns", []),
        "confidence": record.get("confidence", 0.6),
    }


def save_extracted(
    extracted: dict[str, Any],
    history_path: str = "extracted/extracted_history.json",
    source: str = "cli",
    session_id: str = "",
) -> None:
    """追加提取记录到长期记忆文件。"""
    os.makedirs(os.path.dirname(history_path), exist_ok=True)

    history: list[dict] = []
    if os.path.isfile(history_path):
        with open(history_path, "r", encoding="utf-8") as f:
            history = json.load(f)

    record = _normalize(extracted)
    record["timestamp"] = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    record["session_id"] = session_id or datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    record["source"] = source

    history.append(record)
    # 保留最近 300 条
    history = history[-300:]

    with open(history_path, "w", encoding="utf-8") as f:
        json.dump(history, f, ensure_ascii=False, indent=2)

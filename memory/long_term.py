"""长期记忆检索：轻量关键词 + 置信度 + 时间衰减评分（纯 JSON 文件存储）。"""

import json
import os
from datetime import datetime
from typing import Any


EXTRACTED_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "extracted")
HISTORY_FILE = os.path.join(EXTRACTED_DIR, "extracted_history.json")


def _load_history() -> list[dict[str, Any]]:
    """加载长期记忆文件，不存在则返回空列表。"""
    os.makedirs(EXTRACTED_DIR, exist_ok=True)
    if not os.path.isfile(HISTORY_FILE):
        return []
    try:
        with open(HISTORY_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
            return data if isinstance(data, list) else []
    except (json.JSONDecodeError, OSError):
        return []


def _keyword_score(query: str, record: dict[str, Any]) -> float:
    """计算 query 对 record.keywords 的关键词命中得分（0.0 ~ 1.0）。"""
    keywords: list[str] = record.get("keywords", [])
    if not keywords or not query:
        return 0.0
    query_lower = query.lower()
    hits = sum(1 for kw in keywords if kw.lower() in query_lower or query_lower in kw.lower())
    return min(hits / max(len(keywords), 1), 1.0)


def _confidence_weight(record: dict[str, Any]) -> float:
    """置信度加权（0.0 ~ 1.0）。"""
    conf = record.get("confidence", 0.5)
    try:
        return float(conf)
    except (TypeError, ValueError):
        return 0.5


def _time_decay(timestamp_str: str) -> float:
    """时间衰减：越新的记录分越高（0.0 ~ 1.0），半衰期 7 天。"""
    if not timestamp_str:
        return 0.5
    try:
        record_time = datetime.strptime(timestamp_str, "%Y-%m-%d %H:%M:%S")
        elapsed_hours = (datetime.now() - record_time).total_seconds() / 3600
        # 指数衰减，半衰期 168 小时（7 天）
        return 0.5 ** (elapsed_hours / 168)
    except (ValueError, OSError):
        return 0.5


def retrieve_long_term_context(
    query: str, top_k: int = 10
) -> dict[str, list[dict[str, Any]]]:
    """检索与当前 query 相关的长期记忆。

    返回:
        {"error_rules": [...], "preference_rules": [...]}
    """
    history = _load_history()
    if not history:
        return {"error_rules": [], "preference_rules": []}

    records_with_scores = []
    for record in history:
        kw_score = _keyword_score(query, record)
        conf_w = _confidence_weight(record)
        time_w = _time_decay(record.get("timestamp", ""))
        total = kw_score * 0.4 + conf_w * 0.2 + time_w * 0.4
        records_with_scores.append((total, record))

    records_with_scores.sort(key=lambda x: x[0], reverse=True)
    top_records = [r for _, r in records_with_scores[:top_k]]

    error_rules = []
    preference_rules = []

    seen_error = set()
    seen_pref = set()

    for record in top_records:
        # 纠偏规则
        correction = str(record.get("correction", "")).strip()
        if correction and correction not in seen_error:
            error_rules.append({"type": "correction", "rule": correction})
            seen_error.add(correction)

        mistake = str(record.get("mistake", "")).strip()
        if mistake and mistake not in seen_error:
            error_rules.append({"type": "mistake", "rule": mistake})
            seen_error.add(mistake)

        for ap in record.get("anti_patterns", []):
            ap_text = ap if isinstance(ap, str) else ap.get("instruction", "")
            ap_text = str(ap_text).strip()
            if ap_text and ap_text not in seen_error:
                error_rules.append({"type": "anti_pattern", "rule": ap_text})
                seen_error.add(ap_text)

        # 偏好规则
        for pref in record.get("user_preferences", []):
            if isinstance(pref, dict):
                key = pref.get("key", "")
                val = pref.get("value", "")
                tag = f"{key}={val}"
                if tag not in seen_pref:
                    preference_rules.append(pref)
                    seen_pref.add(tag)

        for habit in record.get("user_habits", []):
            if isinstance(habit, dict):
                h_text = habit.get("habit", "")
                if h_text and h_text not in seen_pref:
                    preference_rules.append(habit)
                    seen_pref.add(h_text)

    return {
        "error_rules": error_rules[:10],
        "preference_rules": preference_rules[:12],
    }

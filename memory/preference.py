"""偏好画像管理：从长期记忆中提取并维护用户稳定偏好。"""

import json
import os
from typing import Any

EXTRACTED_DIR = os.path.join(os.path.dirname(os.path.dirname(__file__)), "extracted")
PREFERENCE_FILE = os.path.join(EXTRACTED_DIR, "user_preference_profile.json")


def get_stable_preferences() -> dict[str, dict[str, Any]]:
    """读取用户稳定偏好画像。"""
    os.makedirs(EXTRACTED_DIR, exist_ok=True)
    if not os.path.isfile(PREFERENCE_FILE):
        return {}
    try:
        with open(PREFERENCE_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
            return data if isinstance(data, dict) else {}
    except (json.JSONDecodeError, OSError):
        return {}


def sync_preferences(history: list[dict[str, Any]]) -> None:
    """从长期历史记录中聚合稳定偏好，写入画像文件。

    对于每个 preference key，保留置信度最高的那条。
    """
    profile: dict[str, dict[str, Any]] = {}
    for item in history:
        for pref in item.get("user_preferences", []):
            if not isinstance(pref, dict):
                continue
            key = str(pref.get("key", "")).strip()
            if not key:
                continue
            confidence = float(pref.get("confidence", 0.5))
            prev = profile.get(key)
            if prev is None or confidence >= float(prev.get("confidence", 0.0)):
                profile[key] = {
                    "value": pref.get("value", ""),
                    "confidence": confidence,
                    "updated_at": item.get("timestamp", ""),
                }

    os.makedirs(EXTRACTED_DIR, exist_ok=True)
    with open(PREFERENCE_FILE, "w", encoding="utf-8") as f:
        json.dump(profile, f, ensure_ascii=False, indent=4)

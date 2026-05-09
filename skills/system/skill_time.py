"""Skill: time_skill — 获取当前日期和时间。"""

from datetime import datetime

from pydantic import BaseModel

from skills.base import SkillBase, format_success


class TimeInput(BaseModel):
    """time_skill 无参数，直接调用即可。"""


class TimeSkill(SkillBase):
    name: str = "time_skill"
    description: str = "获取当前日期和时间。直接调用即可，无需先读文档。"
    args_schema: type[BaseModel] = TimeInput

    def _run(self) -> str:
        now = datetime.now()
        return format_success({
            "datetime": now.strftime("%Y-%m-%d %H:%M:%S"),
            "date": now.strftime("%Y-%m-%d"),
            "time": now.strftime("%H:%M:%S"),
            "weekday": now.strftime("%A"),
            "timezone": "Asia/Shanghai",
        })

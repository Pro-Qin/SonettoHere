"""Skill: run_python — 执行 Python 代码。"""

import io
import sys

from pydantic import BaseModel, Field

from skills.base import SkillBase, format_error, format_success


class RunPythonInput(BaseModel):
    get_doc: bool = Field(
        default=False,
        description="设为 true 以获取使用说明和安全限制"
    )
    code: str = Field(
        default="",
        description="要执行的 Python 代码，支持多行"
    )


class RunPythonSkill(SkillBase):
    name: str = "run_python"
    description: str = (
        "在隔离环境中执行 Python 代码，返回 stdout 输出。"
        "用于计算、数据处理、文本转换。★ 首次使用请先 get_doc=true 了解安全限制。"
    )
    args_schema: type[BaseModel] = RunPythonInput

    def _run(self, get_doc: bool = False, code: str = "") -> str:
        if get_doc:
            return self._load_doc()
        if not code:
            return format_error("code 不能为空")

        old_stdout = sys.stdout
        sys.stdout = io.StringIO()

        try:
            exec(code, {"__builtins__": __builtins__})
            output = sys.stdout.getvalue()
            return format_success({"output": output} if output else {"output": "（代码执行完毕，无输出）"})
        except Exception as e:
            return format_error(f"代码执行错误: {e}")
        finally:
            sys.stdout = old_stdout

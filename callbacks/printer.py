"""LangChain Callback Handler — 彩色终端输出。"""

import time
from typing import Any

from colorama import Fore, Style, init
from langchain_core.callbacks import BaseCallbackHandler
from langchain_core.outputs import LLMResult

init(autoreset=True)


class PrinterCallback(BaseCallbackHandler):
    """终端输出回调：思考流、Skill 调用、记忆状态。"""

    def __init__(self):
        super().__init__()
        self._tool_start_time: dict[str, float] = {}
        self._thinking_started = False

    def on_llm_start(
        self, serialized: dict[str, Any], prompts: list[str], **kwargs: Any
    ) -> None:
        self._thinking_started = True
        print(f"\n{Fore.CYAN}{Style.BRIGHT}┌── [Thinking]")
        print(f"{Fore.CYAN}│", end="", flush=True)

    def on_llm_new_token(self, token: str, **kwargs: Any) -> None:
        print(f"{Fore.CYAN}{token}", end="", flush=True)

    def on_llm_end(self, response: LLMResult, **kwargs: Any) -> None:
        if self._thinking_started:
            print(f"\n{Fore.CYAN}└── [End Thinking]")
            self._thinking_started = False

    def on_tool_start(
        self, serialized: dict[str, Any], input_str: str, **kwargs: Any
    ) -> None:
        tool_name = serialized.get("name", "unknown")
        run_id = str(kwargs.get("run_id", ""))
        self._tool_start_time[run_id] = time.time()

        input_display = input_str
        if len(input_display) > 200:
            input_display = input_display[:200] + f"... (共 {len(input_str)} 字符)"

        print(f"\n{Fore.MAGENTA}{Style.BRIGHT}├── [Skill] {Fore.YELLOW}{tool_name}")
        if input_display.strip() and input_display.strip() != "{}":
            print(f"{Fore.MAGENTA}│   └── 参数: {Fore.GREEN}{input_display}")

    def on_tool_end(self, output: str, **kwargs: Any) -> None:
        run_id = str(kwargs.get("run_id", ""))
        elapsed = time.time() - self._tool_start_time.pop(run_id, time.time())

        output_display = str(output) if not isinstance(output, str) else output
        if len(output_display) > 300:
            output_display = output_display[:300] + f"... (共 {len(output_display)} 字符)"

        print(f"{Fore.GREEN}{Style.BRIGHT}└── [End Skill]")
        print(f"{Fore.GREEN}    ├── 耗时: {Fore.CYAN}{elapsed:.2f}s")
        print(f"{Fore.GREEN}    └── 结果: {Fore.WHITE}{output_display}")

    def on_tool_error(self, error: BaseException, **kwargs: Any) -> None:
        run_id = str(kwargs.get("run_id", ""))
        self._tool_start_time.pop(run_id, None)
        print(f"{Fore.RED}    └── 错误: {error}")

    def print_memory(self, current_tokens: int, max_tokens: int) -> None:
        pct = current_tokens / max_tokens * 100 if max_tokens else 0
        print(f"{Fore.CYAN}{Style.BRIGHT}┌── [Memory]")
        print(f"{Fore.CYAN}│   ├── {current_tokens:,} / {max_tokens:,} tokens | {pct:.1f}%")
        print(f"{Fore.CYAN}└── [End Memory]")

    def print_answer(self, answer: str) -> None:
        print(f"\n{Fore.WHITE}{Style.BRIGHT}┌── [Final Answer]")
        print(f"{Fore.WHITE}│")
        for line in answer.split("\n"):
            print(f"{Fore.WHITE}│   {line}")
        print(f"{Fore.WHITE}│")
        print(f"{Fore.WHITE}└── [End Final Answer]")

    def print_tasks(self, tasks: list[str]) -> None:
        print(f"{Fore.RED}{Style.BRIGHT}├── [Task] 任务群已建立")
        for i, task in enumerate(tasks):
            print(f"{Fore.RED}│   ├── 任务 {i + 1}: {task}")

    def print_task_start(self, task: str) -> None:
        print(f"{Fore.RED}│   └── 开始: {task}")

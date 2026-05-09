"""短期记忆 — InMemoryChatMessageHistory + token 自动裁剪。"""

from typing import Any

import tiktoken
from langchain_core.chat_history import InMemoryChatMessageHistory
from langchain_core.messages import BaseMessage


class ShortTermMemory:
    """基于 InMemoryChatMessageHistory 的短期记忆，支持 token 裁剪。"""

    def __init__(self, max_tokens: int = 28000):
        self._store = InMemoryChatMessageHistory()
        self._max_tokens = max_tokens
        try:
            self._encoding = tiktoken.encoding_for_model("gpt-4")
        except Exception:
            self._encoding = tiktoken.get_encoding("cl100k_base")

    @property
    def messages(self) -> list[BaseMessage]:
        return self._store.messages

    def add_message(self, message: BaseMessage) -> None:
        self._store.add_message(message)
        self._trim()

    def add_messages(self, messages: list[BaseMessage]) -> None:
        for m in messages:
            self._store.add_message(m)
        self._trim()

    def _trim(self) -> None:
        msgs = self._store.messages
        while self._count_tokens(msgs) > self._max_tokens and len(msgs) > 2:
            msgs.pop(0)

    def _count_tokens(self, messages: list[BaseMessage]) -> int:
        total = 0
        for m in messages:
            content = m.content if isinstance(m.content, str) else str(m.content)
            total += len(self._encoding.encode(content))
            total += 4
        return total

    def clear(self) -> None:
        self._store.clear()

    def to_dict_list(self) -> list[dict[str, Any]]:
        return [{"role": m.type, "content": m.content} for m in self._store.messages]

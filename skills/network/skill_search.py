"""Skill: smart_search — 智能网络搜索。"""

from pydantic import BaseModel, Field

from skills.base import SkillBase, format_error, format_success


class SmartSearchInput(BaseModel):
    get_doc: bool = Field(default=False, description="设为 true 以获取使用说明和领域知识")
    query: str = Field(default="", description="搜索关键词")
    site: str = Field(default="", description="限制搜索的网站域名")
    filetype: str = Field(default="", description="限制搜索的文件类型")


class SmartSearchSkill(SkillBase):
    name: str = "smart_search"
    description: str = (
        "执行智能网络搜索，返回标题/URL/摘要。支持限定网站和文件类型。"
        "如需全文内容，请用搜索结果中的 URL 单独调用 scrape_webpage。"
        "★ 首次使用先 get_doc=true。"
    )
    args_schema: type[BaseModel] = SmartSearchInput

    def _run(
        self,
        get_doc: bool = False,
        query: str = "",
        site: str = "",
        filetype: str = "",
    ) -> str:
        if get_doc:
            return self._load_doc()
        if not query:
            return format_error("query 不能为空")

        try:
            result = self.client.uapi.zhi_neng_sou_suo.post_search_aggregate(
                query=query,
                site=site,
                filetype=filetype,
            )
            return format_success(result)
        except Exception as e:
            return format_error(f"搜索失败: {e}")

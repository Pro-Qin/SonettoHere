"""Skill: scrape_webpage — 网页内容抓取。"""

from pydantic import BaseModel, Field

from skills.base import SkillBase, format_error, format_success


class ScraperInput(BaseModel):
    get_doc: bool = Field(default=False, description="设为 true 以获取使用说明和领域知识")
    url: str = Field(default="", description="目标网页 URL，须以 http:// 或 https:// 开头")


class WebScraperSkill(SkillBase):
    name: str = "scrape_webpage"
    description: str = (
        "抓取指定网页的标题和正文内容。适合全文阅读搜索结果中的页面。"
        "★ 首次使用先 get_doc=true。"
    )
    args_schema: type[BaseModel] = ScraperInput

    def _run(self, get_doc: bool = False, url: str = "") -> str:
        if get_doc:
            return self._load_doc()
        if not url:
            return format_error("url 不能为空")

        try:
            from bs4 import BeautifulSoup

            headers = {
                "User-Agent": (
                    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                    "AppleWebKit/537.36 (KHTML, like Gecko) "
                    "Chrome/120.0.0.0 Safari/537.36"
                )
            }
            resp = self.client._session.get(url, headers=headers, timeout=10)
            resp.raise_for_status()
            resp.encoding = resp.apparent_encoding

            soup = BeautifulSoup(resp.text, "html.parser")

            title = soup.title.get_text(strip=True) if soup.title else ""

            for element in soup(["script", "style", "nav", "header", "footer", "aside"]):
                element.decompose()

            content = ""
            for selector in [
                "article", "main", ".content", "#content",
                ".post-content", ".article-content",
                'div[class*="content"]', 'div[id*="content"]',
            ]:
                el = soup.select_one(selector)
                if el:
                    content = el.get_text(separator="\n", strip=True)
                    break

            if not content:
                paragraphs = soup.find_all("p")
                content = "\n".join(
                    p.get_text(strip=True) for p in paragraphs if p.get_text(strip=True)
                )

            if not content and soup.body:
                content = soup.body.get_text(separator="\n", strip=True)

            lines = [line.strip() for line in content.split("\n") if line.strip()]
            content = "\n".join(lines)

            if len(content) > 5000:
                content = content[:5000] + "\n\n[内容已截断]"

            return format_success({"title": title, "content": content, "url": url})
        except Exception as e:
            return format_error(f"网页抓取失败: {e}")

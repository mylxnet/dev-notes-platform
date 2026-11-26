"""主题按钮断点回归检查（v1.4.8）。

用法（WSL 内已装 playwright + chromium）：
    BASE=http://127.0.0.1:18080 python3 test/theme_button_check.py

需要顺便检查后台页面（会登录）时再给凭据：
    BASE=http://127.0.0.1:5173 ADMIN_USER=admin ADMIN_PASS=****** python3 test/theme_button_check.py

检查项：
- 390×900：按钮 44×44、圆角 50%、无文字、圆点在按钮内水平居中、title 非空且随点击变化；
- 点击后 html[data-theme] 变为 light（默认跟随系统且系统为浅色时）；
- 各页面滚到底时，按钮与下拉/输入/按钮/链接的真实矩形相交面积必须为 0；
- 1280×900：显示长文案「浅色主题」且有 title；
- 拦截业务 JS 后 html[data-theme] 仍等于预期值，证明首屏 theme-init.js 生效。

输出 JSON：{"errors": [...控制台报错...], "steps": [[名称, 值], ...]}。
"""

import json
import os
from playwright.sync_api import sync_playwright

BASE = os.environ.get("BASE", "http://127.0.0.1:5173")
ADMIN_USER = os.environ.get("ADMIN_USER", "")
ADMIN_PASS = os.environ.get("ADMIN_PASS", "")
SHOT_DIR = os.environ.get("SHOT_DIR", "")

res = {"errors": [], "steps": []}

PUBLIC_PAGES = ["/", "/projects", "/about"]
ADMIN_PAGES = ["/admin/notes", "/admin/notes/new", "/admin/projects", "/admin/settings", "/admin/account"]

JS_COLLECT = """() => {
  return [...document.querySelectorAll('select,input,textarea,button:not(.theme-float),a[href]')]
    .filter(el => { const r = el.getBoundingClientRect();
      return r.width > 4 && r.height > 4 && r.bottom > 0 && r.top < innerHeight; })
    .map(el => { const r = el.getBoundingClientRect();
      return { tag: el.tagName.toLowerCase(),
               label: (el.id || el.className || el.getAttribute('name') || el.textContent || '').toString().trim().slice(0, 24),
               x: r.x, y: r.y, width: r.width, height: r.height };
    });
}"""


def step(name, value):
    res["steps"].append([name, value])


def intersection(a, b):
    """真实矩形相交面积：只有横竖都重叠才算遮挡。"""
    w = min(a["x"] + a["width"], b["x"] + b["width"]) - max(a["x"], b["x"])
    h = min(a["y"] + a["height"], b["y"] + b["height"]) - max(a["y"], b["y"])
    return max(0, w) * max(0, h)


def overlaps(page):
    button = page.evaluate(
        "() => { const r = document.querySelector('.theme-float').getBoundingClientRect();"
        "return { x: r.x, y: r.y, width: r.width, height: r.height }; }"
    )
    hits = []
    for box in page.evaluate(JS_COLLECT):
        area = intersection(button, box)
        if area > 1:
            hits.append([box["tag"], box["label"], round(area),
                         [round(box["x"]), round(box["y"]), round(box["width"]), round(box["height"])]])
    return hits


def shot(page, name, full=False):
    if SHOT_DIR:
        page.screenshot(path=os.path.join(SHOT_DIR, name), full_page=full)


def scroll_and_check(page, path, tag):
    page.goto(BASE + path, wait_until="networkidle")
    page.wait_for_timeout(600)
    page.evaluate("() => window.scrollTo(0, document.body.scrollHeight)")
    page.wait_for_timeout(500)
    step(f"{tag} {path} 滚到底遮挡", overlaps(page))
    step(f"{tag} {path} 横向溢出", page.evaluate(
        "() => document.documentElement.scrollWidth - document.documentElement.clientWidth"))


with sync_playwright() as p:
    browser = p.chromium.launch()
    ctx = browser.new_context(viewport={"width": 390, "height": 900}, color_scheme="light")
    page = ctx.new_page()
    page.on("console", lambda m: res["errors"].append(m.text) if m.type == "error" else None)
    page.on("pageerror", lambda e: res["errors"].append(str(e)))
    try:
        page.goto(BASE + "/", wait_until="networkidle")
        page.wait_for_timeout(700)
        box = page.locator(".theme-float").bounding_box()
        step("390 按钮宽x高", [round(box["width"]), round(box["height"])])
        step("390 按钮文本长度", len(page.locator(".theme-float").inner_text().strip()))
        step("390 圆角", page.evaluate(
            "() => getComputedStyle(document.querySelector('.theme-float')).borderRadius"))
        step("390 title", page.get_attribute(".theme-float", "title"))
        step("390 圆点水平居中偏差", page.evaluate(
            """() => { const b = document.querySelector('.theme-float').getBoundingClientRect();
                       const d = document.querySelector('.theme-dot').getBoundingClientRect();
                       return Math.round(Math.abs((d.x + d.width / 2) - (b.x + b.width / 2)) * 10) / 10; }"""))

        page.click(".theme-float")
        page.wait_for_timeout(250)
        step("390 点一次后 title", page.get_attribute(".theme-float", "title"))
        step("390 点一次后主题", page.get_attribute("html", "data-theme"))
        shot(page, "theme-390-home.png")

        for path in PUBLIC_PAGES:
            scroll_and_check(page, path, "390 公开")

        if ADMIN_USER and ADMIN_PASS:
            page.goto(BASE + "/admin/login", wait_until="networkidle")
            page.fill("input[type=text]", ADMIN_USER)
            page.fill("input[type=password]", ADMIN_PASS)
            page.click("button[type=submit]")
            page.wait_for_url("**/admin/notes", timeout=15000)
            for path in ADMIN_PAGES:
                scroll_and_check(page, path, "390 后台")
            page.goto(BASE + "/admin/settings", wait_until="networkidle")
            page.wait_for_timeout(600)
            shot(page, "theme-390-settings.png")
        else:
            step("390 后台检查", "跳过（未提供 ADMIN_USER / ADMIN_PASS）")

        page.goto(BASE + "/admin/login", wait_until="networkidle")
        page.set_viewport_size({"width": 1280, "height": 900})
        page.wait_for_timeout(500)
        box = page.locator(".theme-float").bounding_box()
        step("1280 按钮宽x高", [round(box["width"]), round(box["height"])])
        step("1280 长标签可见", page.locator(".theme-text").is_visible())
        step("1280 按钮文案", page.locator(".theme-float").inner_text().replace("\n", "").strip())
        step("1280 title", page.get_attribute(".theme-float", "title"))
        shot(page, "theme-1280.png")

        # 拦掉业务 JS，只留 theme-init.js：主题已生效说明首屏初始化脚本真的在 CSP 下执行了。
        page.evaluate("() => localStorage.setItem('dev-notes-theme-mode', 'dark')")
        page.route("**/assets/*.js", lambda r: r.abort())
        page.route("**/src/main.ts", lambda r: r.abort())
        before = len(res["errors"])
        page.goto(BASE + "/", wait_until="domcontentloaded")
        page.wait_for_timeout(800)
        step("拦截业务 JS 后首屏主题", page.get_attribute("html", "data-theme"))
        step("仅初始化脚本时的额外报错数", len(res["errors"]) - before)
    except Exception as exc:
        step("异常", str(exc).split("\n")[0])
    finally:
        browser.close()

print(json.dumps(res, ensure_ascii=False, indent=1))

"""数据备份页端到端检查（v1.4.10）。

用法（WSL 内已装 playwright + chromium）：
    BASE=http://127.0.0.1:18080 ADMIN_USER=admin ADMIN_PASS=****** python3 test/backup_page_check.py

检查项：
- 后台侧栏有「数据备份」入口，页面能打开；
- 统计卡显示数据库/图片/快照三块；
- 点「生成备份包」真的触发下载，且文件名是 .tar.gz；
- 把下载的包重新上传，预检显示「预检通过」；
- 输入 RESTORE 后执行恢复，页面显示「恢复完成」；
- 全程收集控制台报错。

输出 JSON：{"errors": [...], "steps": [[名称, 值], ...]}。
"""

import json
import os

from playwright.sync_api import sync_playwright

BASE = os.environ.get("BASE", "http://127.0.0.1:18080")
ADMIN_USER = os.environ.get("ADMIN_USER", "admin")
ADMIN_PASS = os.environ.get("ADMIN_PASS", "")
DOWNLOAD_PATH = os.environ.get("DOWNLOAD_PATH", "/tmp/backup-e2e.tar.gz")
SHOT_PATH = os.environ.get("SHOT_PATH", "")

result = {"errors": [], "steps": []}


def step(name, value):
    result["steps"].append([name, str(value)])


with sync_playwright() as playwright:
    browser = playwright.chromium.launch()
    context = browser.new_context(viewport={"width": 1280, "height": 900}, accept_downloads=True)
    page = context.new_page()
    page.on("console", lambda message: result["errors"].append(message.text) if message.type == "error" else None)
    page.on("pageerror", lambda error: result["errors"].append(str(error)))

    page.goto(f"{BASE}/admin/login", wait_until="networkidle")
    page.fill("input[type=text]", ADMIN_USER)
    page.fill("input[type=password]", ADMIN_PASS)
    page.click("button[type=submit]")
    page.wait_for_url("**/admin/**", timeout=10000)
    step("登录后地址", page.url)

    page.click("nav.admin-nav >> text=数据备份")
    page.wait_for_url("**/admin/backup")
    page.wait_for_selector(".backup-stats", timeout=10000)
    step("侧栏入口", page.url)
    step("统计卡数量", page.locator(".backup-stat").count())
    step("页面标题", page.locator("h1").first.text_content())

    with page.expect_download(timeout=90000) as download_info:
        page.click(".admin-heading >> text=生成备份包")
    download = download_info.value
    download.save_as(DOWNLOAD_PATH)
    step("下载文件名", download.suggested_filename)
    step("下载大小", os.path.getsize(DOWNLOAD_PATH))
    page.wait_for_selector("text=已生成", timeout=10000)

    page.set_input_files("input.backup-file", DOWNLOAD_PATH)
    page.wait_for_selector("text=预检通过", timeout=90000)
    step("预检结论", page.locator(".backup-picked .tag").first.text_content())

    page.fill(".backup-confirm input", "RESTORE")
    page.click("button:has-text('执行恢复')")
    page.wait_for_selector("text=恢复完成", timeout=180000)
    step("恢复结果", page.locator(".backup-result").last.text_content().strip()[:120])

    if SHOT_PATH:
        page.screenshot(path=SHOT_PATH, full_page=True)
        step("截图", SHOT_PATH)

    browser.close()

print(json.dumps(result, ensure_ascii=False, indent=2))

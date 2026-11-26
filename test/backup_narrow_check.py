"""备份页窄屏与溢出检查（v1.4.11）。

用法（WSL 内已装 playwright + chromium）：
    BASE=http://127.0.0.1:18080 ADMIN_USER=admin ADMIN_PASS=****** python3 test/backup_narrow_check.py

检查项：390×900 下后台「数据备份」页横向溢出为 0、统计卡单列、按钮尺寸不因换行变形、控制台无报错。
"""

import json
import os

from playwright.sync_api import sync_playwright

BASE = os.environ.get("BASE", "http://127.0.0.1:18080")
USER = os.environ.get("ADMIN_USER", "admin")
PASSWORD = os.environ.get("ADMIN_PASS", "")
SHOT = os.environ.get("SHOT_PATH", "/tmp/backup-narrow.png")

out = {}
with sync_playwright() as playwright:
    browser = playwright.chromium.launch()
    context = browser.new_context(viewport={"width": 390, "height": 900})
    page = context.new_page()
    errors = []
    page.on("console", lambda message: errors.append(message.text) if message.type == "error" else None)
    page.on("pageerror", lambda error: errors.append(str(error)))

    page.goto(f"{BASE}/admin/login", wait_until="networkidle")
    page.fill("input[type=text]", USER)
    page.fill("input[type=password]", PASSWORD)
    page.click("button[type=submit]")
    page.wait_for_selector("nav.admin-nav", timeout=10000)

    page.goto(f"{BASE}/admin/backup", wait_until="networkidle")
    page.wait_for_selector(".backup-stats", timeout=10000)
    page.wait_for_timeout(600)

    out["overflow"] = page.evaluate(
        "() => ({ scrollWidth: document.documentElement.scrollWidth, clientWidth: document.documentElement.clientWidth })"
    )
    out["statsColumns"] = page.evaluate(
        "() => getComputedStyle(document.querySelector('.backup-stats')).gridTemplateColumns"
    )
    out["buttons"] = page.evaluate(
        "() => [...document.querySelectorAll('.backup-panel button, .admin-heading button')].map(el => ({ text: el.textContent.trim(), width: Math.round(el.getBoundingClientRect().width), height: Math.round(el.getBoundingClientRect().height) }))"
    )
    page.screenshot(path=SHOT, full_page=True)
    out["errors"] = errors
    browser.close()

print(json.dumps(out, ensure_ascii=False, indent=2))

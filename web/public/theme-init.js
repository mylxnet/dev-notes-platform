// 首屏渲染前确定主题，避免浅色闪一下再切深色（水合后由 ThemeSwitch 接管）。
// 必须放在独立文件里：生产环境 CSP 为 script-src 'self'，内联脚本会被浏览器拦掉。
;(function () {
  var mode = 'system'
  try {
    var saved = localStorage.getItem('dev-notes-theme-mode')
    if (saved === 'system' || saved === 'light' || saved === 'dark') {
      mode = saved
    } else {
      var legacy = localStorage.getItem('dev-notes-theme')
      if (legacy === 'light' || legacy === 'dark') mode = legacy
    }
  } catch (e) {
    // 隐私模式下读不到存储，按跟随系统处理。
  }
  var dark =
    mode === 'dark' ||
    (mode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.dataset.theme = dark ? 'dark' : 'light'
})()

// 站点公开信息的共享状态：整站只拉一次，后台保存后主动刷新，页脚即时跟随。
import { ref } from 'vue'
import { fetchSite } from './api'

export const siteICPNumber = ref('')
let loaded = false

export async function loadSiteInfo(force = false): Promise<void> {
  if (loaded && !force) return
  try {
    const info = await fetchSite()
    siteICPNumber.value = info.icp_number
    loaded = true
  } catch {
    // 公开接口失败时页脚继续显示占位文案，不阻塞页面主体。
  }
}

// 站点公开信息的共享状态：整站只拉一次，后台保存后主动刷新，页脚与首页推荐位即时跟随。
import { ref } from 'vue'
import { fetchSite, type FeaturedNote } from './api'

export const siteICPNumber = ref('')
/** 首页推荐面板的笔记，由后台「站点设置」维护，未配置时为空数组。 */
export const siteFeaturedNotes = ref<FeaturedNote[]>([])
let loaded = false

export async function loadSiteInfo(force = false): Promise<void> {
  if (loaded && !force) return
  try {
    const info = await fetchSite()
    siteICPNumber.value = info.icp_number
    siteFeaturedNotes.value = info.featured_notes ?? []
    loaded = true
  } catch {
    // 公开接口失败时页脚继续显示占位文案、首页不显示推荐面板，不阻塞页面主体。
  }
}

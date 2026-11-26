<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchSettings, listNotes, updateSettings } from '../../admin'
import { ApiError, errorMessageOf, formatDate, type Note } from '../../api'
import { loadSiteInfo } from '../../site'

const icpNumber = ref('')
/** 三个推荐位的 slug，空串表示该位不设置。 */
const featuredSlugs = ref<string[]>(['', '', ''])
/** 可选的已发布笔记，按发布时间倒序。 */
const noteOptions = ref<Note[]>([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')

/** 与服务端一致：可留空清空，填写时 5–100 位汉字/字母/数字/连字符。 */
const valid = computed(
  () => icpNumber.value === '' || /^[\u4e00-\u9fa5A-Za-z0-9-]{5,100}$/.test(icpNumber.value)
)

/** 某个推荐位的候选项：排除其它位已经选中的笔记，避免重复。 */
function optionsFor(index: number): Note[] {
  const taken = featuredSlugs.value.filter((slug, position) => position !== index && slug)
  return noteOptions.value.filter((note) => !taken.includes(note.slug))
}

/** 同一篇笔记只能占一个推荐位：选到别处正在用的，先把那一处清空。 */
function onSlotChange(index: number, event: Event) {
  const value = (event.target as HTMLSelectElement).value
  const next = [...featuredSlugs.value]
  next.forEach((slug, position) => {
    if (slug === value && position !== index) next[position] = ''
  })
  next[index] = value
  featuredSlugs.value = next
  success.value = ''
}

onMounted(async () => {
  try {
    const [settings, notes] = await Promise.all([
      fetchSettings(),
      listNotes({ status: 'published', pageSize: 100 })
    ])
    icpNumber.value = settings.icp_number
    noteOptions.value = notes.items
    const slots = ['', '', '']
    settings.featured_slugs.slice(0, 3).forEach((slug, index) => {
      slots[index] = slug
    })
    featuredSlugs.value = slots
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    loading.value = false
  }
})

async function submit() {
  error.value = ''
  success.value = ''
  if (!valid.value) {
    error.value = '备案号需为 5–100 个汉字、字母、数字或连字符，例如 甘ICP备12345678号-1；留空可恢复占位文案。'
    return
  }

  saving.value = true
  try {
    const saved = await updateSettings({
      icp_number: icpNumber.value.trim(),
      featured_slugs: featuredSlugs.value.filter(Boolean)
    })
    icpNumber.value = saved.icp_number
    const slots = ['', '', '']
    saved.featured_slugs.slice(0, 3).forEach((slug, index) => {
      slots[index] = slug
    })
    featuredSlugs.value = slots
    await loadSiteInfo(true)
    const count = saved.featured_slugs.length
    success.value = count
      ? `设置已保存：首页展示 ${count} 条推荐笔记，页脚备案号已更新。`
      : '设置已保存：推荐位为空，首页隐藏推荐面板。'
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '保存失败，请稍后重试。'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>站点设置</h1>
        <p>维护前台页脚备案号与首页推荐位；推荐位留空时首页不显示该面板。</p>
      </div>
    </div>

    <div class="callout">备案号只在页脚展示，不做外部跳转；推荐位只能选已发布的笔记，最多 3 篇。</div>

    <div class="table-wrap" style="padding: 24px; max-width: 620px">
      <form class="form-grid" novalidate @submit.prevent="submit">
        <label>
          ICP 备案号
          <input
            v-model="icpNumber"
            type="text"
            placeholder="例如 甘ICP备12345678号-1"
            aria-label="ICP 备案号"
            :disabled="loading || saving"
          />
        </label>
        <span class="section-subtitle">可留空；填写时限 5–100 位汉字、字母、数字或连字符。</span>

        <div class="settings-divider" role="separator"></div>

        <div class="settings-group">
          <strong>推荐笔记</strong>
          <span class="section-subtitle">
            显示在首页右上角面板，按推荐位 1→2→3 的顺序展示；留空即少显示，全部留空则隐藏该面板。
          </span>
        </div>

        <label v-for="(slot, index) in featuredSlugs" :key="index">
          推荐位 {{ index + 1 }}
          <select
            :value="slot"
            :disabled="loading || saving"
            :aria-label="`推荐位 ${index + 1}`"
            @change="onSlotChange(index, $event)"
          >
            <option value="">（不设置）</option>
            <option v-for="note in optionsFor(index)" :key="note.slug" :value="note.slug">
              {{ note.title }} · {{ formatDate(note.published_at) }}
            </option>
          </select>
        </label>
        <span class="section-subtitle">
          选项来自已发布笔记，按发布时间倒序；同一篇不能重复选，保存时服务端会再校验一次是否已发布。
        </span>

        <div v-if="error" class="error-banner" role="alert" style="margin-bottom: 0">{{ error }}</div>
        <div v-if="success" class="callout" role="status">{{ success }}</div>
        <button class="btn btn-primary" type="submit" :disabled="loading || saving">
          {{ saving ? '保存中…' : '保存设置' }}
        </button>
      </form>
    </div>
  </div>
</template>

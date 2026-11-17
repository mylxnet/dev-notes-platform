<script setup lang="ts">
// 通用 Markdown 编辑器：工具栏 + 实时分栏预览 + 站内图片上传。
// 正文仍以 Markdown 原文交给父组件保存，预览复用服务端渲染接口，保证与前台一致。
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { previewMarkdown, uploadMedia } from '../admin'
import { errorMessageOf } from '../api'

const props = withDefaults(
  defineProps<{
    modelValue: string
    placeholder?: string
    minHeight?: number
    disabled?: boolean
  }>(),
  {
    placeholder: '支持标题、列表、代码块、表格；不允许原始 HTML 与外链图片',
    minHeight: 340,
    disabled: false
  }
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

type Mode = 'split' | 'edit' | 'preview'

const textarea = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const mode = ref<Mode>('split')
const previewHtml = ref('')
const previewLoading = ref(false)
const previewError = ref('')
const uploadError = ref('')
const uploading = ref(false)

let previewTimer: number | undefined
let requestId = 0

const modes: { value: Mode; label: string }[] = [
  { value: 'split', label: '分栏' },
  { value: 'edit', label: '仅编辑' },
  { value: 'preview', label: '仅预览' }
]

const charCount = computed(() => props.modelValue.length)

/** 预览防抖，避免每敲一个字都打接口；只应用最后一次响应。 */
function schedulePreview() {
  window.clearTimeout(previewTimer)
  previewTimer = window.setTimeout(async () => {
    const id = ++requestId
    if (!props.modelValue.trim()) {
      previewHtml.value = ''
      previewError.value = ''
      previewLoading.value = false
      return
    }
    previewLoading.value = true
    try {
      const html = await previewMarkdown(props.modelValue)
      if (id === requestId) {
        previewHtml.value = html
        previewError.value = ''
      }
    } catch (err) {
      if (id === requestId) previewError.value = errorMessageOf(err)
    } finally {
      if (id === requestId) previewLoading.value = false
    }
  }, 400)
}

watch(() => props.modelValue, schedulePreview, { immediate: true })

onBeforeUnmount(() => {
  window.clearTimeout(previewTimer)
  requestId += 1
})

/** 在光标处替换或插入文本，并可指定插入后选中的区间。 */
function applyInsert(text: string, select?: [number, number]) {
  const el = textarea.value
  if (!el || props.disabled) return
  const start = el.selectionStart
  const end = el.selectionEnd
  emit('update:modelValue', props.modelValue.slice(0, start) + text + props.modelValue.slice(end))
  void nextTick(() => {
    el.focus()
    if (select) {
      el.setSelectionRange(select[0], select[1])
    } else {
      const caret = start + text.length
      el.setSelectionRange(caret, caret)
    }
  })
}

const wrappers: Record<string, [string, string]> = {
  h2: ['\n## ', '\n'],
  h3: ['\n### ', '\n'],
  bold: ['**', '**'],
  italic: ['*', '*'],
  quote: ['\n> ', '\n'],
  ul: ['\n- ', ''],
  ol: ['\n1. ', ''],
  link: ['[链接文字](https://example.com)', ''],
  code: ['`', '`'],
  codeblock: ['\n```\n', '\n```\n'],
  hr: ['\n---\n', '']
}

function onToolbarClick(action: string) {
  if (action === 'image') {
    fileInput.value?.click()
    return
  }
  const [before, after] = wrappers[action]
  if (!before) return

  const el = textarea.value
  const start = el?.selectionStart ?? 0
  const end = el?.selectionEnd ?? 0
  const selected = props.modelValue.slice(start, end)

  if (action === 'link') {
    applyInsert(before, [start + 1, start + 5])
    return
  }
  applyInsert(before + selected + after, selected ? undefined : [start + before.length, start + before.length])
}

function showUploadError(message: string) {
  uploadError.value = message
}

/** 图片上传复用后台接口：校验 → 上传 → 插入语法 → 选中 alt 方便改图注。 */
async function onImageSelected() {
  const file = fileInput.value?.files?.[0]
  if (fileInput.value) fileInput.value.value = ''
  if (!file) return

  uploadError.value = ''
  if (!/\.(png|jpe?g|webp)$/i.test(file.name)) {
    showUploadError('只支持 PNG、JPG、WEBP 图片')
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    showUploadError('图片不能超过 2 MB')
    return
  }

  uploading.value = true
  try {
    const result = await uploadMedia(file)
    const alt = file.name.replace(/\.[^.]+$/, '').replace(/[\[\]]/g, '')
    const insert = `![${alt}](${result.path})`
    const start = textarea.value?.selectionStart ?? props.modelValue.length
    const end = textarea.value?.selectionEnd ?? props.modelValue.length
    applyInsert(insert, [start + 2, start + 2 + alt.length])
  } catch (err) {
    showUploadError(errorMessageOf(err))
  } finally {
    uploading.value = false
  }
}

function setMode(next: Mode) {
  mode.value = next
}
</script>

<template>
  <div class="md-editor" :class="{ 'md-editor-disabled': disabled }">
    <input
      ref="fileInput"
      type="file"
      accept="image/png,image/jpeg,image/webp"
      hidden
      @change="onImageSelected"
    />

    <div class="md-toolbar">
      <div class="md-tools">
        <button type="button" title="二级标题" :disabled="disabled" @click="onToolbarClick('h2')">H2</button>
        <button type="button" title="三级标题" :disabled="disabled" @click="onToolbarClick('h3')">H3</button>
        <span class="md-sep" />
        <button type="button" title="粗体" :disabled="disabled" @click="onToolbarClick('bold')"><b>B</b></button>
        <button type="button" title="斜体" :disabled="disabled" @click="onToolbarClick('italic')"><i>I</i></button>
        <span class="md-sep" />
        <button type="button" title="引用" :disabled="disabled" @click="onToolbarClick('quote')">引用</button>
        <button type="button" title="无序列表" :disabled="disabled" @click="onToolbarClick('ul')">列表</button>
        <button type="button" title="有序列表" :disabled="disabled" @click="onToolbarClick('ol')">1. 列表</button>
        <span class="md-sep" />
        <button type="button" title="链接" :disabled="disabled" @click="onToolbarClick('link')">链接</button>
        <button
          type="button"
          title="上传站内图片"
          :disabled="disabled || uploading"
          @click="onToolbarClick('image')"
        >
          {{ uploading ? '上传中…' : '图片' }}
        </button>
        <span v-if="uploadError" class="md-upload-error" role="alert">{{ uploadError }}</span>
        <span class="md-sep" />
        <button type="button" title="行内代码" :disabled="disabled" @click="onToolbarClick('code')">&lt;/&gt;</button>
        <button type="button" title="代码块" :disabled="disabled" @click="onToolbarClick('codeblock')">{ }</button>
        <button type="button" title="分隔线" :disabled="disabled" @click="onToolbarClick('hr')">—</button>
      </div>
      <div class="md-modes">
        <button
          v-for="item in modes"
          :key="item.value"
          type="button"
          :class="{ active: mode === item.value }"
          @click="setMode(item.value)"
        >
          {{ item.label }}
        </button>
      </div>
    </div>

    <div class="md-split" :class="`md-mode-${mode}`">
      <div v-show="mode !== 'preview'" class="md-pane">
        <div class="md-pane-label">
          <span>Markdown 原文</span><span>{{ charCount }} 字符</span>
        </div>
        <textarea
          ref="textarea"
          :value="modelValue"
          :placeholder="placeholder"
          :disabled="disabled"
          spellcheck="false"
          :style="{ minHeight: `${minHeight}px` }"
          @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>
      <div v-show="mode !== 'edit'" class="md-pane">
        <div class="md-pane-label">
          <span>实时预览（服务端渲染并净化）</span>
          <span v-if="previewLoading" class="md-preview-state">渲染中…</span>
        </div>
        <div class="md-preview article-markdown" v-html="previewHtml" />
        <div v-if="previewError" class="md-upload-error" role="alert">{{ previewError }}（预览失败，正文不受影响）</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.md-editor {
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--surface);
  overflow: hidden;
}
.md-editor-disabled { opacity: 0.65; }
.md-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
}
.md-tools { display: flex; align-items: center; gap: 3px; flex-wrap: wrap; }
.md-tools button,
.md-modes button {
  border: 1px solid transparent;
  background: transparent;
  color: var(--text);
  border-radius: 7px;
  height: 30px;
  min-width: 32px;
  padding: 0 8px;
  font-size: 0.82rem;
  cursor: pointer;
}
.md-tools button:hover:not(:disabled) { background: var(--surface-2); border-color: var(--border); }
.md-tools button:disabled { opacity: 0.5; cursor: not-allowed; }
.md-sep { width: 1px; height: 18px; background: var(--border); margin: 0 3px; }
.md-modes { display: flex; gap: 4px; }
.md-modes button { border-color: var(--border); color: var(--muted); }
.md-modes button.active { background: var(--surface-2); color: var(--text); border-color: var(--accent); }
.md-upload-error { color: var(--red); font-size: 0.78rem; }
.md-split { display: grid; grid-template-columns: 1fr 1fr; }
.md-split.md-mode-edit,
.md-split.md-mode-preview { grid-template-columns: 1fr; }
.md-pane { min-width: 0; padding: 12px; }
.md-pane + .md-pane { border-left: 1px solid var(--border); }
.md-pane-label {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  color: var(--muted);
  font-size: 0.76rem;
  margin-bottom: 8px;
}
.md-preview-state { color: var(--accent); }
.md-pane textarea {
  display: block;
  width: 100%;
  resize: vertical;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  background: var(--surface);
  color: var(--text);
  font-family: Consolas, "Cascadia Code", monospace;
  font-size: 0.82rem;
  line-height: 1.7;
  outline: none;
}
.md-pane textarea:focus { border-color: var(--accent); box-shadow: 0 0 0 3px var(--surface-2); }
.md-preview {
  min-height: 340px;
  border: 1px dashed var(--border);
  border-radius: 8px;
  padding: 12px 16px;
  background: var(--surface);
  overflow-wrap: anywhere;
}
@media (max-width: 760px) {
  .md-split { grid-template-columns: 1fr; }
  .md-pane + .md-pane { border-left: 0; border-top: 1px solid var(--border); }
  .md-toolbar { align-items: flex-start; }
}
</style>

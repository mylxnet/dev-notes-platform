<script setup lang="ts">
// 通用分页条：前台笔记列表与后台笔记/项目列表共用同一套页码规则。
import { computed } from 'vue'

const props = defineProps<{
  /** 当前页，从 1 开始。 */
  page: number
  /** 总页数，由调用方按每页条数与总数算出。 */
  totalPages: number
  /** 总条数，用于右侧「共 N 条」。 */
  total: number
  /** 条数单位：篇 / 个。 */
  unit: string
  label?: string
}>()

const emit = defineEmits<{ change: [target: number] }>()

/** 页码列表：7 页以内全列，更多时首尾 + 当前页左右各一页，其余折叠成省略号。 */
const pageNumbers = computed<(number | 'gap')[]>(() => {
  const last = props.totalPages
  if (last <= 7) return Array.from({ length: last }, (_, index) => index + 1)

  const current = props.page
  const items: (number | 'gap')[] = [1]
  const start = Math.max(2, current - 1)
  const end = Math.min(last - 1, current + 1)
  if (start > 2) items.push('gap')
  for (let index = start; index <= end; index += 1) items.push(index)
  if (end < last - 1) items.push('gap')
  items.push(last)
  return items
})

function goTo(target: number) {
  const clamped = Math.min(Math.max(1, target), props.totalPages)
  if (clamped === props.page) return
  emit('change', clamped)
}
</script>

<template>
  <nav v-if="totalPages > 1" class="pager" :aria-label="label ?? '分页'">
    <button class="pager-btn pager-nav" type="button" :disabled="page === 1" @click="goTo(page - 1)">
      上一页
    </button>
    <template v-for="(item, index) in pageNumbers" :key="`${item}-${index}`">
      <span v-if="item === 'gap'" class="pager-gap">…</span>
      <button
        v-else
        class="pager-btn"
        :class="{ active: item === page }"
        type="button"
        :aria-current="item === page ? 'page' : undefined"
        @click="goTo(item)"
      >
        {{ item }}
      </button>
    </template>
    <button
      class="pager-btn pager-nav"
      type="button"
      :disabled="page === totalPages"
      @click="goTo(page + 1)"
    >
      下一页
    </button>
    <span class="pager-info">第 {{ page }} / {{ totalPages }} 页 · 共 {{ total }} {{ unit }}</span>
  </nav>
</template>

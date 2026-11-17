<script setup lang="ts">
// 站点标识：折角笔记纸 + 字母 N。
// 保留原来「N」的识别度，再补一层「一页笔记」的语义；按 32×32 设计，放大缩小都不失真。
</script>

<template>
  <span class="brand-mark" aria-hidden="true">
    <svg viewBox="0 0 32 32" xmlns="http://www.w3.org/2000/svg" focusable="false">
      <rect width="32" height="32" rx="9" fill="var(--accent)" />
      <path
        d="M20.8 6.2H11.4A2.6 2.6 0 0 0 8.8 8.8v14.4a2.6 2.6 0 0 0 2.6 2.6h9.6a2.6 2.6 0 0 0 2.6-2.6V10.4L20.8 6.2Z"
        fill="#fffaf3"
      />
      <path d="M20.8 6.2 23.6 10.4H20.8Z" fill="color-mix(in srgb, var(--accent) 68%, #000)" />
      <path
        d="M12.6 21.2V13.4L19.4 21.2V13.4"
        fill="none"
        stroke="color-mix(in srgb, var(--accent) 78%, #141413)"
        stroke-width="2.1"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  </span>
</template>

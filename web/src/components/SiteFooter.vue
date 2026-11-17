<script setup lang="ts">
import { siteICPNumber } from '../site'
</script>

<template>
  <footer class="site-footer">
    <div class="container footer-inner">
      <span>{{ siteICPNumber || '甘ICP备xxxxxxxx号（备案通过后显示）' }}</span>
    </div>
  </footer>
</template>

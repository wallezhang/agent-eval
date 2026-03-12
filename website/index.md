---
layout: page
---

<script setup>
import { onMounted } from 'vue'

onMounted(() => {
  const lang = navigator.language || navigator.userLanguage || 'en'
  if (lang.startsWith('zh')) {
    window.location.href = '/zh/'
  } else {
    window.location.href = '/en/'
  }
})
</script>

<div style="display: flex; justify-content: center; align-items: center; min-height: 50vh; gap: 2rem;">
  <a href="/en/" style="font-size: 1.5rem; text-decoration: none;">English</a>
  <span style="color: var(--vp-c-divider);">|</span>
  <a href="/zh/" style="font-size: 1.5rem; text-decoration: none;">简体中文</a>
</div>

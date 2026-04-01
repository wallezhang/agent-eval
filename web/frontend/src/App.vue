<script setup lang="ts">
import { NMessageProvider, NDialogProvider, NConfigProvider } from 'naive-ui'
</script>

<template>
  <NConfigProvider>
    <NMessageProvider>
      <NDialogProvider>
        <router-view />
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

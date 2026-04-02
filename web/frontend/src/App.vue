<script setup lang="ts">
import { Toaster } from '@/components/ui/sonner'
</script>

<template>
  <Toaster position="top-right" :duration="3000" />
  <router-view />
</template>

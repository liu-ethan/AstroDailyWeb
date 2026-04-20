<script setup lang="ts">
import { RouterView } from 'vue-router'
import BaseToast from './components/BaseToast.vue'
</script>

<template>
  <div class="app-shell">
    <RouterView />
    <BaseToast />
  </div>
</template>

<style>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.router-view {
  flex: 1;
}
</style>

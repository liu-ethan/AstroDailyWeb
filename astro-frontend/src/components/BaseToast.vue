<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toastState } = useToast()
</script>

<template>
  <div v-if="toastState.visible" class="toast">{{ toastState.message }}</div>
</template>


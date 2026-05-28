<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  label: string;
  value: number;
  tone?: "default" | "success" | "warning" | "danger";
}>();

const width = computed(() => `${Math.max(0, Math.min(100, Number(props.value) || 0))}%`);
</script>

<template>
  <div class="metric-bar">
    <div class="metric-bar__meta">
      <span>{{ label }}</span>
      <span>{{ value }}%</span>
    </div>
    <div class="metric-bar__track">
      <div class="metric-bar__fill" :class="`metric-bar__fill--${tone ?? 'default'}`" :style="{ width }" />
    </div>
  </div>
</template>

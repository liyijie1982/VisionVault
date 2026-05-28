<script setup lang="ts">
import type { Component } from "vue";

defineProps<{
  label: string;
  value: string | number;
  hint?: string;
  icon?: Component;
}>();
</script>

<template>
  <div class="stat-card">
    <div class="stat-card__label">{{ label }}</div>
    <div class="stat-card__value-row">
      <div v-if="icon" class="stat-card__icon">
        <component :is="icon" />
      </div>
      <div class="stat-card__value">{{ value }}</div>
    </div>
    <div class="stat-card__hint">{{ hint }}</div>
  </div>
</template>

<style scoped>
.stat-card__value-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stat-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.08);
  color: var(--color-text-2);
  flex: 0 0 auto;
  font-size: 16px;
  margin-top: 12px;
}

.stat-card__value {
  line-height: 1;
}
</style>

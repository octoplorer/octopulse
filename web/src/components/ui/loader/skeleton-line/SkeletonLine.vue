<script setup lang="ts">
import { createReusableTemplate } from '@vueuse/core'

defineOptions({
  inheritAttrs: false,
})

const {
  minWidth = 30,
  maxWidth = 100,
  minDuration = 1.3,
  maxDuration = 1.7,
  minDelay = 0,
  maxDelay = 0.5,
  blockHeight,
} = defineProps<Props>()

interface Props {
  minWidth?: number
  maxWidth?: number
  minDuration?: number
  maxDuration?: number
  minDelay?: number
  maxDelay?: number
  blockHeight?: string | number
}

const width = getRandomWidth(minWidth, maxWidth)
const duration = getRandomFloat(minDuration, maxDuration)
const delay = getRandomFloat(minDelay, maxDelay)

const lineStyle = {
  '--skeleton-width': `${width}%`,
  '--shimmer-duration': `${duration}s`,
  '--shimmer-delay': `${delay}s`,
}

function getRandomWidth(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1) + min)
}

function getRandomFloat(min: number, max: number) {
  return (Math.random() * (max - min) + min).toFixed(2)
}

const [DefineLineTemplate, LineTemplate] = createReusableTemplate()
</script>

<template>
  <DefineLineTemplate>
    <div class="skeleton-line" :style="lineStyle" v-bind="$attrs" />
  </DefineLineTemplate>
  <template v-if="blockHeight !== undefined">
    <div flex items-center v-bind="$attrs">
      <LineTemplate />
    </div>
  </template>
  <template v-else>
    <LineTemplate />
  </template>
</template>

<style>
@layer base {
  .skeleton-line {
    position: relative;
    overflow: hidden;
    border-radius: 2px;
    height: 0.5rem;
    width: var(--skeleton-width);
    background-color: var(--color-fill);
  }

  .skeleton-line::after {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    animation: shimmer var(--shimmer-duration, 1.5s) var(--shimmer-delay, 0s) infinite ease-in-out;
    content: '';
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--color-fill-hover) 50%,
      transparent 100%
    );
  }
}
</style>

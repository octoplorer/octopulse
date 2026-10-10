<script setup lang="ts">
const props = withDefaults(defineProps<{ label?: string, step?: number }>(), { label: 'Resize column', step: 10 })
const emit = defineEmits<{ resize: [delta: number], resizeStart: [event: PointerEvent] }>()
function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight')
    return
  event.preventDefault()
  emit('resize', event.key === 'ArrowRight' ? props.step : -props.step)
}
</script>

<template>
  <button type="button" :aria-label="props.label" class="absolute inset-y-0 right-0 m-0 flex w-2.5 touch-none select-none items-center justify-center bg-base p-0 opacity-0 outline-none group-hover:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset" @pointerdown="emit('resizeStart', $event)" @keydown="onKeydown">
    <span class="h-5 w-0.5 rounded bg-hairline" aria-hidden="true" />
  </button>
</template>

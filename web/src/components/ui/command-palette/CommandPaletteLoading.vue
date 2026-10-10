<template>
  <div role="status" class="flex items-center justify-center gap-2 px-3 py-8 text-size-sm text-subtle">
    <span class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" /><slot>Loading…</slot>
  </div>
</template>

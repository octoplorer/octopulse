<script setup lang="ts">
defineProps<{ title: string, description?: string, eyebrow?: string }>()
</script>

<template>
  <header flex="~ justify-between items-center gap-20px" mb="28px" class="[@media(max-width:700px)]:items-start [@media(max-width:700px)]:flex-col [@media(max-width:700px)]:gap-15px [@media(max-width:700px)]:[&_h1]:text-24px [@container_workspace_(max-width:_700px)]:items-start [@container_workspace_(max-width:_700px)]:flex-col [@container_workspace_(max-width:_700px)]:gap-15px">
    <div class="[overflow-wrap:anywhere]" min-w="0">
      <p v-if="eyebrow" class="eyebrow" un-text="12px subtle" font="500" mb="8px">
        {{ eyebrow }}
      </p>
      <h1>{{ title }}</h1>
      <p v-if="description" un-text="13px subtle" mt="6px">
        {{ description }}
      </p>
    </div>
    <div class="[overflow-wrap:anywhere] [@media(max-width:700px)]:w-full [@media(max-width:700px)]:flex-wrap [@container_workspace_(max-width:_700px)]:w-full [@container_workspace_(max-width:_700px)]:flex-wrap" flex="~ items-center gap-9px wrap" min-w="0">
      <slot />
    </div>
  </header>
</template>

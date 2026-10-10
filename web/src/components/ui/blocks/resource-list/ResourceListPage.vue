<script setup lang="ts">
defineProps<{ title?: string, description?: string }>()
</script>

<template>
  <div class="min-h-full w-full bg-overlay text-default">
    <div class="mx-auto flex max-w-1400px flex-col gap-6 p-5 md:p-8 lg:px-10 lg:py-9">
      <header v-if="title || description || $slots.header" class="flex flex-col gap-2">
        <slot name="header">
          <div class="flex items-center gap-2">
            <slot name="icon" /><h1 v-if="title" class="text-size-3xl font-semibold tracking-tight">
              {{ title }}
            </h1>
          </div>
          <p v-if="description" class="text-size-lg text-subtle">
            {{ description }}
          </p>
        </slot>
      </header>
      <div class="flex flex-col-reverse gap-6 xl:flex-row xl:gap-8">
        <div class="min-w-0 grow">
          <slot />
        </div>
        <aside v-if="$slots.usage || $slots.additionalContent" class="flex w-full shrink-0 flex-col gap-4 xl:sticky xl:top-22 xl:h-fit xl:w-380px">
          <slot name="usage" /><slot name="additionalContent" />
        </aside>
      </div>
    </div>
  </div>
</template>

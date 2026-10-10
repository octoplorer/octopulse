<script setup lang="ts">
import { cva } from 'cva'
import { ClipboardText } from '../clipboard-text'
import EmptyDescription from './EmptyDescription.vue'
import EmptyTitle from './EmptyTitle.vue'

const props = withDefaults(defineProps<{ title?: string, description?: string, commandLine?: string, copyLabel?: string, copiedLabel?: string, size?: 'sm' | 'base' | 'lg' }>(), { size: 'base', copyLabel: 'Copy command', copiedLabel: 'Copied' })
const styles = cva({ base: 'empty-state flex w-full min-w-0 flex-col items-center rounded-xl border border-line bg-control text-center text-default [&>div]:max-w-prose', variants: { size: { sm: 'gap-4 px-6 py-8', base: 'gap-6 px-6 py-12 [@container_workspace_(min-width:_700px)]:px-10 [@container_workspace_(min-width:_700px)]:py-16', lg: 'gap-8 px-6 py-12 [@container_workspace_(min-width:_700px)]:px-12 [@container_workspace_(min-width:_700px)]:py-20' } }, defaultVariants: { size: 'base' } })
</script>

<template>
  <div :class="styles({ size: props.size })">
    <slot name="icon" />
    <template v-if="props.title">
      <div class="flex flex-col items-center gap-2.5">
        <EmptyTitle as="h2" :size="props.description ? 'lg' : 'base'" :variant="props.description ? 'heading' : 'secondary'">
          {{ props.title }}
        </EmptyTitle><EmptyDescription v-if="props.description">
          {{ props.description }}
        </EmptyDescription>
      </div>
    </template>
    <ClipboardText v-if="props.commandLine" :text="props.commandLine" :copy-label="props.copyLabel" :copied-label="props.copiedLabel" class="max-w-full">
      <span class="mr-2 select-none text-subtle" aria-hidden="true">$</span>{{ props.commandLine }}
    </ClipboardText>
    <slot />
    <slot name="actions" />
  </div>
</template>

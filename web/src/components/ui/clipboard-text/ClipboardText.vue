<script setup lang="ts">
import { Clipboard } from '@ark-ui/vue/clipboard'
import { Button } from '../button'

withDefaults(defineProps<{ text: string, textToCopy?: string, copyLabel?: string, copiedLabel?: string, inline?: boolean, size?: 'sm' | 'base' | 'lg' }>(), { copyLabel: 'Copy to clipboard', copiedLabel: 'Copied', size: 'base' })
const emit = defineEmits<{ copy: [] }>()
</script>

<template>
  <Clipboard.Root :model-value="textToCopy ?? text" @status-change="$event.copied && emit('copy')">
    <Clipboard.Context v-slot="clipboard">
      <Clipboard.Control class="flex min-w-0 items-center gap-2" :class="inline ? 'inline-flex' : 'rounded-lg border border-solid border-line bg-control pl-3 shadow-control'">
        <span class="min-w-0 flex-1 truncate font-mono text-size-sm"><slot>{{ text }}</slot></span>
        <Clipboard.Trigger as-child>
          <Button :size="size === 'lg' ? 'lg' : 'sm'" shape="square" variant="ghost" :aria-label="clipboard.copied ? copiedLabel : copyLabel">
            <span class="size-4" :class="[clipboard.copied ? 'i-lucide-check text-fg-success' : 'i-lucide-copy']" aria-hidden="true" />
          </Button>
        </Clipboard.Trigger>
      </Clipboard.Control>
      <span class="sr-only" aria-live="polite">{{ clipboard.copied ? copiedLabel : '' }}</span>
    </Clipboard.Context>
  </Clipboard.Root>
</template>

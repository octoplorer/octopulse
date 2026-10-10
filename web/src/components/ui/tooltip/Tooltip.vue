<script setup lang="ts">
import type { TooltipRootProps } from '@ark-ui/vue/tooltip'
import { useDialogContext } from '@ark-ui/vue/dialog'
import { Tooltip as ArkTooltip } from '@ark-ui/vue/tooltip'
import { useTooltipDefaults } from './context'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{ content?: string, placement?: NonNullable<TooltipRootProps['positioning']>['placement'], disabled?: boolean, openDelay?: number, closeDelay?: number, asChild?: boolean, portalled?: boolean, arrow?: boolean }>(), { placement: 'top', asChild: true, portalled: true, arrow: true })
const defaults = useTooltipDefaults()
const dialog = useDialogContext()
const open = defineModel<boolean | undefined>('open', { default: undefined })
</script>

<template>
  <ArkTooltip.Root v-model:open="open" :disabled="props.disabled" :open-delay="props.openDelay ?? defaults.openDelay ?? 400" :close-delay="props.closeDelay ?? defaults.closeDelay ?? 100" :positioning="{ placement: props.placement, gutter: 8, strategy: 'fixed' }">
    <ArkTooltip.Trigger :as-child="props.asChild" :class="props.asChild ? undefined : 'm-0 inline-flex h-auto min-h-0 cursor-default items-center border-none bg-transparent p-0 shadow-none'">
      <slot />
    </ArkTooltip.Trigger>
    <Teleport to="body" :disabled="!props.portalled || !!dialog">
      <ArkTooltip.Positioner class="z-250">
        <ArkTooltip.Content v-bind="$attrs" class="z-250 max-w-72 rounded-lg bg-contrast px-3 py-2 text-size-xs text-inverse shadow-panel">
          <ArkTooltip.Arrow v-if="props.arrow" class="[--arrow-size:8px] [--arrow-background:var(--color-contrast)]">
            <ArkTooltip.ArrowTip />
          </ArkTooltip.Arrow>
          <slot name="content">
            {{ props.content }}
          </slot>
        </ArkTooltip.Content>
      </ArkTooltip.Positioner>
    </Teleport>
  </ArkTooltip.Root>
</template>

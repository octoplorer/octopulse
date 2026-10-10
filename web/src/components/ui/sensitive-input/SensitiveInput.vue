<script setup lang="ts">
import { Clipboard as ArkClipboard } from '@ark-ui/vue/clipboard'
import { useFieldContext } from '@ark-ui/vue/field'
import { PasswordInput as ArkPassword, usePasswordInput } from '@ark-ui/vue/password-input'
import { cva } from 'cva'
import { computed, useAttrs, useId } from 'vue'
import { Input } from '../input'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  label?: string
  name?: string
  description?: string
  error?: string
  disabled?: boolean
  readOnly?: boolean
  required?: boolean
  copyable?: boolean
  size?: 'xs' | 'sm' | 'base' | 'lg'
  variant?: 'default' | 'error'
  placeholder?: string
  showLabel?: string
  hideLabel?: string
  copyLabel?: string
  copiedLabel?: string
}>(), { size: 'base', variant: 'default', copyable: true, showLabel: 'Show value', hideLabel: 'Hide value', copyLabel: 'Copy value', copiedLabel: 'Copied', disabled: undefined, readOnly: undefined, required: undefined })
const emit = defineEmits<{ copy: [] }>()
const value = defineModel<string>({ default: '' })
const visible = defineModel<boolean>('visible', { default: false })
const id = useId()
const attrs = useAttrs()
const field = useFieldContext()
const accessibleLabel = computed(() => typeof attrs['aria-label'] === 'string' ? attrs['aria-label'] : props.label ?? 'Sensitive value')
const descriptionId = `${id}-description`
const errorId = `${id}-error`
const sensitive = usePasswordInput(computed(() => ({ name: props.name, disabled: props.disabled, readOnly: props.readOnly, required: props.required, invalid: !!props.error || props.variant === 'error', visible: visible.value, onVisibilityChange: details => visible.value = details.visible })))
const describedBy = computed(() => [field?.value.ariaDescribedby, props.description && descriptionId, props.error && errorId].filter(Boolean).join(' ') || undefined)
const controlClasses = cva({
  base: 'ui-control-group relative flex min-w-0 items-center border-0 bg-control ring-1 ring-control-border outline-none focus-within:ring-2 focus-within:outline-2 focus-within:outline-solid focus-within:outline-focus',
  variants: {
    size: { xs: 'rounded-sm', sm: 'rounded-md', base: 'rounded-lg', lg: 'rounded-lg' },
    variant: { default: 'focus-within:ring-focus', error: 'ring-action-danger focus-within:ring-action-danger' },
  },
  defaultVariants: { size: 'base', variant: 'default' },
})
</script>

<template>
  <ArkPassword.RootProvider :value="sensitive" class="flex min-w-0 flex-col gap-1.5">
    <ArkPassword.Label v-if="label || $slots.label" class="text-size-base font-medium text-strong">
      <slot name="label">
        {{ label }}
      </slot>
    </ArkPassword.Label>
    <ArkPassword.Control :class="controlClasses({ size, variant: error ? 'error' : variant })">
      <Input v-bind="{ ...sensitive.getInputProps(), ...$attrs }" v-model="value" :size="size" :variant="error ? 'error' : variant" :placeholder="placeholder" :aria-describedby="describedBy" :aria-label="accessibleLabel" class="border-0! bg-transparent! shadow-none! outline-none! focus:ring-0! ring-0!" :class="copyable ? 'pr-18!' : 'pr-10!'" />
      <div class="absolute right-1 flex items-center gap-0.5">
        <button type="button" :disabled="sensitive.disabled" :aria-controls="sensitive.getInputProps().id" :aria-expanded="visible" :aria-label="visible ? hideLabel : showLabel" class="flex size-7 items-center justify-center rounded-md bg-transparent text-subtle outline-none hover:bg-tint focus:text-default focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand! disabled:opacity-50" @click="sensitive.toggleVisible">
          <span :class="visible ? 'i-lucide-eye-off' : 'i-lucide-eye'" class="size-4" aria-hidden="true" />
        </button>
        <ArkClipboard.Root v-if="copyable" :model-value="value" @status-change="$event.copied ? emit('copy') : undefined">
          <ArkClipboard.Context v-slot="clipboard">
            <ArkClipboard.Trigger :disabled="sensitive.disabled || !value" :aria-label="clipboard.copied ? copiedLabel : copyLabel" class="flex size-7 items-center justify-center rounded-md bg-transparent text-subtle outline-none hover:bg-tint focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand! disabled:opacity-50">
              <span :class="clipboard.copied ? 'i-lucide-check text-fg-success' : 'i-lucide-copy'" class="size-4" aria-hidden="true" />
            </ArkClipboard.Trigger>
            <span class="sr-only" role="status">{{ clipboard.copied ? copiedLabel : '' }}</span>
          </ArkClipboard.Context>
        </ArkClipboard.Root>
      </div>
    </ArkPassword.Control>
    <p v-if="description" :id="descriptionId" class="text-size-sm text-subtle">
      {{ description }}
    </p>
    <p v-if="error" :id="errorId" class="text-size-sm text-fg-danger" role="alert">
      {{ error }}
    </p>
  </ArkPassword.RootProvider>
</template>

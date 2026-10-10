<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'
import { computed, shallowRef, useId, watch } from 'vue'
import { Button } from '../button'
import { LayerDialog } from '../layer-dialog'
import { isDeleteConfirmed } from './confirmation'

const props = withDefaults(defineProps<{ resourceType: string, resourceName: string, isDeleting?: boolean, caseSensitive?: boolean, title?: string, description?: string, deleteButtonText?: string, cancelLabel?: string, confirmationLabel?: string, errorMessage?: string, size?: 'sm' | 'base', onDelete?: () => void | Promise<void> }>(), { caseSensitive: true, cancelLabel: 'Cancel' })
const emit = defineEmits<{ confirm: [] }>()
const open = defineModel<boolean>('open', { default: false })
const confirmation = shallowRef('')
const submitting = shallowRef(false)
const actionError = shallowRef<string>()
const inputId = useId()
const busy = computed(() => props.isDeleting || submitting.value)
const confirmed = computed(() => isDeleteConfirmed(confirmation.value, props.resourceName, props.caseSensitive))
watch(() => [open.value, props.resourceName], () => {
  confirmation.value = ''
  actionError.value = undefined
})
async function confirm() {
  if (!confirmed.value || busy.value)
    return
  emit('confirm')
  if (!props.onDelete)
    return
  submitting.value = true
  actionError.value = undefined
  try {
    await props.onDelete()
    open.value = false
  }
  catch (error) {
    actionError.value = error instanceof Error ? error.message : String(error)
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <LayerDialog v-model:open="open" alert :dismiss-disabled="busy" :title="props.title ?? `Delete ${props.resourceName}`" :size="props.size" :description="props.description">
    <div class="space-y-4">
      <p v-if="props.errorMessage || actionError" role="alert" class="rounded-lg bg-danger-tint px-3 py-2 text-size-sm text-fg-danger">
        {{ props.errorMessage ?? actionError }}
      </p>
      <slot>
        <p class="text-size-base text-subtle">
          This action permanently deletes the {{ props.resourceType.toLocaleLowerCase() }} <strong class="font-medium text-default">{{ props.resourceName }}</strong>.
        </p>
      </slot>
      <div class="space-y-2">
        <label :for="inputId" class="block text-size-sm font-medium text-default">{{ props.confirmationLabel ?? `Type ${props.resourceName} to confirm` }}</label><ark.input :id="inputId" :value="confirmation" :disabled="busy" autocomplete="off" spellcheck="false" class="h-9 w-full rounded-lg bg-base px-3 text-size-base text-default outline-none ring-1 ring-line placeholder:text-placeholder focus:ring-[1.5px] focus:ring-focus/50 disabled:opacity-50" @input="confirmation = ($event.target as HTMLInputElement).value" @keydown.enter.prevent="confirm" />
      </div>
    </div>
    <template #footer>
      <Button :disabled="busy" @click="open = false">
        {{ props.cancelLabel }}
      </Button><Button variant="destructive" :disabled="!confirmed || busy" :aria-busy="busy || undefined" @click="confirm">
        <span v-if="busy" class="i-lucide-loader-circle size-4 animate-spin" aria-hidden="true" />{{ props.deleteButtonText ?? `Delete ${props.resourceType}` }}
      </Button>
    </template>
  </LayerDialog>
</template>

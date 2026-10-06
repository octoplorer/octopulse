<script setup lang="ts">
import type { SidebarCollapsibleProps } from './section-types'
import { computed, ref, toRef, useId } from 'vue'
import { provideSidebarCollapsible } from './collapsible-context'
import { useSidebar } from './context'
import { useOpenChangeComplete } from './use-open-change-complete'

const props = withDefaults(defineProps<SidebarCollapsibleProps>(), {
  defaultOpen: false,
  open: undefined,
  autoScrollOnOpen: false,
})
const emit = defineEmits<{
  'update:open': [open: boolean]
  'openChange': [open: boolean]
  'openChangeComplete': [open: boolean]
}>()

const sidebar = useSidebar()
const element = ref<HTMLDivElement | null>(null)
const internalOpen = ref(props.defaultOpen)
const isOpen = computed(() => props.open ?? internalOpen.value)
const isContentShown = computed(() => isOpen.value && (
  sidebar.isMobile.value ? sidebar.openMobile.value : sidebar.state.value !== 'collapsed'
))
let keyboardExpanded = false

function setOpen(open: boolean) {
  internalOpen.value = open
  emit('update:open', open)
  emit('openChange', open)
}

function toggle() {
  keyboardExpanded = false
  setOpen(!isOpen.value)
}

const { complete: completeOpenChange } = useOpenChangeComplete(
  isContentShown,
  sidebar.animationDuration,
  open => emit('openChangeComplete', open),
)

provideSidebarCollapsible({
  contentId: useId(),
  isOpen,
  isCollapsible: true,
  autoScrollOnOpen: toRef(props, 'autoScrollOnOpen'),
  toggle,
  completeOpenChange,
})

function handleFocusIn(event: FocusEvent) {
  if (!isOpen.value && event.target instanceof HTMLElement && event.target.matches(':focus-visible')) {
    keyboardExpanded = true
    setOpen(true)
  }
}

function handleFocusOut(event: FocusEvent) {
  const container = event.currentTarget as HTMLDivElement
  if (
    keyboardExpanded
    && !(event.relatedTarget instanceof Node && container.contains(event.relatedTarget))
    && !container.querySelector('[data-active]')
  ) {
    keyboardExpanded = false
    setOpen(false)
  }
}

defineExpose({ element })
</script>

<template>
  <div
    ref="element"
    :data-open="isOpen || undefined"
    class="min-w-0"
    @focusin="handleFocusIn"
    @focusout="handleFocusOut"
  >
    <slot />
  </div>
</template>

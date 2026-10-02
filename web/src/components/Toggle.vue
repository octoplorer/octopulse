<script setup lang="ts">
import { Switch } from '@ark-ui/vue/switch'

defineProps<{ label: string, description?: string, disabled?: boolean }>()
const value = defineModel<boolean>({ default: false })
</script>

<template>
  <Switch.Root
    class="toggle"
    :checked="value"
    :disabled="disabled"
    @checked-change="value = $event.checked"
  >
    <Switch.HiddenInput /><Switch.Control class="switch-control">
      <Switch.Thumb class="switch-thumb" />
    </Switch.Control><Switch.Label>
      <strong>{{ label }}</strong><small v-if="description">{{ description }}</small>
    </Switch.Label>
  </Switch.Root>
</template>

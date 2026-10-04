<script setup lang="ts">
import { Switch as ArkSwitch } from '@ark-ui/vue/switch'

defineProps<{ label: string, description?: string, disabled?: boolean }>()
const value = defineModel<boolean>({ default: false })
</script>

<template>
  <ArkSwitch.Root
    class="toggle"
    flex="~ items-center gap-10px"
    cursor="pointer"
    un-text="12px"
    :checked="value"
    :disabled="disabled"
    @checked-change="value = $event.checked"
  >
    <ArkSwitch.HiddenInput /><ArkSwitch.Control
      class="switch-control [&[data-state=checked]]:bg-$accent"
      w="32px"
      h="20px"
      rounded="20px"
      bg="$control-border"
      p="3px"
      shrink="0"
    >
      <ArkSwitch.Thumb
        class="block switch-thumb [&[data-state=checked]]:translate-x-12px"
        w="14px"
        h="14px"
        rounded="full"
        bg="white"
        shadow="[0_1px_3px_#0002]"
        transition="all duration-200"
      />
    </ArkSwitch.Control><ArkSwitch.Label>
      <strong class="block" font="500">{{ label }}</strong><small v-if="description" class="block" un-text="12px $muted" mt="4px">{{ description }}</small>
    </ArkSwitch.Label>
  </ArkSwitch.Root>
</template>

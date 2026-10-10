<script setup lang="ts">
import MonitorEditor from '../../../../components/MonitorEditor.vue'

definePage({ meta: { title: 'navigation.newMonitor', contentWidth: 'form', roles: ['admin', 'operator'] } })
</script>

<template>
  <MonitorEditor />
</template>

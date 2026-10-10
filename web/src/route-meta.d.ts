import type { User } from './client/types.gen'
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    contentWidth?: 'wide' | 'compact' | 'form' | 'settings' | 'detail' | 'editor'
    title?: string
    roles?: User['role'][]
  }
}

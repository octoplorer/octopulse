import type { User } from './lib/types'
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    roles?: User['role'][]
  }
}

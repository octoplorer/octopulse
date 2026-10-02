import 'vue-router'
import type { User } from './lib/types'

declare module 'vue-router' {
  interface RouteMeta {
    title?: [string, string]
    roles?: User['role'][]
  }
}

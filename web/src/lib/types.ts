import type * as Contract from '../client/types.gen'
// Editor values are normalized from optional Go configuration fields. The
// public/management DTO contract itself is generated from Go OpenAPI.
type Normalize<T> = T extends (infer U)[] ? Normalize<U>[] : T extends object ? Configuration<T> : T
type Configuration<T> = {
  -readonly [K in keyof T as K extends `${string}Ref` | '$schema' ? never : K]-?: Normalize<
    NonNullable<T[K]>
  >
} & {
  [K in keyof T as K extends `${string}Ref` ? K : never]?: NonNullable<T[K]>
}
export type User = Contract.User
export type Secret = Contract.Secret
export type Channel = Contract.Channel
export type NameValue = Configuration<Contract.NameValue>
export type TLSConfig = Configuration<Contract.TlsConfig>
export type ConnectionConfig = Configuration<Contract.ConnectionConfig>
export type HTTPConfig = Configuration<Contract.HttpConfig>
export type TCPConfig = Configuration<Contract.TcpConfig>
export type DNSConfig = Configuration<Contract.DnsConfig>
export type HeartbeatConfig = Configuration<Contract.HeartbeatConfig>
export type CertificateConfig = Configuration<Contract.CertificateConfig>
export type Monitor = Omit<
  Contract.Monitor,
  'tags' | 'notificationChannelIds' | 'http' | 'tcp' | 'dns' | 'heartbeat' | 'certificate'
> & {
  tags: string[]
  notificationChannelIds: string[]
  http?: HTTPConfig
  tcp?: TCPConfig
  dns?: DNSConfig
  heartbeat?: HeartbeatConfig
  certificate?: CertificateConfig
}
export type IncidentUpdate = Contract.IncidentUpdate
export type Incident = Omit<Contract.Incident, 'monitorIds' | 'pageIds' | 'updates'> & {
  monitorIds: string[]
  pageIds: string[]
  updates: IncidentUpdate[]
}
export type Maintenance = Omit<Contract.Maintenance, 'monitorIds' | 'pageIds'> & {
  monitorIds: string[]
  pageIds: string[]
}
export type Availability = Contract.Availability
export type LatencyPoint = Contract.LatencyPoint
export type Audit = Contract.Audit
export type Settings = Omit<Contract.Settings, 'allowedDomains'> & { allowedDomains: string[] }
export type Attempt = Contract.Attempt
export type Round = Omit<Contract.Round, 'attempts'> & { attempts: Attempt[] }
export type MonitorHistory = Omit<Contract.MonitorHistory, 'rounds' | 'latency'> & {
  rounds: Round[]
  latency: LatencyPoint[]
}
export type PageMonitor = Contract.PageMonitor
export type PageGroup = Omit<Contract.PageGroup, 'monitors'> & { monitors: PageMonitor[] }
export type PageConfig = Omit<Contract.PageConfig, 'groups' | 'links'> & {
  groups: PageGroup[]
  links: Contract.Link[]
}
export type Page = Omit<Contract.Page, 'draft' | 'published'> & {
  draft: PageConfig
  published?: PageConfig
}
export type PublicMonitor = Omit<Contract.PublicMonitor, 'latency'> & { latency: LatencyPoint[] }
export type PublicGroup = Omit<Contract.PublicGroup, 'monitors'> & { monitors: PublicMonitor[] }
export type PublicIncident = Omit<Contract.PublicIncident, 'updates'> & {
  updates: IncidentUpdate[]
}
export type PublicPage = Omit<
  Contract.PublicPage,
  'config' | 'groups' | 'incidents' | 'maintenance'
> & {
  config: PageConfig
  groups: PublicGroup[]
  incidents: PublicIncident[]
  maintenance: Contract.PublicMaintenance[]
}
export type Delivery = Contract.DeliveryView
export type BeszelConfig = Contract.BeszelConfig
export type BeszelSystem = Contract.System
export type BeszelSystems = Omit<Contract.SystemsResponse, 'items'> & { items: BeszelSystem[] }
export type BeszelHistoryPoint = Contract.HistoryPoint
export type BeszelHistory = Omit<Contract.HistoryResponse, 'items'> & {
  items: BeszelHistoryPoint[]
}
export type BeszelContainer = Contract.Container
export type BeszelContainers = Omit<Contract.ContainersResponse, 'items'> & {
  items: BeszelContainer[]
}

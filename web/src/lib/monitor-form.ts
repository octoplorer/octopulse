import type * as Contract from '../client/types.gen'

// Form values have defaults for optional configuration fields and stay mutable.
// Secret references remain optional so an unset reference is omitted on save.
type Normalize<T> = T extends (infer U)[] ? Normalize<U>[] : T extends object ? Configuration<T> : T
type Configuration<T> = {
  -readonly [K in keyof T as K extends `${string}Ref` | '$schema' ? never : K]-?: Normalize<
    NonNullable<T[K]>
  >
} & {
  [K in keyof T as K extends `${string}Ref` ? K : never]?: NonNullable<T[K]>
}

export type NameValueForm = Configuration<Contract.NameValue>
export type TLSConfigForm = Configuration<Contract.TlsConfig>
export type ConnectionConfigForm = Configuration<Contract.ConnectionConfig>
export type HTTPConfigForm = Configuration<Contract.HttpConfig>
export type TCPConfigForm = Configuration<Contract.TcpConfig>
export type DNSConfigForm = Configuration<Contract.DnsConfig>
export type HeartbeatConfigForm = Configuration<Contract.HeartbeatConfig>
export type CertificateConfigForm = Configuration<Contract.CertificateConfig>
export type MonitorForm = Omit<
  Contract.Monitor,
  'tags' | 'notificationChannelIds' | 'http' | 'tcp' | 'dns' | 'heartbeat' | 'certificate'
> & {
  tags: string[]
  notificationChannelIds: string[]
  http?: HTTPConfigForm
  tcp?: TCPConfigForm
  dns?: DNSConfigForm
  heartbeat?: HeartbeatConfigForm
  certificate?: CertificateConfigForm
}

import type {
  CertificateConfig,
  ConnectionConfig,
  DNSConfig,
  HeartbeatConfig,
  HTTPConfig,
  Monitor,
  TCPConfig,
  TLSConfig,
} from './types'
import { t } from '../composables/i18n'

export const monitorTypes = [
  { value: 'http', label: 'monitorTypes.httpHttps' },
  { value: 'tcp', label: 'monitorTypes.tcpConnection' },
  { value: 'dns', label: 'monitorTypes.dnsResolution' },
  { value: 'heartbeat', label: 'monitorTypes.heartbeat' },
  { value: 'certificate', label: 'monitorTypes.certificateExpiry' },
]
function emptyTLS(): TLSConfig {
  return {
    enabled: false,
    insecureSkipVerify: false,
    serverName: '',
    minVersion: '',
    maxVersion: '',
  }
}
function emptyConnection(): ConnectionConfig {
  return {
    proxyUrl: '',
    proxyUsername: '',
    dnsServer: '',
    fixedIp: '',
  }
}
export function emptyHTTP(): HTTPConfig {
  return {
    url: '',
    method: 'GET',
    query: [],
    headers: [],
    host: '',
    body: {
      format: 'none',
      text: '',
      base64: '',
      charset: 'utf-8',
      contentType: '',
      fields: [],
      files: [],
    },
    auth: { type: 'none', username: '', header: 'Authorization', prefix: '' },
    tls: emptyTLS(),
    connection: emptyConnection(),
    redirects: { enabled: true, maxHops: 5, scope: 'same-origin' },
    acceptEncoding: 'gzip',
    requestGzip: false,
    responseCharset: '',
    maxResponseBytes: 2097152,
    assertions: {
      statusCodes: [],
      statusRanges: [{ min: 200, max: 299 }],
      headers: [],
      textContains: [],
      textNotContains: [],
      regex: [],
      json: [],
      maxLatencyMs: 0,
    },
  }
}
export function emptyTCP(): TCPConfig {
  return {
    host: '',
    port: 443,
    tls: emptyTLS(),
    connection: emptyConnection(),
    sendText: '',
    sendBase64: '',
    charset: 'utf-8',
    receiveContains: '',
    receiveRegex: '',
    maxReceiveBytes: 65536,
  }
}
export function emptyDNS(): DNSConfig {
  return {
    name: '',
    recordType: 'A',
    server: '',
    protocol: 'udp',
    expectedRCode: 'NOERROR',
    expectedValues: [],
    matchMode: 'contains',
  }
}
export function emptyHeartbeat(): HeartbeatConfig {
  return {
    periodSeconds: 60,
    graceSeconds: 30,
    lastReceivedAt: 0,
    lastSuccess: false,
    description: '',
  }
}
export function emptyCertificate(): CertificateConfig {
  return {
    host: '',
    port: 443,
    tls: emptyTLS(),
    connection: emptyConnection(),
    warningDays: [30, 14, 7, 1],
    notifyRenewal: true,
    state: 'check_failed',
    expiresAt: 0,
    fingerprint: '',
    daysRemaining: 0,
  }
}
export function newMonitor(): Monitor {
  return {
    id: '',
    name: '',
    description: '',
    type: 'http',
    tags: [],
    group: '',
    enabled: true,
    intervalSeconds: 60,
    timeoutSeconds: 10,
    retries: 2,
    retryDelaySeconds: 1,
    failureThreshold: 1,
    recoveryThreshold: 1,
    notificationChannelIds: [],
    notifyRecovery: true,
    reminderSeconds: 0,
    configVersion: 0,
    state: 'unknown',
    failureCount: 0,
    successCount: 0,
    lastCheckedAt: 0,
    nextCheckAt: 0,
    createdAt: 0,
    updatedAt: 0,
    http: emptyHTTP(),
  }
}
export function targetOf(m: Monitor) {
  return (
    m.http?.url
    || (m.tcp ? `${m.tcp.host}:${m.tcp.port}` : '')
    || m.dns?.name
    || (m.certificate ? `${m.certificate.host}:${m.certificate.port}` : '')
    || t('monitorTypes.heartbeat')
  )
}

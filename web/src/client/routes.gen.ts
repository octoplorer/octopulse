// Generated from Go OpenAPI by scripts/generate-routes.mjs. Do not edit.
export const routes = [
  {
    "method": "POST",
    "pattern": "^/api/heartbeat/([^/]+)/([^/]+)$",
    "parameters": [
      "id",
      "token"
    ],
    "operation": "reportHeartbeat"
  },
  {
    "method": "GET",
    "pattern": "^/api/public/pages/([^/]+)$",
    "parameters": [
      "slug"
    ],
    "operation": "getPublicPage"
  },
  {
    "method": "GET",
    "pattern": "^/api/public/resolve$",
    "parameters": [],
    "operation": "resolvePublicPage"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/assets$",
    "parameters": [],
    "operation": "uploadAsset"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/audit$",
    "parameters": [],
    "operation": "listAudit"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/beszel/config$",
    "parameters": [],
    "operation": "getBeszelConfig"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/beszel/config$",
    "parameters": [],
    "operation": "updateBeszelConfig"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/beszel/systems$",
    "parameters": [],
    "operation": "listBeszelSystems"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/beszel/systems/([^/]+)/containers$",
    "parameters": [
      "id"
    ],
    "operation": "listBeszelContainers"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/beszel/systems/([^/]+)/history$",
    "parameters": [
      "id"
    ],
    "operation": "getBeszelHistory"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/channels$",
    "parameters": [],
    "operation": "listChannels"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/channels$",
    "parameters": [],
    "operation": "createChannels"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/channels/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteChannels"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/channels/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "getChannels"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/channels/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateChannels"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/channels/([^/]+)/test$",
    "parameters": [
      "id"
    ],
    "operation": "testChannel"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/deliveries$",
    "parameters": [],
    "operation": "listDeliveries"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/incidents$",
    "parameters": [],
    "operation": "listIncidents"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/incidents$",
    "parameters": [],
    "operation": "createIncidents"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/incidents/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteIncidents"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/incidents/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "getIncidents"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/incidents/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateIncidents"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/incidents/([^/]+)/updates$",
    "parameters": [
      "id"
    ],
    "operation": "createIncidentUpdate"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/maintenance$",
    "parameters": [],
    "operation": "listMaintenance"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/maintenance$",
    "parameters": [],
    "operation": "createMaintenance"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/maintenance/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteMaintenance"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/maintenance/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "getMaintenance"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/maintenance/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateMaintenance"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/monitors$",
    "parameters": [],
    "operation": "listMonitors"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/monitors$",
    "parameters": [],
    "operation": "createMonitor"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/monitors/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteMonitor"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/monitors/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "getMonitor"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/monitors/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateMonitor"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/monitors/([^/]+)/check$",
    "parameters": [
      "id"
    ],
    "operation": "checkMonitor"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/monitors/([^/]+)/heartbeat/rotate$",
    "parameters": [
      "id"
    ],
    "operation": "rotateHeartbeat"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/monitors/([^/]+)/history$",
    "parameters": [
      "id"
    ],
    "operation": "getMonitorHistory"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/pages$",
    "parameters": [],
    "operation": "listPages"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/pages$",
    "parameters": [],
    "operation": "createPages"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/pages/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deletePages"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/pages/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "getPages"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/pages/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updatePages"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/pages/([^/]+)/preview$",
    "parameters": [
      "id"
    ],
    "operation": "previewPage"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/pages/([^/]+)/publish$",
    "parameters": [
      "id"
    ],
    "operation": "publishPage"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/profile$",
    "parameters": [],
    "operation": "updateProfile"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/secrets$",
    "parameters": [],
    "operation": "listSecrets"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/secrets$",
    "parameters": [],
    "operation": "createSecret"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/secrets/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteSecret"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/secrets/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateSecret"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/session$",
    "parameters": [],
    "operation": "deleteSession"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/session$",
    "parameters": [],
    "operation": "getSession"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/session$",
    "parameters": [],
    "operation": "createSession"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/settings$",
    "parameters": [],
    "operation": "getSettings"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/settings$",
    "parameters": [],
    "operation": "updateSettings"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/setup$",
    "parameters": [],
    "operation": "getSetup"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/setup$",
    "parameters": [],
    "operation": "createSetup"
  },
  {
    "method": "GET",
    "pattern": "^/api/v1/users$",
    "parameters": [],
    "operation": "listUsers"
  },
  {
    "method": "POST",
    "pattern": "^/api/v1/users$",
    "parameters": [],
    "operation": "createUser"
  },
  {
    "method": "DELETE",
    "pattern": "^/api/v1/users/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "deleteUser"
  },
  {
    "method": "PATCH",
    "pattern": "^/api/v1/users/([^/]+)$",
    "parameters": [
      "id"
    ],
    "operation": "updateUser"
  }
] as const

package domain

type NotificationPayload struct {
	MonitorID  string `json:"monitorId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	State      string `json:"state"`
	Generation int64  `json:"generation"`
	CycleID    string `json:"cycleId"`
	CreatedAt  int64  `json:"createdAt"`
	Message    string `json:"message"`
}
type DeliveryMarker struct {
	MonitorID      string `json:"monitorId"`
	ChannelID      string `json:"channelId"`
	CycleID        string `json:"cycleId"`
	DownSentAt     int64  `json:"downSentAt"`
	RecoverySentAt int64  `json:"recoverySentAt"`
}

func DeliveryMarkerID(monitorID, channelID, cycleID string) string {
	return monitorID + ":" + channelID + ":" + cycleID
}

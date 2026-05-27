package resources

type WebhookResponse struct {
	Received  bool   `json:"received"`
	Duplicate bool   `json:"duplicate"`
	EventID   string `json:"event_id,omitempty"`
}

package tunnel

type WebhookRequest struct {
	ID      string            `json:"id"`
	Headers map[string]string `json:"headers"`
	Path    string            `json:"path"`
	Body    string            `json:"body"`
}

type WebhookResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

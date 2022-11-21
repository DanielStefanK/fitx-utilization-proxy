package responses

type StudioResponse struct {
	Content []StudioInfo `json:"content"`
}

type StudioInfo struct {
	ID          uint64 `json:"id"`
	MagiclineId uint64 `json:"magiclineId"`
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
}

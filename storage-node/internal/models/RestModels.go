package models

type RestMessage struct {
	User  string `json:"user"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RestAllValuesMessage struct {
	User   string            `json:"user"`
	Values map[string]string `json:"values"`
}
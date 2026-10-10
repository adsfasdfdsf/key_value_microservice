package models

type RestMessage struct {
	User  string `json:"user"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

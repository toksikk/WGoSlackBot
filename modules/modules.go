package modules

import (
	"log"
	"strings"
)

type Payload struct {
	Channel   string `json:"channel"`
	IconEmoji string `json:"icon_emoji"`
	Text      string `json:"text"`
	Username  string `json:"username"`
}

type WebhookPayload struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ServiceID   string `json:"service_id"`
	TeamDomain  string `json:"team_domain"`
	TeamID      string `json:"team_id"`
	Text        string `json:"text"`
	Timestamp   string `json:"timestamp"`
	Token       string `json:"token"`
	TriggerWord string `json:"trigger_word"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	ResponseURL string `json:"response_url"`
}

var (
	SayCh       chan Payload
	MsgHandlers = make(map[string]func(*WebhookPayload))
	ModParams   = make(map[string]string)
)

func Init(ch chan Payload, params string) {
	SayCh = ch
	for _, param := range strings.Split(params, "!") {
		kv := strings.Split(param, ":")
		ModParams[kv[0]] = kv[1]
		log.Println(kv[0], kv[1])
	}
}

func HandleMessage(payload *WebhookPayload) {
	for _, fn := range MsgHandlers {
		fn(payload)
	}
}

// utility.go
package modules

func GeneratePayload(target string, icon string, text string, botname string) Payload {
	payload := Payload{
		Channel:   target,
		IconEmoji: icon,
		Text:      text,
		Username:  botname,
	}
	return payload
}

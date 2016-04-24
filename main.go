// WGoSlackBot
//
// Copyright (c) 2016 by Daniel Aberger
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"dev.ixab.de/WGSWATBP/WGoSlackBot/modules"
)

var (
	sayCh       chan modules.Payload
	webhookPort = flag.String("webhookport", "61321", "Webhook port")
	slackHook   = flag.String("slackhook", "https://hooks.slack.com/services/", "Slack webhook URL")
	params      = flag.String("params", "", "Module params")
)

func init() {
	flag.Parse()
}

func main() {
	sayCh = make(chan modules.Payload, 1024)
	modules.Init(sayCh, *params)

	go webhook()

	for {
		sendPayload(<-sayCh)
	}
}

func webhook() {
	http.HandleFunc("/hook", webhookHandler)
	log.Fatal(http.ListenAndServe(":"+*webhookPort, nil))
}
func webhookHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	whp := modules.WebhookPayload{
		Token:       req.Form.Get("token"),
		TeamID:      req.Form.Get("team_id"),
		TeamDomain:  req.Form.Get("team_domain"),
		ServiceID:   req.Form.Get("service_id"),
		ChannelID:   req.Form.Get("channel_id"),
		ChannelName: req.Form.Get("channel_name"),
		Timestamp:   req.Form.Get("timestamp"),
		UserID:      req.Form.Get("user_id"),
		UserName:    req.Form.Get("user_name"),
		Text:        req.Form.Get("text"),
		TriggerWord: req.Form.Get("trigger_word"),
		Command:     req.Form.Get("command"),
		ResponseURL: req.Form.Get("response_url"),
	}
	modules.HandleMessage(&whp)
}

func sendPayload(payload modules.Payload) {
	p, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error:", err)
	}
	url := *slackHook

	var jsonStr = p
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("Sent payload. Response Status:", resp.Status)
}

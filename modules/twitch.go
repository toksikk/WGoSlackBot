// flokatirc twitch module rewritten for WGoSlackBot
//
// Copyright (c) 2016 Daniel Aberger <da@ixab.de>

package modules

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"dev.ixab.de/WGSWATBP/WGoSlackBot/util"
)

type TwitchChannelObject struct {
	_id    int `json:"_id"`
	_links struct {
		Chat          string `json:"chat"`
		Commercial    string `json:"commercial"`
		Editors       string `json:"editors"`
		Features      string `json:"features"`
		Follows       string `json:"follows"`
		Self          string `json:"self"`
		StreamKey     string `json:"stream_key"`
		Subscriptions string `json:"subscriptions"`
		Teams         string `json:"teams"`
		Videos        string `json:"videos"`
	} `json:"_links"`
	Background                   interface{} `json:"background"`
	Banner                       string      `json:"banner"`
	BroadcasterLanguage          string      `json:"broadcaster_language"`
	CreatedAt                    string      `json:"created_at"`
	Delay                        interface{} `json:"delay"`
	DisplayName                  string      `json:"display_name"`
	Followers                    int         `json:"followers"`
	Game                         string      `json:"game"`
	Language                     string      `json:"language"`
	Logo                         string      `json:"logo"`
	Mature                       bool        `json:"mature"`
	Name                         string      `json:"name"`
	Partner                      bool        `json:"partner"`
	ProfileBanner                string      `json:"profile_banner"`
	ProfileBannerBackgroundColor string      `json:"profile_banner_background_color"`
	Status                       string      `json:"status"`
	UpdatedAt                    string      `json:"updated_at"`
	URL                          string      `json:"url"`
	VideoBanner                  string      `json:"video_banner"`
	Views                        int         `json:"views"`
}

type TwitchStreamObject struct {
	_links struct {
		Channel string `json:"channel"`
		Self    string `json:"self"`
	} `json:"_links"`
	Stream struct {
		_id    int `json:"_id"`
		_links struct {
			Self string `json:"self"`
		} `json:"_links"`
		AverageFps float64 `json:"average_fps"`
		Channel    struct {
			_id    int `json:"_id"`
			_links struct {
				Chat          string `json:"chat"`
				Commercial    string `json:"commercial"`
				Editors       string `json:"editors"`
				Features      string `json:"features"`
				Follows       string `json:"follows"`
				Self          string `json:"self"`
				StreamKey     string `json:"stream_key"`
				Subscriptions string `json:"subscriptions"`
				Teams         string `json:"teams"`
				Videos        string `json:"videos"`
			} `json:"_links"`
			Background                   interface{} `json:"background"`
			Banner                       string      `json:"banner"`
			BroadcasterLanguage          string      `json:"broadcaster_language"`
			CreatedAt                    string      `json:"created_at"`
			Delay                        interface{} `json:"delay"`
			DisplayName                  string      `json:"display_name"`
			Followers                    int         `json:"followers"`
			Game                         string      `json:"game"`
			Language                     string      `json:"language"`
			Logo                         string      `json:"logo"`
			Mature                       bool        `json:"mature"`
			Name                         string      `json:"name"`
			Partner                      bool        `json:"partner"`
			ProfileBanner                string      `json:"profile_banner"`
			ProfileBannerBackgroundColor string      `json:"profile_banner_background_color"`
			Status                       string      `json:"status"`
			UpdatedAt                    string      `json:"updated_at"`
			URL                          string      `json:"url"`
			VideoBanner                  string      `json:"video_banner"`
			Views                        int         `json:"views"`
		} `json:"channel"`
		CreatedAt  string `json:"created_at"`
		Delay      int    `json:"delay"`
		Game       string `json:"game"`
		IsPlaylist bool   `json:"is_playlist"`
		Preview    struct {
			Large    string `json:"large"`
			Medium   string `json:"medium"`
			Small    string `json:"small"`
			Template string `json:"template"`
		} `json:"preview"`
		VideoHeight int `json:"video_height"`
		Viewers     int `json:"viewers"`
	} `json:"stream"`
}

var (
	announceChannel string
	twitchBotIcon   string
	twitchBotName   string
	twitch          = map[string]bool{
		"rocketbeanstv": false,
		"blizzard":      false,
		"ea":            false,
		"warcraft":      false,
		"starcraft":     false,
		"bobross":       false,
		"pixelj0ckey":   false,
	}
	twitchapiurlstreams  = "https://api.twitch.tv/kraken/streams/"
	twitchapiurlchannels = "https://api.twitch.tv/kraken/channels/"
)

func init() {
	MsgHandlers["twitch"] = twitchHandleMessage
	log.Println("Initializing twitch module")
	go configAndPoll()
}

func configAndPoll() {
	time.Sleep(5 * time.Second)
	announceChannel = ModParams["twitchchan"]
	twitchBotIcon = ModParams["twitchicon"]
	twitchBotName = ModParams["twitchbotname"]
	go pollStreamData()
}

func twitchHandleMessage(payload *WebhookPayload) {
	tok := strings.Split(payload.Text, " ")
	if len(tok) < 1 {
		return
	}
	if payload.Command == "/twitch" {
		log.Print("tok: ")
		log.Println(tok)
		handleCommand(tok, "@"+payload.UserName, true)
	}
	switch tok[0] {
	case "!twitch":
		handleCommand(tok, "#"+payload.ChannelName, false)
	default:
	}
}
func handleCommand(tok []string, target string, isSlashCommand bool) {
	var tokMod int
	if isSlashCommand {
		tokMod = 1
		if tok[1-tokMod] == "" {
			tok = make([]string, 0, 0)
		}
	} else {
		tokMod = 0
	}
	switch len(tok) {
	case 1 - tokMod:
		onlinestreams := 0
		for streamname, _ := range twitch {
			var so TwitchStreamObject
			var co TwitchChannelObject
			so = getTwitchStreamObject(streamname)
			if so.Stream.Game != "" {
				onlinestreams++
				co = getTwitchChannelObject(streamname)
				twitchSendMsg(co, so, target)
				twitch[streamname] = true
			} else {
				twitch[streamname] = false
			}
		}
		if onlinestreams == 0 {
			SayCh <- GeneratePayload(target, twitchBotIcon, "All streams offline", twitchBotName)
		}
	case 2 - tokMod:
		streamname := tok[1-tokMod]
		var so TwitchStreamObject
		var co TwitchChannelObject
		so = getTwitchStreamObject(streamname)
		if so.Stream.Game != "" {
			co = getTwitchChannelObject(streamname)
			twitchSendMsg(co, so, target)
		} else {
			SayCh <- GeneratePayload(target, twitchBotIcon, streamname+" not found or offline", twitchBotName)
		}
	default:
	}
}

func pollStreamData() {
	if announceChannel != "" {
		time.Sleep(10 * time.Second)
		for {
			for streamname, _ := range twitch {
				var so TwitchStreamObject
				var co TwitchChannelObject
				so = getTwitchStreamObject(streamname)
				if so.Stream.Game != "" && !twitch[streamname] {
					co = getTwitchChannelObject(streamname)
					twitchSendMsg(co, so, "")
					twitch[streamname] = true
				} else if so.Stream.Game == "" && twitch[streamname] {
					twitch[streamname] = false
				}
			}
			time.Sleep(180 * time.Second)
		}
	}
}

func getTwitchStreamObject(streamname string) TwitchStreamObject {
	twsurl := twitchapiurlstreams + streamname
	var tobj TwitchStreamObject
	resp, err := http.Get(twsurl)
	if err != nil {
		log.Println(err)
	} else {
		reader, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		} else {
			defer resp.Body.Close()
			json.Unmarshal(reader, &tobj)
			return tobj
		}
	}
	return tobj
}

func getTwitchChannelObject(streamname string) TwitchChannelObject {
	twcurl := twitchapiurlchannels + streamname
	var tcobj TwitchChannelObject
	resp2, err := http.Get(twcurl)
	if err != nil {
		log.Println(err)
	} else {
		reader2, err := ioutil.ReadAll(resp2.Body)
		if err != nil {
			log.Println(err)
		} else {
			defer resp2.Body.Close()
			json.Unmarshal(reader2, &tcobj)
			return tcobj
		}
	}
	return tcobj
}

func twitchSendMsg(tcobj TwitchChannelObject, tso TwitchStreamObject, target string) {
	if target == "" {
		target = announceChannel
	}
	SayCh <- GeneratePayload(target,
		twitchBotIcon,
		"*"+tso.Stream.Channel.DisplayName+
			"*\n*Title:* "+tcobj.Status+
			"\n*Viewers:* "+util.NumberToString(tso.Stream.Viewers, '.')+
			"\n*Playing:* "+tso.Stream.Game+
			"\n"+tcobj.URL+"\n",
		twitchBotName)
}

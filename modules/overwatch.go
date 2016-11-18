package modules

import (
	"log"
	"strings"
	"github.com/leekchan/accounting"
	"strconv"
	"time"
	"dev.ixab.de/WGSWATBP/WGoSlackBot/util"
)

var (
	owbotname    string
	owboticon    string
	owapiurl string = "https://api.lootbox.eu/"
)

func init() {
	MsgHandlers["ow"] = owHandleMessage
	log.Println("Initializing ow module")
	go setOwConfig()
}
func setOwConfig() {
	time.Sleep(5 * time.Second)
	owbotname = ModParams["owbotname"]
	owboticon = ModParams["owboticon"]
}
func owHandleMessage(payload *WebhookPayload) {
	if payload.Command == "/d3" {
		reply := handleOwRequest(payload, true)
		if reply != "" {
			SayCh <- GeneratePayload(payload.ChannelName, d3boticon, reply, d3botname)
		}
	}
	if payload.TriggerWord == "!d3" {
		reply := handleOwRequest(payload, false)
		if reply != "" {
			SayCh <- GeneratePayload(payload.ChannelName, d3boticon, reply, d3botname)
		}
	}
}

func handleOwRequest(payload *WebhookPayload, isCommand bool) string {
	var tokMod int
	if isCommand {
		tokMod = 0
	} else {
		tokMod = 1
	}

	tok := strings.Split(payload.Text, " ")
	var result string
	if len(tok) < 2+tokMod {
		if tok[0+tokMod] == "" {
			SayCh <- GeneratePayload("@"+payload.UserName, owboticon, "Es wurde kein Battletag angegeben.", owbotname)
		} else {
			battletag, err := checkBattleTag(tok[0+tokMod])
			if err != nil {
				log.Println(err)
				SayCh <- GeneratePayload("@"+payload.UserName, owboticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", owbotname)
			}
			profile, err := getD3Profile(battletag)
			if err != nil {
				log.Println(err)
				SayCh <- GeneratePayload("@"+payload.UserName, owboticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", owbotname)
				result = ""
			} else {
				readableBattletag := strings.Replace(battletag, "-", "#", 1)
				lastUpdate := time.Unix(int64(profile.LastUpdated), 0).Format(time.RFC850)
				paragon := util.NumberToString(profile.ParagonLevel, '.')
				paragonHardcore := util.NumberToString(profile.ParagonLevelHardcore, '.')
				paragonSeason := util.NumberToString(profile.ParagonLevelSeason, '.')
				paragonSeasonHardcore := util.NumberToString(profile.ParagonLevelSeasonHardcore, '.')
				var allHeroes string
				for i, hero := range profile.Heroes {
					allHeroes += classToIcon(hero.Class, hero.Gender) + "" + heroTypeToIconString(hero.Seasonal, hero.Hardcore) + " " + hero.Name + " " + strconv.Itoa(hero.Level)
					if i+1 != len(profile.Heroes) {
						allHeroes += ", "
					}
				}
				result +=
					"*" + readableBattletag + "*\n" +
						"*Last Update:* " + lastUpdate + "\n" +
						"*Paragon:* " + paragon + ", :d3hardcore:" + paragonHardcore + ", :d3season:" + paragonSeason + ", :d3seasonhardcore:" + paragonSeasonHardcore + "\n" +
						"*Available Heroes:* " + allHeroes + "\n" +
						"https://eu.battle.net/d3/en/profile/" + battletag + "/"
			}
		}
	}
	if len(tok) >= 2+tokMod {
		battletag, err := checkBattleTag(tok[0+tokMod])
		if err != nil {
			log.Println(err)
			SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", d3botname)
		} else {
			profile, err := getD3Profile(battletag)
			if err != nil {
				log.Println(err)
				SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", d3botname)
			} else {
				hit := false
				for _, currenthero := range profile.Heroes {
					if strings.ToLower(currenthero.Name) == strings.ToLower(tok[1+tokMod]) {
						hero := getD3Hero(battletag, currenthero.ID)
						hit = true
						// TODO: hero data output
						result = "Hero: " + accounting.FormatNumberFloat64(hero.Stats.Damage, 0, ".", ",") + "\n" +
							"https://eu.battle.net/d3/en/profile/" + battletag + "/hero/" + strconv.Itoa(currenthero.ID)
						result = "Hero: " + "https://eu.battle.net/d3/en/profile/" + battletag + "/hero/" + strconv.Itoa(currenthero.ID)
					}
				}
				if !hit {
					SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Es konnte kein Charakter mit dem Namen *"+tok[1+tokMod]+"* unter dem Battletag *"+tok[0+tokMod]+"* gefunden werden.", d3botname)
				}
			}
		}
	}

	return result
}
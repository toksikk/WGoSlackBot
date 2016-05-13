// d3.go
package modules

import (
	"dev.ixab.de/WGSWATBP/WGoSlackBot/util"
	"encoding/json"
	"errors"
	"github.com/leekchan/accounting"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	d3apikey     string
	d3botname    string
	d3boticon    string
	bnetd3apiurl string = "https://eu.api.battle.net/d3/"
)

func init() {
	MsgHandlers["d3"] = d3HandleMessage
	log.Println("Initializing d3 module")
	go setd3Config()
}
func setd3Config() {
	time.Sleep(5 * time.Second)
	d3apikey = ModParams["d3apikey"]
	d3botname = ModParams["d3botname"]
	d3boticon = ModParams["d3boticon"]
}
func d3HandleMessage(payload *WebhookPayload) {
	if payload.Command == "/d3" {
		reply := handleD3Request(payload, true)
		if reply != "" {
			SayCh <- GeneratePayload(payload.ChannelName, d3boticon, reply, d3botname)
		}
	}
	if payload.TriggerWord == "!d3" {
		reply := handleD3Request(payload, false)
		if reply != "" {
			SayCh <- GeneratePayload(payload.ChannelName, d3boticon, reply, d3botname)
		}
	}
}

func handleD3Request(payload *WebhookPayload, isCommand bool) string {
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
			SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Es wurde kein Battletag angegeben.", d3botname)
		} else {
			battletag, err := checkBattleTag(tok[0+tokMod])
			if err != nil {
				log.Println(err)
				SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", d3botname)
			}
			profile, err := getD3Profile(battletag)
			if err != nil {
				log.Println(err)
				SayCh <- GeneratePayload("@"+payload.UserName, d3boticon, "Battletag "+tok[0+tokMod]+" fehlerhaft oder nicht gefunden.", d3botname)
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
func classToIcon(class string, gender int) string {
	genderstring := ""
	if gender == 0 {
		genderstring = "m"
	}
	if gender == 1 {
		genderstring = "f"
	}
	switch class {
	case "witch-doctor":
		return ":d3witchdoctor" + genderstring + ":"
	case "barbarian":
		return ":d3barbarian" + genderstring + ":"
	case "crusader":
		return ":d3crusader" + genderstring + ":"
	case "demon-hunter":
		return ":d3demonhunter" + genderstring + ":"
	case "wizard":
		return ":d3wizard" + genderstring + ":"
	case "monk":
		return ":d3monk" + genderstring + ":"
	default:
		return ""
	}
}
func heroTypeToIconString(seasonal bool, hardcore bool) string {
	result := ""
	if seasonal && hardcore {
		result = ":d3seasonhardcore:"
	} else {
		if seasonal {
			result = ":d3season:"
		}
		if hardcore {
			result = ":d3hardcore:"
		}
	}
	return result
}
func checkBattleTag(battletag string) (tag string, e error) {
	if strings.ContainsRune(battletag, '-') {
		return battletag, nil
	}
	if strings.ContainsRune(battletag, '#') {
		return strings.Replace(battletag, "#", "-", 1), nil
	}
	return "", errors.New("invalid battletag")
}
func getD3Profile(battletag string) (profile d3Profile, e error) {
	requesturl := bnetd3apiurl + "profile/" + battletag + "/?locale=en_GB&apikey=" + d3apikey
	var result d3Profile
	resp, err := http.Get(requesturl)
	if err != nil {
		log.Println(err)
	} else {
		reader, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		} else {
			defer resp.Body.Close()
			json.Unmarshal(reader, &result)
			if result.BattleTag == "" {
				return result, errors.New("Profile not found")
			}
			return result, nil
		}
	}
	return result, nil
}
func getD3Hero(battletag string, heroid int) d3Hero {
	heroidstring := strconv.Itoa(heroid)
	requesturl := bnetd3apiurl + "profile/" + battletag + "/hero/" + heroidstring + "?locale=en_GB&apikey=" + d3apikey
	var result d3Hero
	resp, err := http.Get(requesturl)
	if err != nil {
		log.Println(err)
	} else {
		reader, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		} else {
			defer resp.Body.Close()
			json.Unmarshal(reader, &result)
			return result
		}
	}
	return result
}

type d3Profile struct {
	BattleTag  string `json:"battleTag"`
	Blacksmith struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"blacksmith"`
	BlacksmithHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"blacksmithHardcore"`
	BlacksmithSeason struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"blacksmithSeason"`
	BlacksmithSeasonHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"blacksmithSeasonHardcore"`
	FallenHeroes []interface{} `json:"fallenHeroes"`
	GuildName    string        `json:"guildName"`
	Heroes       []struct {
		Class    string `json:"class"`
		Dead     bool   `json:"dead"`
		Gender   int    `json:"gender"`
		Hardcore bool   `json:"hardcore"`
		ID       int    `json:"id"`
		Kills    struct {
			Elites int `json:"elites"`
		} `json:"kills"`
		Last_updated int    `json:"last-updated"`
		Level        int    `json:"level"`
		Name         string `json:"name"`
		ParagonLevel int    `json:"paragonLevel"`
		Seasonal     bool   `json:"seasonal"`
	} `json:"heroes"`
	HighestHardcoreLevel int `json:"highestHardcoreLevel"`
	Jeweler              struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"jeweler"`
	JewelerHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"jewelerHardcore"`
	JewelerSeason struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"jewelerSeason"`
	JewelerSeasonHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"jewelerSeasonHardcore"`
	Kills struct {
		Elites           int `json:"elites"`
		HardcoreMonsters int `json:"hardcoreMonsters"`
		Monsters         int `json:"monsters"`
	} `json:"kills"`
	LastHeroPlayed int `json:"lastHeroPlayed"`
	LastUpdated    int `json:"lastUpdated"`
	Mystic         struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"mystic"`
	MysticHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"mysticHardcore"`
	MysticSeason struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"mysticSeason"`
	MysticSeasonHardcore struct {
		Level       int    `json:"level"`
		Slug        string `json:"slug"`
		StepCurrent int    `json:"stepCurrent"`
		StepMax     int    `json:"stepMax"`
	} `json:"mysticSeasonHardcore"`
	ParagonLevel               int `json:"paragonLevel"`
	ParagonLevelHardcore       int `json:"paragonLevelHardcore"`
	ParagonLevelSeason         int `json:"paragonLevelSeason"`
	ParagonLevelSeasonHardcore int `json:"paragonLevelSeasonHardcore"`
	Progression                struct {
		Act1 bool `json:"act1"`
		Act2 bool `json:"act2"`
		Act3 bool `json:"act3"`
		Act4 bool `json:"act4"`
		Act5 bool `json:"act5"`
	} `json:"progression"`
	SeasonalProfiles struct {
		Season0 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    float64 `json:"barbarian"`
				Crusader     float64 `json:"crusader"`
				Demon_hunter float64 `json:"demon-hunter"`
				Monk         float64 `json:"monk"`
				Witch_doctor int     `json:"witch-doctor"`
				Wizard       float64 `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season0"`
		Season1 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    int `json:"barbarian"`
				Crusader     int `json:"crusader"`
				Demon_hunter int `json:"demon-hunter"`
				Monk         int `json:"monk"`
				Witch_doctor int `json:"witch-doctor"`
				Wizard       int `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season1"`
		Season2 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    int `json:"barbarian"`
				Crusader     int `json:"crusader"`
				Demon_hunter int `json:"demon-hunter"`
				Monk         int `json:"monk"`
				Witch_doctor int `json:"witch-doctor"`
				Wizard       int `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season2"`
		Season3 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    float64 `json:"barbarian"`
				Crusader     int     `json:"crusader"`
				Demon_hunter int     `json:"demon-hunter"`
				Monk         int     `json:"monk"`
				Witch_doctor int     `json:"witch-doctor"`
				Wizard       int     `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season3"`
		Season4 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    float64 `json:"barbarian"`
				Crusader     int     `json:"crusader"`
				Demon_hunter int     `json:"demon-hunter"`
				Monk         int     `json:"monk"`
				Witch_doctor int     `json:"witch-doctor"`
				Wizard       int     `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season4"`
		Season5 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    int     `json:"barbarian"`
				Crusader     float64 `json:"crusader"`
				Demon_hunter int     `json:"demon-hunter"`
				Monk         int     `json:"monk"`
				Witch_doctor int     `json:"witch-doctor"`
				Wizard       int     `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season5"`
		Season6 struct {
			HighestHardcoreLevel int `json:"highestHardcoreLevel"`
			Kills                struct {
				Elites           int `json:"elites"`
				HardcoreMonsters int `json:"hardcoreMonsters"`
				Monsters         int `json:"monsters"`
			} `json:"kills"`
			ParagonLevel         int `json:"paragonLevel"`
			ParagonLevelHardcore int `json:"paragonLevelHardcore"`
			Progression          struct {
				Act1 bool `json:"act1"`
				Act2 bool `json:"act2"`
				Act3 bool `json:"act3"`
				Act4 bool `json:"act4"`
				Act5 bool `json:"act5"`
			} `json:"progression"`
			SeasonID   int `json:"seasonId"`
			TimePlayed struct {
				Barbarian    int `json:"barbarian"`
				Crusader     int `json:"crusader"`
				Demon_hunter int `json:"demon-hunter"`
				Monk         int `json:"monk"`
				Witch_doctor int `json:"witch-doctor"`
				Wizard       int `json:"wizard"`
			} `json:"timePlayed"`
		} `json:"season6"`
	} `json:"seasonalProfiles"`
	TimePlayed struct {
		Barbarian    float64 `json:"barbarian"`
		Crusader     float64 `json:"crusader"`
		Demon_hunter float64 `json:"demon-hunter"`
		Monk         float64 `json:"monk"`
		Witch_doctor int     `json:"witch-doctor"`
		Wizard       float64 `json:"wizard"`
	} `json:"timePlayed"`
}
type d3Hero struct {
	Class     string `json:"class"`
	Dead      bool   `json:"dead"`
	Followers struct {
		Enchantress struct {
			Items  struct{} `json:"items"`
			Level  int      `json:"level"`
			Skills []struct {
				Skill struct {
					Description       string `json:"description"`
					Icon              string `json:"icon"`
					Level             int    `json:"level"`
					Name              string `json:"name"`
					SimpleDescription string `json:"simpleDescription"`
					SkillCalcID       string `json:"skillCalcId"`
					Slug              string `json:"slug"`
					TooltipURL        string `json:"tooltipUrl"`
				} `json:"skill"`
			} `json:"skills"`
			Slug  string `json:"slug"`
			Stats struct {
				ExperienceBonus int `json:"experienceBonus"`
				GoldFind        int `json:"goldFind"`
				MagicFind       int `json:"magicFind"`
			} `json:"stats"`
		} `json:"enchantress"`
		Scoundrel struct {
			Items struct {
				MainHand struct {
					DisplayColor  string `json:"displayColor"`
					Icon          string `json:"icon"`
					ID            string `json:"id"`
					Name          string `json:"name"`
					TooltipParams string `json:"tooltipParams"`
				} `json:"mainHand"`
				Special struct {
					DisplayColor  string `json:"displayColor"`
					Icon          string `json:"icon"`
					ID            string `json:"id"`
					Name          string `json:"name"`
					TooltipParams string `json:"tooltipParams"`
				} `json:"special"`
			} `json:"items"`
			Level  int `json:"level"`
			Skills []struct {
				Skill struct {
					Description       string `json:"description"`
					Icon              string `json:"icon"`
					Level             int    `json:"level"`
					Name              string `json:"name"`
					SimpleDescription string `json:"simpleDescription"`
					SkillCalcID       string `json:"skillCalcId"`
					Slug              string `json:"slug"`
					TooltipURL        string `json:"tooltipUrl"`
				} `json:"skill"`
			} `json:"skills"`
			Slug  string `json:"slug"`
			Stats struct {
				ExperienceBonus int `json:"experienceBonus"`
				GoldFind        int `json:"goldFind"`
				MagicFind       int `json:"magicFind"`
			} `json:"stats"`
		} `json:"scoundrel"`
		Templar struct {
			Items  struct{} `json:"items"`
			Level  int      `json:"level"`
			Skills []struct {
				Skill struct {
					Description       string `json:"description"`
					Icon              string `json:"icon"`
					Level             int    `json:"level"`
					Name              string `json:"name"`
					SimpleDescription string `json:"simpleDescription"`
					SkillCalcID       string `json:"skillCalcId"`
					Slug              string `json:"slug"`
					TooltipURL        string `json:"tooltipUrl"`
				} `json:"skill"`
			} `json:"skills"`
			Slug  string `json:"slug"`
			Stats struct {
				ExperienceBonus int `json:"experienceBonus"`
				GoldFind        int `json:"goldFind"`
				MagicFind       int `json:"magicFind"`
			} `json:"stats"`
		} `json:"templar"`
	} `json:"followers"`
	Gender   int      `json:"gender"`
	Hardcore bool     `json:"hardcore"`
	ID       int      `json:"id"`
	Items    struct{} `json:"items"`
	Kills    struct {
		Elites int `json:"elites"`
	} `json:"kills"`
	Last_updated    int           `json:"last-updated"`
	LegendaryPowers []interface{} `json:"legendaryPowers"`
	Level           int           `json:"level"`
	Name            string        `json:"name"`
	ParagonLevel    int           `json:"paragonLevel"`
	Progression     struct {
		Act1 struct {
			Completed       bool `json:"completed"`
			CompletedQuests []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"completedQuests"`
		} `json:"act1"`
		Act2 struct {
			Completed       bool `json:"completed"`
			CompletedQuests []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"completedQuests"`
		} `json:"act2"`
		Act3 struct {
			Completed       bool `json:"completed"`
			CompletedQuests []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"completedQuests"`
		} `json:"act3"`
		Act4 struct {
			Completed       bool `json:"completed"`
			CompletedQuests []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"completedQuests"`
		} `json:"act4"`
		Act5 struct {
			Completed       bool `json:"completed"`
			CompletedQuests []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"completedQuests"`
		} `json:"act5"`
	} `json:"progression"`
	SeasonCreated int  `json:"seasonCreated"`
	Seasonal      bool `json:"seasonal"`
	Skills        struct {
		Active []struct {
			Rune struct {
				Description       string `json:"description"`
				Level             int    `json:"level"`
				Name              string `json:"name"`
				Order             int    `json:"order"`
				SimpleDescription string `json:"simpleDescription"`
				SkillCalcID       string `json:"skillCalcId"`
				Slug              string `json:"slug"`
				TooltipParams     string `json:"tooltipParams"`
				Type              string `json:"type"`
			} `json:"rune"`
			Skill struct {
				CategorySlug      string `json:"categorySlug"`
				Description       string `json:"description"`
				Icon              string `json:"icon"`
				Level             int    `json:"level"`
				Name              string `json:"name"`
				SimpleDescription string `json:"simpleDescription"`
				SkillCalcID       string `json:"skillCalcId"`
				Slug              string `json:"slug"`
				TooltipURL        string `json:"tooltipUrl"`
			} `json:"skill"`
		} `json:"active"`
		Passive []struct {
			Skill struct {
				Description string `json:"description"`
				Flavor      string `json:"flavor"`
				Icon        string `json:"icon"`
				Level       int    `json:"level"`
				Name        string `json:"name"`
				SkillCalcID string `json:"skillCalcId"`
				Slug        string `json:"slug"`
				TooltipURL  string `json:"tooltipUrl"`
			} `json:"skill"`
		} `json:"passive"`
	} `json:"skills"`
	Stats struct {
		ArcaneResist      int     `json:"arcaneResist"`
		Armor             int     `json:"armor"`
		AttackSpeed       int     `json:"attackSpeed"`
		BlockAmountMax    int     `json:"blockAmountMax"`
		BlockAmountMin    int     `json:"blockAmountMin"`
		BlockChance       int     `json:"blockChance"`
		ColdResist        int     `json:"coldResist"`
		CritChance        float64 `json:"critChance"`
		CritDamage        float64 `json:"critDamage"`
		Damage            float64 `json:"damage"`
		DamageIncrease    int     `json:"damageIncrease"`
		DamageReduction   int     `json:"damageReduction"`
		Dexterity         int     `json:"dexterity"`
		FireResist        int     `json:"fireResist"`
		GoldFind          int     `json:"goldFind"`
		Healing           float64 `json:"healing"`
		Intelligence      int     `json:"intelligence"`
		Life              int     `json:"life"`
		LifeOnHit         int     `json:"lifeOnHit"`
		LifePerKill       int     `json:"lifePerKill"`
		LifeSteal         int     `json:"lifeSteal"`
		LightningResist   int     `json:"lightningResist"`
		MagicFind         int     `json:"magicFind"`
		PhysicalResist    int     `json:"physicalResist"`
		PoisonResist      int     `json:"poisonResist"`
		PrimaryResource   int     `json:"primaryResource"`
		SecondaryResource int     `json:"secondaryResource"`
		Strength          int     `json:"strength"`
		Thorns            int     `json:"thorns"`
		Toughness         float64 `json:"toughness"`
		Vitality          int     `json:"vitality"`
	} `json:"stats"`
}

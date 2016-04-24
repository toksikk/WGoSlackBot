// ts3.go
package modules

import (
	"bufio"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	users   = map[string]bool{}
	host    string
	port    string
	user    string
	pass    string
	botname string
	channel string
)

func init() {
	MsgHandlers["ts3"] = ts3HandleMessage
	log.Println("Initializing ts3 module")
	go setConfig()
}
func setConfig() {
	time.Sleep(5 * time.Second)
	host = ModParams["tshost"]
	port = ModParams["tsport"]
	user = ModParams["tsuser"]
	pass = ModParams["tspass"]
	botname = ModParams["tsbotname"]
	channel = ModParams["tschan"]
	go loop()
}
func loop() {
	for {
		var err error
		users, err = getUsers()
		if err != nil {
			log.Println(err)
			log.Println("Trying again in 60 seconds.")
			time.Sleep(60 * time.Second)
		} else {
			break
		}
	}

	for {
		if channel == "" { // dont spam if no channel is set
			break
		}
		neu, err := getUsers()
		if err != nil {
			log.Println(err)
			time.Sleep(50 * time.Second)
		} else {
			for i, _ := range neu {
				if users[i] != neu[i] {
					SayCh <- GeneratePayload(channel, ":poop:", i+" joined TS3.", botname)
				}
			}
			for i, _ := range users {
				if users[i] != neu[i] {
					SayCh <- GeneratePayload(channel, ":poop:", i+" left TS3.", botname)
				}
			}
			users = neu
		}

		time.Sleep(10 * time.Second)
	}
}
func ts3HandleMessage(payload *WebhookPayload) {
	log.Println("TS3 users requested by " + payload.UserName)
	users, _ := getUsers()
	var s string
	i := 0
	s += "Current users on TS3 Server:\n"
	for u, _ := range users {
		i++
		s += "- " + u
		if i != len(users) {
			s += "\n"
		}
	}
	SayCh <- GeneratePayload("@"+payload.UserName, ":poop:", s, botname)
}

func getUsers() (u map[string]bool, e error) {
	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for i := 0; i < 2; i++ { // get rid of banner
		_, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("Connection closed by foreign host.")
			}
			return nil, err
		}
	}
	writer.WriteString("login " + user + " " + pass + "\n")
	writer.Flush()
	_, _, err = reader.ReadLine() // get rid of "OK" message
	if err != nil {
		if err.Error() == "EOF" {
			log.Println("Connection closed by foreign host.")
		}
		return nil, err
	}

	writer.WriteString("use sid=1\n")
	writer.Flush()
	_, _, err = reader.ReadLine() // get rid of "OK" message
	if err != nil {
		if err.Error() == "EOF" {
			log.Println("Connection closed by foreign host.")
		}
		return nil, err
	}

	writer.WriteString("clientlist\n")
	writer.Flush()
	line, _, err := reader.ReadLine() // get rid of "OK" message
	if err != nil {
		if err.Error() == "EOF" {
			log.Println("Connection closed by foreign host.")
		}
		return nil, err
	}
	users_raw := string(line)

	re := regexp.MustCompile("client_nickname=(.*?) ")
	users_still_raw := re.FindAllString(users_raw, -1)

	for i := range users_still_raw {
		users_still_raw[i] = strings.Trim(users_still_raw[i], " ")
		users_still_raw[i] = strings.Replace(users_still_raw[i], "client_nickname=", "", 1)
		users_still_raw[i] = strings.Replace(users_still_raw[i], "\\s", " ", -1)
	}
	users := make([]string, 0)
	for i := range users_still_raw {
		if !strings.HasPrefix(users_still_raw[i], "serveradmin from") {
			users = append(users, users_still_raw[i])
		}
	}
	result := make(map[string]bool)
	for i := range users {
		result[users[i]] = true
	}
	conn.Close()
	return result, nil
}

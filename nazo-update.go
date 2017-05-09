package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/CleverbotIO/go-cleverbot.io"
)

// DiscordLogin is a simple struct which contains a username and password for a Discord login
type DiscordLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var cb *cleverbot.Session
var botErr error

func main() {
	loginFile, fileErr := ioutil.ReadFile("./login.json")
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		os.Exit(1)
	}

	var login DiscordLogin
	jsonErr := json.Unmarshal(loginFile, &login)

	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		os.Exit(1)
	}

	dg, err := discordgo.New(login.Username, login.Password)

	if err != nil {
		fmt.Println("Error creating session: ", err)
		os.Exit(1)
	} else {
		fmt.Println("Successfully logged in as " + login.Username)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		os.Exit(1)
	}

	cb, botErr = cleverbot.New("JHowZe3ddT6Da0JU","8NSK1vZVH1lRNMIcTbu4hU6kGEyIDxsW","")
	if botErr != nil {
		log.Fatal(botErr)
	} 

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		msg := r.URL.Query().Get("msg")
		if id != "" && msg != "" {
			dg.ChannelMessageSend(id, "Message from HTTP chat endpoint: \n`"+msg+"`")
			fmt.Fprintf(w, "You are sending: \n"+msg+"\nto channel ID: \n"+id)
		}
	})

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	log.Fatal(http.ListenAndServe(":8080", nil))
	<-make(chan struct{})
	return
}

// messageCreate is the handler function for all incoming Discord messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := strings.Split(m.Content, " ")

	switch msg[0] {
	case "nazoupdate": // Reminds everyone that Nazo is adorable
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "This is your daily reminder that <@!165846085020024832> is adorable.")
		break
	case "deltaspeak": // Echoes the given text from my own user account
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "`ds:` "+m.Content[11:len(m.Content)])
		break
	case "cb": // Communicates with Cleverbot.
		response, botErr2 := cb.Ask(m.Content[2:len(m.Content)])
		if botErr2 != nil {
			log.Fatal(botErr2)
		}

		s.ChannelMessageSend(m.ChannelID, "Cleverbot says: " + response)

		break
	}
}

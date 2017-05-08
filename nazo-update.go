package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"strings"
)

type DiscordLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	msg := m.Content

	if strings.Contains(msg, "nazoupdate") {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "This is your daily reminder that <@!165846085020024832> is adorable.")
	}

}

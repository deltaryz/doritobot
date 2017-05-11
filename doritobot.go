package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CleverbotIO/go-cleverbot.io"
	"github.com/bwmarrin/discordgo"
)

// randomRange gives a random whole integer between the given integers [min, max)
func randomRange(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

// Config is a json file containing the login data and any user-configurable values
type Config struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Bot                 bool   `json:"Bot"`
	BotToken            string `json:"botToken"`
	HTTPEndpointEnabled bool   `json:"httpEndpointEnabled"`
	EchoCommandEnabled  bool   `json:"echoCommandEnabled"`
}

// PvfmMeta is a json file with the current metadata for PonyfilleFM's servers
type PvfmMeta struct {
	Icestats struct {
		Admin              string `json:"admin"`
		Host               string `json:"host"`
		Location           string `json:"location"`
		ServerID           string `json:"server_id"`
		ServerStart        string `json:"server_start"`
		ServerStartIso8601 string `json:"server_start_iso8601"`
		Source             []struct {
			Artist             string      `json:"artist,omitempty"`
			AudioBitrate       int         `json:"audio_bitrate,omitempty"`
			AudioChannels      int         `json:"audio_channels,omitempty"`
			AudioInfo          string      `json:"audio_info"`
			AudioSamplerate    int         `json:"audio_samplerate,omitempty"`
			Channels           int         `json:"channels"`
			Genre              string      `json:"genre"`
			IceBitrate         int         `json:"ice-bitrate,omitempty"`
			ListenerPeak       int         `json:"listener_peak"`
			Listeners          int         `json:"listeners"`
			Listenurl          string      `json:"listenurl"`
			Quality            string      `json:"quality,omitempty"`
			Samplerate         int         `json:"samplerate"`
			ServerDescription  string      `json:"server_description"`
			ServerName         string      `json:"server_name"`
			ServerType         string      `json:"server_type"`
			ServerURL          string      `json:"server_url"`
			StreamStart        string      `json:"stream_start"`
			StreamStartIso8601 string      `json:"stream_start_iso8601"`
			Subtype            string      `json:"subtype,omitempty"`
			Title              string      `json:"title"`
			Dummy              interface{} `json:"dummy"`
			Bitrate            int         `json:"bitrate,omitempty"`
		} `json:"source"`
	} `json:"icestats"`
}

// DerpiResults is a struct to contain Derpibooru search results
type DerpiResults struct {
	Search []struct {
		ID               string        `json:"id"`
		CreatedAt        time.Time     `json:"created_at"`
		UpdatedAt        time.Time     `json:"updated_at"`
		DuplicateReports []interface{} `json:"duplicate_reports"`
		FirstSeenAt      time.Time     `json:"first_seen_at"`
		UploaderID       string        `json:"uploader_id"`
		Score            int           `json:"score"`
		CommentCount     int           `json:"comment_count"`
		Width            int           `json:"width"`
		Height           int           `json:"height"`
		FileName         string        `json:"file_name"`
		Description      string        `json:"description"`
		Uploader         string        `json:"uploader"`
		Image            string        `json:"image"`
		Upvotes          int           `json:"upvotes"`
		Downvotes        int           `json:"downvotes"`
		Faves            int           `json:"faves"`
		Tags             string        `json:"tags"`
		TagIds           []string      `json:"tag_ids"`
		AspectRatio      float64       `json:"aspect_ratio"`
		OriginalFormat   string        `json:"original_format"`
		MimeType         string        `json:"mime_type"`
		Sha512Hash       string        `json:"sha512_hash"`
		OrigSha512Hash   string        `json:"orig_sha512_hash"`
		SourceURL        string        `json:"source_url"`
		Representations  struct {
			ThumbTiny  string `json:"thumb_tiny"`
			ThumbSmall string `json:"thumb_small"`
			Thumb      string `json:"thumb"`
			Small      string `json:"small"`
			Medium     string `json:"medium"`
			Large      string `json:"large"`
			Tall       string `json:"tall"`
			Full       string `json:"full"`
		} `json:"representations"`
		IsRendered  bool `json:"is_rendered"`
		IsOptimized bool `json:"is_optimized"`
	} `json:"search"`
	Total        int           `json:"total"`
	Interactions []interface{} `json:"interactions"`
}

// These need to be global
var cb *cleverbot.Session
var botErr error
var currentID string
var userIDError error
var dg *discordgo.Session
var err error
var login Config

func main() {
	// Read the config.json
	loginFile, fileErr := ioutil.ReadFile("./config.json")
	if fileErr != nil {
		fmt.Println(fileErr.Error())
		os.Exit(1)
	}

	// Parse it
	jsonErr := json.Unmarshal(loginFile, &login)

	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
		os.Exit(1)
	}

	// Log in to discord
	if login.Bot {
		dg, err = discordgo.New("Bot " + login.BotToken)
	} else {
		dg, err = discordgo.New(login.Username, login.Password)
	}

	if err != nil {
		fmt.Println("Error creating session: ", err)
		os.Exit(1)
	} else {
		fmt.Println("Successfully logged in.")
	}

	// Message received handler
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		os.Exit(1)
	}

	// Open cleverbot API
	// TODO: move this to login.json
	cb, botErr = cleverbot.New("JHowZe3ddT6Da0JU", "8NSK1vZVH1lRNMIcTbu4hU6kGEyIDxsW", "")
	if botErr != nil {
		log.Fatal(botErr)
	}

	// HTTP endpoint handler
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		msg := r.URL.Query().Get("msg")
		if id != "" && msg != "" {
			dg.ChannelMessageSend(id, "Message from HTTP chat endpoint: \n`"+msg+"`")
			fmt.Fprintf(w, "You are sending: \n"+msg+"\nto channel ID: \n"+id)
		}
	})

	dgUser, userIDError := dg.User("@me")
	currentID = dgUser.ID

	if userIDError != nil {
		log.Fatal(userIDError)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	if login.HTTPEndpointEnabled {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
	<-make(chan struct{})
	return
}

// messageCreate is the handler function for all incoming Discord messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	var msg []string
	if m.Author.Username == "PonyChat" {
		// this is disabled since my implementation was SHIT
		//msg = strings.Split(m.Content[strings.Index(m.Content, ">")+4:len(m.Content)], " ")
	} else {
		msg = strings.Split(m.Content, " ")
	}

	if len(msg) > 0 {
		switch msg[0] {
		case "nazoupdate": // Reminds everyone that Nazo is adorable
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			s.ChannelMessageSend(m.ChannelID, "This is your daily reminder that <@!165846085020024832> is adorable.")
			break
		case "echo": // Echoes the given text
			if login.EchoCommandEnabled {
				s.ChannelMessageDelete(m.ChannelID, m.ID)
				/*if m.Author.Username == "PonyChat" {
					s.ChannelMessageSend(m.ChannelID, "`ds:` "+m.Content[15:len(m.Content)])
				} else {*/
				s.ChannelMessageSend(m.ChannelID, "`echo:`\n"+m.Content[11:len(m.Content)])
				//}
			}
			break
		case "cb": // Communicates with Cleverbot.
			response, botErr2 := cb.Ask(m.Content[2:len(m.Content)])
			if botErr2 != nil {
				log.Fatal(botErr2)
			}
			s.ChannelMessageSend(m.ChannelID, response)
			break
		case "db": // Grabs an image from Derpibooru results with a given list of tags - always a safe image!
			if len(msg) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Error: not enough arguments")
				break
			}
			derpiTags := strings.Replace(m.Content[3:len(m.Content)], " ", "+", -1)
			resp, derpiHTTPErr := http.Get("https://derpibooru.org/search.json?q=safe," + derpiTags)
			if derpiHTTPErr != nil {
				s.ChannelMessageSend(m.ChannelID, "HTTP error, check console")
				log.Fatal(derpiHTTPErr)
				break
			}
			defer resp.Body.Close()
			derpiBody, derpiErr := ioutil.ReadAll(resp.Body)
			var derpiErr2 error
			results := DerpiResults{}
			derpiErr2 = json.Unmarshal(derpiBody, &results)
			if derpiErr != nil {
				s.ChannelMessageSend(m.ChannelID, "Error reading HTTP response")
				break
			}
			if derpiErr2 != nil {
				s.ChannelMessageSend(m.ChannelID, "Error reading derpibooru API response")
				break
			}
			if len(results.Search) > 0 {
				s.ChannelMessageSend(m.ChannelID, "http:"+results.Search[randomRange(0, len(results.Search))].Image)
			} else {
				s.ChannelMessageSend(m.ChannelID, "Error: no results")
			}
			break
		case "h": // h
			if m.Author.ID != currentID {
				s.ChannelMessageSend(m.ChannelID, "h")
			}
			break
		case "pvfmservers": // Gives a list of all available PVFM streams (direct links)
			pvfmResp, pvfmErr := http.Get("http://dj.bronyradio.com:7090/status-json.xsl")
			if pvfmErr != nil {
				s.ChannelMessageSend(m.ChannelID, "Error receiving station metadata")
				break
			}
			defer pvfmResp.Body.Close()

			pvfmContent, pvfmContentErr := ioutil.ReadAll(pvfmResp.Body)
			if pvfmContentErr != nil {
				s.ChannelMessageSend(m.ChannelID, "Error parsing metadata response")
				break
			}

			currentMeta := PvfmMeta{}

			pvfmJSONErr := json.Unmarshal(pvfmContent, &currentMeta)
			if pvfmJSONErr != nil {
				s.ChannelMessageSend(m.ChannelID, "Error in json unmarshal")
				break
			}

			outputString := "PVFM Servers:\n"

			for _, element := range currentMeta.Icestats.Source {
				outputString += "- " + element.ServerDescription + ": " + strings.Replace(element.Listenurl, "aerial", "dj.bronyradio.com", -1) + "\n"
			}

			outputString += "\nDJ Recordings: http://darkling.darkwizards.com/wang/BronyRadio/?M=D"

			s.ChannelMessageSend(m.ChannelID, outputString)

			break
		}
	}
}
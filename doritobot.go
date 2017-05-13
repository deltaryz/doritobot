package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/CleverbotIO/go-cleverbot.io"
	"github.com/PonyvilleFM/aura/pvfm/station"
	"github.com/bwmarrin/discordgo"
	"github.com/jzelinskie/geddit"
	"github.com/techniponi/doritobot/gitupdate"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// randomRange gives a random whole integer between the given integers [min, max)
func randomRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func updateBot(m string, s *discordgo.Session) error {
	cmd := exec.Command("/bin/bash", "-c", "cd ~/go/src/github.com/techniponi/doritobot/; git pull; go install; cd ~/go/bin")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	s.ChannelMessageSend(m, string(stdout))
	os.Exit(0)
	return nil
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
		if login.HTTPEndpointEnabled {
			id := r.URL.Query().Get("id")
			msg := r.URL.Query().Get("msg")
			if id != "" && msg != "" {
				dg.ChannelMessageSend(id, "Message from HTTP chat endpoint: \n`"+msg+"`")
				fmt.Fprintf(w, "You are sending: \n"+msg+"\nto channel ID: \n"+id)
			}
		}
	})

	http.HandleFunc("/repoupdate", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var update gitupdate.GitUpdate
		err := decoder.Decode(&update)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()
		updateString := "New commit pushed to <https://github.com/techniponi/doritobot> (" + update.HeadCommit.ID + "):\n" + update.HeadCommit.Author.Name + ": " + update.HeadCommit.Message + "\n" + update.HeadCommit.URL

		dg.ChannelMessageSend("298642620849324035", updateString+"\n\nUpdating now...")

		// automatically update bot
		updateErr := updateBot("298642620849324035", dg)
		if updateErr != nil {
			log.Fatal(updateErr)
		}
	})

	dgUser, userIDError := dg.User("@me")
	currentID = dgUser.ID

	if userIDError != nil {
		log.Fatal(userIDError)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
			currentMeta, metaErr := station.GetStats()
			if metaErr != nil {
				s.ChannelMessageSend(m.ChannelID, "Error receiving pvfm metadata")
				break
			}

			outputString := "**PVFM Servers:**\n"

			for _, element := range currentMeta.Icestats.Source {
				outputString += ":musical_note: " + element.ServerDescription + ":\n`" + strings.Replace(element.Listenurl, "aerial", "dj.bronyradio.com", -1) + "`\n"
			}

			outputString += "\n:cd: DJ Recordings:\n`http://darkling.darkwizards.com/wang/BronyRadio/?M=D`"

			s.ChannelMessageSend(m.ChannelID, outputString)

			break
		case "techgore":
			listOptions := geddit.ListingOptions{
				Limit: 50,
			}
			reddit := geddit.NewSession("discordbot")
			results, err := reddit.SubredditSubmissions("techsupportgore", geddit.NewSubmissions, listOptions)
			if err != nil {
				break
			}
			s.ChannelMessageSend(m.ChannelID, results[randomRange(0, len(results))].URL)
			break
		case "snuggle", "cuddle", "hug", "kiss", "boop", "glomp", "nuzzle":
			if len(msg) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Who - Delta, Twisty, or Jac?")
				break
			}
			if msg[1] == "Kappa" {
				s.ChannelMessageSend(m.ChannelID, "https://floof.zone/img/kappagay.png")
				break
			}
			names := map[string]string{
				"awal":    "Twisty",
				"twisty":  "Twisty",
				"delta":   "Delta",
				"dorito":  "Delta",
				"techni":  "Delta",
				"nazo":    "Jac",
				"nuzzles": "Jac",
				"jac":     "Jac",
				"thorax":  "Thorax",
				"shining": "Shiny",
				"shiny":   "Shiny",
				"quartz":  "Quartz",
				"dyed":    "Quartz",
				"rhomb":   "Rhombus",
				"rhombus": "Rhombus",
				"rhomby":  "Rhombus",
				"icebear":  "Ice Bear",
				"ice":  "Ice Bear",
				"bear":  "Ice Bear",
			}
			possibleResponses := []string{
				"snuggles back.",
				"flops over.",
				"blushes profusely.",
				"twitches ears and smiles.",
				"smiles lovingly.",
				"boops you back!",
				"glomps you!",
				"is happy.",
				"jumps with joy!",
				"wasn't expecting that! :heart:",
				"loves you. :heart:",
			}
			characterSpecifics := map[string][]string{
				"Thorax":  {"vibrates his wings in excitement.", "is cheered up from your kindness!"},
				"Shiny":   {"wonders if Cadance is okay with this.", "thinks you would be a great addition to the Sparkle family."},
				"Delta":   {"gets a wingboner.", "vibrates."},
				"Jac":     {"dies of cuteness overload.", "passes out from an extreme overdose of gay.", "can't hold all these husbandos."},
				"Twisty":  {"invites you to his next gig.", "needed that! :heart:"},
				"Quartz":  {"runs away.", "did not like that.", "dyes inside.", "cries.", "is anti-snuggle."},
				"Rhombus": {"giggles like a giddy schoolfilly.", "squeals happily.", "floofs his wings."},
				"Ice Bear": {"doesn't hate your butt.", "has a conspiracy theory.", "has respect. Keep real.", "...sleeps in...fridge...", "will lick your cheeks."},
			}
			if names[msg[1]] == "" {
				s.ChannelMessageSend(m.ChannelID, "I'm afraid I don't know who that is. :c")
				break
			}
			finalMessage := "error" // set to error as default in case of derpage

			if names[msg[1]] == "Quartz" {
				finalMessage = characterSpecifics[names[msg[1]]][randomRange(0, len(characterSpecifics[names[msg[1]]]))]
			} else {

				if randomRange(0, 10) == 7 {
					finalMessage = characterSpecifics[names[msg[1]]][randomRange(0, len(characterSpecifics[names[msg[1]]]))] // this line is a fucking mess
				} else {
					finalMessage = possibleResponses[randomRange(0, len(possibleResponses))]
				}

			}

			s.ChannelMessageSend(m.ChannelID, names[msg[1]]+" "+finalMessage)

			break
		case "gay":
			possibleResponses := []string{
				"floof.zone/img/gaybats.png",
				"floof.zone/img/nazoblep.png",
				"floof.zone/img/awalblep.png",
				"floof.zone/img/floofbat-capall.png",
			}
			s.ChannelMessageSend(m.ChannelID, "http://"+possibleResponses[randomRange(0, len(possibleResponses))])
			break
		case "botupdate":
			err := updateBot(m.ChannelID, s)
			if err != nil {
				log.Fatal(err)
				break
			}
			break
		}
	}
}

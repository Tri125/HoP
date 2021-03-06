package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tri125/HoP/commands"
	"github.com/Tri125/HoP/metrics"
	"github.com/bwmarrin/discordgo"
)

/*
Set this variable with go build with the -ldflags="-X main.version=<value>" parameter.
*/
var version = "undefined"

// Variables used for commands line parameters
var (
	Token string
)

func init() {

	versionFlag := flag.Bool("v", false, "Prints current version")
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	if Token == "" {
		var present bool
		Token, present = os.LookupEnv("HOP_TOKEN")
		if !present {
			log.Fatal("Token not set.")
		}
	}
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		metrics.ErrorEncountered.Add(1)
		log.Println("error creating Discord session,", err)
		return
	}

	//metrics.SetServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/metrics", expvar.Handler())
	srv := &http.Server{Addr: ":" + port, Handler: nil}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildJoin)
	dg.AddHandler(guildRemove)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		metrics.ErrorEncountered.Add(1)
		log.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	log.Println("Version : ", version)
	dg.UpdateStatus(0, "Summoning Singulo")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
	metrics.Close()
	log.Println("Server gracefully stopped.")
}

func guildJoin(s *discordgo.Session, c *discordgo.GuildCreate) {
	metrics.JoinedGuilds.Add(1)
}

func guildRemove(s *discordgo.Session, r *discordgo.GuildDelete) {
	metrics.JoinedGuilds.Add(-1)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by bots, including himself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.Bot || len(m.Content) > 100 {
		return
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		metrics.ErrorEncountered.Add(1)
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		metrics.ErrorEncountered.Add(1)
		// Could not find guild.
		return
	}

	if m.Content == "!grant Captain Access" {
		s.ChannelMessageSend(m.ChannelID, "Go home, Clown.")
	} else {
		command := commands.GetCommand(m.Content)
		switch command := command.(type) {
		default:
			break
		case commands.RemoveType:
			command.RemoveRole(s, g, c, m.Author, m.Content)
			break
		case commands.GrantType:
			command.GrantRole(s, g, c, m.Author, m.Content)
			break
		case commands.JobType:
			command.Jobs(s, g, c, m.Author)
			break
		case commands.HelpType:
			command.HoP(s, m.Author)
		}
	}

	metrics.RequestCounter.Incr(1)

}

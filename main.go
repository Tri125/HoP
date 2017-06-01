package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return
	}
	if m.Content == "!grant Captain Access" {
		s.ChannelMessageSend(m.ChannelID, "Go home, Clown.")
	} else if strings.HasPrefix(m.Content, "!grant") {
		roleRequest := strings.TrimPrefix(m.Content, "!grant")
		roleRequest = strings.TrimSpace((roleRequest))
		grantRole(s, g, c, m.Author, roleRequest)
		return
	} else if strings.HasPrefix(m.Content, "!remove") {
		roleRequest := strings.TrimPrefix(m.Content, "!remove")
		roleRequest = strings.TrimSpace((roleRequest))
		removeRole(s, g, c, m.Author, roleRequest)
		return
	}

}

func grantRole(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User, roleName string) {
	for _, role := range g.Roles {
		if role.Name == roleName {
			err := s.GuildMemberRoleAdd(g.ID, u.ID, role.ID)
			if err != nil {
				fmt.Println("Role Grant failed: ", err)
				return
			}
			s.ChannelMessageSend(c.ID, roleName+" clearance granted to "+u.Mention()+". Have a nice day!")
		}
	}
}

func removeRole(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User, roleName string) {
	for _, role := range g.Roles {
		if role.Name == roleName {
			err := s.GuildMemberRoleRemove(g.ID, u.ID, role.ID)
			if err != nil {
				fmt.Println("Role Removal failed: ", err)
				return
			}
			s.ChannelMessageSend(c.ID, roleName+" clearance removed from "+u.Mention()+".")
		}
	}
}

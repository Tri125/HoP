package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

/*
Set this variable with go build with the -ldflags="-X main.version=<value>" parameter.
*/
var version = "undefined"

// Variables used for command line parameters
var (
	Token       string
	BotCommands []BotCommand
)

type BotCommandType uint8

const (
	GRANT BotCommandType = iota + 1
	REMOVE
	HOP
	JOBS
)

type BotCommand struct {
	description string
	commandType BotCommandType
	userCommand string
}

func (c BotCommand) String() string {
	return c.description
}

func init() {

	versionFlag := flag.Bool("v", false, "Prints current version")
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	BotCommands = make([]BotCommand, 4)

	grant := BotCommand{description: "Grant access to the requested role. You can separate multiple roles by a semicolon.", commandType: GRANT, userCommand: "!grant"}
	remove := BotCommand{description: "Remove access to the requested role. You can separate multiple roles by a semicolon.", commandType: REMOVE, userCommand: "!remove"}
	hoP := BotCommand{description: "List available commands.", commandType: HOP, userCommand: "!HoP"}
	jobs := BotCommand{description: "List available roles.", commandType: JOBS, userCommand: "!jobs"}
	BotCommands[0] = grant
	BotCommands[1] = remove
	BotCommands[2] = hoP
	BotCommands[3] = jobs
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

	// Ignore all messages created by bots, including himself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.Bot {
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
	} else if m.Content == "!HoP" {
		hoP(s, m.Author)
	} else if m.Content == "!jobs" {
		jobs(s, g, c, m.Author)
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
	splitRoles := strings.Split(roleName, ";")
	sort.Strings(splitRoles)
	var granted bool
	for _, role := range g.Roles {
		i := sort.SearchStrings(splitRoles, role.Name)
		if i < len(splitRoles) && splitRoles[i] == role.Name {
			err := s.GuildMemberRoleAdd(g.ID, u.ID, role.ID)
			if err != nil {
				fmt.Println("Role Grant failed: ", err)
				return
			}
			granted = true
		}
	}
	if granted {
		s.ChannelMessageSend(c.ID, strings.Join(splitRoles, " ")+" clearance granted to "+u.Mention()+".\n Have a nice day!")
	}
}

func removeRole(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User, roleName string) {
	splitRoles := strings.Split(roleName, ";")
	sort.Strings(splitRoles)
	var granted bool
	for _, role := range g.Roles {
		i := sort.SearchStrings(splitRoles, role.Name)
		if i < len(splitRoles) && splitRoles[i] == role.Name {
			err := s.GuildMemberRoleRemove(g.ID, u.ID, role.ID)
			if err != nil {
				fmt.Println("Role Removal failed: ", err)
				return
			}
			granted = true
		}
	}
	if granted {
		s.ChannelMessageSend(c.ID, strings.Join(splitRoles, " ")+" clearance removed from "+u.Mention()+".")
	}
}

func hoP(s *discordgo.Session, u *discordgo.User) {
	if len(BotCommands) == 0 {
		return
	}
	var buffer bytes.Buffer
	for _, command := range BotCommands {
		buffer.WriteString("`")
		buffer.WriteString(command.userCommand)
		buffer.WriteString("` : ")
		buffer.WriteString(command.description)
		buffer.WriteString("\n\n")
	}
	c, err := s.UserChannelCreate(u.ID)
	if err != nil {
		return
	}
	s.ChannelMessageSend(c.ID, buffer.String())
}

func jobs(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User) {
	member, err := s.State.Member(g.ID, s.State.User.ID)
	if err != nil {
		fmt.Println("Couldn't get guild member: ", err)
		return
	}
	if len(g.Roles) == 0 || len(member.Roles) == 0 {
		return
	}
	var highestRolePosition int
	for _, roleID := range member.Roles {
		for _, role := range g.Roles {
			if roleID == role.ID {
				if highestRolePosition < role.Position {
					highestRolePosition = role.Position
				}
			}
		}
	}
	var buffer bytes.Buffer
	buffer.WriteString("Here are the available jobs:\n\n")
	for _, role := range g.Roles {
		if role.Position > 0 && role.Position < highestRolePosition {
			buffer.WriteString("``")
			buffer.WriteString(role.Name)
			buffer.WriteString("``\n")
		}
	}
	s.ChannelMessageSend(c.ID, buffer.String())
}

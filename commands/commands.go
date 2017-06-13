package commands

import (
	"bytes"
	"github.com/bwmarrin/discordgo"
	"github.com/tri125/HoP/metrics"
	"log"
	"sort"
	"strings"
)

const (
	grantCommandType  = "GRANT"
	removeCommandType = "REMOVE"
	helpCommandType   = "HELP"
	jobCommandType    = "JOBS"
)

type Command struct {
	description string
	userInput   string
}

func prepareInput(input string, prefix string) string {
	input = strings.TrimPrefix(input, prefix)
	return strings.TrimSpace(input)
}

func (t Command) yolo() {
	return
}

var commands []Command
var grantTest GrantType
var removeTest RemoveType
var helpTest HelpType
var jobTest JobType

type GrantType Command
type RemoveType Command
type HelpType Command
type JobType Command

func (gt GrantType) GrantRole(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User, input string) {
	roleName := prepareInput(input, gt.userInput)
	splitRoles := strings.Split(roleName, ";")
	sort.Strings(splitRoles)
	var granted bool
	for _, role := range g.Roles {
		i := sort.SearchStrings(splitRoles, role.Name)
		if i < len(splitRoles) && splitRoles[i] == role.Name {
			err := s.GuildMemberRoleAdd(g.ID, u.ID, role.ID)
			if err != nil {
				log.Println("Role Grant failed: ", err)
				return
			}
			granted = true
		}
	}
	if granted {
		s.ChannelMessageSend(c.ID, strings.Join(splitRoles, " ")+" clearance granted to "+u.Mention()+".\n Have a nice day!")
	}
}

func (rt RemoveType) RemoveRole(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User, input string) {
	roleName := prepareInput(input, rt.userInput)
	splitRoles := strings.Split(roleName, ";")
	sort.Strings(splitRoles)
	var granted bool
	for _, role := range g.Roles {
		i := sort.SearchStrings(splitRoles, role.Name)
		if i < len(splitRoles) && splitRoles[i] == role.Name {
			err := s.GuildMemberRoleRemove(g.ID, u.ID, role.ID)
			if err != nil {
				log.Println("Role Removal failed: ", err)
				return
			}
			granted = true
		}
	}
	if granted {
		s.ChannelMessageSend(c.ID, strings.Join(splitRoles, " ")+" clearance removed from "+u.Mention()+".")
	}
}

func (ht HelpType) HoP(s *discordgo.Session, u *discordgo.User) {
	if len(commands) == 0 {
		return
	}
	var buffer bytes.Buffer
	for _, command := range commands {
		buffer.WriteString("`")
		buffer.WriteString(command.userInput)
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

func (jt JobType) Jobs(s *discordgo.Session, g *discordgo.Guild, c *discordgo.Channel, u *discordgo.User) {
	member, err := s.State.Member(g.ID, s.State.User.ID)
	if err != nil {
		metrics.ErrorEncountered.Add(1)
		log.Println("Couldn't get guild member: ", err)
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

func init() {
	grantTest = GrantType{"Grant access to the requested role. You can separate multiple roles by a semicolon.", "!grant"}
	removeTest = RemoveType{"Remove access to the requested role. You can separate multiple roles by a semicolon.", "!remove"}
	helpTest = HelpType{"List available commands.", "!HoP"}
	jobTest = JobType{"List available roles.", "!jobs"}

	commands = make([]Command, 4)
	commands[0] = Command(grantTest)
	commands[1] = Command(removeTest)
	commands[2] = Command(helpTest)
	commands[3] = Command(jobTest)

}

func GetCommand(input string) interface{} {
	if strings.HasPrefix(input, grantTest.userInput) {
		return grantTest
	} else if strings.HasPrefix(input, removeTest.userInput) {
		return removeTest
	} else if input == helpTest.userInput {
		return helpTest
	} else if input == jobTest.userInput {
		return jobTest
	} else {
		return nil
	}
}

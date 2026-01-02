package commands

import (
	"context"
	"strings"
	"sync"
	"time"

	base "github.com/Alliance-Community/bots-base"
	"github.com/bwmarrin/discordgo"
	"github.com/emilekm/prism-proxy/prismproxy"
)

const (
	prismCmdName     = "prism"
	banIDCmdName     = "banid"
	timebanIDCmdName = "timebanid"
	unbanNameCmdName = "unbanname"
	unbanIDCmdName   = "unbanid"
)

var (
	defaultCommandsPermission = int64(discordgo.PermissionModerateMembers)
	hashLength                = 32
)

var (
	// Slash Commands
	prismCommand = &discordgo.ApplicationCommand{
		Name:        prismCmdName,
		Description: "PRISM",
		Options: []*discordgo.ApplicationCommandOption{
			banIDCommand,
			timebanIDCommand,
			unbanNameCommand,
			unbanIDCommand,
		},
		DefaultMemberPermissions: &defaultCommandsPermission,
	}

	idOption = &discordgo.ApplicationCommandOption{
		Name:        "id",
		Description: "The hash ID of the user to ban",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
		MinLength:   &hashLength,
		MaxLength:   hashLength,
	}

	reasonOption = &discordgo.ApplicationCommandOption{
		Name:        "reason",
		Description: "The reason for the ban",
		Type:        discordgo.ApplicationCommandOptionString,
		Required:    true,
	}

	banIDCommand = &discordgo.ApplicationCommandOption{
		Name:        banIDCmdName,
		Description: "Ban a user by their ID",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Options: []*discordgo.ApplicationCommandOption{
			idOption,
			reasonOption,
		},
	}

	timebanIDCommand = &discordgo.ApplicationCommandOption{
		Name:        timebanIDCmdName,
		Description: "Temporarily ban a user by their ID",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Options: []*discordgo.ApplicationCommandOption{
			idOption,
			{
				Name:        "duration",
				Description: "The duration of the temporary ban (e.g., 1h, 30m, 2d)",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			reasonOption,
		},
	}

	unbanNameCommand = &discordgo.ApplicationCommandOption{
		Name:        unbanNameCmdName,
		Description: "Unban a user by their name",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "name",
				Description: "The exact name of the user to unban",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	unbanIDCommand = &discordgo.ApplicationCommandOption{
		Name:        unbanIDCmdName,
		Description: "Unban a user by their ID",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Options: []*discordgo.ApplicationCommandOption{
			idOption,
		},
	}
)

type PRCommands struct {
	prism prismproxy.ProxyClient
	mutex sync.Mutex
}

func New(prism prismproxy.ProxyClient) *PRCommands {
	return &PRCommands{
		prism: prism,
	}
}

func (c *PRCommands) Register(baseBot *base.Bot) {
	baseBot.RegisterCommand(prismCommand, c.handlePrismCmd)
}

func (c *PRCommands) handlePrismCmd(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	option := i.ApplicationCommandData().Options[0]
	subCmd := option.Name
	subOptions := option.Options

	switch subCmd {
	case banIDCmdName:
		return c.issueCommand(s, i, subOptions, "!banid")
	case timebanIDCmdName:
		return c.issueCommand(s, i, subOptions, "!timebanid")
	case unbanNameCmdName:
		return c.issueCommand(s, i, subOptions, "!unbanname")
	case unbanIDCmdName:
		return c.issueCommand(s, i, subOptions, "!unbanid")
	default:
		return nil
	}
}

func (c *PRCommands) issueCommand(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption, cmd string) error {
	ctx := context.Background()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	builder.WriteString(cmd)
	for _, option := range options {
		builder.WriteString(" ")
		builder.WriteString(option.StringValue())
	}

	c.mutex.Lock()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	response := ""
	msgs, err := c.prism.ChatMessages(ctx, &prismproxy.Empty{})
	if err == nil {
		go func() {
			defer cancel()
			for {
				msg, err := msgs.Recv()
				if err != nil {
					return
				}

				if msg.Type == prismproxy.ChatMessageType_RESPONSE {
					response = msg.Content
				}
			}
		}()
	}

	resp, err := c.prism.RACommand(ctx, &prismproxy.RACommandReq{
		Command: builder.String(),
	})
	if err != nil {
		c.mutex.Unlock()
		return err
	}

	<-ctx.Done()
	c.mutex.Unlock()

	if response == "" {
		response = resp.Content
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Type:        discordgo.EmbedTypeRich,
				Title:       resp.Content,
				Description: response,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Command:",
						Value: builder.String(),
					},
				},
			},
		},
	})
	return err
}

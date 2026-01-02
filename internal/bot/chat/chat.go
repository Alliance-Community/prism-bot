package chat

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/emilekm/prism-proxy/prismproxy"
)

const (
	tagLength  = 6
	nameLength = 23
)

type discordSession interface {
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (st *discordgo.Message, err error)
	WebhookExecute(webhookID, token string, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (st *discordgo.Message, err error)
}

type Chat struct {
	discord   discordSession
	channelID string
	prism     prismproxy.ProxyClient
}

func NewChat(discord discordSession, prism prismproxy.ProxyClient, channelID string) *Chat {
	return &Chat{
		discord:   discord,
		channelID: channelID,
		prism:     prism,
	}
}

func (c *Chat) Start(ctx context.Context) error {
	messages, err := c.prism.ChatMessages(ctx, &prismproxy.Empty{})
	if err != nil {
		return err
	}

	go func() {
		builder := bytes.Buffer{}

		for {
			msg, err := messages.Recv()
			if err != nil {
				return
			}

			if msg.Content == "" && msg.PlayerName == "" && msg.Channel == "" {
				continue
			}

			builder.WriteString(time.Unix(int64(msg.Timestamp), 0).Format("`[2006-01-02 15:04:05]`"))
			builder.WriteString(" ")

			channelSplit := strings.Split(msg.Channel, " ")

			channelSlots := make([]string, 4)
			for i := range channelSlots {
				channelSlots[i] = ":black_small_square:"
			}
			for _, channelPart := range channelSplit {
				if emoji, slot := channelPartToEmoji(channelPart); emoji != "" {
					channelSlots[slot] = emoji
				}
			}

			for _, slot := range channelSlots {
				builder.WriteString(slot)
			}
			builder.WriteString(" ")

			playerName := msg.PlayerName
			if playerName == "" {
				playerName = " [Server]"
			}

			tagName := strings.SplitN(playerName, " ", 2)
			tag, name := tagName[0], tagName[1]

			builder.WriteString(
				fmt.Sprintf(
					"`% "+strconv.Itoa(tagLength)+"s ",
					tag,
				),
			)

			nameIndentLength := max(nameLength-len(name), 0)

			builder.WriteString(fmt.Sprintf("%s% "+strconv.Itoa(nameIndentLength)+"s`", name, ""))
			builder.WriteString(" : ")

			builder.WriteString(msg.Content)

			if _, err = c.discord.ChannelMessageSend(c.channelID, builder.String()); err != nil {
				slog.Error("failed to send message to discord", "error", err)
			}

			builder.Reset()
		}
	}()

	return nil
}

func channelPartToEmoji(num string) (string, int) {
	switch num {
	case "Admin":
		return ":exclamation:", 0
	case "Global":
		return ":globe_with_meridians:", 0
	case "Game":
		return ":loudspeaker:", 0
	case "BluFor":
		return ":blue_square:", 1
	case "OpFor":
		return ":red_square:", 1
	case "*":
		return ":skull:", 3
	case "1":
		return ":one:", 2
	case "2":
		return ":two:", 2
	case "3":
		return ":three:", 2
	case "4":
		return ":four:", 2
	case "5":
		return ":five:", 2
	case "6":
		return ":six:", 2
	case "7":
		return ":seven:", 2
	case "8":
		return ":eight:", 2
	case "9":
		return ":nine:", 2
	default:
		return "", -1
	}
}

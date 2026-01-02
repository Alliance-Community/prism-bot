package details

import (
	"context"
	"log/slog"
	"strings"
	"text/template"
	"time"

	abase "github.com/Alliance-Community/bots-base"
	"github.com/bwmarrin/discordgo"
	"github.com/emilekm/prism-proxy/prismproxy"
	"github.com/prbf2-tools/prism-bot/internal/config"
)

type ServerDetails struct {
	prism     prismproxy.ProxyClient
	session   *discordgo.Session
	templates map[string]*template.Template
}

func New(channelsConfig []config.Channel, prism prismproxy.ProxyClient) (*ServerDetails, error) {
	templates := make(map[string]*template.Template)

	for _, channel := range channelsConfig {
		tmpl, err := template.New("serverdetails").Parse(channel.Template)
		if err != nil {
			return nil, err
		}
		templates[channel.ID] = tmpl
	}

	return &ServerDetails{
		prism:     prism,
		templates: templates,
	}, nil
}
func (s *ServerDetails) Register(baseBot *abase.Bot) {
	s.session = baseBot.Session()
	s.handleMessages()
}

func (s *ServerDetails) handleMessages() {
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			ctx := context.Background()
			msg, err := s.prism.GetServerDetails(ctx, &prismproxy.Empty{})
			if err != nil {
				slog.Error(err.Error())
				continue
			}

			s.updateServerDetails(msg)
		}
	}()
}

func (s *ServerDetails) updateServerDetails(msg *prismproxy.ServerDetails) {
	if s.session == nil {
		slog.Error("Discord session is nil")
		return
	}

	tmplContext := struct {
		*prismproxy.ServerDetails
		Local localizedMap
	}{
		ServerDetails: msg,
		Local:         locallizeMapDetails(msg.Map),
	}
	for channelID, tmpl := range s.templates {
		var tpl strings.Builder
		err := tmpl.Execute(&tpl, tmplContext)
		if err != nil {
			slog.Error(err.Error(), "op", "ServerDetails.updateServerDetails")
			continue
		}

		_, err = s.session.ChannelEdit(channelID, &discordgo.ChannelEdit{
			Name: tpl.String(),
		})
	}
}

type localizedMap struct {
	Name  string
	Mode  string
	Layer string
	Size  int
}

func locallizeMapDetails(mapDetails *prismproxy.ServerMap) localizedMap {
	name := mapDetails.Name
	mode := mapDetails.Mode
	layer := mapDetails.Layer.String()
	size := 0

	if foundMode, ok := gameModes[mapDetails.Mode]; ok {
		mode = foundMode
	}

	if foundLayer, ok := layers[int(mapDetails.Layer)]; ok {
		layer = foundLayer
	}

	if foundMap, ok := levels[mapDetails.Name]; ok {
		name = foundMap.Name
		size = foundMap.Size
	}

	return localizedMap{
		Name:  name,
		Mode:  mode,
		Layer: layer,
		Size:  size,
	}
}

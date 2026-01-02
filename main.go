package main

import (
	"context"
	"flag"
	"log/slog"

	base "github.com/Alliance-Community/bots-base"

	"github.com/emilekm/prism-proxy/prismproxy"
	"github.com/prbf2-tools/prism-bot/internal/bot/chat"
	"github.com/prbf2-tools/prism-bot/internal/bot/commands"
	"github.com/prbf2-tools/prism-bot/internal/bot/details"
	"github.com/prbf2-tools/prism-bot/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var configFilePath = flag.String("config", "config.yaml", "Path to the configuration file")
var certPath = flag.String("cert", "", "path to TLS certificate")
var proxyAddr = flag.String("proxy", "localhost:50051", "address of the proxy server")
var hostnameOverride = flag.String("hostname-override", "", "override the hostname for TLS verification")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	discordConfig, err := base.GetConfigFromEnv("PRISM_BOT")
	if err != nil {
		return err
	}

	conf, err := config.NewConfig(*configFilePath)
	if err != nil {
		return err
	}

	logger := base.NewLogger(discordConfig)

	slog.SetDefault(logger)

	discordBot, err := base.NewBot(discordConfig, 0, logger)
	if err != nil {
		return err
	}

	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	if *certPath != "" {
		creds, err := credentials.NewClientTLSFromFile(*certPath, *hostnameOverride)
		if err != nil {
			return err
		}
		dialOption = grpc.WithTransportCredentials(creds)
	}

	conn, err := grpc.NewClient(*proxyAddr, dialOption)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := prismproxy.NewProxyClient(conn)

	if conf.Chat != nil {
		err = chat.NewChat(discordBot.Session(), client, conf.Chat.ChannelID).Start(context.Background())
		if err != nil {
			return err
		}
	}

	commands.New(client).Register(discordBot)

	if conf.ServerDetails != nil {
		details, err := details.New(conf.ServerDetails.Channels, client)
		if err != nil {
			return err
		}

		details.Register(discordBot)
	}

	discordBot.Start()
	defer discordBot.Stop()

	base.BlockExit()
	return nil
}

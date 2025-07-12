package discordbot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Bot wraps the Discord session and message handler
// MessageHandler is a function that handles Discord messages
// func(s *discordgo.Session, m *discordgo.MessageCreate)
type MessageHandler func(s *discordgo.Session, m *discordgo.MessageCreate)

type Bot struct {
	Session        *discordgo.Session
	MessageHandler MessageHandler
}

// NewBot creates and configures a new Discord bot
func NewBot(token string, handler MessageHandler) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}

	dg.AddHandler(handler)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	return &Bot{
		Session:        dg,
		MessageHandler: handler,
	}, nil
}

// Run starts the bot and blocks until termination
func (b *Bot) Run() error {
	if err := b.Session.Open(); err != nil {
		return fmt.Errorf("error opening Discord connection: %v", err)
	}
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	b.Session.Close()
	return nil
}

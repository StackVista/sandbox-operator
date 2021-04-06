package slack

import (
	"github.com/slack-go/slack"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification"
)

type Slacker struct {
	client *slack.Client
	config *config.SlackConfig
}

var _ notification.Notifier = (*Slacker)(nil) // Compile-time check

func NewSlacker(config *config.SlackConfig) (*Slacker, error) {
	return &Slacker{
		client: slack.New(config.ApiKey),
		config: config,
	}, nil
}

// Post a message.
// if channelID is given, will post to the given channelID, else it will be posted
// to the default channelID
func (s *Slacker) Notify(channelID string, message string) error {
	msgOpts := s.constructMsgOpts(message)

	channel := channelID
	if channelID != "" {
		channel = s.config.ChannelID
	}

	if _, _, err := s.client.PostMessage(channel, msgOpts...); err != nil {
		return err
	}

	return nil
}

func (s *Slacker) constructMsgOpts(message string) []slack.MsgOption {

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(message, false),
	}

	if s.config.PostAsUser != "" {
		msgOpts = append(msgOpts, slack.MsgOptionUsername(s.config.PostAsUser))
	}

	if s.config.PostAsIconURL != "" {
		msgOpts = append(msgOpts, slack.MsgOptionIconURL(s.config.PostAsIconURL))
	}

	return msgOpts
}

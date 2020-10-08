package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

type Config struct {
	ApiKey        string `split_words:"true" required:"true"`
	ChannelID     string `split_words:"true" required:"false"`
	PostAsUser    string `split_words:"true" required:"false"`
	PostAsIconURL string `split_words:"true" required:"false"`
}

type Slacker struct {
	client *slack.Client
	config *Config
}

func NewSlacker(config *Config) *Slacker {
	return &Slacker{
		client: slack.New(config.ApiKey),
		config: config,
	}
}

// Post a message prefixing it with @user.
// if channelID is given, will post to the given channelID, else it will be posted
// to the default channelID
func (s *Slacker) NotifyUser(slackID string, channelID string, message string) error {
	msgOpts := s.constructMsgOpts(slackID, message)

	channel := channelID
	if channelID != "" {
		channel = s.config.ChannelID
	}

	if _, _, err := s.client.PostMessage(channel, msgOpts...); err != nil {
		return err
	}

	return nil
}

func (s *Slacker) constructMsgOpts(slackID string, message string) []slack.MsgOption {
	msg := fmt.Sprintf("<@%s>, "+message, slackID)

	msgOpts := []slack.MsgOption{
		slack.MsgOptionText(msg, false),
	}

	if s.config.PostAsUser != "" {
		msgOpts = append(msgOpts, slack.MsgOptionUsername(s.config.PostAsUser))
	}

	if s.config.PostAsIconURL != "" {
		msgOpts = append(msgOpts, slack.MsgOptionIconURL(s.config.PostAsIconURL))
	}

	return msgOpts
}

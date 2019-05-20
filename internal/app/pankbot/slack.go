package pankbot

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nlopes/slack"
	"github.com/pendo-io/appwrap"
	"golang.org/x/net/context"
)

var slackUrls = map[string]string{}

var SlackGood = "good"
var SlackWarning = "warning"
var SlackDanger = "danger"

// https://api.slack.com/custom-integrations
type SlackMessage struct {
	Text         string             `json:"text"`
	ResponseType string             `json:"response_type,omitempty"`
	Link         string             `json:"-"` // Not part of slack API
	Channel      string             `json:"channel,omitempty"`
	Markdown     bool               `json:"mrkdwn,omitempty"`
	Attachments  []slack.Attachment `json:"attachments,omitempty"`
	User         string             `json:"user,omitempty"`
}

type UnknownSlackCommand struct {
	err error
}

func (e UnknownSlackCommand) Error() string {
	return fmt.Sprintf("%s", e.err)
}

func (sm SlackMessage) PostMessage(ctx context.Context, log appwrap.Logging, channels []string, command string) error {
	token := slackAuthToken[command]
	api := slack.New(token)
	for _, channel := range channels {
		log.Debugf("Begin sending of slack message %+v in channel %s", sm, channel)
		sm.Channel = channel

		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(sm)
		log.Debugf("Posting slack message")

		options := []slack.MsgOption{
			slack.MsgOptionText(sm.Text, false),
			slack.MsgOptionAttachments(sm.Attachments...),
		}

		if sm.ResponseType == "ephemeral" {
			options = append(options, slack.MsgOptionPostEphemeral(sm.User))
		}

		if _, _, err := api.PostMessage(channel, options...); err != nil {
			log.Debugf("Slack responded: %s", err)
			switch err.(type) {
			case NotJsonContentType:
				// slack is responding with text/html and I don't know why
			default:
				log.Errorf("Slack post failed with %v: %v", err, sm)
				return err
			}
		}
	}
	return nil
}

func (sm *SlackMessage) AddAttachment(title, titleLink, text, color string, mrkdwn []string) *slack.Attachment {
	a := slack.Attachment{
		Title:      title,
		TitleLink:  titleLink,
		Text:       text,
		Color:      color,
		MarkdownIn: mrkdwn,
	}
	sm.Attachments = append(sm.Attachments, a)
	return &sm.Attachments[len(sm.Attachments)-1]
}

func NewSlackMessage(text string, user string, attachmentCount int) *SlackMessage {
	var sm SlackMessage
	sm.Text = text
	sm.Markdown = true
	sm.User = user
	if attachmentCount > 0 {
		sm.Attachments = make([]slack.Attachment, 0, attachmentCount)
	}
	return &sm
}

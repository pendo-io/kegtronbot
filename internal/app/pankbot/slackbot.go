package pankbot

import (
	"net/http"
	"net/url"

	"google.golang.org/appengine"
)

var slackVerificationToken = map[string]string{}
var slackAuthToken = map[string]string{}

type SlackBotMessage struct {
	ChannelId   string
	ChannelName string
	Command     string
	ReponseUrl  string
	TeamDomain  string
	TeamId      string
	Text        string
	Token       string
	UserId      string
	UserName    string
	TriggerId   string
}

func NewSlackBotMessage(form url.Values) *SlackBotMessage {
	m := SlackBotMessage{
		ChannelId:   form["channel_id"][0],
		ChannelName: form["channel_name"][0],
		Command:     form["command"][0],
		ReponseUrl:  form["response_url"][0],
		TeamDomain:  form["team_domain"][0],
		TeamId:      form["team_id"][0],
		Text:        form["text"][0],
		Token:       form["token"][0],
		UserId:      form["user_id"][0],
		UserName:    form["user_name"][0],
		TriggerId:   form["trigger_id"][0],
	}
	return &m
}

func HandleSlack(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log := getLog(ctx)

	fetchAuthInfo(ctx, log)
	if err := assertFormEncodedContentType(r.Header, r.RequestURI); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unhandled data\n"))
		return
	}

	r.ParseForm()
	form := NewSlackBotMessage(r.Form)
	log.Debugf("Form: %v", form)
	if form.Token != slackVerificationToken[form.Command] {
		log.Errorf("Unrecognized token %s", form.Token)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access attempt not authorized\n"))
		return
	}

	handleSlackCommand(ctx, log, w, r, form, false)
}

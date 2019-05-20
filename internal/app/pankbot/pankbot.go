package pankbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/appengine"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/runtime"
	"google.golang.org/appengine/urlfetch"

	"github.com/nlopes/slack"
	"github.com/pendo-io/appwrap"
)

const (
	PANK_RECIPIENT_NAME  = "pankRecipient"
	PANK_TYPE_NAME       = "pankType"
	PANK_REASON_NAME     = "pankReason"
	PANK_IS_PRIVATE_NAME = "pankIsPrivate"

	PANK_COMMAND = "/pank"

	ALREADY_PANKED_ERROR_PREFIX = "You already gave your monthly"
	PANK_YOURSELF_ERROR         = "Don't pank yourself, that's just weird!"

	PANK_DIALOG_CALLBACK = "pank_dialog_callback"
	PANK_BACK_CALLBACK   = "pank_back_callback"

	PANK_BACK_NAME = "pankBack"
	JOIN_PANK_NAME = "joinPank"
)

type Pank struct {
	Giver       string    `json:"giver" datastore:"giver"` // will never change case because of mistake
	Recipient   string    `json:"recipient" datastore:"recipient"`
	RecipientId string    `json:"-" datastore:"-"`
	Type        string    `json:"pank_type" datastore:"pank_type"`
	Reason      string    `json:"reason" datastore:"reason,noindex"`
	Private     bool      `json:"private" datastore:"private,noindex"`
	Timestamp   time.Time `json:"timestamp" datastore:"timestamp"`
}

type PankFollowUp struct {
	Recipient string `json:"recipient"`
	Type      string `json:"type"`
}

func NewPank(giver, recipient, recipientId, pankType, reason string, private bool) *Pank {
	p := Pank{
		Giver:       giver,
		Recipient:   recipient,
		RecipientId: recipientId,
		Type:        pankType,
		Reason:      reason,
		Private:     private,
	}
	p.Timestamp = time.Now()
	return &p
}

var validPankList = []string{
	":pank-life:",
	":pank-customer:",
	":pank-bias-to-act:",
	":pank-data:",
	":pank-honesty:",
	":pank-transparent:",
	":pank-freedom:",
	":pank-on-the-back:",
}
var validPanks = func() map[string]bool {
	p := map[string]bool{}
	for _, pank := range validPankList {
		p[pank] = true
	}
	return p
}()

// these images are being served from imgur
// we should probably change this in the future...  :D
var pankImages = map[string]string{
	":pank-bias-to-act:": "http://i.imgur.com/nEeqsqy.png",
	":pank-customer:":    "http://i.imgur.com/Vcenl5l.png",
	":pank-data:":        "http://i.imgur.com/6tlNYiq.png",
	":pank-freedom:":     "http://i.imgur.com/w8V46q3.png",
	":pank-honesty:":     "http://i.imgur.com/WSzYsom.png",
	":pank-life:":        "http://i.imgur.com/kphu5GF.png",
	":pank-on-the-back:": "http://i.imgur.com/qQC1wt2.png",
	":pank-transparent:": "http://i.imgur.com/pnd8PeL.png",
}

var pankNames = map[string]string{
	":pank-bias-to-act:": "Bias to Act",
	":pank-customer:":    "Maniacal Focus on the Customer",
	":pank-data:":        "Show Me the Data",
	":pank-freedom:":     "Freedom and Responsibility",
	":pank-honesty:":     "Brutal Honesty",
	":pank-life:":        "Promote Life Outside of Work",
	":pank-on-the-back:": "Pank on the Back",
	":pank-transparent:": "Be Transparent",
}

func isPrivatePank(data []string) bool {
	private := strings.ToLower(data[len(data)-1])
	// these are all interpretations that have been made of the help text from time to time….
	if private == "private" || private == "privately" || private == "<private<" || private == "<privately>" || private == "<private(ly)<" || private == "<private(ly)>" || private == "private(ly)" {
		return true
	}
	return false
}

func validatePank(data []string, log appwrap.Logging) error {
	// these are all interpretations that have been made of the help text from time to time….
	if isPrivatePank(data) {
		data = data[:len(data)-1]
	}

	// check for a pank
	pank := data[0]
	if _, exists := validPanks[pank]; !exists {
		return fmt.Errorf("You forgot your pank emoji. Valid emoji are: `%s`", strings.Join(mapKeyList(validPanks), "` `"))
	}

	// check for recipient
	recipient := data[1]
	if recipient[0:1] != "@" {
		return fmt.Errorf("You forgot the recipient handle")
	}

	// check for a reason
	if len(data) <= 2 {
		return fmt.Errorf("Please say why you are panking")
	}
	return nil
}

func ParsePank(command, user string, log appwrap.Logging) (*Pank, error) {
	data := strings.Fields(command)
	if err := validatePank(data, log); err != nil {
		return nil, err
	}
	giver := "@" + user
	recipient := data[1]
	pankType := data[0]
	private := isPrivatePank(data)
	if private {
		data = data[:len(data)-1]
	}
	reason := strings.Join(data[2:], " ")

	return NewPank(giver, recipient, "", pankType, reason, private), nil
}

func postToGoogleSheets(ctx context.Context, pank *Pank) error {
	// Google sheets have been slow to respond
	return runtime.RunInBackground(ctx, func(ctx context.Context) {
		log := getLog(ctx)
		log.Debugf("posting pank %+v to el goog", pank)
		v := url.Values{}
		v.Set("giver", pank.Giver)
		v.Set("recipient", pank.Recipient)
		v.Set("pank_type", pank.Type)
		v.Set("reason", pank.Reason)
		v.Set("private", fmt.Sprintf("%v", pank.Private))
		v.Set("sent_at", pank.Timestamp.Format(time.RFC1123))
		v.Set("timestamp", fmt.Sprintf("%d", pank.Timestamp.UnixNano()))
		client := urlfetch.Client(ctx)
		// should we retry on transient error, or sync up by cron?
		if resp, err := client.PostForm("https://script.google.com/macros/s/AKfycbxCUPTHWstQl2ojurAF7QmYV8QyvOm4gJ2vik8qSphCKsXR4acw/exec", v); err != nil {
			log.Errorf("Failed posting to google sheet: %v", err)
		} else {
			log.Debugf("...finished posting to el goog: %v", resp)
		}

		return
	})
}

func publishToLeftronic(ctx context.Context, pank *Pank) error {
	log := getLog(ctx)
	if pank.Private {
		log.Debugf("not posting private pank %v to leftronic", pank)
		return nil
	}
	log.Debugf("posting pank %v to leftronic", pank)

	html := fmt.Sprintf(`
<div style='font-size:20px;display:flex;align-items:center;justify-content:center;height:100%%;'>
  <img style='display:inline-block;vertical-align:top;margin-right:20px;' width='180' height='180' src="" + image + ""/>
  <span style='text-align:left;vertical-align:top; display:inline-block;'>
    <div style='margin-bottom:20px;'>%s</div>
    <div style='font-size:40px;margin-bottom:8px;margin-bottom:'>%s</div>
    <div>received a pank from %s</div>
    <div style='font-style:italic; font-size:26px;margin-top:20px;'>&quot;%s&quot;</div>
    <img style='height:24px;margin-top:28px;' src="http://i.imgur.com/Js4ssn8.png">
  </span>
</div>`, pank.Timestamp.Format(time.RFC1123), pank.Recipient, pank.Giver, pank.Reason)

	if err := Leftronic(LeftronicAuthInfo["biastopank"], "biastopank", html).Post(ctx); err != nil {
		log.Errorf("Leftronic post failed: %v", err)
		return err
	}
	return nil
}

func quarterStart(ref time.Time) time.Time {
	// Return the first time in the fiscal quarter, which is offset by one month from calendar quarter
	// and so crosses the calendar year, and is set in America/New_York time zone.
	loc, _ := time.LoadLocation("America/New_York")
	ref = ref.In(loc)
	month := ref.Month()
	qmonth := (month - ((month+1)%3+1)%12) + 1 // negative in January, time.Date() handles that
	return time.Date(ref.Year(), qmonth, 1, 0, 0, 0, 0, loc).UTC()
}

func monthStart(ref time.Time) time.Time {
	// returns the start of the current month in America/New_York, our corporate TZ
	loc, _ := time.LoadLocation("America/New_York")
	ref = ref.In(loc)
	month := ref.Month()
	return time.Date(ref.Year(), month, 1, 0, 0, 0, 0, loc).UTC()
}

func handlesSlackPankInteractive(ctx context.Context, log appwrap.Logging, form *SlackBotMessage, defaultUser string, defaultPank string) (string, error) {
	api := slack.New(slackAuthToken[PANK_COMMAND])

	userList := slack.NewUsersSelect(PANK_RECIPIENT_NAME, "Who?")
	userList.Optional = false
	userList.Placeholder = "Choose a pendozer..."
	userList.Value = defaultUser

	panksOptions := make([]slack.DialogSelectOption, len(validPankList))
	for i, pank := range validPankList {
		panksOptions[i] = slack.DialogSelectOption{Label: fmt.Sprintf("%s %s", pank, pankNames[pank]), Value: pank}
	}

	pankList := slack.NewStaticSelectDialogInput(PANK_TYPE_NAME, "What?", panksOptions)
	pankList.Optional = false
	pankList.MinQueryLength = 1
	pankList.Placeholder = "Choose a value..."
	pankList.Value = defaultPank

	reasonText := slack.NewTextInput(PANK_REASON_NAME, "Why?", "")
	reasonText.Optional = false
	reasonText.MinLength = 1

	isPrivateOptions := []slack.DialogSelectOption{{Label: "No", Value: "false"}, {Label: "Yes", Value: "true"}}
	isPrivateList := slack.NewStaticSelectDialogInput(PANK_IS_PRIVATE_NAME, "Is it private?", isPrivateOptions)
	isPrivateList.Optional = false
	isPrivateList.Placeholder = "Is this pank private?"
	isPrivateList.Value = "false"

	dialogTrigger := slack.DialogTrigger{
		TriggerID: form.TriggerId,
		Dialog: slack.Dialog{
			TriggerID:      form.TriggerId,
			CallbackID:     PANK_DIALOG_CALLBACK,
			Title:          "Pank a Pendozer",
			NotifyOnCancel: false,
			SubmitLabel:    "Pank",
			Elements: []slack.DialogElement{
				userList,
				pankList,
				reasonText,
				isPrivateList,
			},
		},
	}

	return openDialog(false, ctx, log, dialogTrigger, api)
}

// Used to get error messages because nlopes/slack doesn't return those errors (yet?)
func openDialog(isDebugMode bool, ctx context.Context, log appwrap.Logging, dialogTrigger slack.DialogTrigger, api *slack.Client) (string, error) {
	if !isDebugMode {
		err := api.OpenDialog(dialogTrigger.TriggerID, dialogTrigger.Dialog)
		if err != nil {
			log.Errorf("Couldn't open dialog %v", err)
			return "", err
		}

		return "", err
	} else {
		encoded, err := json.Marshal(dialogTrigger)
		reqBody := bytes.NewBuffer(encoded)
		req, err := http.NewRequest("POST", "https://slack.com/api/dialog.open", reqBody)
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", slackAuthToken[PANK_COMMAND]))

		client := urlfetch.Client(ctx)
		res, err := client.Do(req)
		if err != nil {
			log.Errorf("client.Do error %+v\n", err)
			return "", err
		}

		data := slack.DialogOpenResponse{}
		responseData, err := ioutil.ReadAll(res.Body)
		if err == nil {
			if res.StatusCode != http.StatusNoContent {
				if jsonErr := json.Unmarshal(responseData, &data); jsonErr != nil {
					log.Errorf("erroneous response: %+v\n", responseData)
					return "", jsonErr
				}
			}
		}

		//
		return "", data.Err()
	}
}

func handleSlackPank(ctx context.Context, log appwrap.Logging, form *SlackBotMessage) (string, error) {
	pank, err := ParsePank(form.Text, form.UserName, log)
	if err != nil {
		return "", err
	}
	ds := appwrap.NewAppengineDatastore(ctx)
	ds = ds.Namespace("PankStore")
	if pank.Giver == pank.Recipient {
		return "", fmt.Errorf(PANK_YOURSELF_ERROR)
	}
	log.Infof("query for pank %+v", pank)
	q := ds.NewQuery("Pank").Filter(
		"giver =", pank.Giver).Filter(
		"pank_type =", pank.Type).Filter(
		"timestamp >", monthStart(time.Now()))
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("%s %s to %s", ALREADY_PANKED_ERROR_PREFIX, pank.Type, p.Recipient)
	}
	key := ds.NewKey("Pank", "", 0, nil)
	if k, err := ds.Put(key, pank); err != nil {
		log.Errorf("failed to put %v: %v", pank, err)
		return "", err
	} else {
		log.Debugf("key %v pank %v", k, pank)
	}
	sm := NewSlackMessage(fmt.Sprintf("<%s> gave %s to <%s> because %s", pank.Giver, pank.Type, pank.Recipient, pank.Reason), pank.Giver, 0)

	pankBackButtonValue, _ := json.Marshal(PankFollowUp{Recipient: form.UserId})
	joinPankButtonValue, _ := json.Marshal(PankFollowUp{Recipient: pank.Recipient, Type: pank.Type})

	sm.ResponseType = "" // Do not set response_type for non-response posts

	// post to giver - no attachments
	if err := sm.PostMessage(ctx, log, []string{pank.Giver}, form.Command); err != nil {
		log.Errorf("Post message to giver call failed: %v", err)
	}

	returnPankAttachment := slack.Attachment{
		CallbackID: PANK_BACK_CALLBACK,
		Actions: []slack.AttachmentAction{
			{Name: PANK_BACK_NAME, Text: fmt.Sprintf("Pank %s back", pank.Giver), Type: "button", Value: string(pankBackButtonValue)},
		},
	}
	sm.Attachments = []slack.Attachment{returnPankAttachment}

	// post to recipient - only pank back
	if err := sm.PostMessage(ctx, log, []string{pank.Recipient}, form.Command); err != nil {
		log.Errorf("Post message to recipient call failed: %v", err)
	}

	if !pank.Private {
		returnPankAttachment.Actions = append(returnPankAttachment.Actions, slack.AttachmentAction{Name: JOIN_PANK_NAME, Text: fmt.Sprintf("Join the %s", pank.Type), Type: "button", Value: string(joinPankButtonValue)})
		sm.Attachments = []slack.Attachment{returnPankAttachment}

		//post to channels - all attachments
		channels := make([]string, 0, 2)
		sm.ResponseType = "in_channel"
		if form.ChannelName != "pank-announcements" {
			// in_channel will send to pank-announcements if not private
			channels = append(channels, "#pank-announcements")
		}
		channels = append(channels, form.ChannelName)
		channels = unique(channels)

		if err := sm.PostMessage(ctx, log, channels, form.Command); err != nil {
			log.Errorf("Post message to channels call failed: %v", err)
		}
	}

	if err := postToGoogleSheets(ctx, pank); err != nil {
		log.Errorf("Google sheets call failed: %v", err)
	}
	if err := publishToLeftronic(ctx, pank); err != nil {
		log.Errorf("Leftronic call failed: %v", err)
	}
	return "", nil
}

func handleSlackPankMe(ctx context.Context, log appwrap.Logging, form *SlackBotMessage) (string, error) {
	sm := NewSlackMessage("Your current pankbot:", form.UserId, 16)
	// "Your" doesn't make sense in a public context
	// Also, this includes PRIVATE pankbot; do not change this to in_channel
	sm.ResponseType = "ephemeral"
	userName := fmt.Sprintf("@%s", form.UserName)
	monthlyPanksGiven := map[string]bool{}
	quarterlyPanksGiven := map[string]bool{}

	ds := appwrap.NewAppengineDatastore(ctx)
	ds = ds.Namespace("PankStore")

	// report on pankbot this quarter
	q := ds.NewQuery("Pank").Filter(
		"giver =", userName).Filter(
		"timestamp >", quarterStart(time.Now())).Order("timestamp")
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return "", err
		}
		sm.Attachments = append(sm.Attachments, slack.Attachment{
			Text:       fmt.Sprintf("You gave <%s> %s because %s", p.Recipient, p.Type, p.Reason),
			MarkdownIn: []string{"text"},
			Color:      SlackGood,
		})
		quarterlyPanksGiven[p.Type] = true
	}

	// limit to pankbot this month
	q = ds.NewQuery("Pank").Filter(
		"giver =", userName).Filter(
		"timestamp >", monthStart(time.Now())).Order("timestamp")
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return "", err
		}
		monthlyPanksGiven[p.Type] = true
	}

	// report on remaining pankbot to give this month
	remainingPanks := []string{}
	for _, pankType := range validPankList {
		if _, exists := monthlyPanksGiven[pankType]; !exists {
			remainingPanks = append(remainingPanks, fmt.Sprintf("%s (`%s`)", pankType, pankType))
		}
	}
	if len(remainingPanks) > 0 {
		sm.Attachments = append(sm.Attachments, slack.Attachment{
			Text:       fmt.Sprintf("This month, you still have to give: %s", strings.Join(remainingPanks, " ")),
			MarkdownIn: []string{"text"},
			Color:      SlackWarning,
		})
	}

	// report on pankbot received this quarter
	q = ds.NewQuery("Pank").Filter(
		"recipient =", userName).Filter(
		"timestamp >", quarterStart(time.Now())).Order("timestamp")
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return "", err
		}
		sm.Attachments = append(sm.Attachments, slack.Attachment{
			Text:       fmt.Sprintf("<%s> gave you %s because %s", p.Giver, p.Type, p.Reason),
			MarkdownIn: []string{"text"},
			Color:      SlackGood,
		})
	}

	err := sm.PostMessage(ctx, log, []string{form.ChannelId}, form.Command)

	return "", err
}

func handleSlackPankReport(ctx context.Context, log appwrap.Logging, form *SlackBotMessage) (string, error) {
	sm := NewSlackMessage("Current public pankbot:", form.UserId, 1)
	// This gets big, don't litter the public channel after all
	sm.ResponseType = "ephemeral"
	text := make([]string, 0, 1000)

	ds := appwrap.NewAppengineDatastore(ctx)
	ds = ds.Namespace("PankStore")
	q := ds.NewQuery("Pank").Filter(
		"timestamp >", quarterStart(time.Now())).Order(
		"-timestamp") // This will exceed slack message limit size, so give most recent first
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return "", err
		}
		if !p.Private {
			text = append(text, fmt.Sprintf("<%s> gave <%s> %s because %s", p.Giver, p.Recipient, p.Type, p.Reason))
		}
	}

	sm.Attachments = append(sm.Attachments, slack.Attachment{
		Text:       strings.Join(text, "\n"),
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	})

	err := sm.PostMessage(ctx, log, []string{form.ChannelId}, form.Command)

	return "", err
}

func handleSlackPankHelpCommand(ctx context.Context, log appwrap.Logging, form *SlackBotMessage) {
	sm := NewSlackMessage("Available `/pank` commands", form.UserId, 0)
	sm.ResponseType = "ephemeral"
	sm.Attachments = []slack.Attachment{{
		Text:       "*`/pank`*\n_Open the `Pank a Pendozer` dialog_",
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	}, {
		Text:       fmt.Sprintf("*`/pank [:pank-emoji:] [@user] [reason] [<private(ly)>]`*\nE.g. `/pank :pank-data: @recipient description of reason`\n`/pank :pank-life: @recipient some personal reason privately`\n_Type `:pank-` and Slack will auto-complete your pank_\nValid emoji are: `%s`\nNote: private pankbot are visible to you, the recipient, administrators, and the culture club, but not posted to #pank-announcements", strings.Join(mapKeyList(validPanks), "` `")),
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	}, {
		Text:       "*`/pank me`*\n_What pankbot you have sent and received (your eyes only)_",
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	}, {
		Text:       "*`/pank report`*\n_Display a report of all *public* pankbot (your eyes only)_",
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	}, {
		Text:       "*`/pank help`*\n_You seem to have figured this one out already, good job!_",
		MarkdownIn: []string{"text"},
		Color:      SlackGood,
	}}
	sm.PostMessage(ctx, log, []string{form.ChannelId}, form.Command)
}

func handleSlackCommand(ctx context.Context, log appwrap.Logging, w http.ResponseWriter, r *http.Request, form *SlackBotMessage, isInteractive bool) {
	if message, err := handleSlackPankCommand(ctx, log, w, r, form); err != nil {
		if isInteractive && err.Error() == PANK_YOURSELF_ERROR {
			errors := slack.DialogInputValidationErrors{}
			errors.Errors = append(errors.Errors, slack.DialogInputValidationError{Name: PANK_RECIPIENT_NAME, Error: err.Error()})
			encoded, _ := json.Marshal(errors)
			reqBody := bytes.NewBuffer(encoded)

			w.Write(reqBody.Bytes())
		} else if isInteractive && strings.HasPrefix(err.Error(), ALREADY_PANKED_ERROR_PREFIX) {
			errors := slack.DialogInputValidationErrors{}
			errors.Errors = append(errors.Errors, slack.DialogInputValidationError{Name: PANK_TYPE_NAME, Error: err.Error()})
			encoded, _ := json.Marshal(errors)
			reqBody := bytes.NewBuffer(encoded)

			w.Write(reqBody.Bytes())
		} else {
			w.Header().Set("Content-Type", "application/json")
			sm := NewSlackMessage(fmt.Sprintf("Try *`%s help`* for instructions", form.Command), form.UserId, 1)
			sm.ResponseType = "ephemeral"
			sm.Attachments = append(sm.Attachments, slack.Attachment{
				Text:       fmt.Sprintf("You typed: `%s %s`", form.Command, form.Text),
				MarkdownIn: []string{"text"},
				Color:      SlackWarning,
			})
			sm.Attachments = append(sm.Attachments, slack.Attachment{
				Text:       fmt.Sprintf("%v", err),
				MarkdownIn: []string{"text"},
				Color:      SlackDanger,
			})
			b, _ := json.Marshal(sm)
			w.Write(b)
		}
	} else if message != "" {
		// Message to be overwritten by later in-channel delayed response
		w.Header().Set("Content-Type", "application/json")
		sm := NewSlackMessage(message, form.UserName, 0)
		sm.ResponseType = "ephemeral"
		b, _ := json.Marshal(sm)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}
}

func handleSlackPankCommand(ctx context.Context, log appwrap.Logging, w http.ResponseWriter, r *http.Request, form *SlackBotMessage) (string, error) {
	if form.Text == "" {
		return handlesSlackPankInteractive(ctx, log, form, "", "")
	}

	args := strings.SplitN(form.Text, " ", 2)

	command := args[0]
	log.Debugf("form: %v", form)
	switch command {
	case "help":
		handleSlackPankHelpCommand(ctx, log, form)
		return "", nil
	case "me":
		return handleSlackPankMe(ctx, log, form)
	case "report":
		return handleSlackPankReport(ctx, log, form)
	case "int":
		return handlesSlackPankInteractive(ctx, log, form, "", "")
	default:
		if message, err := handleSlackPank(ctx, log, form); err != nil {
			return "", err
		} else {
			return message, nil
		}
	}
}

func HandleInteractivePankResponse(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log := getLog(ctx)

	fetchAuthInfo(ctx, log)

	if err := assertFormEncodedContentType(r.Header, r.RequestURI); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unhandled data\n"))
		return
	}

	r.ParseForm()
	byt := []byte(r.Form["payload"][0])
	var callback slack.InteractionCallback
	if err := json.Unmarshal(byt, &callback); err != nil {
		log.Errorf("Could not unmarshall response %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Callback Response\n"))
		return
	}

	log.Debugf("Form: %v", callback)
	if callback.Token != slackVerificationToken[PANK_COMMAND] {
		log.Errorf("Unrecognized token %s", callback.Token)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access attempt not authorized\n"))
		return
	}

	api := slack.New(slackAuthToken[PANK_COMMAND])

	slackMessage := &SlackBotMessage{
		ChannelId:   callback.Channel.ID,
		ChannelName: callback.Channel.Name,
		Command:     "/pank",
		ReponseUrl:  callback.ResponseURL,
		TeamDomain:  callback.Team.Domain,
		TeamId:      callback.Team.ID,
		Text:        "",
		Token:       callback.Token,
		UserId:      callback.User.ID,
		UserName:    callback.User.Name,
		TriggerId:   callback.TriggerID,
	}
	if callback.CallbackID == PANK_DIALOG_CALLBACK {
		recipient, err := api.GetUserInfo(callback.Submission[PANK_RECIPIENT_NAME])
		if err != nil {
			log.Errorf("Error getting user %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting user info\n"))
			return

		}

		text := fmt.Sprintf("%s @%s %s", callback.Submission[PANK_TYPE_NAME], recipient.Name, callback.Submission[PANK_REASON_NAME])
		if callback.Submission[PANK_IS_PRIVATE_NAME] == "true" {
			text = text + " private"
		}

		slackMessage.Text = text

		handleSlackCommand(ctx, log, w, r, slackMessage, true)
	} else if callback.CallbackID == PANK_BACK_CALLBACK {
		actionCallback := callback.ActionCallback

		rePank := PankFollowUp{}
		json.Unmarshal([]byte(actionCallback.Actions[0].Value), &rePank)

		defaultUser := rePank.Recipient
		defaultPank := rePank.Type

		if actionCallback.Actions[0].Name == JOIN_PANK_NAME {
			users, err := api.GetUsers()
			if err != nil {
				log.Errorf("Error getting user %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error getting user info\n"))
				return

			}

			for _, user := range users {
				if user.Name == rePank.Recipient[1:] {
					defaultUser = user.ID
					break
				}
			}
		}

		if message, err := handlesSlackPankInteractive(ctx, log, slackMessage, defaultUser, defaultPank); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.Write([]byte(message))
		}
	}
}

func GetPanks(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//log := getLog(ctx)

	ds := appwrap.NewAppengineDatastore(ctx)
	ds = ds.Namespace("PankStore")

	q := ds.NewQuery("Pank")
	panks := make([]Pank, 0)
	for t := q.Run(); ; {
		var p Pank
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		panks = append(panks, p)
	}

	if panksJson, err := json.Marshal(panks); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write(panksJson)
	}
}

func PostPanks(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log := getLog(ctx)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading request body: %s", err)
		http.Error(w, "500 Internal Server Error\n", http.StatusInternalServerError)
		return
	}
	panksBody := []Pank{}
	json.Unmarshal(body, &panksBody)

	ds := appwrap.NewAppengineDatastore(ctx)
	ds = ds.Namespace("PankStore")

	for _, pank := range panksBody {
		key := ds.NewKey("Pank", "", 0, nil)
		if k, err := ds.Put(key, &pank); err != nil {
			log.Errorf("failed to put %v: %+v", pank, err)
		} else {
			log.Debugf("key %v pank %v", k, pank)
		}
	}

	w.Write([]byte("done"))
}

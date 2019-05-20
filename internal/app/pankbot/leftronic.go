package pankbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

type leftronicPayload struct {
	AccessKey  string `json:"accessKey"`
	StreamName string `json:"streamName"`
	Point      struct {
		Html string `json:"html"`
	} `json:"point"`
}

var leftronicUrl = "https://pushapi.appinsights.com/"

func Leftronic(accessKey, streamName, html string) *leftronicPayload {
	l := &leftronicPayload{
		AccessKey:  accessKey,
		StreamName: streamName,
	}
	l.Point.Html = html
	return l
}

func (l *leftronicPayload) Post(ctx context.Context) error {
	log := getLog(ctx)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(l)
	var statusCode int
	if err := makeJsonRequest(ctx, http.MethodPost, leftronicUrl, b, nil, &statusCode, func(*http.Request) {}); err != nil {
		log.Debugf("Leftronic responded: %s", err)
		switch err.(type) {
		case NotJsonContentType:
			// leftronic is another html response to json perpetrator
		default:
			return err
		}
	}
	if statusCode != http.StatusOK {
		log.Errorf("Leftronic post failed with status %d: %v", statusCode, l)
		return fmt.Errorf("Leftronic post failed with status %d", statusCode)
	}
	return nil
}

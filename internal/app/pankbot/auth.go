package pankbot

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/pendo-io/appwrap"
)

const authCacheTime = 10 * time.Minute

var authInitializedAt time.Time

type GoogleGitHubLoginMap struct {
	GoogleLogin string `datastore:"google_login" json:"google_login"`
	GithubLogin string `datastore:"github_login" json:"github_login"`
}

type SlackAuthData struct {
	BotCommand        string `datastore:"bot_command" json:"bot_command"`
	BotUrl            string `datastore:"bot_url" json:"bot_url"`
	VerificationToken string `datastore:"verification_token" json:"bot_token"`
	OAuthToken        string `datastore:"oauth_token" json:"-"`
}

type leftronicAuthInfo struct {
	AccessKey  string `datastore:"accessKey" json:"accessKey"`
	StreamName string `datastore:"streamName" json:"streamName"`
}

type GoogleAuthInfo struct {
	ClientSecret   string `datastore:"client_secret" json:"client_secret"`
	ClientEmail    string `datastore:"client_email" json:"client_email"`
	DefaultSubject string `datastore:"default_subject" json:"default_subject"`
}

type PendoKey struct {
	KeyName  string `json:"keyName"`
	KeyValue string `json:"keyValue"`
}

var LeftronicAuthInfo = map[string]string{}

var GSuiteJWT = &jwt.Config{}

// `/_ah/start` is not available for automatically scaling instances,
// so each entry point needs to `fetchAuthInfo()` before doing anything else
func fetchAuthInfo(ctx context.Context, log appwrap.Logging) {
	if time.Now().Sub(authInitializedAt) < authCacheTime {
		return
	}
	ds := appwrap.NewAppengineDatastore(ctx).Namespace("AuthStore")

	q := ds.NewQuery("Slack")
	for t := q.Run(); ; {
		s := &SlackAuthData{}
		_, err := t.Next(s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf("failed to get Slack authinfo %v in %v", err, s)
		} else {
			slackUrls[s.BotCommand] = s.BotUrl
			slackVerificationToken[s.BotCommand] = s.VerificationToken
			slackAuthToken[s.BotCommand] = s.OAuthToken
		}
	}

	q = ds.NewQuery("Leftronic")
	for t := q.Run(); ; {
		l := &leftronicAuthInfo{}
		_, err := t.Next(l)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf("failed to get Leftronic authinfo %v in %v", err, l)
		} else {
			LeftronicAuthInfo[l.StreamName] = l.AccessKey
		}
	}

	q = ds.NewQuery("Google")
	g := &GoogleAuthInfo{}
	for t := q.Run(); ; {
		_, err := t.Next(g)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf("failed to get Google authinfo %v in %v", err, g)
		} else {
			GSuiteJWT, err = google.JWTConfigFromJSON([]byte(g.ClientSecret), "https://www.googleapis.com/auth/admin.directory.group.readonly")
			GSuiteJWT.Subject = g.DefaultSubject
			if err != nil {
				log.Errorf("failed to create JWT: %v", err)
			}
			log.Infof("G Suite credentials loaded for %s, key length %d", g.ClientEmail, len(GSuiteJWT.PrivateKey))
		}
	}

	//q = ds.NewQuery("PendoKey").Filter("KeyName =", pendoKeyName_impersonate)
	//for t := q.Run(); ; {
	//	if _, err := t.Next(PendoKeyImpersonate); err == datastore.Done {
	//		break
	//	} else if err != nil {
	//		log.Errorf("failed to get Pendo EnableImpersonate key: %v", err)
	//	}
	//}

	//ds = appwrap.NewAppengineDatastore(ctx).Namespace("IdStore")
	//q = ds.NewQuery("GitHubGoogleLoginMap")
	//u := GoogleGitHubLoginMap{}
	//for t := q.Run(); ; {
	//	_, err := t.Next(&u)
	//	if err == datastore.Done {
	//		break
	//	}
	//	if err != nil {
	//		log.Errorf("failed to get some user id info %v", err)
	//	} else {
	//		GoogleToGitHub[u.GoogleLogin] = u.GithubLogin
	//		GitHubToGoogle[u.GithubLogin] = u.GoogleLogin
	//	}
	//}
	authInitializedAt = time.Now()
}

func HandleSetAuth(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ds := appwrap.NewAppengineDatastore(ctx).Namespace("AuthStore")
	log := getLog(ctx)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading request body: %s", err)
		http.Error(w, "500 Internal Server Error\n", http.StatusInternalServerError)
		return
	}

	var key *datastore.Key

	authType := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"), "auth/")
	switch authType {
	case "github", "google", "leftronic", "slack":
		log.Errorf("Unimplemented authType: %s", authType)
		http.Error(w, "Not implemented\n", http.StatusNotImplemented)
		return
	case "refetch":
		log.Infof("Forcing re-fetch of auth data")
	case "pendo":
		var pendoKey PendoKey
		if err = unmarshalExactJson(body, &pendoKey); err != nil {
			log.Errorf("Malformatted Pendo Key: %s", err)
			http.Error(w, "Malformatted Data", http.StatusBadRequest)
			return
		}
		key = ds.NewKey("PendoKey", pendoKey.KeyName, 0, nil)
		log.Infof("Changing Pendo Key %v in AuthStore: len %d", key, len(pendoKey.KeyValue))
		_, err = ds.Put(key, &pendoKey)
	default:
		log.Errorf("Unknown authType: %s", authType)
		http.Error(w, "404 page not found\n", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Errorf("failed to put new auth data into key %+v: %s", key, err)
		http.Error(w, "500 Internal Server Error\n", http.StatusInternalServerError)
		return
	}

	// Success: force auth re-fetch on next call, return 204
	authInitializedAt = time.Time{}
	w.WriteHeader(http.StatusNoContent)
}

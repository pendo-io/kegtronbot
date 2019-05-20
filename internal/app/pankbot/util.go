package pankbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"

	"github.com/pendo-io/appwrap"
)

type NotJsonContentType struct {
	err error
}

func (e NotJsonContentType) Error() string {
	return fmt.Sprintf("%s", e.err)
}

type NotFormEncodedContentType struct {
	err error
}

func (e NotFormEncodedContentType) Error() string {
	return fmt.Sprintf("%s", e.err)
}

func assertJsonContentType(header http.Header, url string) error {
	contentType := header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0]
	if strings.Compare(contentType, "application/json") != 0 {
		return NotJsonContentType{fmt.Errorf("Unexpected Content-Type %v from %s; expected application/json", contentType, url)}
	}
	return nil
}

func assertFormEncodedContentType(header http.Header, url string) error {
	contentType := header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0]
	if strings.Compare(contentType, "application/x-www-form-urlencoded") != 0 {
		return NotFormEncodedContentType{fmt.Errorf("Unexpected Content-Type %v from %s; expected application/x-www-form-urlencoded", contentType, url)}
	}
	return nil
}

func assertJsonContentResponse(res *http.Response, url string) error {
	if err := assertJsonContentType(res.Header, url); err != nil {
		var b string
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			b = string(body)
		}
		return NotJsonContentType{fmt.Errorf("%s for: %s", err, b)}
	}
	return nil
}

func makeJsonRequestResponse(ctx context.Context, method, url string, body io.Reader, data interface{}, statusCode *int, setAuth func(*http.Request)) (*http.Response, error) {
	return makeJsonRequestResponseWithHeaders(ctx, method, url, body, data, statusCode, setAuth, map[string]string{})
}

func makeJsonRequestResponseWithHeaders(ctx context.Context, method, url string, body io.Reader, data interface{}, statusCode *int, setAuth func(*http.Request), headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &http.Response{}, err
	}

	log := getLog(ctx)

	setAuth(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := urlfetch.Client(ctx)
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("client.Do error %#v\n", err)
		return res, err
	}
	*statusCode = res.StatusCode

	if err = assertJsonContentResponse(res, url); err != nil {
		return res, err
	}

	if responseData, err := ioutil.ReadAll(res.Body); err == nil {
		if data != nil && res.StatusCode != http.StatusNoContent {
			if jsonErr := json.Unmarshal(responseData, data); jsonErr != nil {
				log.Errorf("erroneous response: %#v\n", responseData)
				return res, jsonErr
			}
		}
	}

	return res, err
}

func makeJsonRequest(ctx context.Context, method, url string, body io.Reader, data interface{}, statusCode *int, setAuth func(*http.Request)) error {
	_, err := makeJsonRequestResponse(ctx, method, url, body, data, statusCode, setAuth)
	return err
}

func fileUpload(ctx context.Context, url string, filename string, contents []byte, setAuth func(*http.Request)) (err error) {
	// Prepare a form that you will submit to that URL.
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	part.Write(contents)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	writer.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	setAuth(req)

	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Submit the request
	client := urlfetch.Client(ctx)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func getLog(ctx context.Context) appwrap.Logging {
	return appwrap.NewAppengineLogging(ctx)
}

func mapListKey(l []string) map[string]bool {
	m := make(map[string]bool, len(l))
	for _, k := range l {
		m[k] = true
	}
	return m
}

func mapKeyList(m map[string]bool) []string {
	l := make([]string, len(m))
	i := 0
	for k := range m {
		l[i] = k
		i++
	}
	sort.Sort(sort.StringSlice(l))
	return l
}

func unique(l []string) []string {
	d := make(map[string]bool, len(l))
	for _, item := range l {
		if len(item) > 0 {
			d[item] = true
		}
	}
	return mapKeyList(d)
}

func stringDifference(a, b []string) []string {
	r := make([]string, 0, len(a))
	m := mapListKey(b)
	for _, s := range a {
		if exists, _ := m[s]; !exists {
			r = append(r, s)
		}
	}
	return r
}

func stringIntersection(a, b []string) []string {
	r := make([]string, 0, len(a))
	m := mapListKey(b)
	for _, s := range a {
		if exists, _ := m[s]; exists {
			r = append(r, s)
		}
	}
	return r
}

// because _reasons_
func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func stringSliceContains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func stringSliceToDelimitedString(slice []string, delimiter string) string {
	out := ""
	sliceCnt := len(slice) - 1

	for i, item := range slice {
		out += item
		if i < sliceCnt {
			out += delimiter
		}
	}
	return out
}

// This is a quick hack to make sure all fields in simple types are
//  included in a request. It does not support optional fields:
//  omitempty fields can be left blank but are not advised because zero
//  values submitted to omitempty fields will fail to match.
// martini-contrib/binding provides much better tools
func unmarshalExactJson(data []byte, v interface{}) error {
	var dummy1, dummy2 interface{}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	} else if err = json.Unmarshal(data, &dummy1); err != nil {
		return err
	} else if vBytes, err := json.Marshal(v); err != nil {
		return err
	} else if err = json.Unmarshal(vBytes, &dummy2); err != nil {
		return err
	} else if !reflect.DeepEqual(dummy1, dummy2) {
		return fmt.Errorf("unmarshalExactJson: fields do not match target")
	}
	return nil
}

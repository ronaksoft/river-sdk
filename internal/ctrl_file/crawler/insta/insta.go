package insta

import (
	"encoding/json"
	"errors"
	"github.com/siongui/instago"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
   Creation Time: 2020 - Jul - 02
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

const urlPost = `https://www.instagram.com/p/{{CODE}}/?__a=1`

var (
	RegExExtractCode = regexp.MustCompile("https?://www.instagram.com/(p|tv)/([^\\/]+)(/|$)")
)

// Decode JSON data returned by Instagram post API
type postInfo struct {
	GraphQL struct {
		ShortcodeMedia instago.IGMedia `json:"shortcode_media"`
	} `json:"graphql"`
}

// GetPostInfo Given url of post, return information of the post without login status.
func GetPostInfo(url string) (em instago.IGMedia, err error) {
	parts := RegExExtractCode.FindStringSubmatch(url)
	if len(parts) < 3 {
		return
	}

	postUrl := strings.Replace(urlPost, "{{CODE}}", parts[2], 1)
	b, err := getHTTPResponseNoLogin(postUrl)
	if err != nil {
		return
	}

	pi := postInfo{}
	err = json.Unmarshal(b, &pi)
	if err != nil {
		return
	}
	em = pi.GraphQL.ShortcodeMedia
	return
}

// Send HTTP request and get http response without login.
func getHTTPResponseNoLogin(postUrl string) (b []byte, err error) {
	httpClient := http.Client{
		Timeout: time.Minute,
	}

	resp, err := httpClient.Get(postUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(postUrl +
			"\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return
	}

	return ioutil.ReadAll(resp.Body)
}

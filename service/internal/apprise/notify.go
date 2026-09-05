package apprise

import (
	"net/http"

	apclient "github.com/jamesread/armature-apprise/client"
	aptag "github.com/jamesread/armature-apprise/tag"
	"github.com/jamesread/starapp/service/internal/buildinfo"
)

const (
	// PersonTagPrefix is the Apprise tag prefix; X is the family member (person) id.
	PersonTagPrefix = "starloom_uid_"
)

// Payload is the JSON body accepted by Apprise API /notify/{key} endpoints.
type Payload = apclient.Payload

// UserAgent returns the HTTP User-Agent StarLoom sends to Apprise.
func UserAgent() string {
	return apclient.FormatUserAgent("StarLoom", buildinfo.Version)
}

func notifyConfig() apclient.Config {
	return apclient.DefaultConfig(UserAgent())
}

// PersonTag returns the Apprise tag for a family member (person) id.
func PersonTag(personID int) string {
	return aptag.Person(PersonTagPrefix, personID)
}

// JoinPersonTags builds a comma-separated OR tag expression for the given person ids.
func JoinPersonTags(personIDs []int) string {
	return aptag.Join(PersonTagPrefix, personIDs)
}

// ValidateNotifyURL reports whether url targets a persistent Apprise configuration key.
func ValidateNotifyURL(raw string) error {
	return apclient.ValidateNotifyURL(raw)
}

// Notify POSTs payload to the Apprise API URL with retries. Empty URL is a no-op.
func Notify(client *http.Client, notifyURL string, payload Payload) error {
	return apclient.Notify(notifyConfig(), client, notifyURL, payload)
}

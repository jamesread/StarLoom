package apprise

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonTag(t *testing.T) {
	assert.Equal(t, "starloom_uid_42", PersonTag(42))
	assert.Equal(t, "starloom_uid_1,starloom_uid_3", JoinPersonTags([]int{1, 3}))
	assert.Equal(t, "", JoinPersonTags(nil))
}

func TestNotify_EmptyURLIsNoop(t *testing.T) {
	err := Notify(nil, "", Payload{Title: "t", Body: "b"})
	assert.NoError(t, err)
}

func TestNotify_PostsJSONPayloadWithTag(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotAccept string
	var gotUserAgent string
	var gotBody Payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		gotUserAgent = r.Header.Get("User-Agent")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":null,"details":[]}`))
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL+"/notify/testkey", Payload{
		Title: "Approval",
		Body:  "Please approve",
		Type:  "info",
		Tag:   "starloom_uid_7",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "application/json", gotAccept)
	assert.Equal(t, UserAgent(), gotUserAgent)
	assert.Equal(t, "Approval", gotBody.Title)
	assert.Equal(t, "Please approve", gotBody.Body)
	assert.Equal(t, "info", gotBody.Type)
	assert.Equal(t, "text", gotBody.Format)
	assert.Equal(t, "starloom_uid_7", gotBody.Tag)
}

func TestNotify_DefaultsTypeToInfo(t *testing.T) {
	var gotBody Payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":null,"details":[]}`))
	}))
	defer srv.Close()

	require.NoError(t, Notify(srv.Client(), srv.URL+"/notify/testkey", Payload{Title: "t", Body: "b"}))
	assert.Equal(t, "info", gotBody.Type)
	assert.Equal(t, "text", gotBody.Format)
}

func TestNotify_RetriesOnFailureThenSucceeds(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		if n < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":null,"details":[]}`))
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL+"/notify/testkey", Payload{Title: "t", Body: "b"})
	require.NoError(t, err)
	assert.Equal(t, int32(2), attempts.Load())
}

func TestValidateNotifyURL(t *testing.T) {
	require.NoError(t, ValidateNotifyURL("http://apprise:8000/notify/mykey"))
	require.NoError(t, ValidateNotifyURL("https://apprise.example/notify/mykey/"))

	err := ValidateNotifyURL("http://apprise:8000/notify")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration key")

	err = ValidateNotifyURL("http://apprise:8000/notify/")
	require.Error(t, err)
}

func TestNotify_RejectsNoContentAsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL+"/notify/testkey", Payload{Title: "t", Body: "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 204")
}

func TestNotify_RejectsJSONErrorField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":"delivery failed","details":[]}`))
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL+"/notify/testkey", Payload{Title: "t", Body: "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delivery failed")
}

func TestApprovalURL(t *testing.T) {
	assert.Equal(t, "/family/rewards?redemption=9", ApprovalURL("", 9))
	assert.Equal(t, "https://stars.example/family/rewards?redemption=9", ApprovalURL("https://stars.example/", 9))
}

func TestRenderRedemptionMessage(t *testing.T) {
	body := RenderRedemptionMessage(
		"{{requestor_name}} wants {{reward_name}} ({{stars}}). {{approval_url}} #{{redemption_id}}",
		RedemptionPlaceholders{
			ApprovalURL:   "https://app/family/rewards?redemption=3",
			RequestorName: "Alex",
			RewardName:    "Screen time",
			Stars:         5,
			RedemptionID:  3,
			RequestorID:   12,
		},
	)
	assert.Equal(t, "Alex wants Screen time (5). https://app/family/rewards?redemption=3 #3", body)

	fallback := RenderRedemptionMessage("", RedemptionPlaceholders{
		ApprovalURL:   "/family/rewards?redemption=1",
		RequestorName: "Sam",
		RewardName:    "Ice cream",
		Stars:         2,
	})
	assert.Contains(t, fallback, "Sam")
	assert.Contains(t, fallback, "Ice cream")
	assert.Contains(t, fallback, "/family/rewards?redemption=1")
}

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
	var gotBody Payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL, Payload{
		Title: "Approval",
		Body:  "Please approve",
		Type:  "info",
		Tag:   "starloom_uid_7",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "Approval", gotBody.Title)
	assert.Equal(t, "Please approve", gotBody.Body)
	assert.Equal(t, "info", gotBody.Type)
	assert.Equal(t, "starloom_uid_7", gotBody.Tag)
}

func TestNotify_DefaultsTypeToInfo(t *testing.T) {
	var gotBody Payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.NoError(t, Notify(srv.Client(), srv.URL, Payload{Title: "t", Body: "b"}))
	assert.Equal(t, "info", gotBody.Type)
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
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := Notify(srv.Client(), srv.URL, Payload{Title: "t", Body: "b"})
	require.NoError(t, err)
	assert.Equal(t, int32(2), attempts.Load())
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

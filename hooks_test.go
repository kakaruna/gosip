package gosip

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestHooks(t *testing.T) {
	siteURL := "http://localhost:8989"
	closer, err := startFakeServer(":8989", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// produce an error
		if r.RequestURI == "/_api/error" {
			// intentional 404
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "error": "404 Page not found" }`))
			return
		}
		// faking digest response
		if r.RequestURI == "/_api/ContextInfo" {
			fmt.Fprintf(w, `{"d":{"GetContextWebInformation":{"FormDigestValue":"FAKE","FormDigestTimeoutSeconds":120,"LibraryVersion":"FAKE"}}}`)
			return
		}
		// backoff after 2 retries
		if r.Header.Get("X-Gosip-Retry") == "2" {
			fmt.Fprintf(w, `{ "result": "Cool alfter some retries" }`)
			return
		}
		// intentional 503
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{ "error": "503 Retry Please" }`))
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	// Request counters
	var requestCntrs = struct {
		Errors    int32
		Responses int32
		Retries   int32
		Requests  int32
	}{
		Errors:    0,
		Responses: 0,
		Retries:   0,
		Requests:  0,
	}

	t.Run("Hooks", func(t *testing.T) {
		client := &SPClient{
			AuthCnfg:      &AnonymousCnfg{SiteURL: siteURL},
			RetryPolicies: map[int]int{503: 3},
			Hooks: &HookHandlers{
				OnError: func(e *HookEvent) {
					atomic.AddInt32(&requestCntrs.Errors, 1)
				},
				OnResponse: func(e *HookEvent) {
					atomic.AddInt32(&requestCntrs.Responses, 1)
				},
				OnRetry: func(e *HookEvent) {
					atomic.AddInt32(&requestCntrs.Retries, 1)
				},
				OnRequest: func(e *HookEvent) {
					atomic.AddInt32(&requestCntrs.Requests, 1)
				},
			},
		}

		if err := simpleCall(client, "/_api/get"); err != nil {
			t.Error(err)
		}

		if err := simpleCall(client, "/_api/error"); err == nil {
			t.Error("should be an error response")
		}

		// 4 requests
		if requestCntrs.Requests != 4 {
			t.Error("wrong number of requests")
		}

		// 2 retries
		if requestCntrs.Retries != 2 {
			t.Error("wrong number of retries")
		}

		// 2 response
		if requestCntrs.Responses != 2 {
			t.Error("wrong number of responses")
		}

		// 1 error
		if requestCntrs.Errors != 1 {
			t.Error("wrong number of errors")
		}

	})

}

func simpleCall(client *SPClient, uri string) error {
	req, err := http.NewRequest("GET", client.AuthCnfg.GetSiteURL()+uri, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Execute(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return fmt.Errorf("can't retry a request")
	}
	return nil
}
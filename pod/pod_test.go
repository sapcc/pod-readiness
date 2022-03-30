package pod

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

var testServer *Pod

const jsonNotRDY = "{\"ready\":false}\n"
const jsonRDY = "{\"ready\":true}\n"
const queryParamMissing = "query parameter `key` required"
const queryParamNotDefault = "query parameter `key` must not be 'default'"

func TestGetHealthy(t *testing.T) {
	setup()
	execGetRequest(t, "healthy", 200, "READY")
}

func TestGetNotHealthy(t *testing.T) {
	setup()
	execGetRequest(t, "healthy", 200, "READY")
	execPatchRequest(t, "", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "healthy", 500, "NOT READY")
}

func TestGetReadiness(t *testing.T) {
	setup()
	execGetRequest(t, "pod/readiness", 200, jsonRDY)
}

func TestPatchReadiness(t *testing.T) {
	setup()
	execPatchRequest(t, "", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "", 200, jsonRDY, []byte(jsonRDY))
	execGetRequest(t, "pod/readiness", 200, jsonRDY)
}

func TestPatchReadinessWithKeyDefaultError(t *testing.T) {
	setup()
	execPatchRequest(t, "default", 400, queryParamNotDefault, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonRDY)
}

func TestPatchReadinessWithKey(t *testing.T) {
	setup()
	execPatchRequest(t, "test", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "test", 200, jsonRDY, []byte(jsonRDY))
	execGetRequest(t, "pod/readiness", 200, jsonRDY)
}

func TestPatchReadinessWithKeyThenWithoutNotReady(t *testing.T) {
	setup()
	execPatchRequest(t, "test", 200, jsonNotRDY, []byte(jsonNotRDY))
	execPatchRequest(t, "test2", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "", 200, jsonNotRDY, []byte(jsonRDY))
}

func TestPatchReadinessWithKeyThenWithoutReady(t *testing.T) {
	setup()
	execPatchRequest(t, "test", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "test", 200, jsonRDY, []byte(jsonRDY))
	execGetRequest(t, "pod/readiness", 200, jsonRDY)
	execPatchRequest(t, "", 200, jsonNotRDY, []byte(jsonNotRDY))
	execPatchRequest(t, "", 200, jsonRDY, []byte(jsonRDY))
}

func TestPatchReadinessThenWithKeyNotReady(t *testing.T) {
	setup()
	execPatchRequest(t, "", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "test", 200, jsonNotRDY, []byte(jsonNotRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
	execPatchRequest(t, "test", 200, jsonNotRDY, []byte(jsonRDY))
	execGetRequest(t, "pod/readiness", 200, jsonNotRDY)
}

func setup() {
	if testServer == nil {
		testServer = New()
		go testServer.StartServer()
	} else {
		testServer.Readiness.Ready = true
		testServer.Readiness.status = make(map[string]bool)
	}
}

func execGetRequest(t *testing.T, path string, expStatus int, expResponse string) {
	execRequest(t, http.MethodGet, path, "", expStatus, expResponse, nil)
}

func execPatchRequest(t *testing.T, key string, expStatus int, expResponse string, body []byte) {
	execRequest(t, http.MethodPatch, "pod/readiness", key, expStatus, expResponse, body)
}

func execRequest(t *testing.T, method string, path string, key string, expStatus int, expResponse string, body []byte) {

	urlString := fmt.Sprintf("http://localhost:8080/%s", path)
	if key != "" {
		url, _ := url.Parse(urlString)
		params := url.Query()
		params.Set("key", key)
		url.RawQuery = params.Encode()
		urlString = url.String()
	}
	request, err := http.NewRequest(method, urlString, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("could not create request for %s", path)
		t.FailNow()
	}

	if method == http.MethodPatch {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("request failed for %s", path)
	}

	if response.StatusCode != expStatus {
		t.Errorf("%s:%s:expected status code: %v, actual status code: %v", method, path, expStatus, response.StatusCode)
		t.FailNow()
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal("failed to read response body")
		t.FailNow()
	}

	if string(responseBody) != expResponse {
		t.Errorf("expected response: %s, actual response: %s", expResponse, string(responseBody))
		t.FailNow()
	}
}

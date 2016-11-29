package checkin

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/groob/plist"
	"github.com/micromdm/checkin/service/mock"
	"github.com/micromdm/mdm"
	"golang.org/x/net/context"
)

func TestHTTPCheckin_Authenticate(t *testing.T) {
	client := setup(t)
	defer client.Close()

	var httpTests = []struct {
		name         string
		method       mock.CheckinFunc
		request      io.Reader
		expectStatus int
	}{
		{
			name:         "happy_path",
			method:       mock.SucceedCheckin,
			request:      mustMarshalCheckinRequest(t, "Authenticate"),
			expectStatus: http.StatusOK,
		},
		{
			name:         "fail_checkin",
			method:       mock.FailCheckin,
			request:      mustMarshalCheckinRequest(t, "Authenticate"),
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "limit_reader",
			method:       mock.SucceedCheckin,
			request:      neverEnding('a'),
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range httpTests {
		t.Run(tt.name, func(t *testing.T) {
			client.svc.AuthenticateFunc = tt.method
			resp := client.Do(t, "PUT", tt.request)
			if want, have := tt.expectStatus, resp.StatusCode; want != have {
				t.Fatalf("want %d, have %d", want, have)
			}
			if !client.svc.AuthenticateInvoked &&
				resp.StatusCode == http.StatusOK {
				t.Errorf("request suceeded without invoking service method.")
			}
		})
	}
}

func TestHTTPCheckin_TokenUpdate(t *testing.T) {
	client := setup(t)
	defer client.Close()

	var httpTests = []struct {
		name         string
		method       mock.CheckinFunc
		request      io.Reader
		expectStatus int
	}{
		{
			name:         "happy_path",
			method:       mock.SucceedCheckin,
			request:      mustMarshalCheckinRequest(t, "TokenUpdate"),
			expectStatus: http.StatusOK,
		},
		{
			name:         "fail_checkin",
			method:       mock.FailCheckin,
			request:      mustMarshalCheckinRequest(t, "TokenUpdate"),
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "limit_reader",
			method:       mock.SucceedCheckin,
			request:      neverEnding('a'),
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range httpTests {
		t.Run(tt.name, func(t *testing.T) {
			client.svc.TokenUpdateFunc = tt.method
			resp := client.Do(t, "PUT", tt.request)
			if want, have := tt.expectStatus, resp.StatusCode; want != have {
				t.Fatalf("want %d, have %d", want, have)
			}
			if !client.svc.TokenUpdateInvoked &&
				resp.StatusCode == http.StatusOK {
				t.Errorf("request suceeded without invoking service method.")
			}
		})
	}
}

func TestHTTPCheckin_CheckOut(t *testing.T) {
	client := setup(t)
	defer client.Close()

	var httpTests = []struct {
		name         string
		method       mock.CheckinFunc
		request      io.Reader
		expectStatus int
	}{
		{
			name:         "happy_path",
			method:       mock.SucceedCheckin,
			request:      mustMarshalCheckinRequest(t, "CheckOut"),
			expectStatus: http.StatusOK,
		},
		{
			name:         "fail_checkin",
			method:       mock.FailCheckin,
			request:      mustMarshalCheckinRequest(t, "CheckOut"),
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid_messageType",
			method:       mock.FailCheckin,
			request:      mustMarshalCheckinRequest(t, "UnknownMessageType"),
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "limit_reader",
			method:       mock.SucceedCheckin,
			request:      neverEnding('a'),
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range httpTests {
		t.Run(tt.name, func(t *testing.T) {
			client.svc.CheckoutFunc = tt.method
			resp := client.Do(t, "PUT", tt.request)
			if want, have := tt.expectStatus, resp.StatusCode; want != have {
				t.Fatalf("want %d, have %d", want, have)
			}
			if !client.svc.CheckOutInvoked &&
				resp.StatusCode == http.StatusOK {
				t.Errorf("request suceeded without invoking service method.")
			}
		})
	}
}

type client struct {
	*httptest.Server
	svc    *mock.CheckinService
	client *http.Client
}

func mustMarshalCheckinRequest(t *testing.T, messageType string) *bytes.Buffer {
	req := mdm.CheckinCommand{
		MessageType: messageType,
		UDID:        "some-device",
	}
	buf := new(bytes.Buffer)
	err := plist.NewEncoder(buf).Encode(&req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return buf

}

func (s client) Do(t *testing.T, method string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, s.URL, body)
	if err != nil {
		t.Fatalf("failed to create http request, err = %v", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatalf("http request failed: err = %v", err)
	}
	return resp
}

func setup(t *testing.T) client {
	svc := &mock.CheckinService{}
	e := Endpoints{
		CheckinEndpoint: MakeCheckinEndpoint(svc),
	}
	h := MakeHTTPHandlers(
		context.Background(),
		e,
		httptransport.ServerErrorEncoder(EncodeError),
	)
	s := httptest.NewServer(h.CheckinHandler)
	return client{s, svc, http.DefaultClient}
}

// a never ending io.Reader for testing that the server terminates a request
// with a too large body.
type neverEnding byte

func (b neverEnding) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

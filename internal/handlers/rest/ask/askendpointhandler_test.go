package ask

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	asksrv "github.com/sagar-shaw-bits/CrediScout/internal/services/ask"
)

// fakeService is a hand-written stand-in for the ask service, matching the
// repo's existing fake-based style (no testify). It returns canned values and
// records that it was invoked.
type fakeService struct {
	resp   asksrv.AskResponse
	err    error
	called bool
}

func (f *fakeService) Ask(_ context.Context, _ asksrv.AskRequest) (asksrv.AskResponse, error) {
	f.called = true
	return f.resp, f.err
}

// newTestContext builds a gin context wired to a POST /api/ask request with the
// given raw body, plus the recorder that captures the response.
func newTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/ask", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func TestAskEndpointHandler(t *testing.T) {
	const validBody = `{"question":"What is the joining fee?","top_k":5}`

	tests := map[string]struct {
		body          string
		svc           *fakeService
		wantStatus    int
		wantSvcCalled bool
		wantAnswer    string // asserted only when non-empty
	}{
		"success returns 200 with answer": {
			body: validBody,
			svc: &fakeService{resp: asksrv.AskResponse{
				Answer:  "The joining fee is 1000.",
				Sources: []asksrv.Source{{URL: "u1", Title: "Fees", Score: 0.9}},
			}},
			wantStatus:    http.StatusOK,
			wantSvcCalled: true,
			wantAnswer:    "The joining fee is 1000.",
		},
		"malformed json returns 400 without calling service": {
			body:          `{"question":`, // truncated json
			svc:           &fakeService{},
			wantStatus:    http.StatusBadRequest,
			wantSvcCalled: false,
		},
		"missing required question returns 400 without calling service": {
			body:          `{"top_k":5}`, // question has binding:"required"
			svc:           &fakeService{},
			wantStatus:    http.StatusBadRequest,
			wantSvcCalled: false,
		},
		"empty-question sentinel maps to 400": {
			body:          validBody,
			svc:           &fakeService{err: asksrv.ErrEmptyQuestion},
			wantStatus:    http.StatusBadRequest,
			wantSvcCalled: true,
		},
		"generic service error maps to 500": {
			body:          validBody,
			svc:           &fakeService{err: errors.New("retrieve: boom")},
			wantStatus:    http.StatusInternalServerError,
			wantSvcCalled: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			h := NewAskEndpointHandler(tt.svc, nil)
			c, w := newTestContext(tt.body)

			h.AskEndpointHandler()(c)

			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
			assert.Equal(t, tt.wantSvcCalled, tt.svc.called, "service-called mismatch")
			if tt.wantAnswer != "" {
				var got asksrv.AskResponse
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
				assert.Equal(t, tt.wantAnswer, got.Answer)
			}
			// Error responses must not leak internal detail on 500.
			if tt.wantStatus == http.StatusInternalServerError {
				assert.NotContains(t, w.Body.String(), "boom", "500 body leaked internal error")
			}
		})
	}
}

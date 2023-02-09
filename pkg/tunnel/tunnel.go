package tunnel

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/corbado/cli/pkg/ansi"
)

var ErrUnauthorized = errors.New("Invalid credentials")
var ErrSessionExists = errors.New("Active session already exists")
var ErrInternal = errors.New("Tunnel returned internal error. Please try again later")
var ErrConnectionClosed = errors.New("Connection closed")

type Tunnel struct {
	ansi          *ansi.Ansi
	tunnelAddress string
	localAddress  string
	conn          *websocket.Conn
	httpClient    *http.Client

	stopLock        sync.Mutex
	shutdownContext context.Context
	cancel          context.CancelFunc
}

// New returns new tunnel instance
func New(ansi *ansi.Ansi, tunnelAddress string) *Tunnel {
	shutdownContext, cancel := context.WithCancel(context.Background())
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Tunnel{
		ansi:            ansi,
		tunnelAddress:   tunnelAddress,
		httpClient:      httpClient,
		shutdownContext: shutdownContext,
		cancel:          cancel,
	}
}

// Connect connects to tunnel server with given project ID and CLI secret
func (t *Tunnel) Connect(projectID string, cliSecret string) error {
	conn, resp, err := websocket.DefaultDialer.Dial(t.tunnelAddress, t.basicAuth(projectID, cliSecret)) //nolint:bodyclose
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusUnauthorized {
				return ErrUnauthorized
			}

			if resp.StatusCode == http.StatusConflict {
				return ErrSessionExists
			}

			if resp.StatusCode == http.StatusInternalServerError {
				return ErrInternal
			}
		}

		return errors.WithStack(err)
	}

	t.conn = conn

	return nil
}

// Start starts getting webhook requests from tunnel server
func (t *Tunnel) Start(localAddress string) error {
	defer func() {
		if err := t.Stop(); err != nil {
			panic(errors.WithStack(err))
		}
	}()

	go t.handleSignals()

	t.localAddress = localAddress

	for {
		select {
		case <-t.shutdownContext.Done():
			return nil

		default:
			_, req, err := t.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, 1000, 1006) {
					return ErrConnectionClosed
				}

				if strings.Contains(err.Error(), "use of closed network connection") {
					return ErrConnectionClosed
				}

				return errors.Errorf("error reading from tunnel server: %+v", err)
			}

			if err := t.processWebsocketRequest(req); err != nil {
				return err
			}
		}
	}
}

// Stop stops getting webhook requests from tunnel server
func (t *Tunnel) Stop() error {
	t.stopLock.Lock()
	defer t.stopLock.Unlock()

	t.cancel()

	if t.conn != nil {
		if err := t.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
			return errors.WithStack(err)
		}

		if err := t.conn.Close(); err != nil {
			return errors.WithStack(err)
		}

		t.conn = nil
	}

	return nil
}

func (t *Tunnel) handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	if err := t.Stop(); err != nil {
		fmt.Printf("Failed to gracefully stop the tunnel: %+v\n", err)
	}
}

func (t *Tunnel) basicAuth(projectID string, cliSecret string) http.Header {
	auth := base64.StdEncoding.EncodeToString([]byte(projectID + ":" + cliSecret))

	return http.Header{
		"Authorization": {fmt.Sprintf("Basic %s", auth)},
	}
}

func (t *Tunnel) processWebsocketRequest(req []byte) error {
	wreq := &WebhookRequest{}
	if err := json.Unmarshal(req, wreq); err != nil {
		return errors.Errorf("Received invalid payload from tunnel server: %s", string(req))
	}

	wresp, err := t.processWebhookRequest(wreq)
	if err != nil {
		if errResp := t.conn.WriteJSON(&WebhookResponse{
			ID:     wreq.ID,
			Status: http.StatusInternalServerError,
			Body:   fmt.Sprintf("%+v", err),
		}); errResp != nil {
			return errors.WithStack(errResp)
		}

		return err
	}

	if err := t.conn.WriteJSON(wresp); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (t *Tunnel) processWebhookRequest(req *WebhookRequest) (*WebhookResponse, error) {
	httpRequest := &http.Request{}
	httpRequest.Method = http.MethodPost
	httpRequest.Header = make(http.Header)
	for name, value := range req.Headers {
		httpRequest.Header.Add(name, value)
	}

	u, err := url.Parse(fmt.Sprintf("%s%s", t.localAddress, req.Path))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	httpRequest.URL = u
	httpRequest.Body = io.NopCloser(strings.NewReader(req.Body))

	httpResponse, err := t.httpClient.Do(httpRequest)
	if err != nil {
		if os.IsTimeout(err) {
			t.printMessage(
				httpRequest.Method,
				httpRequest.URL.String(),
				len(req.Body),
				t.httpClient.Timeout,
				http.StatusGatewayTimeout,
				0,
			)

			return &WebhookResponse{
				ID:     req.ID,
				Status: http.StatusGatewayTimeout,
				Body: fmt.Sprintf(
					"%s %s timed out (%s)",
					httpRequest.Method,
					httpRequest.URL.String(),
					t.httpClient.Timeout,
				),
			}, nil
		}

		return nil, errors.WithStack(err)
	}
	defer httpResponse.Body.Close()

	respBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t.printMessage(
		httpRequest.Method,
		httpRequest.URL.String(),
		len(req.Body),
		0,
		httpResponse.StatusCode,
		len(respBytes),
	)

	return &WebhookResponse{
		ID:      req.ID,
		Status:  httpResponse.StatusCode,
		Headers: t.headersToMap(httpResponse.Header),
		Body:    string(respBytes),
	}, nil
}

func (t *Tunnel) headersToMap(headers http.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for name, values := range headers {
		result[name] = values[0]
	}

	return result
}

func (t *Tunnel) printMessage(method string, url string, requestBodyLen int, timeout time.Duration, responseHTTPStatusCode int, responseBodyLen int) {
	if responseHTTPStatusCode == http.StatusGatewayTimeout {
		fmt.Printf(
			"[%s] Corbado issued request > Received through tunnel > Local: %s %s (body: %s) > Timeout (%s) HTTP status %s (body: %s), sent it through tunnel > Corbado got response\n",
			time.Now().Format("2006-01-02 15:04:05"),
			t.ansi.Bold(method),
			url,
			formatBytes(requestBodyLen),
			timeout,
			t.ansi.ColorizeHTTPStatusCode(responseHTTPStatusCode),
			formatBytes(responseBodyLen),
		)

		return
	}

	fmt.Printf(
		"[%s] Corbado issued request > Received through tunnel > Local: %s %s (body: %s) > Got HTTP status %s (body: %s), sent it through tunnel > Corbado got response\n",
		time.Now().Format("2006-01-02 15:04:05"),
		t.ansi.Bold(method),
		url,
		formatBytes(requestBodyLen),
		t.ansi.ColorizeHTTPStatusCode(responseHTTPStatusCode),
		formatBytes(responseBodyLen),
	)
}

func formatBytes(bytes int) string {
	return fmt.Sprintf("%.2f Kb", float64(bytes)/1024)
}

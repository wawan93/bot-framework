package bot_framework

import (
	"testing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"fmt"
	"net/url"
	"path"
	"net/http/httptest"
)

type rewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, `{"ok":true}`)
}

func getBot() *BotFramework {
	server := httptest.NewServer(http.HandlerFunc(okHandler))
	sUrl, err := url.Parse(server.URL)
	http.DefaultClient.Transport = rewriteTransport{URL: sUrl}
	api, err := tgbotapi.NewBotAPIWithClient(
		"token",
		http.DefaultClient,
	)
	if err != nil {
		log.Panic(err)
	}
	api.Debug = true
	return NewBotFramework(api)
}

func TestNewBotFramework(t *testing.T) {
	bot := getBot()
	if bot == nil {
		t.Fatal(bot)
	}
}

func TestBotFramework_handleUpdate(t *testing.T) {
	t.Parallel()
	bot := getBot()

	u := &tgbotapi.Update{}
	err := bot.handleUpdate(u)
	if err == nil || err.Error() != "no message" {
		t.Error("empty update must not be handled")
	}
}

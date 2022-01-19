package httpd

import (
	"net/url"
	"testing"
)

func TestResolveURL(t *testing.T) {
	u, err := url.Parse("css/../events.html")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := "/events.html"

	resolved, err := resolve(u)
	if err != nil {
		t.Fatalf("Error resolving URL (%v)", err)
	}

	if resolved != expected {
		t.Errorf("URL %v incorrectly resolved - expected:%v, got:%v", u, expected, resolved)
	}
}

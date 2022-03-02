package commands

import (
	"strings"
	"testing"
)

func TestGenKeys(t *testing.T) {
	keys, err := genkeys()
	if err != nil {
		t.Fatalf("Unexpected error creating CA (%v)", err)
	}

	f := func(v interface{}) string {
		return string(encode(v))
	}

	if s := f(keys.CA.privateKey); !strings.HasPrefix(s, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Errorf("Invalid CA private key (%v)", s)
	}

	if s := f(keys.CA.certificate); !strings.HasPrefix(s, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("Invalid CA certificate (%v)", s)
	}

	if s := f(keys.server.privateKey); !strings.HasPrefix(s, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Errorf("Invalid server private key (%v)", s)
	}

	if s := f(keys.server.certificate); !strings.HasPrefix(s, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("Invalid server certificate (%v)", s)
	}

	if s := f(keys.client.privateKey); !strings.HasPrefix(s, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Errorf("Invalid client private key (%v)", s)
	}

	if s := f(keys.client.certificate); !strings.HasPrefix(s, "-----BEGIN CERTIFICATE-----") {
		t.Errorf("Invalid client certificate (%v)", s)
	}
}

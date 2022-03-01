package commands

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"
)

type certificates struct {
	CA     keyset
	server keyset
}

type keyset struct {
	privateKey  *rsa.PrivateKey
	certificate []byte
}

var certs = struct {
	CA    x509.Certificate
	httpd x509.Certificate
}{
	CA: x509.Certificate{
		SerialNumber: nonce(),
		Subject: pkix.Name{
			Organization: []string{"uhppoted-httpd"},
			Country:      []string{"uhppoted"},
			Province:     []string{"httpd"},
			Locality:     []string{"localhost"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	},

	httpd: x509.Certificate{
		SerialNumber: nonce(),
		Subject: pkix.Name{
			Organization: []string{"uhppoted-httpd"},
			Country:      []string{"uhppoted"},
			Province:     []string{"httpd"},
			Locality:     []string{"localhost"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	},
}

func genkeys() (*certificates, error) {
	// ... create CA key and certificate
	CA, err := genCA()
	if err != nil {
		return nil, err
	} else if CA == nil {
		return nil, fmt.Errorf("Invalid CA (%v)", CA)
	}

	// ... create server key and certificate
	httpd, err := genServerKey(CA.privateKey)
	if err != nil {
		return nil, err
	} else if httpd == nil {
		return nil, fmt.Errorf("Invalid TLS server key and certificate (%v)", httpd)
	}

	return &certificates{
		CA:     *CA,
		server: *httpd,
	}, nil
}

func genCA() (*keyset, error) {
	// ... create CA key and certificate
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certs.CA, &certs.CA, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	var u bytes.Buffer
	pem.Encode(&u, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return &keyset{
		privateKey:  key,
		certificate: cert,
	}, nil
}

func genServerKey(capk *rsa.PrivateKey) (*keyset, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certs.httpd, &certs.CA, &key.PublicKey, capk)
	if err != nil {
		return nil, err
	}

	return &keyset{
		privateKey:  key,
		certificate: cert,
	}, nil
}

func encode(p interface{}) []byte {
	var b bytes.Buffer

	switch v := p.(type) {

	case *rsa.PrivateKey:
		pem.Encode(&b, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(v),
		})

	case []byte:
		pem.Encode(&b, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: v,
		})

	default:
		log.Fatalf("Invalid TLS key or certificate (%T)", p)
	}

	return b.Bytes()
}

func nonce() *big.Int {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	N, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatalf("Error generating TLS nonce (%v)", err)
	}

	return N
}

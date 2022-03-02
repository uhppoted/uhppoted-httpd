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
	client keyset
}

type keyset struct {
	privateKey  *rsa.PrivateKey
	certificate []byte
}

var certs = struct {
	CA     x509.Certificate
	httpd  x509.Certificate
	client x509.Certificate
}{
	CA: x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"uhppoted"},
			Country:      []string{"uhppoted-httpd"},
			Province:     []string{"httpd"},
			CommonName:   "uhppoted-httpd-CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	},

	httpd: x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"uhppoted"},
			Country:      []string{"uhppoted-httpd"},
			Province:     []string{"httpd"},
			CommonName:   "localhost",
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	},

	client: x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"uhppoted-httpd"},
			Country:      []string{"uhppoted"},
			Province:     []string{"httpd"},
			CommonName:   "Don Duque",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
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

	// ... create client key and certificate
	client, err := genClientKey(CA.privateKey)
	if err != nil {
		return nil, err
	} else if httpd == nil {
		return nil, fmt.Errorf("Invalid TLS client key and certificate (%v)", httpd)
	}

	return &certificates{
		CA:     *CA,
		server: *httpd,
		client: *client,
	}, nil
}

func genCA() (*keyset, error) {
	// ... create CA key and certificate
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certs.CA, &certs.CA, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return &keyset{
		privateKey:  key,
		certificate: cert,
	}, nil
}

func genServerKey(CA *rsa.PrivateKey) (*keyset, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certs.httpd, &certs.CA, &key.PublicKey, CA)
	if err != nil {
		return nil, err
	}

	return &keyset{
		privateKey:  key,
		certificate: cert,
	}, nil
}

func genClientKey(CA *rsa.PrivateKey) (*keyset, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certs.client, &certs.CA, &key.PublicKey, CA)
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
		if v == nil {
			log.Fatalf("Invalid TLS key (%v)", v)
		} else {
			pem.Encode(&b, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(v),
			})
		}

	case []byte:
		if v == nil {
			log.Fatalf("Invalid TLS certificate (%v)", v)
		} else {
			pem.Encode(&b, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: v,
			})
		}

	default:
		log.Fatalf("Invalid TLS key or certificate (%T)", p)
	}

	return b.Bytes()
}

package cctest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/attrmgr"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// ClientIdentity describes the identity exposed to Fabric client identity helpers.
type ClientIdentity struct {
	MSPID      string
	ID         string
	CertPEM    []byte
	Attributes map[string]string
}

// Clone returns an isolated copy.
func (i ClientIdentity) Clone() ClientIdentity {
	out := ClientIdentity{
		MSPID:      i.MSPID,
		ID:         i.ID,
		CertPEM:    cloneBytes(i.CertPEM),
		Attributes: map[string]string{},
	}
	for key, value := range i.Attributes {
		out.Attributes[key] = value
	}
	return out
}

type mockClientIdentity struct {
	identity ClientIdentity
	cert     *x509.Certificate
}

func newMockClientIdentity(identity ClientIdentity) *mockClientIdentity {
	cert, _ := parseCertificate(identity.CertPEM)
	return &mockClientIdentity{identity: identity.Clone(), cert: cert}
}

func (i *mockClientIdentity) GetID() (string, error) {
	return i.identity.ID, nil
}

func (i *mockClientIdentity) GetMSPID() (string, error) {
	return i.identity.MSPID, nil
}

func (i *mockClientIdentity) GetAttributeValue(name string) (string, bool, error) {
	value, ok := i.identity.Attributes[name]
	return value, ok, nil
}

func (i *mockClientIdentity) AssertAttributeValue(name, value string) error {
	actual, ok, err := i.GetAttributeValue(name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("attribute %q was not found", name)
	}
	if actual != value {
		return fmt.Errorf("attribute %q equals %q, not %q", name, actual, value)
	}
	return nil
}

func (i *mockClientIdentity) GetX509Certificate() (*x509.Certificate, error) {
	return i.cert, nil
}

func serializedIdentity(identity ClientIdentity) ([]byte, []byte, error) {
	certPEM := cloneBytes(identity.CertPEM)
	if len(certPEM) == 0 {
		generated, err := generateCertificate(identity)
		if err != nil {
			return nil, nil, err
		}
		certPEM = generated
	}

	sid := &msp.SerializedIdentity{
		Mspid:   identity.MSPID,
		IdBytes: certPEM,
	}
	raw, err := proto.Marshal(sid)
	if err != nil {
		return nil, nil, err
	}
	return raw, certPEM, nil
}

func generateCertificate(identity ClientIdentity) ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	attrs := struct {
		Attrs map[string]string `json:"attrs"`
	}{Attrs: identity.Attributes}
	attrValue, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   identity.ID,
			Organization: []string{identity.MSPID},
		},
		Issuer: pkix.Name{
			CommonName: "cctest-ca",
		},
		NotBefore: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:  time.Date(2034, 1, 1, 0, 0, 0, 0, time.UTC),
		ExtraExtensions: []pkix.Extension{
			{Id: asn1.ObjectIdentifier(attrmgr.AttrOID), Critical: false, Value: attrValue},
		},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}

func parseCertificate(certPEM []byte) (*x509.Certificate, error) {
	if len(certPEM) == 0 {
		return nil, nil
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}

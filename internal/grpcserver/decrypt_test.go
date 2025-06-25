package grpcserver_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/grpcserver"
	"github.com/nekr0z/muhame/internal/proto"
)

var (
	update bool

	privateKey *rsa.PrivateKey

	privateKeyFile    = filepath.Join("testdata", "private.key")
	publicKeyFile     = filepath.Join("testdata", "public.key")
	cipherMsgFile     = filepath.Join("testdata", "ciphermsg")
	cipherBulkMsgFile = filepath.Join("testdata", "cipherbulkmsg")
)

func TestDecrypt(t *testing.T) {
	met := &proto.Metric{
		Name: "test",
		Value: &proto.Metric_Counter{
			Counter: &proto.Counter{
				Delta: 1,
			},
		},
	}

	interceptor := grpcserver.DecryptInterceptor(privateKey)

	t.Run("single", func(t *testing.T) {
		msg := &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: met,
			},
		}

		message, err := pb.Marshal(msg)
		require.NoError(t, err)

		var ciphermsg []byte

		if update {
			ciphermsg, err = crypt.Encrypt(message, &privateKey.PublicKey)
			require.NoError(t, err)

			err := os.WriteFile(cipherMsgFile, ciphermsg, 0644)
			require.NoError(t, err)
		} else {
			ciphermsg, err = os.ReadFile(cipherMsgFile)
			require.NoError(t, err)
		}

		req := &proto.MetricRequest{
			Payload: &proto.MetricRequest_Data{
				Data: ciphermsg,
			},
		}

		res, err := interceptor(context.Background(), req, nil, handler)
		require.NoError(t, err)

		got, ok := res.(*proto.MetricRequest)
		require.True(t, ok)
		assert.Equal(t, met.GetCounter().GetDelta(), got.GetMetric().GetCounter().GetDelta())
		assert.Equal(t, met.GetName(), got.GetMetric().GetName())
	})

	t.Run("bulk", func(t *testing.T) {
		message, err := pb.Marshal(&proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{
					Metrics: []*proto.Metric{
						met,
					},
				},
			},
		})
		require.NoError(t, err)

		var ciphermsg []byte

		if update {
			ciphermsg, err = crypt.Encrypt(message, &privateKey.PublicKey)
			require.NoError(t, err)

			err := os.WriteFile(cipherBulkMsgFile, ciphermsg, 0644)
			require.NoError(t, err)
		} else {
			ciphermsg, err = os.ReadFile(cipherBulkMsgFile)
			require.NoError(t, err)
		}

		req := &proto.BulkRequest{
			Payload: &proto.BulkRequest_Data{
				Data: ciphermsg,
			},
		}

		res, err := interceptor(context.Background(), req, nil, handler)
		require.NoError(t, err)

		got, ok := res.(*proto.BulkRequest)
		require.True(t, ok)

		mm := got.GetMetrics().GetMetrics()
		require.Equal(t, 1, len(mm))
		assert.Equal(t, met.GetCounter().GetDelta(), mm[0].GetCounter().GetDelta())
		assert.Equal(t, met.GetName(), mm[0].GetName())
	})
}

func handler(_ context.Context, req any) (any, error) {
	return req, nil
}

func TestMain(m *testing.M) {
	flag.BoolVar(&update, "update", false, "update keys and testdata")
	flag.Parse()

	var err error

	if update {
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			panic(err)
		}

		err = saveKeys(privateKey)
		if err != nil {
			panic(err)
		}
	} else {
		privateKey, err = crypt.LoadPrivateKey(privateKeyFile)
		if err != nil {
			panic(err)
		}
	}

	os.Exit(m.Run())
}

func saveKeys(key *rsa.PrivateKey) error {
	var privateKeyPEM bytes.Buffer
	err := pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(privateKeyFile, privateKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	})
	if err != nil {
		return err
	}

	return os.WriteFile(publicKeyFile, publicKeyPEM.Bytes(), 0644)
}

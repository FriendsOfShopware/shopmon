package mail

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type smtpServer struct {
	ln     net.Listener
	wg     sync.WaitGroup
	msgs   []string
	closed chan struct{}
}

func newSMTPServer(t *testing.T) *smtpServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := &smtpServer{
		ln:     ln,
		closed: make(chan struct{}),
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(s.closed)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			s.handleConn(conn)
		}
	}()

	return s
}

func (s *smtpServer) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	write := func(msg string) {
		fmt.Fprint(conn, msg+"\r\n")
	}

	write("220 localhost ESMTP Test")

	var dataMode bool
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if dataMode {
			if line == "." {
				dataMode = false
				s.msgs = append(s.msgs, data.String())
				data.Reset()
				write("250 OK")
			} else {
				data.WriteString(line + "\r\n")
			}
			continue
		}

		cmd := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO"):
			write("250-localhost")
			write("250 AUTH LOGIN PLAIN")
		case strings.HasPrefix(cmd, "AUTH"):
			write("235 Authentication successful")
		case strings.HasPrefix(cmd, "MAIL FROM"):
			write("250 OK")
		case strings.HasPrefix(cmd, "RCPT TO"):
			write("250 OK")
		case strings.HasPrefix(cmd, "DATA"):
			write("354 Start mail input")
			dataMode = true
		case cmd == "QUIT":
			write("221 Bye")
			return
		default:
			write("500 Unknown command")
		}
	}
}

func (s *smtpServer) close() {
	s.ln.Close()
	<-s.closed
	s.wg.Wait()
}

func (s *smtpServer) addr() string {
	return s.ln.Addr().String()
}

// generateTLSCert creates a CA and a server certificate signed by it for the given host.
// Returns the server certificate and the CA cert pool (for client-side verification).
func generateTLSCert(t *testing.T, host string) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Test CA"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP(host)},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	keyDER, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	return cert, pool
}

func TestNewService(t *testing.T) {
	cfg := SMTPConfig{
		Host: "smtp.example.com",
		Port: "587",
		From: "noreply@example.com",
	}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
	assert.Equal(t, cfg, svc.cfg)
}

func TestSendNonSecure(t *testing.T) {
	srv := newSMTPServer(t)
	defer srv.close()

	host, port, _ := net.SplitHostPort(srv.addr())

	svc := NewService(SMTPConfig{
		Host:    host,
		Port:    port,
		From:    "sender@example.com",
		ReplyTo: "reply@example.com",
	})

	err := svc.Send("recipient@example.com", "Test Subject", "<p>Hello</p>")
	require.NoError(t, err)

	require.Len(t, srv.msgs, 1)
	msg := srv.msgs[0]
	assert.Contains(t, msg, "From: sender@example.com")
	assert.Contains(t, msg, "To: recipient@example.com")
	assert.Contains(t, msg, "Subject: Test Subject")
	assert.Contains(t, msg, "Reply-To: reply@example.com")
	assert.Contains(t, msg, "MIME-Version: 1.0")
	assert.Contains(t, msg, "Content-Type: text/html; charset=UTF-8")
	assert.Contains(t, msg, "<p>Hello</p>")
}

func TestSendNonSecureNoReplyTo(t *testing.T) {
	srv := newSMTPServer(t)
	defer srv.close()

	host, port, _ := net.SplitHostPort(srv.addr())

	svc := NewService(SMTPConfig{
		Host: host,
		Port: port,
		From: "sender@example.com",
	})

	err := svc.Send("recipient@example.com", "Test Subject", "<p>Hello</p>")
	require.NoError(t, err)

	require.Len(t, srv.msgs, 1)
	assert.NotContains(t, srv.msgs[0], "Reply-To:")
}

func TestSendWithAuth(t *testing.T) {
	srv := newSMTPServer(t)
	defer srv.close()

	host, port, _ := net.SplitHostPort(srv.addr())

	svc := NewService(SMTPConfig{
		Host: host,
		Port: port,
		User: "user",
		Pass: "pass",
		From: "sender@example.com",
	})

	err := svc.Send("recipient@example.com", "Subject", "body")
	require.NoError(t, err)
	require.Len(t, srv.msgs, 1)
}

func TestSendSecure(t *testing.T) {
	host := "127.0.0.1"
	serverCert, caPool := generateTLSCert(t, host)

	ln, err := tls.Listen("tcp", host+":0", &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	})
	require.NoError(t, err)
	defer ln.Close()

	var msgs []string
	done := make(chan struct{})

	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		write := func(msg string) { fmt.Fprint(conn, msg+"\r\n") }

		write("220 localhost ESMTP Test")

		var dataMode bool
		var data strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			if dataMode {
				if line == "." {
					dataMode = false
					msgs = append(msgs, data.String())
					data.Reset()
					write("250 OK")
				} else {
					data.WriteString(line + "\r\n")
				}
				continue
			}

			cmd := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO"):
				write("250 OK")
			case strings.HasPrefix(cmd, "MAIL FROM"), strings.HasPrefix(cmd, "RCPT TO"):
				write("250 OK")
			case strings.HasPrefix(cmd, "DATA"):
				write("354 Start mail input")
				dataMode = true
			case cmd == "QUIT":
				write("221 Bye")
				return
			default:
				write("500 Unknown command")
			}
		}
	}()

	_, port, _ := net.SplitHostPort(ln.Addr().String())

	svc := NewService(SMTPConfig{
		Host:   host,
		Port:   port,
		Secure: true,
		From:   "sender@example.com",
		TLSConfig: &tls.Config{
			RootCAs: caPool,
		},
	})

	err = svc.Send("recipient@example.com", "Secure Subject", "<p>Secure body</p>")
	require.NoError(t, err)

	<-done

	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0], "Subject: Secure Subject")
	assert.Contains(t, msgs[0], "<p>Secure body</p>")
}

func TestSendConnectionError(t *testing.T) {
	svc := NewService(SMTPConfig{
		Host:   "127.0.0.1",
		Port:   "1",
		Secure: false,
		From:   "sender@example.com",
	})

	err := svc.Send("recipient@example.com", "Subject", "body")
	assert.Error(t, err)
}

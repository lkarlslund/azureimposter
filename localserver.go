package azureimposter

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func codeVerifier() (codeVerifier string, challenge string, err error) {
	cvBytes := make([]byte, 32)
	if _, err = rand.Read(cvBytes); err != nil {
		return
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(cvBytes)
	// for PKCE, create a hash of the code verifier
	cvh := sha256.Sum256([]byte(codeVerifier))
	challenge = base64.RawURLEncoding.EncodeToString(cvh[:])
	return
}

type Server struct {
	Addr     string
	Port     int
	TLS      bool
	Cert     tls.Certificate
	s        *http.Server
	ResultCh chan Result
	reqState string
}

func Serve(emulateurl string) (*Server, error) {
	serv := &Server{
		s: &http.Server{
			// Addr: "localhost:0",
		},
		// reqState: reqState,
		ResultCh: make(chan Result, 1),
	}

	var port int
	var l net.Listener
	var err error

	if emulateurl != "" {
		u, err := url.Parse(emulateurl)
		if err != nil {
			return nil, err
		}

		if u.Port() == "" {
			if u.Scheme == "http" {
				port = 80
			} else if u.Scheme == "https" {
				port = 443
				cert, err := GenerateCert(u.Hostname())
				if err != nil {
					return nil, err
				}
				serv.Cert = cert
				serv.s.TLSConfig = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}
				serv.TLS = true
			} else {
				return nil, errors.New("Unknown scheme " + u.Scheme)
			}
		} else {
			// Ignore errors, it will then default to 0
			port64, _ := strconv.ParseInt(u.Port(), 10, 64)
			port = int(port64)
		}

		l, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			return nil, err
		}
		serv.Port = port
	} else {
		// Autogenerated - find a free port
		for i := 0; i < 10; i++ {
			l, err = net.Listen("tcp", "localhost:0")
			if err != nil {
				continue
			}
			addr := l.Addr().String()
			port64, _ := strconv.ParseInt(addr[strings.LastIndex(addr, ":")+1:], 10, 64)
			port = int(port64)

			serv.Addr = fmt.Sprintf("http://localhost:%v", port)

			break
		}
	}

	if err != nil {
		return nil, err
	}

	serv.s.Handler = http.HandlerFunc(serv.handler)

	if err := serv.start(l); err != nil {
		return nil, err
	}

	return serv, nil
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	headerErr := q.Get("error")
	if headerErr != "" {
		desc := q.Get("error_description")
		// Note: It is a little weird we handle some errors by not going to the failPage. If they all should,
		// change this to s.error() and make s.error() write the failPage instead of an error code.
		_, _ = w.Write([]byte(fmt.Sprintf(failPage, headerErr, desc)))
		s.ResultCh <- Result{Err: fmt.Errorf(desc)}
		return
	}

	respState := q.Get("state")
	switch respState {
	case s.reqState:
	case "":
		s.error(w, http.StatusInternalServerError, "server didn't send OAuth state")
		return
	default:
		s.error(w, http.StatusInternalServerError, "mismatched OAuth state, req(%s), resp(%s)", s.reqState, respState)
		return
	}

	code := q.Get("code")
	if code == "" {
		s.error(w, http.StatusInternalServerError, "authorization code missing in query string")
		return
	}

	_, _ = w.Write(okPage)
	s.ResultCh <- Result{Code: code}
}

func (s *Server) start(l net.Listener) error {
	go func() {
		var err error
		if s.TLS {
			err = s.s.ServeTLS(l, "", "")
		} else {
			err = s.s.Serve(l)
		}
		if err != nil {
			select {
			case s.ResultCh <- Result{Err: err}:
			default:
			}
		}
	}()

	return nil
}

func (s *Server) error(w http.ResponseWriter, code int, str string, i ...interface{}) {
	err := fmt.Errorf(str, i...)
	http.Error(w, err.Error(), code)
	s.ResultCh <- Result{Err: err}
}

type Result struct {
	// Code is the code sent by the authority server.
	Code string
	// Err is set if there was an error.
	Err error
}

var okPage = []byte(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Authentication Complete</title>
</head>
<body>
    <p>Authentication complete. You can return to the application. Feel free to close this browser tab.</p>
</body>
</html>
`)

const failPage = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Authentication Failed</title>
</head>
<body>
	<p>Authentication failed. You can return to the application. Feel free to close this browser tab.</p>
	<p>Error details: error %s error_description: %s</p>
</body>
</html>
`

package azureimposter

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/go-resty/resty/v2"
	"github.com/lkarlslund/lorca"
	"github.com/lkarlslund/stringsplus"
	mozcertificate "github.com/mozilla/tls-observatory/certificate"
)

func GetToken(authority, clientID, redirectURI string, scope string) (*TokenResult, error) {
	var extraargs []string

	resultchan := make(chan Result)
	var interceptlogs bool
	if stringsplus.EqualFoldHasPrefix(redirectURI, "http") {
		u, err := url.Parse(redirectURI)
		if err == nil {
			// Try to subvert the final redirect to our own server
			extraargs = append(extraargs,
				"--host-resolver-rules",
				fmt.Sprintf("MAP %v 127.0.0.1", u.Host),
			)
		}
	}
	if stringsplus.EqualFoldHasPrefix(redirectURI, "urn:ietf:wg:oauth:2.0:oob") {
		// TODO Figure out this one :-)
		interceptlogs = true
	} else {
		srv, err := Serve(redirectURI)
		if err != nil {
			return nil, err
		}
		if redirectURI == "" {
			redirectURI = srv.Addr
		}
		if srv.TLS {
			if c, e := x509.ParseCertificate(srv.Cert.Certificate[0]); e == nil {
				extraargs = append(extraargs,
					"--ignore-certificate-errors-spki-list",
					mozcertificate.PKPSHA256Hash(c),
				)
			}
		}

		// Override results
		resultchan = srv.ResultCh
	}

	c, err := public.New(clientID, public.WithAuthority(authority))
	if err != nil {
		return nil, err
	}

	// Get the URL for interactive login
	ctx := context.Background()
	loginurl, err := c.CreateAuthCodeURL(ctx, clientID, redirectURI, []string{scope})
	if err != nil {
		return nil, err
	}

	// Launch browser
	l, err := lorca.New("", "", 400, 600, extraargs...)
	if err != nil {
		return nil, err
	}

	if interceptlogs {
		err := l.AddScriptToEvaluateOnNewDocument(`console.stdlog = console.log.bind(console);
console.logs = [];
console.log = function(){
    console.logs.push(Array.from(arguments));
    console.stdlog.apply(console, arguments);
}`)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	resulturl := ""
	go func() {
		ticker := time.NewTicker(time.Millisecond * 20)
	loop:
		for {
			select {
			case <-l.Done():
				break loop
			case result := <-resultchan:
				if result.Err == nil {
					resulturl = result.Code
				}
				l.Close()
				break loop
			case <-ticker.C:
				// val := l.Eval("document.title")
				// title := val.String()
				// if strings.Contains(title, "code=") {
				// 	fmt.Println(title)
				// }
				log := l.Eval("if (console.logs.length ==1) console.logs[0]").String()
				if log != "" {
					fmt.Println(log)
					l.Eval("console.logs = []")
				}

				// title := val.String()
				// if strings.Contains(title, "code=") {
				// 	fmt.Println(title)
				// }

				if redirectURI != "" {
					currenturl := l.Eval("document.URL").String()
					if stringsplus.EqualFoldHasPrefix(currenturl, redirectURI) {
						// We got it ... let's GO GO GO
						resulturl = currenturl
						l.Close()
					}
				}
			}
		}
	}()

	// Load the URL and wait for browser to either redirect correctly or close
	l.Load(loginurl)
	<-l.Done()

	u, err := url.Parse(resulturl)
	if err != nil {
		return nil, err
	}
	code := u.Query().Get("code")
	if code != "" {
		return RequestToken(clientID, code, redirectURI)
	}
	return nil, errors.New("No code returned, can't get token")
}

func RequestToken(clientid, code, redirecturi string) (*TokenResult, error) {
	client := resty.New()
	resp, err := client.R().
		SetFormData(map[string]string{
			"client_id":    clientid,
			"grant_type":   "authorization_code",
			"code":         code,
			"redirect_uri": redirecturi,
		}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("https://login.microsoftonline.com/common/oauth2/token")
	if err != nil {
		return nil, err
	}

	var result TokenResult
	err = json.Unmarshal([]byte(resp.String()), &result)
	return &result, err
}

func RefreshToken(clientid, refreshtoken string, scope string) (*TokenResult, error) {
	client := resty.New()
	resp, err := client.R().
		SetFormData(map[string]string{
			"client_id":     clientid,
			"grant_type":    "refresh_token",
			"refresh_token": refreshtoken,
			"scope":         scope,
		}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	if err != nil {
		return nil, err
	}

	var result TokenResult
	err = json.Unmarshal([]byte(resp.String()), &result)
	return &result, err
}

type TokenResult struct {
	TokenType    string      `json:"token_type"`
	Scope        string      `json:"scope"`
	ExpiresIn    interface{} `json:"expires_in"`
	ExtExpiresIn interface{} `json:"ext_expires_in"`
	ExpiresOn    string      `json:"expires_on"`
	NotBefore    string      `json:"not_before"`
	Resource     string      `json:"resource"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	Foci         string      `json:"foci"`
	IdToken      string      `json:"id_token"`
}

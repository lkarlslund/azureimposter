package main

import (
	"flag"
	"fmt"

	"github.com/lkarlslund/azureimposter"
)

func main() {
	wkc := flag.String("wkc", "Graph", "Use a predefined well known client name")

	authority := flag.String("authority", "https://login.windows.net/common/", "OAuth authority request URL")
	clientIDoverride := flag.String("clientID", "", "clientID to use")
	redirectURIorverride := flag.String("redirectURI", "", "redirectURI to use")
	scopeOverride := flag.String("scopes", "", "scope to request")

	pretender := azureimposter.WellKnownClients[*wkc]

	clientID := pretender.ClientId
	redirectURI := pretender.RedirectURI
	scope := pretender.Scope

	if *clientIDoverride != "" {
		clientID = *clientIDoverride
	}

	if *redirectURIorverride != "" {
		redirectURI = *redirectURIorverride
	}

	if *scopeOverride != "" {
		scope = *scopeOverride
	}

	token, err := azureimposter.GetToken(
		*authority,
		clientID,
		redirectURI,
		scope,
	)

	if err != nil {
		fmt.Println("Problem getting token:", err)
	}

	fmt.Println("Access token:")
	fmt.Println(token.AccessToken)
}

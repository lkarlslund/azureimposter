package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/lkarlslund/azureimposter"
)

func main() {
	wkc := flag.String("wkc", "Az", "Use a predefined well known client name")

	authority := flag.String("authority", "https://login.windows.net/common/", "OAuth authority request URL")
	clientIDoverride := flag.String("clientID", "", "clientID to use")
	redirectURIorverride := flag.String("redirectURI", "", "redirectURI to use")
	scopesOverride := flag.String("scopes", "", "comma list separated list of scopes to request")

	pretender := azureimposter.WellKnownClients[*wkc]

	clientID := pretender.ClientId
	redirectURI := pretender.RedirectURI
	scopes := pretender.DefaultScopes

	if *clientIDoverride != "" {
		clientID = *clientIDoverride
	}

	if *redirectURIorverride != "" {
		redirectURI = *redirectURIorverride
	}

	if *scopesOverride != "" {
		scopes = strings.Split(*scopesOverride, ",")
	}

	token, err := azureimposter.GetToken(
		*authority,
		clientID,
		redirectURI,
		scopes,
	)

	if err != nil {
		fmt.Println("Problem getting token:", err)
	}

	fmt.Println(token)
}

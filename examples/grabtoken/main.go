package main

import (
	"flag"
	"fmt"

	"github.com/lkarlslund/azureimposter"
)

func main() {
	wkc := flag.String("wkc", "Graph", "Use a predefined well known client name")
	authority := flag.String("authority", "https://login.windows.net/common/", "OAuth authority request URL")

	authinfo := azureimposter.WellKnownClients[*wkc]

	token, err := azureimposter.AcquireToken(
		*authority,
		authinfo,
	)

	if err != nil {
		fmt.Println("Problem getting token:", err)
	}

	fmt.Println("Access token:")
	fmt.Println(token.AccessToken)
}

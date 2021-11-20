package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lkarlslund/azureimposter"
	"github.com/lkarlslund/azureimposter/api/msgraphbeta"
)

type AzureUser struct {
	ID                           string          `json:"id,omitempty"`
	DisplayName                  string          `json:"displayName,omitempty"`
	Mail                         string          `json:"mail,omitempty"`
	OnPremisesDistinguishedName  string          `json:"onPremisesDistinguishedName,omitempty"`
	OnPremisesDomainName         string          `json:"onPremisesDomainName,omitempty"`
	OnPremisesImmutableId        string          `json:"onPremisesImmutableId,omitempty"`
	OnPremisesLastSyncDateTime   *time.Time      `json:"onPremisesLastSyncDateTime,omitempty"`
	OnPremisesSamAccountName     string          `json:"onPremisesSamAccountName,omitempty"`
	OnPremisesSecurityIdentifier string          `json:"onPremisesSecurityIdentifier,omitempty"`
	OnPremisesSyncEnabled        bool            `json:"onPremisesSyncEnabled,omitempty"`
	OnPremisesUserPrincipalName  string          `json:"onPremisesUserPrincipalName,omitempty"`
	SignInActivity               *SignInActivity `json:"signInActivity", version:"beta,omitempty"`
}

type SignInActivity struct {
	LastSignInDateTime               *time.Time `json:"lastSignInDateTime,omitempty"`
	LastSignInRequestID              string     `json:"lastSignInRequestId,omitempty"`
	LastNonInteractiveSignInDateTime *time.Time `json:"lastNonInteractiveSignInDateTime,omitempty"`
	LastNonInteractiveRequestID      string     `json:"lastNonInteractiveRequestId,omitempty"`
}

func main() {
	authority := flag.String("authority", "https://login.windows.net/common/", "OAuth authority request URL")

	graphinfo := azureimposter.WellKnownClients["Graph"]

	token, err := azureimposter.AcquireToken(
		*authority,
		graphinfo.RedirectURI,
		graphinfo.ClientId,
		"https://graph.microsoft.com//.default",
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := azureimposter.NewClient(*token)

	req := client.R()
	req.SetQueryParam("$select", "id,displayName,mail,onPremisesDistinguishedName,onPremisesDomainName,onPremisesImmutableId,onPremisesLastSyncDateTime,onPremisesSamAccountName,onPremisesSecurityIdentifier,onPremisesSyncEnabled,onPremisesUserPrincipalName,signInActivity")
	req.Method = "GET"
	req.URL = "https://graph.microsoft.com/beta/users"

	var users []msgraphbeta.MicrosoftGraphUser

	if err = req.GetChunkedData(func(data []byte) error {
		var userchunk []msgraphbeta.MicrosoftGraphUser
		err = json.Unmarshal(data, &userchunk)
		if err != nil {
			return err
		}
		users = append(users, userchunk...)
		return nil
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Output users to stdout
	j, _ := json.MarshalIndent(users, "", "  ")
	fmt.Print(string(j))
}

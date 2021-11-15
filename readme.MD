# azureimposter

Go module that pretends to be any clientID and grabs an authentication token from Azure using interactive login (w/mfa if enabled) and returns the token to the caller. This can then be used to enumerate users, groups etc. depending on what scope you've requested.

Requires Chrome browser, and should work on Windows / Linux / OSX.

Se the "example" folder for a quick demo.

Inspired by the techniques used in [AADinternals PowerShell module](https://github.com/Gerenios/AADInternals) module

# lastuserlogon

Simple demo that requests interactive token and queries Microsoft Graph via the beta API for user information - including attributes for last login

If you get a Authentication_RequestFromUnsupportedUserRole error, it means you're probably missing the AuditLog.Read.All role - more info here: [https://docs.microsoft.com/da-dk/graph/api/user-list?view=graph-rest-beta&tabs=http#example-3-get-users-including-their-last-sign-in-time]
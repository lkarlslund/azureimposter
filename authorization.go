package azureimposter

type Authorization struct {
	ClientID string
	Scope    string

	Token        string
	RefreshToken string
}

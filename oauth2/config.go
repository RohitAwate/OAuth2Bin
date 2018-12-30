package oauth2

// AuthCodeConfig defines the variables required in the OAuth 2.0 Authorization Code flow
type AuthCodeConfig struct {
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	AuthGrant    string
	AccessToken  string
}

// ImplicitConfig defines the variables required in the OAuth 2.0 Implicit flow
type ImplicitConfig struct {
	AuthURL     string
	ClientID    string
	AccessToken string
}

// OA2Config defines the configurations for all the flows in OAuth 2.0
type OA2Config struct {
	AuthCodeCnfg AuthCodeConfig
	ImplicitCnfg ImplicitConfig
}

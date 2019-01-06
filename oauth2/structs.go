package oauth2

// AuthCodeConfig defines the variables required in the OAuth 2.0 Authorization Code flow
type AuthCodeConfig struct {
	AuthURL      string `json:"authURL"`
	TokenURL     string `json:"tokenURL"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

// ImplicitConfig defines the variables required in the OAuth 2.0 Implicit flow
type ImplicitConfig struct {
	AuthURL  string `json:"authURL"`
	ClientID string `json:"clientID"`
}

// OA2Config defines the configurations for all the flows in OAuth 2.0
type OA2Config struct {
	AuthCodeCnfg AuthCodeConfig `json:"authCode"`
	ImplicitCnfg ImplicitConfig `json:"implicit"`
}

// Used as response for failed token requests.
// Using the necessary structures mentioned in RFC 6749 Section 4.1.2.1 (https://tools.ietf.org/html/rfc6749#section-4.1.2.1)
// error_uri is ignored since this is not a real API and has no documentation.
// state is ignored because it is ignored by flowHandlers.
type requestError struct {
	Error string `json:"error"`
	Desc  string `json:"error_description"`
}

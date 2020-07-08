package trustpilot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

var (
	// Compile time check whether Client value implements Trustpilot interface.
	_ Trustpilot = (*Client)(nil)

	ErrNotFound        = errors.New("NOT_FOUND")
	ErrInvalidArgument = errors.New("INVALID_ARGUMENT")
	ErrInternal        = errors.New("INTERNAL")
	ErrUnauthenticated = errors.New("UNAUTHENTICATED")

	apiBaseURL = &url.URL{
		Host:   "api.trustpilot.com",
		Scheme: "https",
		Path:   "/v1/",
	}

	invitationAPIBaseURL = &url.URL{
		Host:   "invitations-api.trustpilot.com",
		Scheme: "https",
		Path:   "/v1/",
	}

	tokenURL = &url.URL{
		Host:   "api.trustpilot.com",
		Scheme: "https",
		Path:   "/v1/oauth/oauth-business-users-for-applications/accesstoken",
	}
)

const (
	invitationPathf = "private/business-units/%s/email-invitations"
)

// AuthenticationConfig configuration for the oauth2 password grant type. Use
// TokenURL client option to set oauth2 token url.
type PasswordGrantConfig struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

// ClientOption is a functional option for configuring the API client.
type ClientOption func(*Client) error

// APIBaseURL allows to change the default API base url.
func APIBaseURL(u *url.URL) ClientOption {
	return func(c *Client) error {
		c.apiBaseURL = u
		return nil
	}
}

// TokenURL allows to change the default API oauth 2.0 token url.
func TokenURL(u *url.URL) ClientOption {
	return func(c *Client) error {
		c.tokenURL = u
		return nil
	}
}

// InvitationAPIBaseURL allows to change the default Invitation API base url.
func InvitationAPIBaseURL(u *url.URL) ClientOption {
	return func(c *Client) error {
		c.invitationAPIBaseURL = u
		return nil
	}
}

// AuthConfig is a functional option for configuring oauth 2.0 password grant credentials.
func AuthConfig(conf *PasswordGrantConfig) ClientOption {
	return func(c *Client) error {
		c.authConfig = conf
		return nil
	}
}

// Debug is a functional option for configuring client debug.
func Debug(b bool) ClientOption {
	return func(c *Client) error {
		c.debug = b
		return nil
	}
}

// HTTPClient is a functional option for configuring http client.
func HTTPClient(h *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = h
		return nil
	}
}

func (c *Client) applyOptions(opts ...ClientOption) error {
	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil
}

// Client implements Trustpilot API.
type Client struct {
	debug                bool
	apiBaseURL           *url.URL
	invitationAPIBaseURL *url.URL
	tokenURL             *url.URL
	authConfig           *PasswordGrantConfig
	tokenSource          oauth2.TokenSource

	httpClient *http.Client
}

// NewClient sets up a new Trustpilot client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		apiBaseURL:           apiBaseURL,
		invitationAPIBaseURL: invitationAPIBaseURL,
		tokenURL:             tokenURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Assume authentication is not required for the API.
	if c.authConfig == nil {
		return c, nil
	}

	// Set up oauth2 password grant flow.
	oauthConfig := &oauth2.Config{
		ClientID:     c.authConfig.ClientID,
		ClientSecret: c.authConfig.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: c.tokenURL.String(),
		},
	}

	t, err := oauthConfig.PasswordCredentialsToken(context.Background(), c.authConfig.Username, c.authConfig.Password)
	if err != nil {
		return nil, err
	}

	c.tokenSource = oauth2.ReuseTokenSource(t, oauthConfig.TokenSource(context.Background(), t))

	return c, nil
}

func (c *Client) CreateInvitation(ctx context.Context, r *CreateInvitationRequest) (*CreateInvitationResponse, error) {
	path := fmt.Sprintf(invitationPathf, r.BusinessUnitID)

	req, err := c.newInvitationAPIRequest("POST", path, r)
	if err != nil {
		return nil, err
	}

	var res CreateInvitationResponse
	resp, err := c.do(ctx, req, &res)
	if err != nil {
		return nil, err
	}

	if err := parseError(resp); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) ListTemplates(ctx context.Context, r *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	path := fmt.Sprintf(invitationPathf, r.BusinessUnitID)

	req, err := c.newInvitationAPIRequest("GET", path, r)
	if err != nil {
		return nil, err
	}

	var res ListTemplatesResponse
	resp, err := c.do(ctx, req, &res)
	if err != nil {
		return nil, err
	}

	if err := parseError(resp); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) newAPIRequest(method, path string, body interface{}) (*http.Request, error) {
	u := c.apiBaseURL.ResolveReference(&url.URL{Path: path})
	return c.newRequest(method, u, body)
}

func (c *Client) newInvitationAPIRequest(method, path string, body interface{}) (*http.Request, error) {
	u := c.invitationAPIBaseURL.ResolveReference(&url.URL{Path: path})
	return c.newRequest(method, u, body)
}

func (c *Client) newRequest(method string, u *url.URL, body interface{}) (*http.Request, error) {
	var b []byte
	if body != nil {
		jb, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = jb
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.tokenSource == nil {
		return req, nil
	}

	// Use the token for authentication.
	tkn, err := c.tokenSource.Token()
	if err != nil {
		var e error

		// Try to figure out what kind of error it is.
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			switch rErr.Response.StatusCode {
			case http.StatusUnauthorized:
				e = fmt.Errorf("%s: %w", rErr.Error(), ErrUnauthenticated)
			case http.StatusBadRequest:
				e = fmt.Errorf("%s: %w", rErr.Error(), ErrInvalidArgument)
			case http.StatusNotFound:
				e = fmt.Errorf("%s: %w", rErr.Error(), ErrNotFound)
			default:
				e = fmt.Errorf("%s: %w", rErr.Error(), ErrInternal)
			}
		}

		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+tkn.AccessToken)

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if c.debug {
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", reqDump)
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.debug {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", respDump)
	}

	if resp.ContentLength != 0 && v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func parseError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return ErrInvalidArgument
	case http.StatusUnauthorized:
		return ErrUnauthenticated
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return ErrInternal
	}
}

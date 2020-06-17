package trustpilot_test

import (
	"github.com/dietdoctor/go-trustpilot"
)

func ExampleNewClient() {
	authConfig := &trustpilot.PasswordGrantConfig{
		ClientID:     "1a9SsZ2XJKraw1Prl8m+gvq",
		ClientSecret: "AxLoJhlAcbWiVA8cHW2fINep",
		Username:     "apiuser@example.com",
		Password:     "6X2hQa1lPHu9dVjikUr0FbRt",
	}

	client, err := trustpilot.NewClient(
		trustpilot.AuthConfig(authConfig),
		trustpilot.Debug(true),
	)

	if err != nil {
		// TODO handle the error.
	}

	_ = client
}

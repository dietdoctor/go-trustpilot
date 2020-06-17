# go-trustpilot

Trustpilot API client library for Go.

![ci](https://github.com/dietdoctor/go-trustpilot/workflows/build-go/badge.svg) [![GoDoc](https://godoc.org/github.com/dietdoctor/go-trustpilot?status.svg)](https://godoc.org/github.com/dietdoctor/go-trustpilot)

This library does not implement full Trustpilot API.

## Getting Started

The client library can be configured using `ClientOption` as functional options.

```go
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
```

## Authentication

This client supports Trustpilot's oauth 2.0 `password` grant type. Other
Trustpilot-supported grant types may be added.

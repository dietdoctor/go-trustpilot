package trustpilot_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dietdoctor/go-trustpilot"
	"github.com/stretchr/testify/assert"
)

func setup() (*http.ServeMux, *trustpilot.Client, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	u, _ := url.Parse(server.URL)
	client, _ := trustpilot.NewClient(trustpilot.APIBaseURL(u), trustpilot.InvitationAPIBaseURL(u), trustpilot.Debug(true))

	teardown := func() {
		server.Close()
	}

	return mux, client, teardown
}

func fixture(path string) string {
	b, err := ioutil.ReadFile("testdata/fixtures/" + path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestCreateInvitation(t *testing.T) {
	mux, client, teardown := setup()
	defer teardown()

	businessUnitID := "123"

	mux.HandleFunc(fmt.Sprintf("/private/business-units/%s/email-invitations", businessUnitID),
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, fixture("error_response.json"))
		},
	)

	_, err := client.CreateInvitation(context.Background(), &trustpilot.CreateInvitationRequest{
		BusinessUnitID:  businessUnitID,
		ConsumerEmail:   "foobar@invalid",
		ReferenceNumber: "users/12344",
		ConsumerName:    "John Doe",
		Locale:          "en-US",
	})

	assert.EqualError(t, err, "INVALID_ARGUMENT: 'Consumer Email' is not in the correct format.")
}

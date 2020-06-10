package fakeserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dietdoctor/go-trustpilot"
)

type FakeServer struct {
	BusinessUnitID           string
	CreateInvitationResponse *trustpilot.CreateInvitationResponse
	ListTemplatesResponse    *trustpilot.ListTemplatesResponse
}

func NewServer() *FakeServer {
	return &FakeServer{}
}

func (s *FakeServer) createInvitationHandler(w http.ResponseWriter, r *http.Request) {
	res := s.CreateInvitationResponse

	b, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		panic(err)
	}
}

func (s *FakeServer) listTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	res := s.ListTemplatesResponse

	b, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		panic(err)
	}
}

func (s *FakeServer) Serve(listen string) error {
	createInvitationPath := fmt.Sprintf("/v1/private/business-units/%s/email-invitations", s.BusinessUnitID)
	http.HandleFunc(createInvitationPath, s.createInvitationHandler)

	listTemplatesPath := fmt.Sprintf("/v1/private/business-units/%s/templates", s.BusinessUnitID)
	http.HandleFunc(listTemplatesPath, s.listTemplatesHandler)

	return http.ListenAndServe(listen, nil)
}

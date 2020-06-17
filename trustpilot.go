package trustpilot

import (
	"context"
)

// Trustpilot defines the Trustpilot API client interface.
type Trustpilot interface {
	CreateInvitation(ctx context.Context, r *CreateInvitationRequest) (*CreateInvitationResponse, error)
	ListTemplates(ctx context.Context, r *ListTemplatesRequest) (*ListTemplatesResponse, error)
}

type CreateInvitationRequest struct {
	BusinessUnitID  string // Required.
	ConsumerEmail   string `json:"consumerEmail"` // Required.
	ReplyTo         string `json:"replyTo,omitempty"`
	ReferenceNumber string `json:"referenceNumber,omitempty"` // The customer’s internal reference number.
	ConsumerName    string `json:"consumerName,omitempty"`
	Locale          string `json:"locale,omitempty"`
	LocationID      string `json:"locationId,omitempty"`
	SenderEmail     string `json:"senderEmail,omitempty"`
	SenderName      string `json:"senderName,omitempty"`

	ServiceReviewInvitation *ServiceReviewInvitation `json:"serviceReviewInvitation,omitempty"`
}

type ServiceReviewInvitation struct {
	PreferredSendTime string   `json:"preferredSendTime,omitempty"` // ISO8601 UTC.
	RedirectURI       string   `json:"redirectUri,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	TemplateID        string   `json:"templateId,omitempty"`
}

type CreateInvitationResponse struct {
}

type ListTemplatesRequest struct {
	BusinessUnitID string // Required.
}

type ListTemplatesResponse struct {
	Templates []Template `json:"templates"`
}

type Template struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	IsDefaultTemplate bool   `json:"isDefaultTemplate"`
	Locale            string `json:"locale,omitempty"`
	Type              string `json:"type"`
}

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kwinso/medods-test-task/internal/db"
	"net/http"
	"net/netip"
	"net/url"
)

type ReportService interface {
	SendIPChangeReport(auth db.Auth, newIP netip.Addr) error
}

type WebhookReportsService struct {
	reportEndpoint url.URL
}

func NewWebhookReportsService(reportEndpoint url.URL) WebhookReportsService {
	return WebhookReportsService{
		reportEndpoint: reportEndpoint,
	}
}

var (
	MismatchedResponseStatusErrFormat = "expected response status to be %d, but got %s"
)

type ipChangeReport struct {
	Guid      string `json:"guid"`
	UserAgent string `json:"user_agent"`
	OldIP     string `json:"old_ip"`
	NewIP     string `json:"new_ip"`
}

func (s *WebhookReportsService) SendIPChangeReport(auth db.Auth, newIP netip.Addr) error {
	ipReport := &ipChangeReport{
		Guid:      auth.Guid,
		UserAgent: auth.UserAgent,
		OldIP:     newIP.String(),
		NewIP:     newIP.String(),
	}
	content, err := json.Marshal(ipReport)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.reportEndpoint.String(), "application/json", bytes.NewBuffer(content))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(MismatchedResponseStatusErrFormat, 200, resp.Status)
	}

	return nil
}

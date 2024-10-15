package gopchromedp

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
)

type Item struct {
	Url                        string
	FileName                   string
	ScreenShotFileName         string
	HTMLFileName               string
	RemoteIP                   string
	RemotePort                 string
	ResponseStatus             string
	ResponseStatusText         string
	ResponseHeaders            []string
	ResponseProtocol           string
	ResponseSecure             bool
	ResponseIssuer             string
	ResponseCertificateSubject string
	ResponseCertificateSanList []string
	Resources                  []Item
}

func NewItem(url string) (item Item) {
	item = Item{
		Url:                url,
		FileName:           GetFileName(url),
		ScreenShotFileName: GetScreenshotFileName(url),
		HTMLFileName:       GetHTMLFileName(url),
		ResponseSecure:     false,
	}

	return item
}

// GetHTMLFileName Compute the filename based on the URL.
func GetFileName(url string) string {
	filename := strings.ReplaceAll(url, ":", "-")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, ".", "_")
	filename = strings.ReplaceAll(filename, "?", "-")

	return filename
}

func getHTTPResponseHeaders(ctx context.Context, item *Item) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventRequestWillBeSent); ok {
			event := (*network.EventRequestWillBeSent)(ev)
			if event.RedirectResponse != nil {
				response := event.RedirectResponse
				if response.URL == item.Url || response.URL == item.Url+"/" {
					item.RemoteIP = fmt.Sprint(response.RemoteIPAddress)
					item.RemotePort = fmt.Sprint(response.RemotePort)
					item.ResponseStatus = fmt.Sprint(response.Status)
					item.ResponseStatusText = response.StatusText

					for header, value := range response.Headers {
						item.ResponseHeaders = append(item.ResponseHeaders, fmt.Sprintf("%s: %s", header, value))
					}

					if response.SecurityDetails != nil {
						item.ResponseProtocol = response.SecurityDetails.Protocol
						item.ResponseIssuer = response.SecurityDetails.Issuer
						item.ResponseCertificateSubject = response.SecurityDetails.SubjectName
						item.ResponseCertificateSanList = response.SecurityDetails.SanList
					}
				}
			}
		}

		if ev, ok := ev.(*network.EventResponseReceived); ok {
			event := (*network.EventResponseReceived)(ev)
			response := event.Response

			if response.URL == item.Url || response.URL == item.Url+"/" {
				item.RemoteIP = fmt.Sprint(response.RemoteIPAddress)
				item.RemotePort = fmt.Sprint(response.RemotePort)
				item.ResponseStatus = fmt.Sprint(response.Status)
				item.ResponseStatusText = response.StatusText

				for header, value := range response.Headers {
					item.ResponseHeaders = append(item.ResponseHeaders, fmt.Sprintf("%s: %s", header, value))
				}

				if response.SecurityState == security.StateSecure {
					item.ResponseSecure = true
				}
				if response.SecurityDetails != nil {
					item.ResponseProtocol = response.SecurityDetails.Protocol
					item.ResponseIssuer = response.SecurityDetails.Issuer
					item.ResponseCertificateSubject = response.SecurityDetails.SubjectName
					item.ResponseCertificateSanList = response.SecurityDetails.SanList
				}
			} else {
				resource := Item{
					Url:                fmt.Sprint(response.URL),
					RemoteIP:           fmt.Sprint(response.RemoteIPAddress),
					RemotePort:         fmt.Sprint(response.RemotePort),
					ResponseStatus:     fmt.Sprint(response.Status),
					ResponseStatusText: item.ResponseStatusText,
				}

				for header, value := range response.Headers {
					resource.ResponseHeaders = append(resource.ResponseHeaders, fmt.Sprintf("%s: %s\n", header, value))
				}

				if response.SecurityDetails != nil {
					item.ResponseProtocol = response.SecurityDetails.Protocol
					item.ResponseIssuer = response.SecurityDetails.Issuer
					item.ResponseCertificateSubject = response.SecurityDetails.SubjectName
					item.ResponseCertificateSanList = response.SecurityDetails.SanList
				}

				item.Resources = append(item.Resources, resource)
			}
		}
	})
}

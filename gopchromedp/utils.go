package gopchromedp

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Item struct {
	Url                        string
	RemoteIP                   string
	RemotePort                 string
	ResponseStatus             string
	ResponseStatusText         string
	ResponseHeaders            []string
	ResponseProtocol           string
	ResponseIssuer             string
	ResponseCertificateSubject string
	ResponseCertificateSanList []string
	Resources                  []Item
}

func getHTTPResponseHeaders(ctx context.Context, item *Item) {
	resource := Item{}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		// fmt.Println("Listen input :", reflect.TypeOf(ev))
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			event := (*network.EventResponseReceived)(ev)

			if event.Response.URL == item.Url || event.Response.URL == item.Url+"/" {
				item.RemoteIP = fmt.Sprint(event.Response.RemoteIPAddress)
				item.RemotePort = fmt.Sprint(event.Response.RemotePort)
				item.ResponseStatus = fmt.Sprint(event.Response.Status)
				item.ResponseStatusText = event.Response.StatusText

				for header, value := range event.Response.Headers {
					item.ResponseHeaders = append(item.ResponseHeaders, fmt.Sprintf("%s: %s", header, value))
				}

				item.ResponseProtocol = event.Response.SecurityDetails.Protocol
				item.ResponseIssuer = event.Response.SecurityDetails.Issuer
				item.ResponseCertificateSubject = event.Response.SecurityDetails.SubjectName
				item.ResponseCertificateSanList = event.Response.SecurityDetails.SanList
			} else {
				resource.Url = fmt.Sprint(event.Response.URL)
				resource.RemoteIP = fmt.Sprint(event.Response.RemoteIPAddress)
				resource.RemotePort = fmt.Sprint(event.Response.RemotePort)
				resource.ResponseStatus = fmt.Sprint(event.Response.Status)
				resource.ResponseStatusText = event.Response.StatusText

				for header, value := range event.Response.Headers {
					resource.ResponseHeaders = append(resource.ResponseHeaders, fmt.Sprintf("%s: %s\n", header, value))
				}

				item.Resources = append(item.Resources, resource)
			}
		}
	})
}

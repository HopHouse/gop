package target

// Code generated by cdproto-gen. DO NOT EDIT.

import (
	"github.com/chromedp/cdproto/cdp"
)

// ID [no description].
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-TargetID
type ID string

// String returns the ID as string value.
func (t ID) String() string {
	return string(t)
}

// SessionID unique identifier of attached debugging session.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-SessionID
type SessionID string

// String returns the SessionID as string value.
func (t SessionID) String() string {
	return string(t)
}

// Info [no description].
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-TargetInfo
type Info struct {
	TargetID         ID                   `json:"targetId"`
	Type             string               `json:"type"`
	Title            string               `json:"title"` // List of types: https://source.chromium.org/chromium/chromium/src/+/main:content/browser/devtools/devtools_agent_host_impl.cc?ss=chromium&q=f:devtools%20-f:out%20%22::kTypeTab%5B%5D%22
	URL              string               `json:"url"`
	Attached         bool                 `json:"attached"`                // Whether the target has an attached client.
	OpenerID         ID                   `json:"openerId,omitempty"`      // Opener target Id
	CanAccessOpener  bool                 `json:"canAccessOpener"`         // Whether the target has access to the originating window.
	OpenerFrameID    cdp.FrameID          `json:"openerFrameId,omitempty"` // Frame id of originating window (is only set if target has an opener).
	BrowserContextID cdp.BrowserContextID `json:"browserContextId,omitempty"`
	Subtype          string               `json:"subtype,omitempty"` // Provides additional details for specific target types. For example, for the type of "page", this may be set to "portal" or "prerender".
}

// FilterEntry a filter used by target query/discovery/auto-attach
// operations.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-FilterEntry
type FilterEntry struct {
	Exclude bool   `json:"exclude,omitempty"` // If set, causes exclusion of matching targets from the list.
	Type    string `json:"type,omitempty"`    // If not present, matches any type.
}

// Filter the entries in TargetFilter are matched sequentially against
// targets and the first entry that matches determines if the target is included
// or not, depending on the value of exclude field in the entry. If filter is
// not specified, the one assumed is [{type: "browser", exclude: true}, {type:
// "tab", exclude: true}, {}] (i.e. include everything but browser and tab).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-TargetFilter
type Filter []struct {
	Exclude bool   `json:"exclude,omitempty"` // If set, causes exclusion of matching targets from the list.
	Type    string `json:"type,omitempty"`    // If not present, matches any type.
}

// RemoteLocation [no description].
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Target#type-RemoteLocation
type RemoteLocation struct {
	Host string `json:"host"`
	Port int64  `json:"port"`
}

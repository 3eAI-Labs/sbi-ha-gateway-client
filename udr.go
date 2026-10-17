// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import "encoding/json"

// Values of DataRequest.Service: the API name of the request URI.
const (
	ServiceNudrDR  = "nudr-dr"  // Nudr_DataRepository (TS 29.504, TS 29.505)
	ServiceNudmSDM = "nudm-sdm" // Nudm_SubscriberDataManagement
	ServiceNudmPP  = "nudm-pp"  // Nudm_ParameterProvision
	ServiceNudmEE  = "nudm-ee"  // Nudm_EventExposure
)

// DataRequest is the request on SubjectDataQuery and SubjectDataUpdate: one call of a
// data service, described by its URI and HTTP method. Service and Resource identify the
// resource; for example Nudm_SDM GET /nudm-sdm/v1/{supi}/am-data?plmn-id=... arrives as
// Service "nudm-sdm", Resource "am-data", Query "plmn-id=...".
type DataRequest struct {
	// Service is the API name, one of the Service* constants.
	Service string `json:"service"`
	// UeId is the UE identifier of the request URI, as sent: a SUPI (Nudr, SDM, PP), a GPSI
	// (Nudr, PP, EE), an External Group Id (PP, EE) or "anyUE" (EE). Empty when the URI has
	// none (shared data).
	UeId string `json:"ueId"`
	// DataSet is, for Nudr_DataRepository only, the first path segment after
	// /nudr-dr/v1/, for example "subscription-data"; empty for the Nudm services.
	DataSet string `json:"dataSet"`
	// Resource is the path after the UE identifier, or after the API version (Nudr: after
	// the data set) when the URI has none, without a leading slash: "am-data",
	// "sdm-subscriptions/1", "pp-data", "ee-subscriptions", "shared-data", "shared-data/x",
	// "authentication-data/authentication-subscription"; "" for the UE's document itself
	// (SDM GET /{supi}, Nudr GET /subscription-data/{ueId}).
	Resource string `json:"resource"`
	// Method is the HTTP method: GET on SubjectDataQuery; PUT, PATCH, POST or DELETE on
	// SubjectDataUpdate.
	Method string `json:"method"`
	// Query is the request URI's query string, unchanged, without "?"; parse it with
	// net/url.ParseQuery. Some reads need it, such as shared-data-ids on shared-data.
	Query string `json:"query,omitempty"`
	// Payload is the request body for PUT, PATCH and POST, unchanged.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// DataResponse is the success reply on SubjectDataQuery and SubjectDataUpdate: the gateway
// answers the network function 200 with Payload as the body. Refuse a request with an
// ErrorReply.
type DataResponse struct {
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NotificationPayload is published on SubjectDataNotify.
type NotificationPayload struct {
	// CallbackUri is the network function's notification URI, tried first.
	CallbackUri string `json:"callbackUri"`
	// AlternateCallbackUris are tried in order when CallbackUri fails with a 5xx, a 429 or
	// a transport error, in up to the gateway's MAX_RETRIES passes (3 by default). Any other
	// 4xx ends the delivery.
	AlternateCallbackUris []string `json:"alternateCallbackUris,omitempty"`
	// ChangeItem is POSTed unchanged as the notification body.
	ChangeItem json.RawMessage `json:"changeItem"`
}

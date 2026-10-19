// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import "encoding/json"

// API names: Request.Service, the <apiName> of the request URI.
const (
	ServiceNudmUEAU = "nudm-ueau" // Nudm_UEAuthentication (TS 29.503)
	ServiceNudmUECM = "nudm-uecm" // Nudm_UEContextManagement (TS 29.503)
	ServiceNudmSDM  = "nudm-sdm"  // Nudm_SubscriberDataManagement (TS 29.503)
	ServiceNudmEE   = "nudm-ee"   // Nudm_EventExposure (TS 29.503)
	ServiceNudmPP   = "nudm-pp"   // Nudm_ParameterProvision (TS 29.503)
	ServiceNudmMT   = "nudm-mt"   // Nudm_MT (TS 29.503)
	ServiceNudmRSDS = "nudm-rsds" // Nudm_ReportSMDeliveryStatus (TS 29.503)
	ServiceNudrDR   = "nudr-dr"   // Nudr_DataRepository, subscription data (TS 29.504, TS 29.505)
)

// Request is the request on every gateway → application subject, SubjectAuthRequest to
// SubjectDataUpdate: one operation a network function called, in the specification's terms.
// The body is the network function's, unchanged: decode it with the data types of the
// API's OpenAPI definition (3GPP TS 29.503, TS 29.505).
//
// For example, an AMF's GET /nudm-sdm/v2/imsi-001010000000001/am-data?plmn-id=... arrives as
// Service "nudm-sdm", Operation "GetAmData", Method "GET", Path "/imsi-001010000000001/am-data",
// Params {"supi": "imsi-001010000000001"}, Query "plmn-id=...".
type Request struct {
	// Service is the API name, one of the Service* constants.
	Service string `json:"service"`
	// Operation is the operationId of the API's OpenAPI definition: "GenerateAuthData",
	// "3GppRegistration", "GetAmData", "Subscribe". Dispatch on it.
	Operation string `json:"operation"`
	// Method is the HTTP method.
	Method string `json:"method"`
	// Path is the request URI's path after /{apiName}/{apiVersion}, as the network function
	// sent it (percent-encoded): "/imsi-001010000000001/am-data".
	Path string `json:"path"`
	// Params are the path variables, by the names the specification gives them ("supi",
	// "supiOrSuci", "ueId", "ueIdentity", "pduSessionId", "subscriptionId", ...), decoded.
	// UE identifiers have been checked against the kinds the operation accepts.
	Params map[string]string `json:"params,omitempty"`
	// Query is the request URI's query string, unchanged, without "?"; parse it with
	// net/url.ParseQuery.
	Query string `json:"query,omitempty"`
	// ContentType is the media type of Body: "application/json",
	// "application/merge-patch+json" or "application/json-patch+json". Empty without a body.
	ContentType string `json:"contentType,omitempty"`
	// Body is the request body, unchanged: a JSON object or array, or absent.
	Body json.RawMessage `json:"body,omitempty"`
}

// Reply is the success reply on every gateway → application subject. Refuse a request with
// an ErrorReply instead.
type Reply struct {
	// Status is the HTTP status the network function gets. It must be one of the success
	// statuses the specification gives the operation; it may be left 0 when there is only
	// one. For example a registration (PUT) is 201 when it created the registration, 200
	// or 204 when it replaced one.
	Status int `json:"status,omitempty"`
	// ResourceId is the identifier of the resource a POST created, for the operations whose
	// 201 carries a Location to a new resource under the request URI (ConfirmAuth,
	// Subscribe, SubscribeToSharedData, CreateEeSubscription, the Nudr subscriptions): the
	// gateway sends Location <request URI>/<ResourceId>. One path segment.
	ResourceId string `json:"resourceId,omitempty"`
	// Body is the response body, unchanged: required with 200 and 201, ignored with 204.
	Body json.RawMessage `json:"body,omitempty"`
}

// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

// ErrorReply refuses a request on any of the gateway → application request subjects
// (SubjectAuthRequest to SubjectDataUpdate). The gateway answers the network function with
// ProblemDetails.Status as the HTTP status and ProblemDetails as the
// application/problem+json body, and emits no audit event: nothing changed.
//
// A reply is an ErrorReply when it has a problemDetails member; otherwise it is the
// operation's result. Use the application errors TS 29.503 defines for the service — Tables
// 6.1.7.3-1 (SDM), 6.2.7.3-1 (UECM), 6.3.7.3-1 (UEAU), 6.4.7.3-1 (EE), 6.5.7.3-1 (PP) and
// 6.7.7.3-1 (MT), for example USER_NOT_FOUND with 404 — or the common ones of TS 29.500
// Table 5.2.7.2-1.
type ErrorReply struct {
	ProblemDetails ProblemDetails `json:"problemDetails"`
}

// ProblemDetails is TS 29.571 ProblemDetails, as far as a UDM or UDR uses it. In an
// ErrorReply, Status is required and must be 400–599; the gateway answers 500 otherwise.
type ProblemDetails struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title,omitempty"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	// Instance identifies this occurrence of the problem (a URI reference).
	Instance string `json:"instance,omitempty"`
	// Cause is the machine-readable application error, such as "USER_NOT_FOUND".
	Cause string `json:"cause,omitempty"`
	// InvalidParams lists the parameters that made the request invalid; TS 29.500 asks
	// for it with the 400 causes about IEs and query parameters.
	InvalidParams     []InvalidParam `json:"invalidParams,omitempty"`
	SupportedFeatures string         `json:"supportedFeatures,omitempty"`
}

// InvalidParam is TS 29.571 InvalidParam: Param is a JSON Pointer to the attribute in
// the request body, or the name of a query parameter or header.
type InvalidParam struct {
	Param  string `json:"param"`
	Reason string `json:"reason,omitempty"`
}

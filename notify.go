// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import "encoding/json"

// NotificationPayload is published on SubjectDataNotify.
type NotificationPayload struct {
	// CallbackUri is the network function's notification URI, tried first.
	CallbackUri string `json:"callbackUri"`
	// AlternateCallbackUris are tried in order when CallbackUri fails with a 5xx, a 429 or
	// a transport error, in up to the gateway's MAX_RETRIES passes (3 by default). Any other
	// 4xx ends the delivery. A network function that keeps failing, or that answered with
	// Retry-After, is not called again until it has rested or the time has passed.
	AlternateCallbackUris []string `json:"alternateCallbackUris,omitempty"`
	// ChangeItem is POSTed unchanged as the notification body.
	ChangeItem json.RawMessage `json:"changeItem"`
}

// NotificationFailure is published on SubjectDataNotifyFailed when the gateway gives up on a
// NotificationPayload: no network function got the notification. A Status of 404, for
// example, tells the back end that the network function no longer has the subscription.
type NotificationFailure struct {
	// Notification is the NotificationPayload as the gateway received it.
	Notification NotificationPayload `json:"notification"`
	// Result says why the gateway stopped: one of the Notify* constants.
	Result string `json:"result"`
	// Status is the last HTTP status a network function answered; 0 when none answered.
	Status int `json:"status,omitempty"`
	// ProblemDetails is the body of that answer when it is a JSON object, usually a
	// ProblemDetails (TS 29.571).
	ProblemDetails json.RawMessage `json:"problemDetails,omitempty"`
	// Cause describes the last failure for people: "HTTP_503", "TRANSPORT_ERROR",
	// "RETRY_AFTER_TOO_LONG", "SHUTDOWN", ...
	Cause string `json:"cause,omitempty"`
	// Attempts is the number of requests the gateway sent.
	Attempts int `json:"attempts"`
}

// Values of NotificationFailure.Result.
const (
	// NotifyClientError: a network function answered a 4xx other than 429. The gateway
	// neither retries it nor tries the alternates.
	NotifyClientError = "client_error"
	// NotifyExhausted: every attempt failed with a 5xx, a 429 or a transport error, a
	// Retry-After asked for a longer wait than the gateway holds a notification, or the
	// gateway stopped first.
	NotifyExhausted = "exhausted"
	// NotifyNotAllowed: OUTBOUND_ALLOWED_DESTINATIONS allows none of the callback URIs.
	NotifyNotAllowed = "not_allowed"
	// NotifyNotImplemented: the gateway's OUTBOUND_PROFILE cannot deliver notifications.
	NotifyNotImplemented = "not_implemented"
)

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
	// 4xx ends the delivery.
	AlternateCallbackUris []string `json:"alternateCallbackUris,omitempty"`
	// ChangeItem is POSTed unchanged as the notification body.
	ChangeItem json.RawMessage `json:"changeItem"`
}

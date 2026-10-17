// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import "time"

// --- Serving SMSF lookup (SubjectUecmRequest) ------------------------------------------

// Values of NudmUecmMessage.MessageType and NudmUecmMessage.Operation.
const (
	UecmMessageType    = "SBI_NUDM_UECM"
	UecmOpGetUeContext = "GET_UE_CONTEXT" // request
	UecmOpResponse     = "RESPONSE"       // reply
)

// NudmUecmMessage is both the request and the reply on SubjectUecmRequest.
//
// The request sets MessageType UecmMessageType, Operation UecmOpGetUeContext, Msisdn and
// CorrelationId. The reply echoes CorrelationId with Operation UecmOpResponse and an
// HTTP-style status: 200 with Supi, SmsfFqdn and SmsfPlmnId; 404 when no SMSF is known
// for the subscriber; 400 for another Operation; 503 when the lookup failed.
// ProblemDetails is then a JSON object with a "cause" member. A request that is not JSON
// gets no reply.
type NudmUecmMessage struct {
	MessageType    string `json:"messageType"`
	Operation      string `json:"operation"`
	Msisdn         string `json:"msisdn"`
	CorrelationId  string `json:"correlationId"`
	Supi           string `json:"supi,omitempty"`
	SmsfFqdn       string `json:"smsfFqdn,omitempty"`
	SmsfPlmnId     string `json:"smsfPlmnId,omitempty"`
	HttpStatus     int    `json:"httpStatus,omitempty"`
	ProblemDetails string `json:"problemDetails,omitempty"`
}

// --- SMS through the serving SMSF (SubjectSmsfSendRequest) -----------------------------

// SbiOtaMessage is the message on SubjectSmsfSendRequest.
//
// The gateway handles Header.Service "Nsmsf_SMService" with Header.Operation "SendSMS"
// only. The DeliveryEvent's TransactionID comes from Metadata "tx_id" (or
// "transactionId", "transaction_id"), and is "unknown" when none is set.
type SbiOtaMessage struct {
	Header   SbiHeader         `json:"header"`
	Routing  SbiRouting        `json:"routing"`
	Payload  SS7PayloadView    `json:"payload"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SbiHeader names the SBI service operation.
type SbiHeader struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Version   string `json:"version"`
}

// SbiRouting says where the SMS goes: the subscriber and the FQDN (or URI) of the SMSF
// serving it, as returned by SubjectUecmRequest. A FQDN without a scheme is reached
// over https.
type SbiRouting struct {
	Supi     string `json:"supi"`
	SmsfFqdn string `json:"smsfFqdn"`
	Gpsi     string `json:"gpsi,omitempty"`
}

// SS7PayloadView carries the SMS TPDU: Encoding "BINARY", Content base64. The gateway
// does not decode it.
type SS7PayloadView struct {
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// Values of DeliveryEvent.Status.
const (
	DeliveryDelivered = "DELIVERED"
	DeliveryFailed    = "FAILED"
	// DeliveryPending is a transient failure: the gateway also asks JetStream to
	// redeliver the SbiOtaMessage.
	DeliveryPending = "PENDING"
)

// DeliveryEvent is published on SubjectDeliveryEvent for each SbiOtaMessage the gateway
// processed. Bearer is "SBI"; BearerMeta carries "supi" and "smsfFqdn".
type DeliveryEvent struct {
	TransactionID string            `json:"tx_id"`
	Bearer        string            `json:"bearer"`
	Status        string            `json:"status"`
	Cause         string            `json:"cause,omitempty"`
	BearerMeta    map[string]string `json:"bearerMeta,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

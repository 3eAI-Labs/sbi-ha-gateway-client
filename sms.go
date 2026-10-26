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
// HTTP-style status: 200 with Supi, SmsfPlmnId and SmsfFqdn, SmsfInstanceId or both; 404 when no SMSF is known
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
	SmsfInstanceId string `json:"smsfInstanceId,omitempty"`
	SmsfPlmnId     string `json:"smsfPlmnId,omitempty"`
	HttpStatus     int    `json:"httpStatus,omitempty"`
	ProblemDetails string `json:"problemDetails,omitempty"`
}

// --- SMS through the serving SMSF (SubjectSmsfSendRequest) -----------------------------

// SmsMessage is the message on SubjectSmsfSendRequest: an MT SMS the gateway forwards to
// the serving SMSF with Nsmsf_SMService MtForwardSm (TS 29.540: POST
// .../ue-contexts/{supi}/send-mt-sms, a multipart/related body whose binary part is
// Payload).
//
// The gateway handles Header.Service "Nsmsf_SMService" with Header.Operation "MtForwardSm"
// ("SendSMS" is accepted as its former name). The DeliveryEvent's TransactionID comes from
// Metadata "tx_id" (or "transactionId", "transaction_id"), and is "unknown" when none is
// set.
type SmsMessage struct {
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

// SbiRouting says where the SMS goes: the subscriber and the SMSF serving it, as returned by
// SubjectUecmRequest.
//
// SmsfInstanceId is the SMSF's NF instance ID, as the SMSF registered it with the UDM
// (TS 29.503 SmsfRegistration). When it is set and the gateway registers an SMS_GMSC
// profile, the gateway finds the SMSF's nsmsf-sms service with NRF discovery (TS 29.510),
// sends the SMS there and asks for an access token for that instance alone; SmsfFqdn is not
// used. Otherwise the SMS goes to SmsfFqdn (or URI); a FQDN without a scheme is reached over
// https.
type SbiRouting struct {
	Supi           string `json:"supi"`
	SmsfFqdn       string `json:"smsfFqdn"`
	SmsfInstanceId string `json:"smsfInstanceId,omitempty"`
	Gpsi           string `json:"gpsi,omitempty"`
}

// SS7PayloadView carries the SMS: Encoding EncodingRPData, Content the RP-DATA (network to
// mobile station, TS 24.011 clause 7.3.1.1) in base64 — RP-Message Reference, the Service
// Centre address as RP-Originator Address, an empty RP-Destination Address and the
// SMS-DELIVER TPDU as RP-User Data. MtForwardSm's JSON part has no Service Centre address,
// so the payload carries it: this is the gateway's reading of TS 29.540 clause 6.1.6.4.2,
// which states the encoding only as "TS 23.040 and TS 24.011". The gateway checks the
// message type and sends the bytes unchanged; a payload in any other encoding, such as a
// bare TPDU, fails.
type SS7PayloadView struct {
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// EncodingRPData is SS7PayloadView.Encoding for an RP-DATA in base64.
const EncodingRPData = "RP_DATA"

// Values of DeliveryEvent.Status. For an SMS the SMSF accepted, the status comes from the
// mobile station's report in the MtForwardSm response: RP-ACK is DeliveryDelivered,
// RP-ERROR DeliveryFailed with Cause "RP_CAUSE_<n>" (TS 24.011 Table 8.4). The report
// itself, base64, is in BearerMeta "deliveryReport".
const (
	DeliveryDelivered = "DELIVERED"
	DeliveryFailed    = "FAILED"
	// DeliveryPending is a transient failure: the gateway also asks JetStream to
	// redeliver the SmsMessage. On the last delivery a transient failure is
	// DeliveryFailed instead, with Cause "RETRIES_EXHAUSTED" and the transient cause in
	// BearerMeta "lastCause".
	DeliveryPending = "PENDING"
)

// DeliveryEvent is published on SubjectDeliveryEvent for each SmsMessage the gateway
// processed. Bearer is "SBI"; BearerMeta carries "supi" and "smsfFqdn", and "smsfInstanceId"
// when the SmsMessage has one.
type DeliveryEvent struct {
	TransactionID string            `json:"tx_id"`
	Bearer        string            `json:"bearer"`
	Status        string            `json:"status"`
	Cause         string            `json:"cause,omitempty"`
	BearerMeta    map[string]string `json:"bearerMeta,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

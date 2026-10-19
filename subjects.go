// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

// Gateway → application, core NATS request/reply. The gateway sends a Request when a
// network function calls one of the operations it serves and waits for the reply, up to its
// request timeout, to build the HTTP response. Subscribe with a queue group so that one
// instance of the application answers each request.
//
// Every reply is a JSON object: a Reply, or an ErrorReply. No reply in time, or no
// application subscribed, is answered 504 TARGET_NF_NOT_REACHABLE; a reply the gateway
// cannot use is answered 500.
const (
	// SubjectAuthRequest carries the Nudm_UEAU operations other than auth-events
	// (GenerateAuthData and the other generate-av operations). Their replies carry
	// authentication vectors (key material): restrict this subject with NATS permissions to
	// the gateway and the authentication back end.
	SubjectAuthRequest = "5g.auth.request"
	// SubjectAuthConfirm carries the Nudm_UEAU auth-events operations, ConfirmAuth and
	// DeleteAuth.
	SubjectAuthConfirm = "5g.auth.confirm"
	// SubjectMobilityRegistration carries the Nudm_UECM operations that change a
	// registration: PUT, PATCH, DELETE, and the POSTs other than SendRoutingInfoSm.
	SubjectMobilityRegistration = "5g.mobility.registration"
	// SubjectMobilityQuery carries the Nudm_UECM reads: the GETs and SendRoutingInfoSm.
	SubjectMobilityQuery = "5g.mobility.query"
	// SubjectLocationInfo carries Nudm_MT: QueryUeInfo and ProvideLocationInfo.
	SubjectLocationInfo = "5g.mt.loc-info"
	// SubjectDataQuery carries the reads (HTTP GET) of Nudm_SDM, Nudm_EE, Nudm_PP and
	// Nudr_DataRepository subscription data.
	SubjectDataQuery = "5g.data.query"
	// SubjectDataUpdate carries the changes (HTTP PUT, PATCH, POST, DELETE) of the same APIs,
	// and Nudm_ReportSMDeliveryStatus.
	SubjectDataUpdate = "5g.data.update"
)

// Application → gateway.
const (
	// SubjectDataNotify is a core NATS publish of a NotificationPayload: the gateway POSTs
	// its ChangeItem to the callback URI, trying the alternates in order when the primary
	// fails. The gateway subscribes with queue group "sbi-gateway", so each notification is
	// delivered by one gateway instance.
	SubjectDataNotify = "5g.data.notify"
	// SubjectUecmRequest is core NATS request/reply: NudmUecmMessage (Operation
	// "GET_UE_CONTEXT") → NudmUecmMessage (Operation "RESPONSE") naming the SMSF that
	// serves the subscriber, or an HTTP status and problem cause when there is none.
	SubjectUecmRequest = "sbi.outbound.nudm.uecm.request"
	// SubjectSmsfSendRequest is a JetStream work-queue subject carrying an SbiOtaMessage:
	// the gateway sends the SMS with Nsmsf_SMService SendSMS (TS 29.540) and publishes a
	// DeliveryEvent. The stream (StreamSmsfSend) belongs to the application; the gateway
	// only creates its durable consumer (ConsumerSmsfSend) on it.
	SubjectSmsfSendRequest = "sbi.outbound.nsmsf.send.request"
)

// Gateway → application, published once per SbiOtaMessage it has processed.
const (
	// SubjectDeliveryEvent carries a DeliveryEvent. The gateway publishes through JetStream
	// when it is available, so a stream on this subject is expected.
	SubjectDeliveryEvent = "ota.delivery.event"
)

// JetStream names used with SubjectSmsfSendRequest.
const (
	// StreamSmsfSend is the stream, created by the application, that captures
	// SubjectSmsfSendRequest.
	StreamSmsfSend = "OTA_OUTBOUND"
	// ConsumerSmsfSend is the durable pull consumer the gateway creates on StreamSmsfSend.
	ConsumerSmsfSend = "sbi-gateway-smsf"
)

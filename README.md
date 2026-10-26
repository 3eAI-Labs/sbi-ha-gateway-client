# sbi-ha-gateway-client

NATS wire contract for [sbi-ha-gateway](https://github.com/3eAI-Labs/sbi-ha-gateway), a gateway that
serves the Nudm and Nudr APIs of a 5G core's service-based interface (3GPP TS 29.503, TS 29.504)
over HTTP/2, calls the SMSF's Nsmsf_SMService (TS 29.540), and bridges both to applications over
NATS.

Licensed under the Apache License 2.0, separately from the gateway (AGPL-3.0-only), so a UDM/UDR back
end, an SMS platform or any other application can link it under any licence.

```bash
go get github.com/3eAI-Labs/sbi-ha-gateway-client
```

The module has no dependencies outside the Go standard library.

## What it covers

The subjects and JSON envelopes of every NATS interface the gateway has.

### Network function → gateway → application

The gateway serves the operations of the Nudm APIs it routes (UEAU, UECM, SDM, EE, PP, MT, RSDS; TS 29.503)
and of Nudr_DataRepository subscription data (TS 29.504, TS 29.505), as their OpenAPI definitions
give them. For each call it checks the URI, its UE identifiers and the request body's media type and
JSON, and sends a core NATS request; it builds the HTTP response from the reply. The bodies are the
network function's and your application's, unchanged: decode them with the data types of the API's
OpenAPI definition. Answer with a core subscription (`msg.Respond`), and keep these subjects out of
JetStream streams: a stream that captures a subject answers every request with its own publish
acknowledgement before your application can, and the gateway answers the network function 500.

Every request is a `Request`: `service` (the API name), `operation` (the operationId of its OpenAPI
definition, such as `GenerateAuthData`, `3GppRegistration` or `GetAmData`), `method`, `path` (after
`/{apiName}/{apiVersion}`), `params` (the path variables, by the specification's names), `query`,
`contentType` and `body`.

| Operations | Subject |
|------------|---------|
| Nudm_UEAU, except auth-events: they return authentication vectors | `5g.auth.request` |
| Nudm_UEAU auth-events (`ConfirmAuth`, `DeleteAuth`) | `5g.auth.confirm` |
| Nudm_UECM changes (PUT, PATCH, DELETE, and the POSTs other than `SendRoutingInfoSm`) | `5g.mobility.registration` |
| Nudm_UECM reads (GET, `SendRoutingInfoSm`) | `5g.mobility.query` |
| Nudm_MT | `5g.mt.loc-info` |
| Nudm_SDM, Nudm_EE, Nudm_PP, Nudr_DataRepository reads (GET) | `5g.data.query` |
| The same APIs' changes (PUT, PATCH, POST, DELETE), and Nudm_ReportSMDeliveryStatus | `5g.data.update` |

Reply with a `Reply` — `status`, `resourceId`, `body` — or an `ErrorReply`. The gateway answers the
network function:

| Reply | HTTP response | Audit event |
|-------|---------------|-------------|
| `Reply` | `status`: one of the success statuses the specification gives the operation (0: its only one), `body` as the `application/json` body (required except with 204). A 201 that creates a resource under the request URI needs `resourceId`: `Location` is the request URI plus `/<resourceId>`; a 201 whose resource is the request URI gets that as `Location` | for every operation but GET, with the operationId as the action |
| `ErrorReply` `{"problemDetails": {...}}` | `problemDetails.status` (400–599), `problemDetails` as the `application/problem+json` body | none |
| None within the gateway's request timeout, or no application subscribed | 504, cause `TARGET_NF_NOT_REACHABLE` | none |
| A status the operation does not allow, a missing body or `resourceId`, a `problemDetails` without a 4xx/5xx `status`, anything that is not a JSON object, or a JetStream acknowledgement | 500, cause `SYSTEM_FAILURE` | none |

Use the application errors TS 29.503 defines for the API (`USER_NOT_FOUND` with 404, for example) or
the common ones of TS 29.500 Table 5.2.7.2-1.

### Application → gateway

| Purpose | Subject | Message |
|---------|---------|---------|
| Deliver a notification to a network function's callback URI | `5g.data.notify` (core publish) | `NotificationPayload` |
| A notification the gateway could not deliver, with the network function's last answer | `5g.data.notify.failed` (core publish, from the gateway) | `NotificationFailure` |
| Ask which SMSF serves a subscriber | `sbi.outbound.nudm.uecm.request` (core request/reply) | `NudmUecmMessage` → `NudmUecmMessage` |
| Send an MT SMS through the serving SMSF (Nsmsf_SMService MtForwardSm; the SMS as an RP-DATA) | `sbi.outbound.nsmsf.send.request` (JetStream work queue) | `SmsMessage` |
| Result of each SMS the gateway processed | `sbi.nsmsf.delivery.event` (the gateway's `SMS_DELIVERY_EVENT_SUBJECT` sets another) | `DeliveryEvent` |

`sbi.outbound.nsmsf.send.request` is the one subject that is meant to be in a stream. The gateway
creates its durable consumer `sbi-gateway-smsf` on the stream that captures it, whatever its name;
when none does, it creates `SBI_SMSF_SEND` (its `SMS_STREAM` names another), a work queue with one
replica. Create the stream yourself to choose its replicas and limits. The delivery events go
through JetStream when a stream captures their subject, and as core publishes otherwise. The gateway asks for redelivery (up to three deliveries) when the SMSF fails
transiently, and reports `PENDING` in the `DeliveryEvent` each time but the last. A transient
failure on the third delivery is reported `FAILED` with cause `RETRIES_EXHAUSTED` and the transient
cause in `BearerMeta` `lastCause`, so every SMS ends with `DELIVERED` or `FAILED`.

## Usage

A UDM back end answering generate-auth-data:

```go
nc, err := nats.Connect(natsURL)
if err != nil { /* ... */ }

respond := func(m *nats.Msg, v any) {
	data, _ := json.Marshal(v)
	_ = m.Respond(data)
}

_, err = nc.QueueSubscribe(sbiclient.SubjectAuthRequest, "udm", func(m *nats.Msg) {
	var req sbiclient.Request
	if err := json.Unmarshal(m.Data, &req); err != nil || req.Operation != "GenerateAuthData" {
		respond(m, sbiclient.ErrorReply{ProblemDetails: sbiclient.ProblemDetails{
			Status: 501, Detail: "operation not supported",
		}})
		return
	}
	var info AuthenticationInfoRequest // TS 29.503, from your OpenAPI models
	if err := json.Unmarshal(req.Body, &info); err != nil {
		respond(m, sbiclient.ErrorReply{ProblemDetails: sbiclient.ProblemDetails{
			Status: 400, Cause: "INVALID_MSG_FORMAT", Detail: err.Error(),
		}}) // the AUSF gets a 400 with this ProblemDetails
		return
	}
	result := authenticationInfoResult(req.Params["supiOrSuci"], info) // authType, authenticationVector, supi
	body, _ := json.Marshal(result)
	respond(m, sbiclient.Reply{Body: body}) // 200: the operation's only success status
})
```

Subscribe with a queue group so that one instance of your application answers each request.

## Security

The `5g.auth.request` reply carries 5G authentication vectors (RAND, AUTN, XRES\*, K_AUSF). Restrict
the subject with NATS permissions — only the gateway publishes requests, only the authentication back
end subscribes — run NATS with TLS, and never log these payloads.

## Versioning

Below 1.0, a minor version may change the schema; patch versions do not. `TestWireSchema` pins every
field name and JSON type of the envelopes.

## Licence

[Apache License 2.0](LICENSE). Contributions require the
[3eAI Labs Contributor Licence Agreement](CLA.md).

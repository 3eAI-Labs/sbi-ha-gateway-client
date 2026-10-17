# sbi-ha-gateway-client

NATS wire contract for [sbi-ha-gateway](https://github.com/3eAI-Labs/sbi-ha-gateway), a gateway that
serves part of a 5G core's service-based interface (Nudm, Nudr, Nsmsf; 3GPP TS 29.503, TS 29.504,
TS 29.540) over HTTP/2 and bridges it to applications over NATS.

Licensed under the Apache License 2.0, separately from the gateway (AGPL-3.0-only), so a UDM/UDR back
end, an SMS platform or any other application can link it under any licence.

```bash
go get github.com/3eAI-Labs/sbi-ha-gateway-client
```

The module has no dependencies outside the Go standard library.

## What it covers

The subjects and JSON envelopes of every NATS interface the gateway has.

### Network function → gateway → application

The gateway sends a core NATS request when a network function calls its HTTP/2 API, and builds the
HTTP response from the reply. Answer with a core subscription (`msg.Respond`), and keep these subjects
out of JetStream streams: a stream that captures a subject answers every request with its own publish
acknowledgement before your application can, and the gateway answers the network function 500.

| SBI operation | Subject | Request → reply |
|---------------|---------|-----------------|
| Nudm_UEAU generate-auth-data | `5g.auth.request` | `AuthRequest` → `AuthenticationInfoResult` |
| Nudm_UEAU auth-events | `5g.auth.confirm` | `AuthConfirmRequest` → `AuthConfirmReply` (`authEventId`) |
| Nudm_UECM PUT, PATCH (AMF 3GPP access, SMF), DELETE (SMF, SMSF 3GPP access) | `5g.mobility.registration` | `MobilityRequest` → `MobilityReply` (`{}`, or `{"updated": true}` when a PUT replaced a registration) |
| Nudm_UECM GET (registration data sets, AMF, SMSF, SMF registrations) | `5g.mobility.query` | `MobilityQuery` → a JSON object, returned to the network function |
| Nudm_MT provide-loc-info | `5g.mt.loc-info` | `LocationInfoRequest` → `ProvideLocInfoResult` |
| Nudr_DR, Nudm_SDM, Nudm_PP, Nudm_EE reads (GET) | `5g.data.query` | `DataRequest` → `DataResponse` |
| The same services' writes (PUT, PATCH, POST, DELETE) | `5g.data.update` | `DataRequest` → `DataResponse` |

Every reply is a JSON object: the result in the table (`{}` where the gateway needs none), or an
`ErrorReply` that refuses the request. The gateway answers the network function:

| Reply | HTTP response | Audit event |
|-------|---------------|-------------|
| The result | 200, 201 or 204, as the operation defines | for state changes |
| `ErrorReply` `{"problemDetails": {...}}` | `problemDetails.status` (400–599), `problemDetails` as the `application/problem+json` body | none |
| None within the gateway's request timeout, or no application subscribed | 504, cause `TARGET_NF_NOT_REACHABLE` | none |
| Not a JSON object, a `problemDetails` without a 4xx/5xx `status`, or a JetStream acknowledgement | 500, cause `SYSTEM_FAILURE` | none |

Use the application errors TS 29.503 defines for the service (`USER_NOT_FOUND` with 404, for
example) or the common ones of TS 29.500 Table 5.2.7.2-1.

### Application → gateway

| Purpose | Subject | Message |
|---------|---------|---------|
| Deliver a notification to a network function's callback URI | `5g.data.notify` (core publish) | `NotificationPayload` |
| Ask which SMSF serves a subscriber | `sbi.outbound.nudm.uecm.request` (core request/reply) | `NudmUecmMessage` → `NudmUecmMessage` |
| Send an SMS through the serving SMSF (Nsmsf_SMService SendSMS) | `sbi.outbound.nsmsf.send.request` (JetStream work queue) | `SbiOtaMessage` |
| Result of each SMS the gateway processed | `ota.delivery.event` | `DeliveryEvent` |

`sbi.outbound.nsmsf.send.request` is the one subject that is meant to be in a stream: your
application creates the stream `OTA_OUTBOUND` on it, and the gateway creates its durable consumer
`sbi-gateway-smsf`. The gateway asks for redelivery (up to three deliveries) when the SMSF fails
transiently, and reports `PENDING` in the `DeliveryEvent` each time.

## Usage

A UDM back end answering generate-auth-data:

```go
nc, err := nats.Connect(natsURL)
if err != nil { /* ... */ }

_, err = nc.QueueSubscribe(sbiclient.SubjectAuthRequest, "udm", func(m *nats.Msg) {
	var req sbiclient.AuthRequest
	if err := json.Unmarshal(m.Data, &req); err != nil {
		data, _ := json.Marshal(sbiclient.ErrorReply{ProblemDetails: sbiclient.ProblemDetails{
			Status: 400, Cause: "INVALID_MSG_FORMAT", Detail: err.Error(),
		}})
		_ = m.Respond(data) // the AUSF gets a 400 with this ProblemDetails
		return
	}
	res := sbiclient.AuthenticationInfoResult{
		Supi:       resolve(req.Supi), // req.Supi may be a SUCI
		AuthVector: vectorFor(req.Supi, req.ServingNetworkName),
	}
	data, _ := json.Marshal(res)
	_ = m.Respond(data)
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

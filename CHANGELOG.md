# Changelog

All notable changes to `sbi-ha-gateway-client` are documented in this file.
Format loosely follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versioning follows [SemVer](https://semver.org/) (pre-1.0: a minor version may be breaking).

## [0.4.0] — 2026-10-26

### Changed — breaking

- `SbiOtaMessage` is now `SmsMessage`. Its JSON is unchanged.
- `SubjectDeliveryEvent` is `sbi.nsmsf.delivery.event` and `StreamSmsfSend` is `SBI_SMSF_SEND`:
  defaults now, which a gateway 0.5.0 sets otherwise with `SMS_DELIVERY_EVENT_SUBJECT` and
  `SMS_STREAM`. An application that kept the earlier names sets the gateway's variables to them.
  Gateway 0.5.0 uses the stream that captures `sbi.outbound.nsmsf.send.request` whatever its name,
  and creates one only when none does.

### Added

- `example/`: a runnable back end (its own module) that answers the gateway's requests on NATS: a
  starting point, not a UDM.
- `SbiRouting.SmsfInstanceId` (`smsfInstanceId`, optional): the SMSF's NF instance ID, as the SMSF
  registered it with the UDM (TS 29.503 SmsfRegistration). A gateway that registers an SMS_GMSC
  profile finds the SMSF with NRF discovery and sends the SMS there, `SmsfFqdn` unused; without
  one, the SMS goes to `SmsfFqdn`. The `DeliveryEvent`'s `BearerMeta` then carries
  `smsfInstanceId`.
- `NudmUecmMessage.SmsfInstanceId` (`smsfInstanceId`, optional): the serving-SMSF lookup's reply
  may name the SMSF by instance ID, with or instead of `SmsfFqdn`.

## [0.3.0] — 2026-10-19

### Added

- `SubjectDataNotifyFailed` (`5g.data.notify.failed`) and `NotificationFailure`. Gateway 0.3.0
  publishes one for each `NotificationPayload` it did not deliver: the payload, `result` (one of
  the `Notify*` constants), the last HTTP `status` and its `problemDetails`, `cause` and the
  number of `attempts`. Gateway 0.2.x logged such a notification and dropped it.

### Documentation

- `SubjectSmsfSendRequest` names MtForwardSm, which the gateway calls since v0.2.0, not SendSMS.
- `NotificationPayload.AlternateCallbackUris`: a network function that keeps failing, or answered
  `Retry-After`, is not called until it has rested or the time has passed.

- With gateway 0.3.0, a transient failure on the last of the three deliveries of an
  SMS message is reported `FAILED`, cause `RETRIES_EXHAUSTED`, the transient cause in
  `BearerMeta` `lastCause`. Before, the last event was `PENDING` and nothing followed it.

## [0.2.0] — 2026-10-19

### Changed — breaking

- **The contract carries the specification's bodies verbatim.** One request type, `Request`, on
  every gateway → application subject: `service`, `operation` (the operationId of the API's OpenAPI
  definition), `method`, `path`, `params` (the path variables by the specification's names), `query`,
  `contentType` and `body`, the network function's body unchanged. One reply type, `Reply`: `status`
  (one of the operation's success statuses; 0 for its only one), `resourceId` (for a 201 that creates
  a resource under the request URI) and `body`, returned to the network function unchanged.
  `ErrorReply` is unchanged.
- Removed: `AuthRequest`, `AuthenticationInfoRequest`, `AuthenticationInfoResult`,
  `AuthenticationVector`, `ResynchronizationInfo`, `AuthConfirmRequest`, `AuthConfirmReply`,
  `AuthEvent`, `MobilityRequest`, `MobilityReply`, `MobilityQuery`, `Amf3GppAccessRegistration`,
  `SmfRegistration`, `SmsfRegistration`, `LocationInfoRequest`, `ProvideLocInfoRequest`,
  `ProvideLocInfoResult`, `UserLocation`, `EutraLocation`, `NrLocation`, `N3gaLocation`, `Ecgi`,
  `Ncgi`, `Tai`, `Guami`, `PlmnId`, `Snssai`, `DataRequest`, `DataResponse`, and the constants
  `OpRegister`, `OpUpdate`, `OpDeregister`, `NfSMF` and `NfSMSF`. They re-modelled TS 29.503 by
  hand and did not match it: the authentication result went to the AUSF as `authVector`/`kAusf`
  without `authType` or `avType` where TS 29.503 Tables 6.3.6.2.3-1 and 6.3.6.2.5-1 define
  `authenticationVector`/`kausf` and make both mandatory; information elements they did not model were
  dropped; elements the specification defines as objects were strings. Decode bodies with the API's
  OpenAPI data types instead.
- **An MT SMS on `sbi.outbound.nsmsf.send.request` is an RP-DATA.** The gateway forwards it with
  Nsmsf_SMService MtForwardSm (`send-mt-sms`, `multipart/related`), which TS 29.540 Table 5.2.1-1
  gives the SMS-GMSC; v0.1.0's gateway used UplinkSMS (`sendsms`), the AMF's, with a JSON body.
  - `SS7PayloadView` must carry `Encoding` `EncodingRPData` and the RP-DATA towards the UE
    (TS 24.011 clause 7.3.1.1): the Service Centre address travels in it, since MtForwardSm's
    JSON part has none.
  - A bare TPDU fails with cause `PAYLOAD_NOT_RP_DATA`.
  - `Header.Operation` is `MtForwardSm`; `SendSMS` is still accepted.
  - The `DeliveryEvent` takes its status from the mobile station's report (RP-ACK: delivered;
    RP-ERROR: failed, cause `RP_CAUSE_<n>`) and carries the report in `BearerMeta`
    `deliveryReport`.
- `Service*` constants for every API the gateway serves: `ServiceNudmUEAU`, `ServiceNudmUECM`,
  `ServiceNudmMT` and `ServiceNudmRSDS` added.

## [0.1.0] — 2026-10-18

### Added

- First release of the gateway's NATS wire contract as its own module, package `sbiclient`: the
  subject constants and the JSON envelopes of every NATS interface of sbi-ha-gateway — the Nudm
  (UEAU, UECM, MT) and Nudr/SDM/PP/EE request/reply subjects, the data-change notification, and the
  serving-SMSF lookup, SMS send and delivery-event subjects.
- The starting point is the JSON gateway 1.0.0 sent and read: the request envelopes that the gateway
  built as untyped maps are now types (`AuthRequest`, `AuthConfirmRequest`, `MobilityRequest`,
  `MobilityQuery`, `LocationInfoRequest`) with the same field names; pass-through bodies are
  `json.RawMessage`. The items below change it, before this first release.
- `TestWireSchema`: a golden file pinning every JSON field name and type.
- `AuthRequest` embeds `AuthenticationInfoRequest`, the AUSF's whole request body:
  `resynchronizationInfo` (`rand`, `auts`) now reaches the back end, with `supportedFeatures`,
  `cellCagInfo`, `n5gcInd`, `nswoInd` and `disasterRoamingInd`.
- `DataRequest` tells the calls apart: `service` (`nudr-dr`, `nudm-sdm`, `nudm-pp`, `nudm-ee`),
  `resource` always the path after the subscriber, and `query`, the URI's query string. Gateway 1.0.0
  sent PP, EE and SDM-subscription calls with empty `dataSet` and `resource`, and no query.
- Nudm_UECM as TS 29.503 defines it: `MobilityQuery` carries `ueId` (a SUPI or a GPSI; it was
  `supi`), `resource` (the registration asked for, or all) and `query`; `MobilityReply.updated` tells
  the gateway a PUT replaced a registration (200) rather than created one (201 with `Location`).
  `OpDeregister` is no longer sent for the AMF, nor `OpUpdate` for the SMSF: the specification has
  neither.
- `AuthConfirmReply.authEventId` is required: the gateway answers the AUSF 201 with the event's URI.
- `DataRequest.ueId` replaces `supi`: it may be a SUPI, a GPSI, an External Group Id or `anyUE`.
- `ErrorReply` (`{"problemDetails": ProblemDetails}`, TS 29.571) refuses a request on any request
  subject: the network function gets its status and body, and nothing is audited. It replaces
  `DataResponse.Status` `"ERROR"` and `Code`, which could only produce a 400, and gives
  `5g.auth.confirm` and `5g.mobility.registration` an error reply they did not have.
- Every reply is a JSON object, `{}` where the gateway needs no result. No reply within the
  gateway's request timeout, or no subscriber, is answered 504 `TARGET_NF_NOT_REACHABLE`; a reply
  that is not a JSON object, or a JetStream publish acknowledgement, is answered 500.

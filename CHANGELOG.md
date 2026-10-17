# Changelog

All notable changes to `sbi-ha-gateway-client` are documented in this file.
Format loosely follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versioning follows [SemVer](https://semver.org/) (pre-1.0: a minor version may be breaking).

## [Unreleased]

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

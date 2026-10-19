// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

// Package sbiclient is the NATS wire contract of sbi-ha-gateway: the subjects and the
// JSON message schema the gateway uses between a 5G core's service-based interface
// (Nudm, Nudr, Nsmsf; 3GPP TS 29.503, TS 29.504, TS 29.540) and the applications behind it.
//
// This module is licensed under the Apache License 2.0, separately from the gateway
// server (AGPL-3.0-only), so a UDM/UDR back end, an SMS platform or any other
// application can link it without taking on the server's licence. The dependency runs
// one way: the server imports this module for the wire contract, this module never
// imports the server.
//
// # Wire model
//
// All bodies are JSON.
//
//   - Gateway → application, core NATS request/reply. When a network function calls one of
//     the operations the gateway serves, the gateway sends a [Request] on a 5g.* subject — the
//     operation in the terms of its OpenAPI definition, with the network function's body
//     unchanged — and builds the HTTP response from the reply: a [Reply], whose body reaches
//     the network function unchanged, or an [ErrorReply] carrying the ProblemDetails the
//     network function gets. See [SubjectAuthRequest] and the others. Answer with a core
//     subscription, and keep these subjects out of JetStream streams: a stream that captures
//     a subject answers every request with its own publish acknowledgement before the
//     application can, and the gateway answers the network function 500.
//   - Application → gateway, core NATS publish: [SubjectDataNotify] asks the gateway to
//     deliver a notification to a network function's callback URI.
//   - Application → gateway, core NATS request/reply: [SubjectUecmRequest] asks for the
//     SMSF serving a subscriber.
//   - Application → gateway, JetStream work queue: [SubjectSmsfSendRequest] asks the
//     gateway to send an SMS through the serving SMSF. The application owns the stream; the
//     gateway answers with a [DeliveryEvent] on [SubjectDeliveryEvent].
//
// # Security
//
// The [SubjectAuthRequest] replies carry 5G authentication vectors (RAND, AUTN, XRES*,
// K_AUSF). Restrict the subject with NATS permissions so that only the gateway can publish
// requests and only the authentication back end can subscribe, run NATS with TLS, and never
// log these payloads.
package sbiclient

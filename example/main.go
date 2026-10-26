// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

// Command example is a back end for sbi-ha-gateway: it answers the gateway's requests on NATS
// so that the gateway can be tried end to end. It is a starting point, not a UDM: it keeps the
// AMF registrations it is sent in memory, answers subscriber data with a fixed body, and
// refuses everything else with 501.
//
//	go run . -nats nats://127.0.0.1:4222
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	sbiclient "github.com/3eAI-Labs/sbi-ha-gateway-client"
	"github.com/nats-io/nats.go"
)

// queue is the queue group: run several instances and each request is answered once.
const queue = "example-backend"

func main() {
	natsURL := flag.String("nats", nats.DefaultURL, "the NATS server the gateway uses")
	flag.Parse()

	nc, err := nats.Connect(*natsURL, nats.Name("sbi-ha-gateway example back end"))
	if err != nil {
		log.Fatalf("connect %s: %v", *natsURL, err)
	}
	defer nc.Close()

	b := &backend{registrations: map[string]json.RawMessage{}}
	for _, subject := range []string{
		sbiclient.SubjectAuthRequest, sbiclient.SubjectAuthConfirm,
		sbiclient.SubjectMobilityRegistration, sbiclient.SubjectMobilityQuery,
		sbiclient.SubjectLocationInfo, sbiclient.SubjectDataQuery, sbiclient.SubjectDataUpdate,
	} {
		if _, err := nc.QueueSubscribe(subject, queue, b.handle); err != nil {
			log.Fatalf("subscribe %s: %v", subject, err)
		}
	}
	log.Printf("answering the gateway's requests on %s", *natsURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	_ = nc.Drain()
}

type backend struct {
	mu            sync.Mutex
	registrations map[string]json.RawMessage // AMF registrations (Amf3GppAccessRegistration), by UE
}

func (b *backend) handle(m *nats.Msg) {
	var req sbiclient.Request
	if err := json.Unmarshal(m.Data, &req); err != nil {
		respond(m, problem(http.StatusBadRequest, "INVALID_MSG_FORMAT", err.Error()))
		return
	}
	// The operation alone: the path and the body carry subscriber identifiers, which do not
	// belong in a log.
	log.Printf("%s %s", req.Service, req.Operation)

	switch req.Operation {
	case "3GppRegistration": // PUT /nudm-uecm/v1/{ueId}/registrations/amf-3gpp-access
		b.mu.Lock()
		_, replaced := b.registrations[req.Params["ueId"]]
		b.registrations[req.Params["ueId"]] = req.Body
		b.mu.Unlock()
		status := http.StatusCreated
		if replaced {
			status = http.StatusOK
		}
		respond(m, sbiclient.Reply{Status: status, Body: req.Body})

	case "Get3GppRegistration": // GET /nudm-uecm/v1/{ueId}/registrations/amf-3gpp-access
		b.mu.Lock()
		body, ok := b.registrations[req.Params["ueId"]]
		b.mu.Unlock()
		if !ok {
			respond(m, problem(http.StatusNotFound, "CONTEXT_NOT_FOUND", "no AMF registration"))
			return
		}
		respond(m, sbiclient.Reply{Body: body})

	case "GetAmData": // GET /nudm-sdm/v2/{supi}/am-data: AccessAndMobilitySubscriptionData
		respond(m, sbiclient.Reply{Body: json.RawMessage(`{
			"subscribedUeAmbr": {"uplink": "1 Gbps", "downlink": "2 Gbps"},
			"nssai": {"defaultSingleNssais": [{"sst": 1}]}
		}`)})

	default:
		respond(m, problem(http.StatusNotImplemented, "", "the example back end does not implement "+req.Operation))
	}
}

// problem is an error reply: the network function gets status and the ProblemDetails.
func problem(status int, cause, detail string) sbiclient.ErrorReply {
	return sbiclient.ErrorReply{ProblemDetails: sbiclient.ProblemDetails{Status: status, Cause: cause, Detail: detail}}
}

func respond(m *nats.Msg, v any) {
	data, err := json.Marshal(v)
	if err == nil {
		err = m.Respond(data)
	}
	if err != nil {
		log.Printf("reply: %v", err)
	}
}

// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import "encoding/json"

// --- Nudm_UEAU -------------------------------------------------------------------------

// AuthRequest is the request on SubjectAuthRequest: the subscriber from the request URI
// and the AUSF's request body, whose members appear at the top level.
type AuthRequest struct {
	// Supi is the SUPI or SUCI from the request URI, as the AUSF sent it.
	Supi string `json:"supi"`
	AuthenticationInfoRequest
}

// AuthenticationInfoRequest is the body of generate-auth-data (TS 29.503 Table
// 6.3.6.2.2-1), as the AUSF sent it.
type AuthenticationInfoRequest struct {
	ServingNetworkName string `json:"servingNetworkName"`
	AusfInstanceId     string `json:"ausfInstanceId"`
	// ResynchronizationInfo is present when the UE rejected the last vector for a
	// sequence number out of range: generate a new vector after re-synchronising the
	// sequence number from AUTS (TS 33.501 clause 6.1.3.3.2).
	ResynchronizationInfo *ResynchronizationInfo `json:"resynchronizationInfo,omitempty"`
	SupportedFeatures     string                 `json:"supportedFeatures,omitempty"`
	// CellCagInfo is the CAG list of the cell; absent, the UE is not accessing from a CAG
	// cell.
	CellCagInfo        []string `json:"cellCagInfo,omitempty"`
	N5gcInd            bool     `json:"n5gcInd,omitempty"`
	NswoInd            bool     `json:"nswoInd,omitempty"`
	DisasterRoamingInd bool     `json:"disasterRoamingInd,omitempty"`
}

// ResynchronizationInfo is RAND and AUTS, hex-encoded. AUTS is derived from the
// subscriber's key: never log it.
type ResynchronizationInfo struct {
	Rand string `json:"rand"`
	Auts string `json:"auts"`
}

// AuthenticationInfoResult is the reply on SubjectAuthRequest; the gateway returns it to
// the AUSF as the body of a 200 response (TS 29.503 AuthenticationInfoResult).
type AuthenticationInfoResult struct {
	AuthVector        *AuthenticationVector `json:"authVector,omitempty"`
	Supi              string                `json:"supi,omitempty"`
	SupportedFeatures string                `json:"supportedFeatures,omitempty"`
}

// AuthenticationVector is a 5G HE AKA vector, hex-encoded. Key material: never log it.
type AuthenticationVector struct {
	Rand     string `json:"rand"`
	Autn     string `json:"autn"`
	XresStar string `json:"xresStar"`
	KAusf    string `json:"kAusf"`
}

// AuthConfirmRequest is the request on SubjectAuthConfirm.
type AuthConfirmRequest struct {
	Supi  string    `json:"supi"`
	Event AuthEvent `json:"event"`
}

// AuthConfirmReply is the success reply on SubjectAuthConfirm. AuthEventId is required:
// the identifier the UDM assigns to the stored event; the gateway answers the AUSF 201 with
// the event and Location .../{supi}/auth-events/{authEventId} (TS 29.503).
type AuthConfirmReply struct {
	AuthEventId string `json:"authEventId"`
}

// AuthEvent is the AUSF's authentication result (TS 29.503 AuthEvent).
type AuthEvent struct {
	NfInstanceId       string `json:"nfInstanceId"`
	Success            bool   `json:"success"`
	TimeStamp          string `json:"timeStamp"`
	AuthType           string `json:"authType"`
	ServingNetworkName string `json:"servingNetworkName"`
	AuthRemovalInd     bool   `json:"authRemovalInd,omitempty"`
}

// --- Nudm_UECM -------------------------------------------------------------------------

// Registration operations in MobilityRequest.Op.
const (
	OpRegister   = "REGISTER"   // HTTP PUT: Data is the full registration
	OpUpdate     = "UPDATE"     // HTTP PATCH (AMF, SMF): Data is the JSON object the network function sent
	OpDeregister = "DEREGISTER" // HTTP DELETE (SMF, SMSF): no Data
)

// Registering network functions in MobilityRequest.Nf. An AMF registration (3GPP access)
// has no Nf.
const (
	NfSMF  = "SMF"
	NfSMSF = "SMSF"
)

// MobilityRequest is the request on SubjectMobilityRegistration.
//
// Data depends on Nf and Op: for OpRegister an Amf3GppAccessRegistration (no Nf),
// SmfRegistration (NfSMF) or SmsfRegistration (NfSMSF); for OpUpdate the modification
// object the network function sent, unchanged; for OpDeregister nothing.
type MobilityRequest struct {
	Supi string `json:"supi"`
	Op   string `json:"op"`
	Nf   string `json:"nf,omitempty"`
	// PduSessionId is the PDU session ID from the request URI, for NfSMF only.
	PduSessionId string          `json:"pduSessionId,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
}

// MobilityReply is the success reply on SubjectMobilityRegistration. For OpRegister, set
// Updated when the registration replaced an existing one: the gateway answers 200;
// otherwise 201 with the registration's URI in Location (TS 29.503 PUT responses). Other
// operations: reply {}.
type MobilityReply struct {
	Updated bool `json:"updated,omitempty"`
}

// MobilityQuery is the request on SubjectMobilityQuery: a Nudm_UECM GET.
type MobilityQuery struct {
	// UeId is the URI's {ueId}: a SUPI or a GPSI (msisdn-…, extid-…).
	UeId string `json:"ueId"`
	// Resource is the path after .../registrations, without a leading slash: "" for the
	// registration data sets, "amf-3gpp-access", "smsf-3gpp-access", "smf-registrations"
	// or "smf-registrations/{pduSessionId}".
	Resource string `json:"resource,omitempty"`
	// Query is the URI's query string, unchanged, without "?"; GET registrations carries
	// registration-dataset-names in it.
	Query string `json:"query,omitempty"`
}

// Amf3GppAccessRegistration is an AMF registration for 3GPP access (TS 29.503).
type Amf3GppAccessRegistration struct {
	AmfInstanceId     string   `json:"amfInstanceId"`
	SupportedFeatures string   `json:"supportedFeatures,omitempty"`
	PurgeId           string   `json:"purgeId,omitempty"`
	RatType           string   `json:"ratType"`
	DeregCallbackUri  string   `json:"deregCallbackUri"`
	Guami             *Guami   `json:"guami,omitempty"`
	BackupAmfInfo     []string `json:"backupAmfInfo,omitempty"`
	DratType          string   `json:"dratType,omitempty"`
	UeSrvccCapability bool     `json:"ueSrvccCapability,omitempty"`
}

// Guami is a globally unique AMF identifier (TS 29.571).
type Guami struct {
	PlmnId PlmnId `json:"plmnId"`
	AmfId  string `json:"amfId"`
}

// PlmnId is a PLMN identity (TS 29.571).
type PlmnId struct {
	Mcc string `json:"mcc"`
	Mnc string `json:"mnc"`
}

// SmfRegistration is an SMF registration for one PDU session (TS 29.503).
type SmfRegistration struct {
	SmfInstanceId     string  `json:"smfInstanceId"`
	SmfSetId          string  `json:"smfSetId,omitempty"`
	SupportedFeatures string  `json:"supportedFeatures,omitempty"`
	PduSessionId      int     `json:"pduSessionId"`
	SingleNssai       *Snssai `json:"singleNssai"`
	Dnn               string  `json:"dnn,omitempty"`
	EmergencyServices bool    `json:"emergencyServices,omitempty"`
	PcscfRestoration  string  `json:"pcscfRestorationCallbackUri,omitempty"`
	PlmnId            *PlmnId `json:"plmnId,omitempty"`
	PgwFqdn           string  `json:"pgwFqdn,omitempty"`
	PgwIpAddr         string  `json:"pgwIpAddr,omitempty"`
}

// Snssai is a single network slice selection assistance information (TS 29.571).
type Snssai struct {
	Sst int    `json:"sst"`
	Sd  string `json:"sd,omitempty"`
}

// SmsfRegistration is an SMSF registration (TS 29.503).
type SmsfRegistration struct {
	SmsfInstanceId      string   `json:"smsfInstanceId"`
	SmsfSetId           string   `json:"smsfSetId,omitempty"`
	SupportedFeatures   string   `json:"supportedFeatures,omitempty"`
	PlmnId              *PlmnId  `json:"plmnId,omitempty"`
	SmsfMAPAddress      string   `json:"smsfMAPAddress,omitempty"`
	SmsfDiameterAddress string   `json:"smsfDiameterAddress,omitempty"`
	BackupAmfInfo       []string `json:"backupAmfInfo,omitempty"`
}

// --- Nudm_MT ---------------------------------------------------------------------------

// LocationInfoRequest is the request on SubjectLocationInfo.
type LocationInfoRequest struct {
	Supi string                `json:"supi"`
	Req  ProvideLocInfoRequest `json:"req"`
}

// ProvideLocInfoRequest is the body the network function sent (TS 29.503 LocationInfoRequest).
type ProvideLocInfoRequest struct {
	Req5gsLoc         bool   `json:"req5gsLoc,omitempty"`
	ReqCurrentLoc     bool   `json:"reqCurrentLoc,omitempty"`
	ReqRatType        bool   `json:"reqRatType,omitempty"`
	ReqTimeZone       bool   `json:"reqTimeZone,omitempty"`
	SupportedFeatures string `json:"supportedFeatures,omitempty"`
}

// ProvideLocInfoResult is the reply on SubjectLocationInfo; the gateway returns it to the
// network function as the body of a 200 response (TS 29.503 LocationInfoResult).
type ProvideLocInfoResult struct {
	CurrentLoc        bool          `json:"currentLoc,omitempty"`
	Location          *UserLocation `json:"location,omitempty"`
	RatType           string        `json:"ratType,omitempty"`
	Timezone          string        `json:"timezone,omitempty"`
	Supi              string        `json:"supi,omitempty"`
	SupportedFeatures string        `json:"supportedFeatures,omitempty"`
}

// UserLocation is a UE location (TS 29.571).
type UserLocation struct {
	NrLocation    *NrLocation    `json:"nrLocation,omitempty"`
	EutraLocation *EutraLocation `json:"eutraLocation,omitempty"`
	N3gaLocation  *N3gaLocation  `json:"n3gaLocation,omitempty"`
}

// NrLocation is an NR location (TS 29.571).
type NrLocation struct {
	Tai                      *Tai   `json:"tai"`
	Ncgi                     *Ncgi  `json:"ncgi"`
	AgeOfLocationInformation int    `json:"ageOfLocationInformation,omitempty"`
	UeLocationTimestamp      string `json:"ueLocationTimestamp,omitempty"`
}

// EutraLocation is an E-UTRA location (TS 29.571).
type EutraLocation struct {
	Tai                      *Tai   `json:"tai"`
	Ecgi                     *Ecgi  `json:"ecgi"`
	AgeOfLocationInformation int    `json:"ageOfLocationInformation,omitempty"`
	UeLocationTimestamp      string `json:"ueLocationTimestamp,omitempty"`
}

// N3gaLocation is a non-3GPP access location (TS 29.571).
type N3gaLocation struct {
	Tai        *Tai   `json:"tai,omitempty"`
	N3gppTai   *Tai   `json:"n3gppTai,omitempty"`
	N3IwfId    string `json:"n3IwfId,omitempty"`
	UeIpv4Addr string `json:"ueIpv4Addr,omitempty"`
	UeIpv6Addr string `json:"ueIpv6Addr,omitempty"`
	PortNumber int    `json:"portNumber,omitempty"`
}

// Tai is a tracking area identity (TS 29.571).
type Tai struct {
	PlmnId *PlmnId `json:"plmnId"`
	Tac    string  `json:"tac"`
}

// Ncgi is an NR cell global identity (TS 29.571).
type Ncgi struct {
	PlmnId   *PlmnId `json:"plmnId"`
	NrCellId string  `json:"nrCellId"`
}

// Ecgi is an E-UTRA cell global identity (TS 29.571).
type Ecgi struct {
	PlmnId      *PlmnId `json:"plmnId"`
	EutraCellId string  `json:"eutraCellId"`
}

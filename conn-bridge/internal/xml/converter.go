package xml

import (
	"encoding/xml"
	"fmt"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
)

// ========== ENTRY CONVERTERS ==========

// CreateEntryRequestToXML converts gRPC CreateEntryRequest to XML bytes
func CreateEntryRequestToXML(req *pb.CreateEntryRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Key == nil {
		return nil, fmt.Errorf("key cannot be nil")
	}
	if req.Account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	xmlReq := &XMLCreateEntryRequest{
		Entry: XMLEntry{
			Key:     req.Key.KeyValue,
			KeyType: keyTypeToXML(req.Key.KeyType),
			Account: accountToXML(req.Account),
			// Owner will be populated from account holder info in real implementation
		},
		RequestId: req.RequestId,
	}

	return marshalXML(xmlReq)
}

// CreateEntryResponseFromXML converts XML bytes to gRPC CreateEntryResponse
func CreateEntryResponseFromXML(xmlData []byte) (*pb.CreateEntryResponse, error) {
	var xmlResp XMLCreateEntryResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.CreateEntryResponse{
		EntryId:    xmlResp.Entry.Key,
		ExternalId: xmlResp.CorrelationId,
		Status:     commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
	}, nil
}

// UpdateEntryRequestToXML converts gRPC UpdateEntryRequest to XML bytes
func UpdateEntryRequestToXML(req *pb.UpdateEntryRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLUpdateEntryRequest{
		Key:        req.EntryId,
		KeyType:    "CPF", // TODO: Get from existing entry
		NewAccount: accountToXML(req.NewAccount),
		RequestId:  req.RequestId,
	}

	return marshalXML(xmlReq)
}

// UpdateEntryResponseFromXML converts XML bytes to gRPC UpdateEntryResponse
func UpdateEntryResponseFromXML(xmlData []byte) (*pb.UpdateEntryResponse, error) {
	var xmlResp XMLUpdateEntryResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.UpdateEntryResponse{
		EntryId: xmlResp.Entry.Key,
		Account: accountFromXML(&xmlResp.Entry.Account),
	}, nil
}

// DeleteEntryRequestToXML converts gRPC DeleteEntryRequest to XML bytes
func DeleteEntryRequestToXML(req *pb.DeleteEntryRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLDeleteEntryRequest{
		Key:       req.EntryId,
		KeyType:   "CPF", // TODO: Get from existing entry
		RequestId: req.RequestId,
	}

	return marshalXML(xmlReq)
}

// DeleteEntryResponseFromXML converts XML bytes to gRPC DeleteEntryResponse
func DeleteEntryResponseFromXML(xmlData []byte) (*pb.DeleteEntryResponse, error) {
	var xmlResp XMLDeleteEntryResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.DeleteEntryResponse{
		Deleted: xmlResp.Deleted,
		// DeletedAt will be set by the caller
	}, nil
}

// GetEntryRequestToXML converts gRPC GetEntryRequest to XML bytes
func GetEntryRequestToXML(req *pb.GetEntryRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLGetEntryRequest{
		RequestId: req.RequestId,
	}

	// Handle oneof identifier
	switch id := req.Identifier.(type) {
	case *pb.GetEntryRequest_EntryId:
		xmlReq.EntryId = id.EntryId
	case *pb.GetEntryRequest_ExternalId:
		xmlReq.EntryId = id.ExternalId
	case *pb.GetEntryRequest_KeyQuery:
		xmlReq.Key = id.KeyQuery.KeyValue
		xmlReq.KeyType = keyTypeToXML(id.KeyQuery.KeyType)
	default:
		return nil, fmt.Errorf("identifier is required")
	}

	return marshalXML(xmlReq)
}

// GetEntryResponseFromXML converts XML bytes to gRPC GetEntryResponse
func GetEntryResponseFromXML(xmlData []byte) (*pb.GetEntryResponse, error) {
	var xmlResp XMLGetEntryResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.GetEntryResponse{
		EntryId:    xmlResp.Entry.Key,
		ExternalId: xmlResp.CorrelationId,
		Key: &commonv1.DictKey{
			KeyType:  keyTypeFromXML(xmlResp.Entry.KeyType),
			KeyValue: xmlResp.Entry.Key,
		},
		Account: accountFromXML(&xmlResp.Entry.Account),
		Status:  commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
	}, nil
}

// ========== CLAIM CONVERTERS ==========

// CreateClaimRequestToXML converts gRPC CreateClaimRequest to XML bytes
func CreateClaimRequestToXML(req *pb.CreateClaimRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLCreateClaimRequest{
		Claim: XMLClaim{
			Key:     req.KeyValue,
			KeyType: keyTypeToXML(req.KeyType),
			// Type and other fields will be populated in real implementation
		},
	}

	return marshalXML(xmlReq)
}

// CreateClaimResponseFromXML converts XML bytes to gRPC CreateClaimResponse
func CreateClaimResponseFromXML(xmlData []byte) (*pb.CreateClaimResponse, error) {
	var xmlResp XMLCreateClaimResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.CreateClaimResponse{
		ClaimId:    xmlResp.Claim.ClaimId,
		ExternalId: xmlResp.CorrelationId,
		Status:     claimStatusFromXML(xmlResp.Claim.Status),
	}, nil
}

// CancelClaimRequestToXML converts gRPC CancelClaimRequest to XML bytes
func CancelClaimRequestToXML(req *pb.CancelClaimRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLCancelClaimRequest{
		ClaimId: req.ClaimId,
		Reason:  "USER_REQUESTED_CANCEL", // Default
	}

	return marshalXML(xmlReq)
}

// CancelClaimResponseFromXML converts XML bytes to gRPC CancelClaimResponse
func CancelClaimResponseFromXML(xmlData []byte) (*pb.CancelClaimResponse, error) {
	var xmlResp XMLCancelClaimResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.CancelClaimResponse{
		ClaimId: xmlResp.Claim.ClaimId,
		Status:  claimStatusFromXML(xmlResp.Claim.Status),
	}, nil
}

// CompleteClaimRequestToXML converts gRPC CompleteClaimRequest to XML bytes
func CompleteClaimRequestToXML(req *pb.CompleteClaimRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := &XMLCompleteClaimRequest{
		ClaimId: req.ClaimId,
	}

	return marshalXML(xmlReq)
}

// CompleteClaimResponseFromXML converts XML bytes to gRPC CompleteClaimResponse
func CompleteClaimResponseFromXML(xmlData []byte) (*pb.CompleteClaimResponse, error) {
	var xmlResp XMLCompleteClaimResponse
	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.CompleteClaimResponse{
		ClaimId: xmlResp.Claim.ClaimId,
		Status:  claimStatusFromXML(xmlResp.Claim.Status),
	}, nil
}

// GetClaimRequestToXML converts gRPC GetClaimRequest to XML bytes
func GetClaimRequestToXML(req *pb.GetClaimRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Determine ClaimId from identifier
	var claimId string
	switch id := req.Identifier.(type) {
	case *pb.GetClaimRequest_ClaimId:
		claimId = id.ClaimId
	case *pb.GetClaimRequest_ExternalId:
		claimId = id.ExternalId
	default:
		return nil, fmt.Errorf("identifier is required")
	}

	// Construct XML GET request (simplified - Bacen may use URL params instead)
	xmlReq := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<GetClaimRequest>
  <ClaimId>%s</ClaimId>
  <RequestId>%s</RequestId>
</GetClaimRequest>`, claimId, req.RequestId)

	return []byte(xmlReq), nil
}

// GetClaimResponseFromXML converts XML bytes to gRPC GetClaimResponse
func GetClaimResponseFromXML(xmlData []byte) (*pb.GetClaimResponse, error) {
	// For GET operations, Bacen may return just the Claim object
	var xmlResp struct {
		XMLName xml.Name `xml:"GetClaimResponse"`
		Claim   XMLClaim `xml:"Claim"`
	}

	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.GetClaimResponse{
		ClaimId:              xmlResp.Claim.ClaimId,
		ExternalId:           xmlResp.Claim.ClaimId, // May need mapping
		EntryId:              "entry-id-from-xml",    // Extract from XML if available
		Status:               claimStatusFromXML(xmlResp.Claim.Status),
		CompletionPeriodDays: 30,
		ClaimerIspb:          xmlResp.Claim.ClaimerAccount.Participant,
		OwnerIspb:            xmlResp.Claim.DonorParticipant,
		Found:                true,
	}, nil
}

// ========== PORTABILITY CONVERTERS ==========

// InitiatePortabilityRequestToXML converts gRPC InitiatePortabilityRequest to XML bytes
func InitiatePortabilityRequestToXML(req *pb.InitiatePortabilityRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Key == nil {
		return nil, fmt.Errorf("key cannot be nil")
	}
	if req.NewAccount == nil {
		return nil, fmt.Errorf("new_account cannot be nil")
	}

	// Build XML for portability initiation
	xmlReq := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<InitiatePortabilityRequest>
  <EntryId>%s</EntryId>
  <Key>
    <Type>%s</Type>
    <Value>%s</Value>
  </Key>
  <NewAccount>
    <Participant>%s</Participant>
    <Branch>%s</Branch>
    <AccountNumber>%s</AccountNumber>
    <AccountType>%s</AccountType>
  </NewAccount>
  <IdempotencyKey>%s</IdempotencyKey>
  <RequestId>%s</RequestId>
</InitiatePortabilityRequest>`,
		req.EntryId,
		keyTypeToXML(req.Key.KeyType),
		req.Key.KeyValue,
		req.NewAccount.Ispb,
		req.NewAccount.BranchCode,
		req.NewAccount.AccountNumber,
		accountTypeToXML(req.NewAccount.AccountType),
		req.IdempotencyKey,
		req.RequestId)

	return []byte(xmlReq), nil
}

// InitiatePortabilityResponseFromXML converts XML bytes to gRPC InitiatePortabilityResponse
func InitiatePortabilityResponseFromXML(xmlData []byte) (*pb.InitiatePortabilityResponse, error) {
	var xmlResp struct {
		XMLName           xml.Name `xml:"InitiatePortabilityResponse"`
		PortabilityId     string   `xml:"PortabilityId"`
		EntryId           string   `xml:"EntryId"`
		Status            string   `xml:"Status"`
		ResponseTime      string   `xml:"ResponseTime"`
		CorrelationId     string   `xml:"CorrelationId"`
	}

	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.InitiatePortabilityResponse{
		PortabilityId:      xmlResp.PortabilityId,
		EntryId:            xmlResp.EntryId,
		Status:             commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING,
		BacenTransactionId: xmlResp.CorrelationId,
	}, nil
}

// ConfirmPortabilityRequestToXML converts gRPC ConfirmPortabilityRequest to XML bytes
func ConfirmPortabilityRequestToXML(req *pb.ConfirmPortabilityRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.NewAccount == nil {
		return nil, fmt.Errorf("new_account cannot be nil")
	}

	xmlReq := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ConfirmPortabilityRequest>
  <EntryId>%s</EntryId>
  <PortabilityId>%s</PortabilityId>
  <NewAccount>
    <Participant>%s</Participant>
    <Branch>%s</Branch>
    <AccountNumber>%s</AccountNumber>
    <AccountType>%s</AccountType>
  </NewAccount>
  <IdempotencyKey>%s</IdempotencyKey>
  <RequestId>%s</RequestId>
</ConfirmPortabilityRequest>`,
		req.EntryId,
		req.PortabilityId,
		req.NewAccount.Ispb,
		req.NewAccount.BranchCode,
		req.NewAccount.AccountNumber,
		accountTypeToXML(req.NewAccount.AccountType),
		req.IdempotencyKey,
		req.RequestId)

	return []byte(xmlReq), nil
}

// ConfirmPortabilityResponseFromXML converts XML bytes to gRPC ConfirmPortabilityResponse
func ConfirmPortabilityResponseFromXML(xmlData []byte) (*pb.ConfirmPortabilityResponse, error) {
	var xmlResp struct {
		XMLName           xml.Name   `xml:"ConfirmPortabilityResponse"`
		EntryId           string     `xml:"EntryId"`
		PortabilityId     string     `xml:"PortabilityId"`
		Status            string     `xml:"Status"`
		Account           XMLAccount `xml:"Account"`
		ResponseTime      string     `xml:"ResponseTime"`
		CorrelationId     string     `xml:"CorrelationId"`
	}

	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.ConfirmPortabilityResponse{
		EntryId:            xmlResp.EntryId,
		PortabilityId:      xmlResp.PortabilityId,
		Status:             commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED,
		Account:            accountFromXML(&xmlResp.Account),
		BacenTransactionId: xmlResp.CorrelationId,
	}, nil
}

// CancelPortabilityRequestToXML converts gRPC CancelPortabilityRequest to XML bytes
func CancelPortabilityRequestToXML(req *pb.CancelPortabilityRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	xmlReq := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<CancelPortabilityRequest>
  <EntryId>%s</EntryId>
  <PortabilityId>%s</PortabilityId>
  <Reason>%s</Reason>
  <IdempotencyKey>%s</IdempotencyKey>
  <RequestId>%s</RequestId>
</CancelPortabilityRequest>`,
		req.EntryId,
		req.PortabilityId,
		req.Reason,
		req.IdempotencyKey,
		req.RequestId)

	return []byte(xmlReq), nil
}

// CancelPortabilityResponseFromXML converts XML bytes to gRPC CancelPortabilityResponse
func CancelPortabilityResponseFromXML(xmlData []byte) (*pb.CancelPortabilityResponse, error) {
	var xmlResp struct {
		XMLName           xml.Name `xml:"CancelPortabilityResponse"`
		EntryId           string   `xml:"EntryId"`
		PortabilityId     string   `xml:"PortabilityId"`
		Status            string   `xml:"Status"`
		ResponseTime      string   `xml:"ResponseTime"`
		CorrelationId     string   `xml:"CorrelationId"`
	}

	if err := xml.Unmarshal(xmlData, &xmlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return &pb.CancelPortabilityResponse{
		EntryId:            xmlResp.EntryId,
		PortabilityId:      xmlResp.PortabilityId,
		Status:             commonv1.EntryStatus_ENTRY_STATUS_ACTIVE, // Reverts to ACTIVE
		BacenTransactionId: xmlResp.CorrelationId,
	}, nil
}

// ========== HELPER FUNCTIONS ==========

// marshalXML marshals any XML struct to bytes with header
func marshalXML(v interface{}) ([]byte, error) {
	xmlData, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}

	xmlHeader := []byte(xml.Header)
	return append(xmlHeader, xmlData...), nil
}

// accountToXML converts gRPC Account to XML Account
func accountToXML(acc *commonv1.Account) XMLAccount {
	if acc == nil {
		return XMLAccount{}
	}

	return XMLAccount{
		Participant:   acc.Ispb,
		Branch:        acc.BranchCode,
		AccountNumber: acc.AccountNumber,
		AccountType:   accountTypeToXML(acc.AccountType),
	}
}

// accountFromXML converts XML Account to gRPC Account
func accountFromXML(acc *XMLAccount) *commonv1.Account {
	if acc == nil {
		return nil
	}

	return &commonv1.Account{
		Ispb:          acc.Participant,
		BranchCode:    acc.Branch,
		AccountNumber: acc.AccountNumber,
		AccountType:   accountTypeFromXML(acc.AccountType),
	}
}

// ========== ENUM CONVERTERS ==========

// keyTypeToXML converts gRPC KeyType to XML string
func keyTypeToXML(kt commonv1.KeyType) string {
	switch kt {
	case commonv1.KeyType_KEY_TYPE_CPF:
		return "CPF"
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		return "CNPJ"
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		return "EMAIL"
	case commonv1.KeyType_KEY_TYPE_PHONE:
		return "PHONE"
	case commonv1.KeyType_KEY_TYPE_EVP:
		return "EVP"
	default:
		return ""
	}
}

// keyTypeFromXML converts XML string to gRPC KeyType
func keyTypeFromXML(s string) commonv1.KeyType {
	switch s {
	case "CPF":
		return commonv1.KeyType_KEY_TYPE_CPF
	case "CNPJ":
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case "EMAIL":
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case "PHONE":
		return commonv1.KeyType_KEY_TYPE_PHONE
	case "EVP":
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

// accountTypeToXML converts gRPC AccountType to XML string
func accountTypeToXML(at commonv1.AccountType) string {
	switch at {
	case commonv1.AccountType_ACCOUNT_TYPE_CHECKING:
		return "CHECKING"
	case commonv1.AccountType_ACCOUNT_TYPE_SAVINGS:
		return "SAVINGS"
	case commonv1.AccountType_ACCOUNT_TYPE_PAYMENT:
		return "PAYMENT"
	default:
		return ""
	}
}

// accountTypeFromXML converts XML string to gRPC AccountType
func accountTypeFromXML(s string) commonv1.AccountType {
	switch s {
	case "CHECKING":
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case "SAVINGS":
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case "PAYMENT":
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

// claimStatusFromXML converts XML string to gRPC ClaimStatus
func claimStatusFromXML(s string) commonv1.ClaimStatus {
	switch s {
	case "OPEN":
		return commonv1.ClaimStatus_CLAIM_STATUS_OPEN
	case "WAITING_RESOLUTION":
		return commonv1.ClaimStatus_CLAIM_STATUS_WAITING_RESOLUTION
	case "CONFIRMED":
		return commonv1.ClaimStatus_CLAIM_STATUS_CONFIRMED
	case "CANCELLED":
		return commonv1.ClaimStatus_CLAIM_STATUS_CANCELLED
	case "COMPLETED":
		return commonv1.ClaimStatus_CLAIM_STATUS_COMPLETED
	default:
		return commonv1.ClaimStatus_CLAIM_STATUS_UNSPECIFIED
	}
}
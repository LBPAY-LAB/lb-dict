package xml

import "encoding/xml"

// Namespace BACEN para DICT
const BacenDICTNamespace = "https://www.bcb.gov.br/pi/pacs.002/1.8"

// ========== SHARED STRUCTURES ==========

// XMLAccount representa a conta transacional
type XMLAccount struct {
	Participant   string `xml:"Participant"`
	Branch        string `xml:"Branch"`
	AccountNumber string `xml:"AccountNumber"`
	AccountType   string `xml:"AccountType"`
	OpeningDate   string `xml:"OpeningDate,omitempty"`
}

// XMLOwner representa o dono da chave
type XMLOwner struct {
	Type        string `xml:"Type"`
	TaxIdNumber string `xml:"TaxIdNumber"`
	Name        string `xml:"Name"`
	TradeName   string `xml:"TradeName,omitempty"` // Nome fantasia (apenas PJ)
}

// ========== CREATE ENTRY ==========

// XMLCreateEntryRequest representa a requisição para criar chave DICT
type XMLCreateEntryRequest struct {
	XMLName   xml.Name `xml:"CreateEntryRequest"`
	Signature string   `xml:"Signature,omitempty"`
	Entry     XMLEntry `xml:"Entry"`
	Reason    string   `xml:"Reason,omitempty"`
	RequestId string   `xml:"RequestId"`
}

// XMLEntry representa o vínculo chave → conta → owner
type XMLEntry struct {
	Key     string     `xml:"Key"`
	KeyType string     `xml:"KeyType"`
	Account XMLAccount `xml:"Account"`
	Owner   XMLOwner   `xml:"Owner"`
}

// XMLCreateEntryResponse representa a resposta do BACEN
type XMLCreateEntryResponse struct {
	XMLName       xml.Name         `xml:"CreateEntryResponse"`
	Signature     string           `xml:"Signature,omitempty"`
	ResponseTime  string           `xml:"ResponseTime"`
	CorrelationId string           `xml:"CorrelationId"`
	Entry         XMLExtendedEntry `xml:"Entry"`
}

// XMLExtendedEntry representa o Entry com campos adicionais do BACEN
type XMLExtendedEntry struct {
	Key              string     `xml:"Key"`
	KeyType          string     `xml:"KeyType"`
	Account          XMLAccount `xml:"Account"`
	Owner            XMLOwner   `xml:"Owner"`
	CreationTime     string     `xml:"CreationTime"`
	KeyOwnershipDate string     `xml:"KeyOwnershipDate"`
	LastModifiedDate string     `xml:"LastModifiedDate,omitempty"`
}

// ========== UPDATE ENTRY ==========

// XMLUpdateEntryRequest representa a requisição para atualizar chave DICT
type XMLUpdateEntryRequest struct {
	XMLName    xml.Name   `xml:"UpdateEntryRequest"`
	Signature  string     `xml:"Signature,omitempty"`
	Key        string     `xml:"Key"`
	KeyType    string     `xml:"KeyType"`
	NewAccount XMLAccount `xml:"NewAccount"`
	Reason     string     `xml:"Reason,omitempty"`
	RequestId  string     `xml:"RequestId"`
}

// XMLUpdateEntryResponse representa a resposta do BACEN
type XMLUpdateEntryResponse struct {
	XMLName       xml.Name         `xml:"UpdateEntryResponse"`
	Signature     string           `xml:"Signature,omitempty"`
	ResponseTime  string           `xml:"ResponseTime"`
	CorrelationId string           `xml:"CorrelationId"`
	Entry         XMLExtendedEntry `xml:"Entry"`
}

// ========== DELETE ENTRY ==========

// XMLDeleteEntryRequest representa a requisição para deletar chave DICT
type XMLDeleteEntryRequest struct {
	XMLName   xml.Name `xml:"DeleteEntryRequest"`
	Signature string   `xml:"Signature,omitempty"`
	Key       string   `xml:"Key"`
	KeyType   string   `xml:"KeyType"`
	Reason    string   `xml:"Reason,omitempty"`
	RequestId string   `xml:"RequestId"`
}

// XMLDeleteEntryResponse representa a resposta do BACEN
type XMLDeleteEntryResponse struct {
	XMLName       xml.Name `xml:"DeleteEntryResponse"`
	Signature     string   `xml:"Signature,omitempty"`
	ResponseTime  string   `xml:"ResponseTime"`
	CorrelationId string   `xml:"CorrelationId"`
	Deleted       bool     `xml:"Deleted"`
	Key           string   `xml:"Key"`
	KeyType       string   `xml:"KeyType"`
}

// ========== GET ENTRY ==========

// XMLGetEntryRequest representa a requisição para buscar chave DICT
type XMLGetEntryRequest struct {
	XMLName   xml.Name `xml:"GetEntryRequest"`
	Key       string   `xml:"Key,omitempty"`
	KeyType   string   `xml:"KeyType,omitempty"`
	EntryId   string   `xml:"EntryId,omitempty"`
	RequestId string   `xml:"RequestId"`
}

// XMLGetEntryResponse representa a resposta do BACEN
type XMLGetEntryResponse struct {
	XMLName       xml.Name         `xml:"GetEntryResponse"`
	Signature     string           `xml:"Signature,omitempty"`
	ResponseTime  string           `xml:"ResponseTime"`
	CorrelationId string           `xml:"CorrelationId"`
	Entry         XMLExtendedEntry `xml:"Entry"`
}

// ========== CLAIM STRUCTURES ==========

// XMLClaim represents a portability or ownership claim
type XMLClaim struct {
	ClaimId             string     `xml:"ClaimId"`
	Type                string     `xml:"Type"` // PORTABILITY, OWNERSHIP
	Key                 string     `xml:"Key"`
	KeyType             string     `xml:"KeyType"`
	Status              string     `xml:"Status"` // OPEN, WAITING_RESOLUTION, CONFIRMED, CANCELLED, COMPLETE
	DonorParticipant    string     `xml:"DonorParticipant"`
	ClaimerAccount      XMLAccount `xml:"ClaimerAccount"`
	Claimer             XMLOwner   `xml:"Claimer"`
	CompletionPeriodEnd string     `xml:"CompletionPeriodEnd"` // ISO 8601
	ResolutionPeriodEnd string     `xml:"ResolutionPeriodEnd"` // ISO 8601
	LastModified        string     `xml:"LastModified"`        // ISO 8601
	CreationTime        string     `xml:"CreationTime"`        // ISO 8601
}

// XMLCreateClaimRequest represents POST /claims/ request
type XMLCreateClaimRequest struct {
	XMLName   xml.Name `xml:"CreateClaimRequest"`
	Signature string   `xml:"Signature,omitempty"`
	Claim     XMLClaim `xml:"Claim"`
}

// XMLCreateClaimResponse represents POST /claims/ response
type XMLCreateClaimResponse struct {
	XMLName       xml.Name `xml:"CreateClaimResponse"`
	Signature     string   `xml:"Signature,omitempty"`
	ResponseTime  string   `xml:"ResponseTime"`
	CorrelationId string   `xml:"CorrelationId"`
	Claim         XMLClaim `xml:"Claim"`
}

// XMLConfirmClaimRequest represents POST /claims/{ClaimId}/confirm request
type XMLConfirmClaimRequest struct {
	XMLName   xml.Name `xml:"ConfirmClaimRequest"`
	Signature string   `xml:"Signature,omitempty"`
	ClaimId   string   `xml:"ClaimId"`
}

// XMLConfirmClaimResponse represents POST /claims/{ClaimId}/confirm response
type XMLConfirmClaimResponse struct {
	XMLName       xml.Name `xml:"ConfirmClaimResponse"`
	Signature     string   `xml:"Signature,omitempty"`
	ResponseTime  string   `xml:"ResponseTime"`
	CorrelationId string   `xml:"CorrelationId"`
	Claim         XMLClaim `xml:"Claim"`
}

// XMLCancelClaimRequest represents POST /claims/{ClaimId}/cancel request
type XMLCancelClaimRequest struct {
	XMLName   xml.Name `xml:"CancelClaimRequest"`
	Signature string   `xml:"Signature,omitempty"`
	ClaimId   string   `xml:"ClaimId"`
	Reason    string   `xml:"Reason"`
}

// XMLCancelClaimResponse represents POST /claims/{ClaimId}/cancel response
type XMLCancelClaimResponse struct {
	XMLName       xml.Name `xml:"CancelClaimResponse"`
	Signature     string   `xml:"Signature,omitempty"`
	ResponseTime  string   `xml:"ResponseTime"`
	CorrelationId string   `xml:"CorrelationId"`
	Claim         XMLClaim `xml:"Claim"`
}

// XMLCompleteClaimRequest represents POST /claims/{ClaimId}/complete request
type XMLCompleteClaimRequest struct {
	XMLName   xml.Name `xml:"CompleteClaimRequest"`
	Signature string   `xml:"Signature,omitempty"`
	ClaimId   string   `xml:"ClaimId"`
}

// XMLCompleteClaimResponse represents POST /claims/{ClaimId}/complete response
type XMLCompleteClaimResponse struct {
	XMLName       xml.Name `xml:"CompleteClaimResponse"`
	Signature     string   `xml:"Signature,omitempty"`
	ResponseTime  string   `xml:"ResponseTime"`
	CorrelationId string   `xml:"CorrelationId"`
	Claim         XMLClaim `xml:"Claim"`
}

// ========== INFRACTION STRUCTURES ==========

// XMLContactInformation representa as informações de contato
type XMLContactInformation struct {
	Email string `xml:"Email"`
	Phone string `xml:"Phone"`
}

// XMLCreateInfractionReportRequest representa o request do manual (seção 4.2)
type XMLCreateInfractionReportRequest struct {
	XMLName          xml.Name             `xml:"CreateInfractionReportRequest"`
	Signature        string               `xml:"Signature,omitempty"`
	Participant      string               `xml:"Participant"`
	InfractionReport XMLInfractionReport  `xml:"InfractionReport"`
}

// XMLInfractionReport representa o InfractionReport do manual
type XMLInfractionReport struct {
	TransactionId      string                  `xml:"TransactionId"`
	Reason             string                  `xml:"Reason"`
	SituationType      string                  `xml:"SituationType"`
	ReportDetails      string                  `xml:"ReportDetails,omitempty"`
	ContactInformation *XMLContactInformation  `xml:"ContactInformation"`
}

// XMLCreateInfractionReportResponse representa a resposta do manual (seção 4.3)
type XMLCreateInfractionReportResponse struct {
	XMLName          xml.Name                `xml:"CreateInfractionReportResponse"`
	Signature        string                  `xml:"Signature,omitempty"`
	ResponseTime     string                  `xml:"ResponseTime"`
	CorrelationId    string                  `xml:"CorrelationId"`
	InfractionReport XMLInfractionReportFull `xml:"InfractionReport"`
}

// XMLInfractionReportFull representa o InfractionReport completo da resposta
type XMLInfractionReportFull struct {
	TransactionId      string                 `xml:"TransactionId"`
	Reason             string                 `xml:"Reason"`
	SituationType      string                 `xml:"SituationType"`
	ReportDetails      string                 `xml:"ReportDetails"`
	ContactInformation *XMLContactInformation `xml:"ContactInformation"`
	Id                 string                 `xml:"Id"`
	Status             string                 `xml:"Status"`
	CreationTime       string                 `xml:"CreationTime"`
	LastModified       string                 `xml:"LastModified"`
}
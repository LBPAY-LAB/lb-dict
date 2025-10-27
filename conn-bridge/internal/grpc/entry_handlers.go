package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-bridge/internal/xml"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Bacen DICT API endpoints
	endpointCreateEntry = "/api/v1/dict/entries"
	endpointGetEntry    = "/api/v1/dict/entries"
	endpointUpdateEntry = "/api/v1/dict/entries"
	endpointDeleteEntry = "/api/v1/dict/entries"
)

// CreateEntry handles the CreateEntry RPC call
// Flow: gRPC Request → XML → Sign XML → SOAP Envelope → mTLS POST → Parse Response → gRPC Response
func (s *Server) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"requestId": req.RequestId,
		"keyType":   req.Key.KeyType,
		"keyValue":  maskKey(req.Key.KeyValue),
	}).Info("CreateEntry called")

	// Step 1: Validate request
	if err := s.validateCreateEntryRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// Step 2: Convert gRPC request to XML
	xmlData, err := xml.CreateEntryRequestToXML(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert request to XML")
		return nil, status.Errorf(codes.Internal, "failed to convert request to XML: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Converted gRPC request to XML")

	// Step 3: Sign XML with ICP-Brasil A3 (via Java XML Signer service)
	signedXML, err := s.xmlSigner.SignXML(ctx, string(xmlData))
	if err != nil {
		s.logger.WithError(err).Error("Failed to sign XML")
		return nil, status.Errorf(codes.Internal, "failed to sign XML: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"signedXMLSize": len(signedXML),
	}).Debug("XML signed successfully with ICP-Brasil A3")

	// Step 4: Build SOAP envelope
	soapEnvelope, err := s.soapClient.BuildSOAPEnvelope(signedXML, "")
	if err != nil {
		s.logger.WithError(err).Error("Failed to build SOAP envelope")
		return nil, status.Errorf(codes.Internal, "failed to build SOAP envelope: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"soapEnvelopeSize": len(soapEnvelope),
	}).Debug("Built SOAP envelope")

	// Step 5: Send SOAP request to Bacen via HTTPS + mTLS
	soapResponse, err := s.soapClient.SendSOAPRequest(ctx, endpointCreateEntry, soapEnvelope)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send SOAP request to Bacen")
		return nil, status.Errorf(codes.Unavailable, "failed to send request to Bacen: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"responseSize": len(soapResponse),
	}).Debug("Received SOAP response from Bacen")

	// Step 6: Parse SOAP response
	bodyXML, err := s.soapClient.ParseSOAPResponse(soapResponse)
	if err != nil {
		s.logger.WithError(err).Error("Failed to parse SOAP response")
		return nil, status.Errorf(codes.Internal, "failed to parse SOAP response: %v", err)
	}

	// Step 7: Convert XML response to gRPC response
	response, err := xml.CreateEntryResponseFromXML(bodyXML)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert XML response to gRPC")
		return nil, status.Errorf(codes.Internal, "failed to convert response: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"entryId":    response.EntryId,
		"externalId": response.ExternalId,
		"status":     response.Status,
	}).Info("CreateEntry completed successfully")

	return response, nil
}

// GetEntry handles the GetEntry RPC call
// Flow: gRPC Request → XML → Sign XML → SOAP Envelope → mTLS GET → Parse Response → gRPC Response
func (s *Server) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"requestId": req.RequestId,
	}).Info("GetEntry called")

	// Step 1: Validate request
	if req.GetEntryId() == "" && req.GetExternalId() == "" && req.GetKeyQuery() == nil {
		return nil, status.Error(codes.InvalidArgument, "one of entry_id, external_id, or key_query is required")
	}

	// Step 2: Convert gRPC request to XML
	xmlData, err := xml.GetEntryRequestToXML(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert request to XML")
		return nil, status.Errorf(codes.Internal, "failed to convert request to XML: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Converted gRPC request to XML")

	// Step 3: Sign XML with ICP-Brasil A3
	signedXML, err := s.xmlSigner.SignXML(ctx, string(xmlData))
	if err != nil {
		s.logger.WithError(err).Error("Failed to sign XML")
		return nil, status.Errorf(codes.Internal, "failed to sign XML: %v", err)
	}

	s.logger.Debug("XML signed successfully")

	// Step 4: Build SOAP envelope
	soapEnvelope, err := s.soapClient.BuildSOAPEnvelope(signedXML, "")
	if err != nil {
		s.logger.WithError(err).Error("Failed to build SOAP envelope")
		return nil, status.Errorf(codes.Internal, "failed to build SOAP envelope: %v", err)
	}

	// Step 5: Send SOAP request to Bacen
	soapResponse, err := s.soapClient.SendSOAPRequest(ctx, endpointGetEntry, soapEnvelope)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send SOAP request to Bacen")
		return nil, status.Errorf(codes.Unavailable, "failed to query Bacen: %v", err)
	}

	s.logger.Debug("Received SOAP response from Bacen")

	// Step 6: Parse SOAP response
	bodyXML, err := s.soapClient.ParseSOAPResponse(soapResponse)
	if err != nil {
		s.logger.WithError(err).Error("Failed to parse SOAP response")
		return nil, status.Errorf(codes.Internal, "failed to parse SOAP response: %v", err)
	}

	// Step 7: Convert XML response to gRPC response
	response, err := xml.GetEntryResponseFromXML(bodyXML)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert XML response to gRPC")
		return nil, status.Errorf(codes.Internal, "failed to convert response: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"entryId":    response.EntryId,
		"externalId": response.ExternalId,
		"keyValue":   maskKey(response.Key.KeyValue),
	}).Info("GetEntry completed successfully")

	return response, nil
}

// UpdateEntry handles the UpdateEntry RPC call
// Flow: gRPC Request → XML → Sign XML → SOAP Envelope → mTLS PUT → Parse Response → gRPC Response
func (s *Server) UpdateEntry(ctx context.Context, req *pb.UpdateEntryRequest) (*pb.UpdateEntryResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"requestId": req.RequestId,
		"entryId":   req.EntryId,
	}).Info("UpdateEntry called")

	// Step 1: Validate request
	if req.EntryId == "" {
		return nil, status.Error(codes.InvalidArgument, "entry_id is required")
	}
	if req.NewAccount == nil {
		return nil, status.Error(codes.InvalidArgument, "new_account is required")
	}

	// Step 2: Convert gRPC request to XML
	xmlData, err := xml.UpdateEntryRequestToXML(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert request to XML")
		return nil, status.Errorf(codes.Internal, "failed to convert request to XML: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Converted gRPC request to XML")

	// Step 3: Sign XML with ICP-Brasil A3
	signedXML, err := s.xmlSigner.SignXML(ctx, string(xmlData))
	if err != nil {
		s.logger.WithError(err).Error("Failed to sign XML")
		return nil, status.Errorf(codes.Internal, "failed to sign XML: %v", err)
	}

	s.logger.Debug("XML signed successfully")

	// Step 4: Build SOAP envelope
	soapEnvelope, err := s.soapClient.BuildSOAPEnvelope(signedXML, "")
	if err != nil {
		s.logger.WithError(err).Error("Failed to build SOAP envelope")
		return nil, status.Errorf(codes.Internal, "failed to build SOAP envelope: %v", err)
	}

	// Step 5: Send SOAP request to Bacen
	soapResponse, err := s.soapClient.SendSOAPRequest(ctx, endpointUpdateEntry, soapEnvelope)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send SOAP request to Bacen")
		return nil, status.Errorf(codes.Unavailable, "failed to update entry in Bacen: %v", err)
	}

	s.logger.Debug("Received SOAP response from Bacen")

	// Step 6: Parse SOAP response
	bodyXML, err := s.soapClient.ParseSOAPResponse(soapResponse)
	if err != nil {
		s.logger.WithError(err).Error("Failed to parse SOAP response")
		return nil, status.Errorf(codes.Internal, "failed to parse SOAP response: %v", err)
	}

	// Step 7: Convert XML response to gRPC response
	response, err := xml.UpdateEntryResponseFromXML(bodyXML)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert XML response to gRPC")
		return nil, status.Errorf(codes.Internal, "failed to convert response: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"entryId":            response.EntryId,
		"bacenTransactionId": response.BacenTransactionId,
	}).Info("UpdateEntry completed successfully")

	return response, nil
}

// DeleteEntry handles the DeleteEntry RPC call
// Flow: gRPC Request → XML → Sign XML → SOAP Envelope → mTLS DELETE → Parse Response → gRPC Response
func (s *Server) DeleteEntry(ctx context.Context, req *pb.DeleteEntryRequest) (*pb.DeleteEntryResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"requestId": req.RequestId,
		"entryId":   req.EntryId,
	}).Info("DeleteEntry called")

	// Step 1: Validate request
	if req.EntryId == "" {
		return nil, status.Error(codes.InvalidArgument, "entry_id is required")
	}

	// Step 2: Convert gRPC request to XML
	xmlData, err := xml.DeleteEntryRequestToXML(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert request to XML")
		return nil, status.Errorf(codes.Internal, "failed to convert request to XML: %v", err)
	}

	s.logger.WithFields(logrus.Fields{
		"xmlSize": len(xmlData),
	}).Debug("Converted gRPC request to XML")

	// Step 3: Sign XML with ICP-Brasil A3
	signedXML, err := s.xmlSigner.SignXML(ctx, string(xmlData))
	if err != nil {
		s.logger.WithError(err).Error("Failed to sign XML")
		return nil, status.Errorf(codes.Internal, "failed to sign XML: %v", err)
	}

	s.logger.Debug("XML signed successfully")

	// Step 4: Build SOAP envelope
	soapEnvelope, err := s.soapClient.BuildSOAPEnvelope(signedXML, "")
	if err != nil {
		s.logger.WithError(err).Error("Failed to build SOAP envelope")
		return nil, status.Errorf(codes.Internal, "failed to build SOAP envelope: %v", err)
	}

	// Step 5: Send SOAP request to Bacen
	soapResponse, err := s.soapClient.SendSOAPRequest(ctx, endpointDeleteEntry, soapEnvelope)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send SOAP request to Bacen")
		return nil, status.Errorf(codes.Unavailable, "failed to delete entry in Bacen: %v", err)
	}

	s.logger.Debug("Received SOAP response from Bacen")

	// Step 6: Parse SOAP response
	bodyXML, err := s.soapClient.ParseSOAPResponse(soapResponse)
	if err != nil {
		s.logger.WithError(err).Error("Failed to parse SOAP response")
		return nil, status.Errorf(codes.Internal, "failed to parse SOAP response: %v", err)
	}

	// Step 7: Convert XML response to gRPC response
	response, err := xml.DeleteEntryResponseFromXML(bodyXML)
	if err != nil {
		s.logger.WithError(err).Error("Failed to convert XML response to gRPC")
		return nil, status.Errorf(codes.Internal, "failed to convert response: %v", err)
	}

	// Set deletion timestamp
	response.DeletedAt = timestamppb.New(time.Now())

	s.logger.WithFields(logrus.Fields{
		"entryId":            req.EntryId,
		"deleted":            response.Deleted,
		"bacenTransactionId": response.BacenTransactionId,
	}).Info("DeleteEntry completed successfully")

	return response, nil
}

// validateCreateEntryRequest validates the CreateEntry request
func (s *Server) validateCreateEntryRequest(req *pb.CreateEntryRequest) error {
	if req.Key == nil {
		return fmt.Errorf("key is required")
	}
	if req.Key.KeyType == commonv1.KeyType_KEY_TYPE_UNSPECIFIED {
		return fmt.Errorf("key type is required")
	}
	if req.Key.KeyValue == "" {
		return fmt.Errorf("key value is required")
	}
	if req.Account == nil {
		return fmt.Errorf("account is required")
	}
	if req.Account.Ispb == "" {
		return fmt.Errorf("ISPB is required")
	}
	if req.Account.AccountNumber == "" {
		return fmt.Errorf("account number is required")
	}
	if req.RequestId == "" {
		return fmt.Errorf("request_id is required")
	}

	return nil
}

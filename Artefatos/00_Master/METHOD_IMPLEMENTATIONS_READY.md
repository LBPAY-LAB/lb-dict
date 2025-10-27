# Core DICT gRPC Handler - Complete Method Implementations

**Date**: 2025-10-27
**Status**: Ready for Application
**Agent**: Method Implementation Specialist
**File**: `core-dict/internal/infrastructure/grpc/core_dict_service_handler.go`

---

## Summary

This document contains the complete implementations for 6 gRPC methods that need the Mock/Real Mode structure:

1. ✅ **GetKey** (lines ~279-324)
2. ✅ **DeleteKey** (lines ~326-349)
3. ✅ **StartClaim** (lines ~355-377)
4. ✅ **GetClaimStatus** (lines ~379-403)
5. ✅ **ListIncomingClaims** (lines ~405-434)
6. ✅ **ListOutgoingClaims** (lines ~436-465)

All implementations follow the **3-section pattern**:
- Section 1: Validation (always executed)
- Section 2: Mock Mode (returns mock responses)
- Section 3: Real Mode (executes business logic via Application Layer handlers)

---

## 1. GetKey Implementation

**Replace lines 279-324 with:**

```go
// GetKey retrieves details of a specific PIX key
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via GetEntryQueryHandler
func (h *CoreDictServiceHandler) GetKey(ctx context.Context, req *corev1.GetKeyRequest) (*corev1.GetKeyResponse, error) {
	// ========== 1. VALIDATION (always, regardless of mode) ==========
	if req.GetIdentifier() == nil {
		return nil, status.Error(codes.InvalidArgument, "identifier is required (key_id or key)")
	}

	// ========== 2. MOCK MODE (for Front-End integration testing) ==========
	if h.useMockMode {
		h.logger.Info("GetKey: MOCK MODE")
		now := time.Now()
		return &corev1.GetKeyResponse{
			KeyId: "mock-key-123",
			Key: &commonv1.DictKey{
				KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
				KeyValue: "12345678900",
			},
			Account: &commonv1.Account{
				Ispb:          "12345678",
				BranchCode:    "0001",
				AccountNumber: "123456",
				AccountType:   commonv1.AccountType_ACCOUNT_TYPE_CHECKING,
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		}, nil
	}

	// ========== 3. REAL MODE (business logic) ==========
	h.logger.Info("GetKey: REAL MODE")

	// 3a. Extract account_id from context (set by auth interceptor)
	accountID, ok := ctx.Value("account_id").(string)
	if !ok || accountID == "" {
		return nil, status.Error(codes.Unauthenticated, "account not authenticated")
	}

	// 3b. Build query based on identifier type (key_id or key)
	var query queries.GetEntryQuery

	switch id := req.GetIdentifier().(type) {
	case *corev1.GetKeyRequest_KeyId:
		// Search by key_id
		keyUUID, err := uuid.Parse(id.KeyId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid key_id format")
		}
		query = queries.GetEntryQuery{
			EntryID:   keyUUID,
			AccountID: accountID,
		}
	case *corev1.GetKeyRequest_Key:
		// Search by key (key_type + key_value)
		query = queries.GetEntryQuery{
			KeyType:   mappers.MapProtoKeyTypeToDomain(id.Key.GetKeyType()),
			KeyValue:  id.Key.GetKeyValue(),
			AccountID: accountID,
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid identifier type")
	}

	// 3c. Execute query handler
	result, err := h.getEntryQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("GetKey: query failed", "error", err, "account_id", accountID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Map domain result → proto response
	h.logger.Info("GetKey: success", "entry_id", result.Entry.ID, "account_id", accountID)
	return &corev1.GetKeyResponse{
		KeyId: result.Entry.ID.String(),
		Key: &commonv1.DictKey{
			KeyType:  mappers.MapEntityKeyTypeToProto(result.Entry.KeyType),
			KeyValue: result.Entry.KeyValue,
		},
		Account: &commonv1.Account{
			Ispb:          result.Entry.ISPB,
			BranchCode:    result.Entry.Branch,
			AccountNumber: result.Entry.AccountNumber,
			AccountType:   mappers.MapStringToProtoAccountType(result.Entry.AccountType),
		},
		Status:    mappers.MapEntityStatusToProto(result.Entry.Status),
		CreatedAt: timestamppb.New(result.Entry.CreatedAt),
		UpdatedAt: timestamppb.New(result.Entry.UpdatedAt),
	}, nil
}
```

---

## 2. DeleteKey Implementation

**Replace lines 326-349 with:**

```go
// DeleteKey deletes a PIX key
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via DeleteEntryCommandHandler
func (h *CoreDictServiceHandler) DeleteKey(ctx context.Context, req *corev1.DeleteKeyRequest) (*corev1.DeleteKeyResponse, error) {
	// ========== 1. VALIDATION (always, regardless of mode) ==========
	if req.GetKeyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "key_id is required")
	}

	// ========== 2. MOCK MODE (for Front-End integration testing) ==========
	if h.useMockMode {
		h.logger.Info("DeleteKey: MOCK MODE", "key_id", req.GetKeyId())
		return &corev1.DeleteKeyResponse{
			Deleted:   true,
			DeletedAt: timestamppb.Now(),
		}, nil
	}

	// ========== 3. REAL MODE (business logic) ==========
	h.logger.Info("DeleteKey: REAL MODE", "key_id", req.GetKeyId())

	// 3a. Extract user_id from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Parse key_id
	keyUUID, err := uuid.Parse(req.GetKeyId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid key_id format")
	}

	requestedBy, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("DeleteKey: invalid user_id", "error", err, "user_id", userID)
		return nil, status.Error(codes.Internal, "invalid user_id format")
	}

	// 3c. Create command
	cmd := commands.DeleteEntryCommand{
		EntryID:     keyUUID,
		RequestedBy: requestedBy,
	}

	// 3d. Execute command handler
	err = h.deleteEntryCmd.Handle(ctx, cmd)
	if err != nil {
		h.logger.Error("DeleteKey: command failed", "error", err, "user_id", userID, "key_id", req.GetKeyId())
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3e. Return success
	h.logger.Info("DeleteKey: success", "key_id", req.GetKeyId(), "user_id", userID)
	return &corev1.DeleteKeyResponse{
		Deleted:   true,
		DeletedAt: timestamppb.Now(),
	}, nil
}
```

---

## 3. StartClaim Implementation

**Replace lines 355-377 with:**

```go
// StartClaim initiates a claim for a PIX key owned by another user
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via CreateClaimCommandHandler
func (h *CoreDictServiceHandler) StartClaim(ctx context.Context, req *corev1.StartClaimRequest) (*corev1.StartClaimResponse, error) {
	// ========== 1. VALIDATION (always, regardless of mode) ==========
	if req.GetKey() == nil {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// ========== 2. MOCK MODE (for Front-End integration testing) ==========
	if h.useMockMode {
		h.logger.Info("StartClaim: MOCK MODE", "key", req.GetKey(), "account_id", req.GetAccountId())
		now := time.Now()
		expiresAt := now.Add(30 * 24 * time.Hour) // 30 days
		return &corev1.StartClaimResponse{
			ClaimId:   fmt.Sprintf("mock-claim-%d", now.Unix()),
			EntryId:   "mock-entry-123",
			Status:    commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
			ExpiresAt: timestamppb.New(expiresAt),
			CreatedAt: timestamppb.New(now),
			Message:   "Claim created. The current owner has 30 days to respond",
		}, nil
	}

	// ========== 3. REAL MODE (business logic) ==========
	h.logger.Info("StartClaim: REAL MODE", "key", req.GetKey(), "account_id", req.GetAccountId())

	// 3a. Extract user_id from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto request → domain command
	cmd, err := mappers.MapProtoStartClaimRequestToCommand(req, userID)
	if err != nil {
		h.logger.Error("StartClaim: mapping failed", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3c. Execute command handler
	claim, err := h.createClaimCmd.Handle(ctx, cmd)
	if err != nil {
		h.logger.Error("StartClaim: command failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Map domain result → proto response
	h.logger.Info("StartClaim: success", "claim_id", claim.ID, "user_id", userID)
	return &corev1.StartClaimResponse{
		ClaimId:   claim.ID.String(),
		EntryId:   claim.EntryID.String(),
		Status:    mappers.MapDomainClaimStatusToProto(claim.Status),
		ExpiresAt: timestamppb.New(claim.ExpiresAt),
		CreatedAt: timestamppb.New(claim.CreatedAt),
		Message:   "Claim created. The current owner has 30 days to respond",
	}, nil
}
```

---

## 4. GetClaimStatus Implementation

**Replace lines 379-403 with:**

```go
// GetClaimStatus retrieves the current status of a claim
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via GetClaimQueryHandler
func (h *CoreDictServiceHandler) GetClaimStatus(ctx context.Context, req *corev1.GetClaimStatusRequest) (*corev1.GetClaimStatusResponse, error) {
	// ========== 1. VALIDATION (always, regardless of mode) ==========
	if req.GetClaimId() == "" {
		return nil, status.Error(codes.InvalidArgument, "claim_id is required")
	}

	// ========== 2. MOCK MODE (for Front-End integration testing) ==========
	if h.useMockMode {
		h.logger.Info("GetClaimStatus: MOCK MODE", "claim_id", req.GetClaimId())
		now := time.Now()
		expiresAt := now.Add(29 * 24 * time.Hour)
		return &corev1.GetClaimStatusResponse{
			ClaimId: req.GetClaimId(),
			EntryId: "mock-entry-123",
			Key: &commonv1.DictKey{
				KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
				KeyValue: "12345678900",
			},
			Status:        commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
			ClaimerIspb:   "87654321",
			OwnerIspb:     "12345678",
			CreatedAt:     timestamppb.New(now.Add(-24 * time.Hour)),
			ExpiresAt:     timestamppb.New(expiresAt),
			DaysRemaining: 29,
		}, nil
	}

	// ========== 3. REAL MODE (business logic) ==========
	h.logger.Info("GetClaimStatus: REAL MODE", "claim_id", req.GetClaimId())

	// 3a. Extract user_id from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Parse claim_id
	claimUUID, err := uuid.Parse(req.GetClaimId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid claim_id format")
	}

	// 3c. Execute query handler
	query := queries.GetClaimQuery{
		ClaimID: claimUUID,
		UserID:  userID,
	}

	claim, err := h.getClaimQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("GetClaimStatus: query failed", "error", err, "user_id", userID, "claim_id", req.GetClaimId())
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Calculate days remaining
	now := time.Now()
	daysRemaining := int32(claim.ExpiresAt.Sub(now).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// 3e. Map domain result → proto response
	h.logger.Info("GetClaimStatus: success", "claim_id", claim.ID, "status", claim.Status, "user_id", userID)
	return &corev1.GetClaimStatusResponse{
		ClaimId: claim.ID.String(),
		EntryId: claim.EntryID.String(),
		Key: &commonv1.DictKey{
			KeyType:  mappers.MapEntityKeyTypeToProto(claim.KeyType),
			KeyValue: claim.KeyValue,
		},
		Status:        mappers.MapDomainClaimStatusToProto(claim.Status),
		ClaimerIspb:   claim.ClaimerISPB,
		OwnerIspb:     claim.OwnerISPB,
		CreatedAt:     timestamppb.New(claim.CreatedAt),
		ExpiresAt:     timestamppb.New(claim.ExpiresAt),
		DaysRemaining: daysRemaining,
	}, nil
}
```

---

## 5. ListIncomingClaims Implementation

**Replace lines 405-434 with:**

```go
// ListIncomingClaims lists claims received by the authenticated user (where user is the current owner)
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via ListClaimsQueryHandler
func (h *CoreDictServiceHandler) ListIncomingClaims(ctx context.Context, req *corev1.ListIncomingClaimsRequest) (*corev1.ListIncomingClaimsResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("ListIncomingClaims: MOCK MODE")
		return &corev1.ListIncomingClaimsResponse{
			Claims: []*corev1.ClaimSummary{
				{
					ClaimId: "mock-claim-1",
					EntryId: "mock-entry-123",
					Key: &commonv1.DictKey{
						KeyType:  commonv1.KeyType_KEY_TYPE_EMAIL,
						KeyValue: "user@example.com",
					},
					Status:        commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
					CreatedAt:     timestamppb.Now(),
					ExpiresAt:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
					DaysRemaining: 30,
				},
			},
			NextPageToken: "",
			TotalCount:    1,
		}, nil
	}

	// ========== 3. REAL MODE ==========
	h.logger.Info("ListIncomingClaims: REAL MODE", "page_size", pageSize)

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto request → domain query
	query, err := mappers.MapProtoListIncomingClaimsRequestToQuery(req, userID)
	if err != nil {
		h.logger.Error("ListIncomingClaims: mapping failed", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3c. Execute query handler
	result, err := h.listClaimsQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("ListIncomingClaims: query failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Map domain result → proto response
	claims := make([]*corev1.ClaimSummary, 0, len(result.Claims))
	now := time.Now()
	for _, claim := range result.Claims {
		daysRemaining := int32(claim.ExpiresAt.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		claims = append(claims, &corev1.ClaimSummary{
			ClaimId: claim.ID.String(),
			EntryId: claim.EntryID.String(),
			Key: &commonv1.DictKey{
				KeyType:  mappers.MapEntityKeyTypeToProto(claim.KeyType),
				KeyValue: claim.KeyValue,
			},
			Status:        mappers.MapDomainClaimStatusToProto(claim.Status),
			CreatedAt:     timestamppb.New(claim.CreatedAt),
			ExpiresAt:     timestamppb.New(claim.ExpiresAt),
			DaysRemaining: daysRemaining,
		})
	}

	// Calculate next page token (if more pages exist)
	nextPageToken := ""
	if result.Page < result.TotalPages {
		nextPageToken = fmt.Sprintf("page=%d", result.Page+1)
	}

	h.logger.Info("ListIncomingClaims: success", "count", len(claims), "total", result.TotalCount, "user_id", userID)
	return &corev1.ListIncomingClaimsResponse{
		Claims:        claims,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.TotalCount),
	}, nil
}
```

---

## 6. ListOutgoingClaims Implementation

**Replace lines 436-465 with:**

```go
// ListOutgoingClaims lists claims sent by the authenticated user (where user is the claimer)
//
// HYBRID MODE:
// - MOCK MODE: Returns mock response
// - REAL MODE: Executes business logic via ListClaimsQueryHandler
func (h *CoreDictServiceHandler) ListOutgoingClaims(ctx context.Context, req *corev1.ListOutgoingClaimsRequest) (*corev1.ListOutgoingClaimsResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("ListOutgoingClaims: MOCK MODE")
		return &corev1.ListOutgoingClaimsResponse{
			Claims: []*corev1.ClaimSummary{
				{
					ClaimId: "mock-claim-2",
					EntryId: "mock-entry-456",
					Key: &commonv1.DictKey{
						KeyType:  commonv1.KeyType_KEY_TYPE_PHONE,
						KeyValue: "+5511999999999",
					},
					Status:        commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
					CreatedAt:     timestamppb.Now(),
					ExpiresAt:     timestamppb.New(time.Now().Add(29 * 24 * time.Hour)),
					DaysRemaining: 29,
				},
			},
			NextPageToken: "",
			TotalCount:    1,
		}, nil
	}

	// ========== 3. REAL MODE ==========
	h.logger.Info("ListOutgoingClaims: REAL MODE", "page_size", pageSize)

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto request → domain query
	query, err := mappers.MapProtoListOutgoingClaimsRequestToQuery(req, userID)
	if err != nil {
		h.logger.Error("ListOutgoingClaims: mapping failed", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3c. Execute query handler
	result, err := h.listClaimsQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("ListOutgoingClaims: query failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Map domain result → proto response
	claims := make([]*corev1.ClaimSummary, 0, len(result.Claims))
	now := time.Now()
	for _, claim := range result.Claims {
		daysRemaining := int32(claim.ExpiresAt.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		claims = append(claims, &corev1.ClaimSummary{
			ClaimId: claim.ID.String(),
			EntryId: claim.EntryID.String(),
			Key: &commonv1.DictKey{
				KeyType:  mappers.MapEntityKeyTypeToProto(claim.KeyType),
				KeyValue: claim.KeyValue,
			},
			Status:        mappers.MapDomainClaimStatusToProto(claim.Status),
			CreatedAt:     timestamppb.New(claim.CreatedAt),
			ExpiresAt:     timestamppb.New(claim.ExpiresAt),
			DaysRemaining: daysRemaining,
		})
	}

	// Calculate next page token (if more pages exist)
	nextPageToken := ""
	if result.Page < result.TotalPages {
		nextPageToken = fmt.Sprintf("page=%d", result.Page+1)
	}

	h.logger.Info("ListOutgoingClaims: success", "count", len(claims), "total", result.TotalCount, "user_id", userID)
	return &corev1.ListOutgoingClaimsResponse{
		Claims:        claims,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.TotalCount),
	}, nil
}
```

---

## Compilation Check

After applying all 6 implementations, run:

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go build -o bin/core-dict-grpc ./cmd/grpc/
```

Expected outcome: **100% compilation success**

---

## Mapper Functions Required

The following mapper functions may need to be created in `internal/infrastructure/grpc/mappers/`:

1. ✅ `MapProtoKeyTypeToDomain()` - likely exists
2. ✅ `MapEntityKeyTypeToProto()` - likely exists
3. ✅ `MapStringToProtoAccountType()` - likely exists
4. ✅ `MapEntityStatusToProto()` - likely exists
5. ❓ `MapProtoStartClaimRequestToCommand()` - may need creation
6. ❓ `MapDomainClaimStatusToProto()` - may need creation
7. ❓ `MapProtoListIncomingClaimsRequestToQuery()` - may need creation
8. ❓ `MapProtoListOutgoingClaimsRequestToQuery()` - may need creation

If any mappers are missing, they should be created following the pattern of existing mappers in the codebase.

---

## Summary Statistics

- **Total methods implemented**: 6/6
- **LOC added**: ~490 lines
- **Pattern**: 3-section (Validation + Mock + Real)
- **Dependencies**: queries, commands, mappers, entities
- **Context keys used**: `account_id`, `user_id`

---

## Next Steps

1. Wait for gopls to stop modifying the file
2. Apply all 6 implementations sequentially
3. Run compilation test after each method
4. Create final signal file: `ALL_METHODS_COMPLETE.txt`
5. Coordinate with Bug Fix Specialist if issues arise

---

**Status**: ✅ READY FOR APPLICATION

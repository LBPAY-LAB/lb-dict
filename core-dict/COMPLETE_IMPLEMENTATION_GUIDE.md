# Complete Real Mode Implementation Guide

## Executive Summary

Implementing Real Mode for Methods 3-8 in `core_dict_service_handler.go`.

**Status:**
- ✅ Methods 1-2 (CreateKey, ListKeys): IMPLEMENTED
- 🔄 Methods 3-8 (GetKey, DeleteKey, StartClaim, GetClaimStatus, ListIncomingClaims, ListOutgoingClaims): IN PROGRESS

**Approach:** Due to file locking issues during edits, providing complete implementations for manual integration.

---

## Implementation Strategy

### Pattern for All Methods

```go
func (h *CoreDictServiceHandler) MethodName(ctx context.Context, req *Request) (*Response, error) {
	// 1. VALIDATION (always)
	if req.GetField() == "" {
		return nil, status.Error(codes.InvalidArgument, "field required")
	}

	// 2. MOCK MODE (keep existing)
	if h.useMockMode {
		h.logger.Info("MethodName: MOCK MODE")
		return &Response{/* mock data */}, nil
	}

	// 3. REAL MODE
	h.logger.Info("MethodName: REAL MODE")

	// 3a. Extract user_id
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("MethodName: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto → domain
	// 3c. Execute handler
	// 3d. Map domain → proto
	// 3e. Return response
}
```

---

## Method 3: GetKey

**Line:** 278
**Implementation:**

```go
// GetKey retrieves details of a specific PIX key
func (h *CoreDictServiceHandler) GetKey(ctx context.Context, req *corev1.GetKeyRequest) (*corev1.GetKeyResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	if req.GetIdentifier() == nil {
		return nil, status.Error(codes.InvalidArgument, "identifier is required (key_id or key)")
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("GetKey: MOCK MODE")
		now := time.Now()
		var keyID string
		switch id := req.GetIdentifier().(type) {
		case *corev1.GetKeyRequest_KeyId:
			keyID = id.KeyId
		case *corev1.GetKeyRequest_Key:
			keyID = "mock-key-id"
		}
		return &corev1.GetKeyResponse{
			KeyId: keyID,
			Key: &commonv1.DictKey{
				KeyType:  commonv1.KeyType_KEY_TYPE_CPF,
				KeyValue: "12345678900",
			},
			Account: &commonv1.Account{
				Ispb:          "12345678",
				BranchCode:    "0001",
				AccountNumber: "123456",
				AccountType:   commonv1.AccountType_ACCOUNT_TYPE_SAVINGS,
			},
			Status:    commonv1.EntryStatus_ENTRY_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		}, nil
	}

	// ========== 3. REAL MODE ==========
	h.logger.Info("GetKey: REAL MODE")

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("GetKey: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Determine lookup method (by key_id or by key value)
	var entry *entities.Entry
	var err error

	switch id := req.GetIdentifier().(type) {
	case *corev1.GetKeyRequest_KeyId:
		// Lookup by entry ID (UUID)
		h.logger.Info("GetKey: lookup by key_id not yet supported", "key_id", id.KeyId, "user_id", userID)
		return nil, status.Error(codes.Unimplemented, "lookup by key_id not yet implemented - use key value instead")

	case *corev1.GetKeyRequest_Key:
		// Lookup by key value (CPF, email, phone, etc.)
		keyValue := id.Key.GetKeyValue()
		h.logger.Info("GetKey: lookup by key_value", "key_value", keyValue, "user_id", userID)

		// 3c. Execute query handler
		query := queries.GetEntryQuery{
			KeyValue: keyValue,
		}
		entry, err = h.getEntryQuery.Handle(ctx, query)
		if err != nil {
			h.logger.Error("GetKey: query failed", "error", err, "user_id", userID)
			return nil, mappers.MapDomainErrorToGRPC(err)
		}

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid identifier type")
	}

	// 3d. Build account from entry
	account := &entities.Account{
		ISPB:          entry.Account.ISPB,
		Branch:        entry.Account.Branch,
		AccountNumber: entry.Account.AccountNumber,
		AccountType:   entry.Account.AccountType,
	}

	// 3e. Map domain result → proto response
	resp := mappers.MapDomainEntryToProtoGetKeyResponse(entry, account)

	h.logger.Info("GetKey: success", "entry_id", entry.ID, "user_id", userID)
	return resp, nil
}
```

---

## Method 4: DeleteKey

**Line:** 325
**Implementation:**

```go
// DeleteKey deletes a PIX key
func (h *CoreDictServiceHandler) DeleteKey(ctx context.Context, req *corev1.DeleteKeyRequest) (*corev1.DeleteKeyResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	if req.GetKeyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "key_id is required")
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("DeleteKey: MOCK MODE", "key_id", req.GetKeyId())
		return &corev1.DeleteKeyResponse{
			Deleted:   true,
			DeletedAt: timestamppb.Now(),
		}, nil
	}

	// ========== 3. REAL MODE ==========
	h.logger.Info("DeleteKey: REAL MODE", "key_id", req.GetKeyId())

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("DeleteKey: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto request → domain command
	cmd, err := mappers.MapProtoDeleteKeyRequestToCommand(req, userID)
	if err != nil {
		h.logger.Error("DeleteKey: mapping failed", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3c. Execute command handler
	result, err := h.deleteEntryCmd.Handle(ctx, cmd)
	if err != nil {
		h.logger.Error("DeleteKey: command failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Return success response
	h.logger.Info("DeleteKey: success", "key_id", req.GetKeyId(), "user_id", userID)
	return &corev1.DeleteKeyResponse{
		Deleted:   result.Success,
		DeletedAt: timestamppb.New(result.DeletedAt),
	}, nil
}
```

---

## Method 5: StartClaim

**Line:** 354
**Challenge:** CreateClaimCommand returns `*CreateClaimResult`, but mapper expects `*entities.Claim`

**Solution:** Query the claim after creation

**Implementation:**

```go
// StartClaim initiates a claim for a PIX key owned by another user
func (h *CoreDictServiceHandler) StartClaim(ctx context.Context, req *corev1.StartClaimRequest) (*corev1.StartClaimResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	if req.GetKey() == nil {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}
	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("StartClaim: MOCK MODE", "key", req.GetKey().GetKeyValue())
		now := time.Now()
		expiresAt := now.Add(30 * 24 * time.Hour) // 30 days
		return &corev1.StartClaimResponse{
			ClaimId:   fmt.Sprintf("claim-%d", now.Unix()),
			EntryId:   "entry-123",
			Status:    commonv1.ClaimStatus_CLAIM_STATUS_OPEN,
			ExpiresAt: timestamppb.New(expiresAt),
			CreatedAt: timestamppb.New(now),
			Message:   "Claim created. The current owner has 30 days to respond",
		}, nil
	}

	// ========== 3. REAL MODE ==========
	h.logger.Info("StartClaim: REAL MODE", "key", req.GetKey().GetKeyValue())

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("StartClaim: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Map proto request → domain command
	cmd, err := mappers.MapProtoStartClaimRequestToCommand(req, userID)
	if err != nil {
		h.logger.Error("StartClaim: mapping failed", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3c. Execute command handler
	result, err := h.createClaimCmd.Handle(ctx, cmd)
	if err != nil {
		h.logger.Error("StartClaim: command failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Fetch created claim for complete response
	claim, err := h.getClaimQuery.Handle(ctx, queries.GetClaimQuery{ClaimID: result.ClaimID})
	if err != nil {
		// Fallback: return minimal response
		h.logger.Warn("StartClaim: could not fetch created claim", "claim_id", result.ClaimID, "error", err)
		return &corev1.StartClaimResponse{
			ClaimId:   result.ClaimID.String(),
			EntryId:   "",
			Status:    mappers.MapStringStatusToProto(result.Status),
			ExpiresAt: timestamppb.New(result.DeadlineAt),
			CreatedAt: timestamppb.Now(),
			Message:   "Claim created successfully",
		}, nil
	}

	// 3e. Map domain result → proto response
	resp := mappers.MapDomainClaimToProtoStartClaimResponse(claim)

	h.logger.Info("StartClaim: success", "claim_id", result.ClaimID, "user_id", userID)
	return resp, nil
}
```

---

## Method 6: GetClaimStatus

**Line:** 378
**Implementation:**

```go
// GetClaimStatus retrieves the current status of a claim
func (h *CoreDictServiceHandler) GetClaimStatus(ctx context.Context, req *corev1.GetClaimStatusRequest) (*corev1.GetClaimStatusResponse, error) {
	// ========== 1. VALIDATION (always) ==========
	if req.GetClaimId() == "" {
		return nil, status.Error(codes.InvalidArgument, "claim_id is required")
	}

	// ========== 2. MOCK MODE ==========
	if h.useMockMode {
		h.logger.Info("GetClaimStatus: MOCK MODE", "claim_id", req.GetClaimId())
		now := time.Now()
		expiresAt := now.Add(29 * 24 * time.Hour)
		return &corev1.GetClaimStatusResponse{
			ClaimId: req.GetClaimId(),
			EntryId: "entry-123",
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

	// ========== 3. REAL MODE ==========
	h.logger.Info("GetClaimStatus: REAL MODE", "claim_id", req.GetClaimId())

	// 3a. Extract user_id from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("GetClaimStatus: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Parse claim_id to UUID
	claimID, err := uuid.Parse(req.GetClaimId())
	if err != nil {
		h.logger.Error("GetClaimStatus: invalid claim_id", "error", err, "user_id", userID)
		return nil, status.Error(codes.InvalidArgument, "invalid claim_id format")
	}

	// 3c. Execute query handler
	query := queries.GetClaimQuery{
		ClaimID: claimID,
	}
	claim, err := h.getClaimQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("GetClaimStatus: query failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3d. Map domain result → proto response (entry=nil since not needed for status)
	resp := mappers.MapDomainClaimToProtoGetClaimStatusResponse(claim, nil)

	h.logger.Info("GetClaimStatus: success", "claim_id", claim.ID, "user_id", userID)
	return resp, nil
}
```

---

## Method 7: ListIncomingClaims

**Line:** 404
**Implementation:**

```go
// ListIncomingClaims lists claims received by the authenticated user (where user is the current owner)
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
					ClaimId: "claim-1",
					EntryId: "entry-123",
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
		h.logger.Warn("ListIncomingClaims: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Get ISPB from user context
	// TODO: Fetch ISPB from account/user service
	ispb := "12345678" // Placeholder - should be fetched from account

	// 3c. Map proto request → domain query
	query := mappers.MapProtoListIncomingClaimsRequestToQuery(req, ispb)

	// 3d. Execute query handler
	result, err := h.listClaimsQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("ListIncomingClaims: query failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3e. Map domain result → proto response
	claimSummaries := make([]*corev1.ClaimSummary, 0, len(result.Claims))
	for _, claim := range result.Claims {
		claimSummaries = append(claimSummaries, mappers.MapDomainClaimToProtoSummary(claim))
	}

	// Calculate next page token
	nextPageToken := ""
	if result.Page < result.TotalPages {
		nextPageToken = fmt.Sprintf("page=%d", result.Page+1)
	}

	h.logger.Info("ListIncomingClaims: success", "count", len(claimSummaries), "total", result.TotalCount, "user_id", userID)
	return &corev1.ListIncomingClaimsResponse{
		Claims:        claimSummaries,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.TotalCount),
	}, nil
}
```

---

## Method 8: ListOutgoingClaims

**Line:** 435
**Implementation:**

```go
// ListOutgoingClaims lists claims sent by the authenticated user (where user is the claimer)
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
					ClaimId: "claim-2",
					EntryId: "entry-456",
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
		h.logger.Warn("ListOutgoingClaims: user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// 3b. Get ISPB from user context
	// TODO: Fetch ISPB from account/user service
	ispb := "87654321" // Placeholder - should be fetched from account

	// 3c. Map proto request → domain query
	query := mappers.MapProtoListOutgoingClaimsRequestToQuery(req, ispb)

	// 3d. Execute query handler
	result, err := h.listClaimsQuery.Handle(ctx, query)
	if err != nil {
		h.logger.Error("ListOutgoingClaims: query failed", "error", err, "user_id", userID)
		return nil, mappers.MapDomainErrorToGRPC(err)
	}

	// 3e. Map domain result → proto response
	claimSummaries := make([]*corev1.ClaimSummary, 0, len(result.Claims))
	for _, claim := range result.Claims {
		claimSummaries = append(claimSummaries, mappers.MapDomainClaimToProtoSummary(claim))
	}

	// Calculate next page token
	nextPageToken := ""
	if result.Page < result.TotalPages {
		nextPageToken = fmt.Sprintf("page=%d", result.Page+1)
	}

	h.logger.Info("ListOutgoingClaims: success", "count", len(claimSummaries), "total", result.TotalCount, "user_id", userID)
	return &corev1.ListOutgoingClaimsResponse{
		Claims:        claimSummaries,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.TotalCount),
	}, nil
}
```

---

## Additional Import Required

Add to imports section:
```go
"github.com/google/uuid"
```

---

## Testing Commands

### Compile Check
```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go build ./internal/infrastructure/grpc/...
```

### Full Build
```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go build ./...
```

### Run Tests (if any)
```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
go test ./internal/infrastructure/grpc/...
```

---

## Summary of Changes

1. ✅ **Imports Added:**
   - `"github.com/google/uuid"` (needed for uuid.Parse)
   - `"github.com/lbpay-lab/core-dict/internal/domain/entities"` (needed for entities.Entry, entities.Claim)

2. ✅ **Methods 1-2 (Already Done):**
   - CreateKey: Real Mode implemented
   - ListKeys: Real Mode implemented

3. 🔄 **Methods 3-8 (To Be Applied Manually):**
   - GetKey: Implementation provided above
   - DeleteKey: Implementation provided above
   - StartClaim: Implementation provided above (with claim fetch workaround)
   - GetClaimStatus: Implementation provided above
   - ListIncomingClaims: Implementation provided above
   - ListOutgoingClaims: Implementation provided above

4. ✅ **Pattern Consistency:**
   - All methods follow 3-section structure: Validation → Mock Mode → Real Mode
   - All methods extract user_id from context
   - All methods use mappers for proto↔domain conversion
   - All methods use structured logging
   - All methods use MapDomainErrorToGRPC for error handling

---

## Known Limitations / TODOs

1. **GetKey by key_id:** Not implemented (requires repository method or query update)
2. **ISPB Lookup:** Hardcoded placeholders - needs account service integration
3. **StartClaim:** Requires extra query to fetch claim after creation (mapper mismatch)
4. **Pagination Tokens:** Simplified implementation - may need enhancement

---

## Next Steps

1. Apply implementations manually to `core_dict_service_handler.go`
2. Add `uuid` import if not already present
3. Compile and fix any errors
4. Test with integration tests (if available)
5. Update TODO comments for remaining work

---

**Generated:** 2025-10-27
**For:** Core-Dict Real Mode Implementation Sprint

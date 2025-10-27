package com.lbpay.xmlsigner.controller;

import com.lbpay.xmlsigner.model.SignRequest;
import com.lbpay.xmlsigner.model.SignResponse;
import com.lbpay.xmlsigner.service.XmlSignerService;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

/**
 * REST Controller for XML signing operations
 */
@RestController
@RequestMapping("/api/v1/xml-signer")
public class XmlSignerController {

    private static final Logger logger = LoggerFactory.getLogger(XmlSignerController.class);

    @Autowired
    private XmlSignerService xmlSignerService;

    /**
     * Sign XML document
     *
     * POST /api/v1/xml-signer/sign
     *
     * @param request Sign request with XML and certificate info
     * @return Signed XML response
     */
    @PostMapping("/sign")
    public ResponseEntity<SignResponse> signXml(@Valid @RequestBody SignRequest request) {
        logger.info("Received sign request for certificate: {}", request.getCertificatePath());

        SignResponse response = xmlSignerService.signXml(request);

        if (response.isSuccess()) {
            return ResponseEntity.ok(response);
        } else {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(response);
        }
    }

    /**
     * Verify XML signature
     *
     * POST /api/v1/xml-signer/verify
     *
     * @param payload Map containing signed XML
     * @return Verification result
     */
    @PostMapping("/verify")
    public ResponseEntity<Map<String, Object>> verifySignature(@RequestBody Map<String, String> payload) {
        String signedXml = payload.get("signedXml");

        if (signedXml == null || signedXml.isEmpty()) {
            Map<String, Object> error = new HashMap<>();
            error.put("success", false);
            error.put("message", "Signed XML is required");
            return ResponseEntity.badRequest().body(error);
        }

        boolean isValid = xmlSignerService.verifySignature(signedXml);

        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("valid", isValid);
        result.put("message", isValid ? "Signature is valid" : "Signature is invalid");

        return ResponseEntity.ok(result);
    }

    /**
     * Health check endpoint
     *
     * GET /api/v1/xml-signer/health
     *
     * @return Health status
     */
    @GetMapping("/health")
    public ResponseEntity<Map<String, Object>> health() {
        Map<String, Object> health = new HashMap<>();
        health.put("status", "UP");
        health.put("service", "xml-signer");
        health.put("timestamp", System.currentTimeMillis());

        return ResponseEntity.ok(health);
    }

    /**
     * Get service info
     *
     * GET /api/v1/xml-signer/info
     *
     * @return Service information
     */
    @GetMapping("/info")
    public ResponseEntity<Map<String, Object>> info() {
        Map<String, Object> info = new HashMap<>();
        info.put("service", "XML Digital Signature Service");
        info.put("version", "1.0.0");
        info.put("description", "Digital signature service for XML documents using ICP-Brasil certificates");
        info.put("supportedAlgorithms", new String[]{
                "RSA-SHA256", "RSA-SHA512", "RSA-SHA1"
        });
        info.put("supportedFormats", new String[]{
                "PKCS12 (.p12, .pfx)", "JKS (.jks)"
        });

        return ResponseEntity.ok(info);
    }

    /**
     * Exception handler for validation errors
     */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<Map<String, Object>> handleException(Exception e) {
        logger.error("Error processing request", e);

        Map<String, Object> error = new HashMap<>();
        error.put("success", false);
        error.put("message", "Error processing request");
        error.put("error", e.getMessage());

        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(error);
    }
}

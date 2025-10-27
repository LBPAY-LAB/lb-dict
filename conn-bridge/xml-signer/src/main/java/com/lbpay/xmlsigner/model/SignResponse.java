package com.lbpay.xmlsigner.model;

import java.time.LocalDateTime;

/**
 * Response model for XML signing operations
 */
public class SignResponse {

    private boolean success;
    private String signedXml;
    private String message;
    private LocalDateTime timestamp;
    private String certificateInfo;
    private String error;

    // Constructors
    public SignResponse() {
        this.timestamp = LocalDateTime.now();
    }

    public SignResponse(boolean success, String signedXml, String message) {
        this.success = success;
        this.signedXml = signedXml;
        this.message = message;
        this.timestamp = LocalDateTime.now();
    }

    public static SignResponse success(String signedXml, String certificateInfo) {
        SignResponse response = new SignResponse();
        response.setSuccess(true);
        response.setSignedXml(signedXml);
        response.setMessage("XML signed successfully");
        response.setCertificateInfo(certificateInfo);
        return response;
    }

    public static SignResponse error(String errorMessage) {
        SignResponse response = new SignResponse();
        response.setSuccess(false);
        response.setMessage("Failed to sign XML");
        response.setError(errorMessage);
        return response;
    }

    // Getters and Setters
    public boolean isSuccess() {
        return success;
    }

    public void setSuccess(boolean success) {
        this.success = success;
    }

    public String getSignedXml() {
        return signedXml;
    }

    public void setSignedXml(String signedXml) {
        this.signedXml = signedXml;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public LocalDateTime getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LocalDateTime timestamp) {
        this.timestamp = timestamp;
    }

    public String getCertificateInfo() {
        return certificateInfo;
    }

    public void setCertificateInfo(String certificateInfo) {
        this.certificateInfo = certificateInfo;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }

    @Override
    public String toString() {
        return "SignResponse{" +
                "success=" + success +
                ", message='" + message + '\'' +
                ", timestamp=" + timestamp +
                ", certificateInfo='" + certificateInfo + '\'' +
                ", error='" + error + '\'' +
                '}';
    }
}

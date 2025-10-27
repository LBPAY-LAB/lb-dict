package com.lbpay.xmlsigner.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

/**
 * Request model for XML signing operations
 */
public class SignRequest {

    @NotBlank(message = "XML content is required")
    private String xmlContent;

    @NotBlank(message = "Certificate path is required")
    private String certificatePath;

    private String certificatePassword;

    @NotNull(message = "Development mode flag is required")
    private Boolean devMode = false;

    private String keyAlias;

    private String signatureMethod = "RSA-SHA256";

    private String canonicalizationMethod = "http://www.w3.org/2001/10/xml-exc-c14n#";

    // Constructors
    public SignRequest() {}

    public SignRequest(String xmlContent, String certificatePath, String certificatePassword) {
        this.xmlContent = xmlContent;
        this.certificatePath = certificatePath;
        this.certificatePassword = certificatePassword;
    }

    // Getters and Setters
    public String getXmlContent() {
        return xmlContent;
    }

    public void setXmlContent(String xmlContent) {
        this.xmlContent = xmlContent;
    }

    public String getCertificatePath() {
        return certificatePath;
    }

    public void setCertificatePath(String certificatePath) {
        this.certificatePath = certificatePath;
    }

    public String getCertificatePassword() {
        return certificatePassword;
    }

    public void setCertificatePassword(String certificatePassword) {
        this.certificatePassword = certificatePassword;
    }

    public Boolean getDevMode() {
        return devMode;
    }

    public void setDevMode(Boolean devMode) {
        this.devMode = devMode;
    }

    public String getKeyAlias() {
        return keyAlias;
    }

    public void setKeyAlias(String keyAlias) {
        this.keyAlias = keyAlias;
    }

    public String getSignatureMethod() {
        return signatureMethod;
    }

    public void setSignatureMethod(String signatureMethod) {
        this.signatureMethod = signatureMethod;
    }

    public String getCanonicalizationMethod() {
        return canonicalizationMethod;
    }

    public void setCanonicalizationMethod(String canonicalizationMethod) {
        this.canonicalizationMethod = canonicalizationMethod;
    }

    @Override
    public String toString() {
        return "SignRequest{" +
                "certificatePath='" + certificatePath + '\'' +
                ", devMode=" + devMode +
                ", keyAlias='" + keyAlias + '\'' +
                ", signatureMethod='" + signatureMethod + '\'' +
                ", canonicalizationMethod='" + canonicalizationMethod + '\'' +
                '}';
    }
}

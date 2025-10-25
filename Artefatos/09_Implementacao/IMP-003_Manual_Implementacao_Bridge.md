# IMP-003: Manual de Implementação - RSFN Bridge

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Bridge (SOAP/mTLS Adapter)
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: BACKEND (AI Agent - Backend Developer)

---

## Sumário Executivo

Este manual fornece instruções passo-a-passo para configurar, implementar e executar o **RSFN Bridge**, o módulo adaptador que comunica com o Bacen DICT via SOAP/XML com autenticação mTLS usando certificados ICP-Brasil A3.

---

## Índice

1. [Pré-requisitos](#1-pré-requisitos)
2. [Setup do Repositório](#2-setup-do-repositório)
3. [Configuração de Certificados mTLS (ICP-Brasil A3)](#3-configuração-de-certificados-mtls-icp-brasil-a3)
4. [Setup do XML Signer (Java)](#4-setup-do-xml-signer-java)
5. [Configuração SOAP Client](#5-configuração-soap-client)
6. [Configuração da Aplicação](#6-configuração-da-aplicação)
7. [Execução Local](#7-execução-local)
8. [Testes de Integração Bacen](#8-testes-de-integração-bacen)
9. [Checklist de Implementação](#9-checklist-de-implementação)

---

## 1. Pré-requisitos

### 1.1. Software Necessário

| Software | Versão Mínima | Propósito |
|----------|---------------|-----------|
| **Go** | 1.22+ | Linguagem de programação (gRPC Server) |
| **Java JDK** | 11+ | XML Signer (ICP-Brasil) |
| **OpenSSL** | 1.1+ | Manipulação de certificados |
| **Docker** | 20.10+ | Containerização |
| **Wiremock** | 2.35+ | Mock de API Bacen (testes) |

### 1.2. Certificados ICP-Brasil A3

- **Certificado A3 e-CNPJ** (arquivo .pfx ou .p12)
- **Senha do certificado**
- **Cadeia de certificados ICP-Brasil** (AC Raiz, AC Intermediária)

### 1.3. Variáveis de Ambiente

Criar arquivo `.env`:

```bash
# Application Configuration
APP_ENV=development
APP_NAME=rsfn-bridge
GRPC_PORT=50051
HTTP_PORT=8081

# Bacen DICT Configuration
BACEN_DICT_URL=https://dict.bacen.gov.br/api/v1
BACEN_DICT_TIMEOUT=30s
BACEN_DICT_ISPB=12345678

# mTLS Certificate (ICP-Brasil A3)
MTLS_CERT_PATH=/certs/cert.pem
MTLS_KEY_PATH=/certs/key.pem
MTLS_CA_PATH=/certs/ca-chain.pem
MTLS_CERT_PASSWORD=your_cert_password_here

# Java XML Signer
JAVA_HOME=/usr/lib/jvm/java-11-openjdk
XML_SIGNER_JAR_PATH=/app/xml-signer.jar
XML_SIGNER_KEYSTORE_PATH=/certs/keystore.p12
XML_SIGNER_KEYSTORE_PASSWORD=your_keystore_password

# Circuit Breaker
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_TIMEOUT=60s

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
ENABLE_TRACING=true
```

---

## 2. Setup do Repositório

### 2.1. Criar Estrutura do Projeto

```bash
# Criar diretório do projeto
mkdir -p rsfn-bridge
cd rsfn-bridge

# Inicializar módulo Go
go mod init github.com/lbpay/rsfn-bridge

# Criar estrutura de diretórios
mkdir -p cmd/server
mkdir -p internal/{domain,application,infrastructure}
mkdir -p internal/infrastructure/{grpc,soap,mtls,signer}
mkdir -p pkg/{logger,circuit_breaker}
mkdir -p certs
mkdir -p java-signer/src/main/java/com/lbpay/xmlsigner
mkdir -p config
mkdir -p tests/mocks
```

### 2.2. Estrutura de Diretórios

```
rsfn-bridge/
├── cmd/
│   └── server/
│       └── main.go                     # Entrypoint (gRPC Server)
├── internal/
│   ├── domain/
│   │   ├── entry.go                    # Domain model: Entry
│   │   └── claim.go                    # Domain model: Claim
│   ├── application/
│   │   ├── entry_service.go            # Entry business logic
│   │   └── claim_service.go            # Claim business logic
│   ├── infrastructure/
│   │   ├── grpc/
│   │   │   ├── server.go               # gRPC Server
│   │   │   └── handlers/
│   │   │       ├── entry_handler.go    # Entry gRPC handler
│   │   │       └── claim_handler.go    # Claim gRPC handler
│   │   ├── soap/
│   │   │   ├── client.go               # SOAP Client (Bacen)
│   │   │   ├── templates/
│   │   │   │   ├── create_entry.xml    # SOAP template
│   │   │   │   └── create_claim.xml    # SOAP template
│   │   │   └── parser.go               # XML Response parser
│   │   ├── mtls/
│   │   │   └── transport.go            # mTLS HTTP Transport
│   │   └── signer/
│   │       └── xml_signer.go           # Java XML Signer caller
│   └── ports/
│       └── bacen_client.go             # Bacen client interface
├── pkg/
│   ├── logger/
│   │   └── logger.go                   # Logger implementation
│   └── circuit_breaker/
│       └── circuit_breaker.go          # Circuit Breaker pattern
├── certs/
│   ├── cert.pem                        # Certificado mTLS (ICP-Brasil)
│   ├── key.pem                         # Private key
│   ├── ca-chain.pem                    # CA chain (ICP-Brasil)
│   └── keystore.p12                    # Keystore Java
├── java-signer/
│   ├── pom.xml                         # Maven dependencies
│   └── src/main/java/com/lbpay/xmlsigner/
│       └── XmlSigner.java              # XML Signer (ICP-Brasil)
├── config/
│   └── config.yaml                     # Application config
├── tests/
│   └── mocks/
│       └── wiremock/
│           └── mappings/
│               └── bacen_dict.json     # Wiremock mappings
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yaml
├── .env
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 2.3. Instalar Dependências Go

```bash
# gRPC + Protobuf
go get google.golang.org/grpc
go get google.golang.org/protobuf

# HTTP Client
go get github.com/go-resty/resty/v2

# XML Processing
go get github.com/beevik/etree

# Circuit Breaker
go get github.com/sony/gobreaker

# Configuration
go get github.com/spf13/viper

# Logging
go get github.com/sirupsen/logrus

# OpenTelemetry
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
```

---

## 3. Configuração de Certificados mTLS (ICP-Brasil A3)

### 3.1. Extrair Certificado e Chave Privada do arquivo .pfx

```bash
# Assumindo que você tem o certificado em formato .pfx (ICP-Brasil A3)
cd certs

# Extrair certificado (.pem)
openssl pkcs12 -in certificado-a3.pfx -clcerts -nokeys -out cert.pem
# Senha: <digite a senha do certificado A3>

# Extrair chave privada (.pem)
openssl pkcs12 -in certificado-a3.pfx -nocerts -out key-encrypted.pem
# Senha: <digite a senha do certificado A3>

# Remover senha da chave privada (para uso em servidor)
openssl rsa -in key-encrypted.pem -out key.pem
# Senha: <digite a senha novamente>

# Limpar arquivo temporário
rm key-encrypted.pem
```

### 3.2. Obter Cadeia de Certificados ICP-Brasil

```bash
# Baixar cadeia de certificados ICP-Brasil
# Fonte: http://acraiz.icpbrasil.gov.br/

# AC Raiz v10 (vigente)
wget http://acraiz.icpbrasil.gov.br/credenciadas/RAIZ/ICP-Brasilv10.crt -O ac-raiz-v10.crt

# AC Intermediária (exemplo: AC Serasa)
# Baixe de acordo com a sua AC emissora

# Converter .crt para .pem
openssl x509 -inform DER -in ac-raiz-v10.crt -out ac-raiz-v10.pem

# Concatenar cadeia
cat ac-intermediaria.pem ac-raiz-v10.pem > ca-chain.pem
```

### 3.3. Criar Keystore Java (para XML Signer)

```bash
# Converter .pfx para .p12 (formato compatível com Java)
openssl pkcs12 -export -in cert.pem -inkey key.pem -out keystore.p12 \
  -name "lbpay-dict" -CAfile ca-chain.pem -caname root

# Senha: <defina uma senha para o keystore>
```

### 3.4. Validar Certificados

```bash
# Verificar certificado
openssl x509 -in cert.pem -text -noout

# Verificar validade
openssl x509 -in cert.pem -noout -dates

# Verificar cadeia de certificados
openssl verify -CAfile ca-chain.pem cert.pem
```

---

## 4. Setup do XML Signer (Java)

### 4.1. Criar Projeto Maven

**Arquivo**: `java-signer/pom.xml`

```xml
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.lbpay</groupId>
    <artifactId>xml-signer</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>

    <dependencies>
        <!-- Apache Santuario (XML Security) -->
        <dependency>
            <groupId>org.apache.santuario</groupId>
            <artifactId>xmlsec</artifactId>
            <version>2.3.3</version>
        </dependency>

        <!-- BouncyCastle (ICP-Brasil) -->
        <dependency>
            <groupId>org.bouncycastle</groupId>
            <artifactId>bcprov-jdk15on</artifactId>
            <version>1.70</version>
        </dependency>

        <dependency>
            <groupId>org.bouncycastle</groupId>
            <artifactId>bcpkix-jdk15on</artifactId>
            <version>1.70</version>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-assembly-plugin</artifactId>
                <version>3.3.0</version>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>com.lbpay.xmlsigner.XmlSigner</mainClass>
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

### 4.2. Implementar XML Signer

**Arquivo**: `java-signer/src/main/java/com/lbpay/xmlsigner/XmlSigner.java` (pseudocode)

```java
package com.lbpay.xmlsigner;

import org.apache.xml.security.signature.XMLSignature;
import org.apache.xml.security.transforms.Transforms;
import org.apache.xml.security.utils.Constants;
import org.w3c.dom.Document;

import javax.xml.parsers.DocumentBuilderFactory;
import java.io.FileInputStream;
import java.security.KeyStore;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;

public class XmlSigner {
    public static void main(String[] args) throws Exception {
        // Args: [0] XML file path, [1] Keystore path, [2] Keystore password, [3] Alias

        if (args.length < 4) {
            System.err.println("Uso: XmlSigner <xml-file> <keystore> <password> <alias>");
            System.exit(1);
        }

        String xmlFile = args[0];
        String keystorePath = args[1];
        String keystorePassword = args[2];
        String alias = args[3];

        // Carregar XML
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware(true);
        Document doc = dbf.newDocumentBuilder().parse(xmlFile);

        // Carregar Keystore
        KeyStore ks = KeyStore.getInstance("PKCS12");
        ks.load(new FileInputStream(keystorePath), keystorePassword.toCharArray());

        // Obter chave privada e certificado
        PrivateKey privateKey = (PrivateKey) ks.getKey(alias, keystorePassword.toCharArray());
        X509Certificate cert = (X509Certificate) ks.getCertificate(alias);

        // Inicializar Apache Santuario
        org.apache.xml.security.Init.init();

        // Criar assinatura XML
        XMLSignature signature = new XMLSignature(doc, "",
            XMLSignature.ALGO_ID_SIGNATURE_RSA_SHA256);

        doc.getDocumentElement().appendChild(signature.getElement());

        // Configurar transformações
        Transforms transforms = new Transforms(doc);
        transforms.addTransform(Transforms.TRANSFORM_ENVELOPED_SIGNATURE);
        signature.addDocument("", transforms, Constants.ALGO_ID_DIGEST_SHA256);

        // Adicionar certificado
        signature.addKeyInfo(cert);

        // Assinar
        signature.sign(privateKey);

        // Serializar XML assinado para stdout
        javax.xml.transform.TransformerFactory.newInstance()
            .newTransformer()
            .transform(new javax.xml.transform.dom.DOMSource(doc),
                      new javax.xml.transform.stream.StreamResult(System.out));
    }
}
```

### 4.3. Compilar XML Signer

```bash
cd java-signer

# Compilar com Maven
mvn clean package

# JAR gerado em: target/xml-signer-1.0.0-jar-with-dependencies.jar
cp target/xml-signer-1.0.0-jar-with-dependencies.jar ../xml-signer.jar
```

---

## 5. Configuração SOAP Client

### 5.1. Implementar mTLS HTTP Transport

**Arquivo**: `internal/infrastructure/mtls/transport.go` (pseudocode)

```go
package mtls

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "net/http"
)

func NewMTLSTransport(certFile, keyFile, caFile string) (*http.Transport, error) {
    // Carregar certificado e chave privada
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    // Carregar CA chain
    caCert, err := ioutil.ReadFile(caFile)
    if err != nil {
        return nil, err
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // Configurar TLS
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
    }

    // Criar Transport HTTP com mTLS
    transport := &http.Transport{
        TLSClientConfig: tlsConfig,
    }

    return transport, nil
}
```

### 5.2. Implementar SOAP Client

**Arquivo**: `internal/infrastructure/soap/client.go` (pseudocode)

```go
package soap

import (
    "bytes"
    "context"
    "io/ioutil"
    "net/http"
    "text/template"
)

type SOAPClient struct {
    httpClient *http.Client
    baseURL    string
    signer     *signer.XMLSigner
}

func NewSOAPClient(baseURL string, transport *http.Transport, signer *signer.XMLSigner) *SOAPClient {
    return &SOAPClient{
        httpClient: &http.Client{Transport: transport},
        baseURL:    baseURL,
        signer:     signer,
    }
}

func (c *SOAPClient) CreateEntry(ctx context.Context, entry *domain.Entry) (*CreateEntryResponse, error) {
    // 1. Gerar XML SOAP a partir de template
    xmlPayload, err := c.generateSOAPPayload("create_entry.xml", entry)
    if err != nil {
        return nil, err
    }

    // 2. Assinar XML com certificado ICP-Brasil
    signedXML, err := c.signer.SignXML(xmlPayload)
    if err != nil {
        return nil, err
    }

    // 3. Enviar requisição SOAP via mTLS
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/dict/entries", bytes.NewReader(signedXML))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "text/xml; charset=utf-8")
    req.Header.Set("SOAPAction", "CreateEntry")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 4. Parsear resposta XML
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return parseCreateEntryResponse(body)
}

func (c *SOAPClient) generateSOAPPayload(templateName string, data interface{}) ([]byte, error) {
    tmpl, err := template.ParseFiles("internal/infrastructure/soap/templates/" + templateName)
    if err != nil {
        return nil, err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

### 5.3. Template SOAP

**Arquivo**: `internal/infrastructure/soap/templates/create_entry.xml`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:dict="http://bacen.gov.br/dict/v1">
    <soapenv:Header/>
    <soapenv:Body>
        <dict:CreateEntry>
            <dict:ISPB>{{.ParticipantISPB}}</dict:ISPB>
            <dict:KeyType>{{.KeyType}}</dict:KeyType>
            <dict:KeyValue>{{.KeyValue}}</dict:KeyValue>
            <dict:AccountType>{{.AccountType}}</dict:AccountType>
            <dict:Branch>{{.Branch}}</dict:Branch>
            <dict:Account>{{.Account}}</dict:Account>
            <dict:OwnerName>{{.OwnerName}}</dict:OwnerName>
            <dict:OwnerTaxID>{{.OwnerTaxID}}</dict:OwnerTaxID>
        </dict:CreateEntry>
    </soapenv:Body>
</soapenv:Envelope>
```

---

## 6. Configuração da Aplicação

### 6.1. Main gRPC Server

**Arquivo**: `cmd/server/main.go` (pseudocode)

```go
package main

import (
    "log"
    "net"
    "google.golang.org/grpc"
    "github.com/lbpay/rsfn-bridge/internal/infrastructure/grpc/handlers"
    "github.com/lbpay/rsfn-bridge/internal/infrastructure/mtls"
    "github.com/lbpay/rsfn-bridge/internal/infrastructure/soap"
    "github.com/lbpay/rsfn-bridge/internal/infrastructure/signer"
)

func main() {
    // Configurar mTLS Transport
    transport, err := mtls.NewMTLSTransport(
        "/certs/cert.pem",
        "/certs/key.pem",
        "/certs/ca-chain.pem",
    )
    if err != nil {
        log.Fatalf("Erro ao configurar mTLS: %v", err)
    }

    // Configurar XML Signer
    xmlSigner := signer.NewXMLSigner("/app/xml-signer.jar", "/certs/keystore.p12", "password")

    // Configurar SOAP Client
    soapClient := soap.NewSOAPClient("https://dict.bacen.gov.br/api/v1", transport, xmlSigner)

    // Criar gRPC Server
    grpcServer := grpc.NewServer()
    handlers.RegisterBridgeService(grpcServer, soapClient)

    // Iniciar servidor
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Erro ao escutar porta: %v", err)
    }

    log.Println("Bridge gRPC Server rodando na porta 50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Erro ao iniciar servidor: %v", err)
    }
}
```

---

## 7. Execução Local

### 7.1. Docker Compose (com Wiremock)

**Arquivo**: `docker/docker-compose.yaml`

```yaml
version: '3.8'

services:
  rsfn-bridge:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "50051:50051"
    volumes:
      - ../certs:/certs:ro
      - ../xml-signer.jar:/app/xml-signer.jar:ro
    environment:
      - BACEN_DICT_URL=http://wiremock:8080
      - MTLS_CERT_PATH=/certs/cert.pem
      - MTLS_KEY_PATH=/certs/key.pem
      - MTLS_CA_PATH=/certs/ca-chain.pem
      - XML_SIGNER_JAR_PATH=/app/xml-signer.jar
      - XML_SIGNER_KEYSTORE_PATH=/certs/keystore.p12

  wiremock:
    image: wiremock/wiremock:latest
    ports:
      - "8080:8080"
    volumes:
      - ../tests/mocks/wiremock:/home/wiremock
```

```bash
# Iniciar serviços
cd docker
docker-compose up -d
```

---

## 8. Testes de Integração Bacen

### 8.1. Configurar Wiremock (Mock Bacen)

**Arquivo**: `tests/mocks/wiremock/mappings/bacen_dict.json`

```json
{
  "request": {
    "method": "POST",
    "urlPath": "/dict/entries",
    "headers": {
      "Content-Type": { "contains": "text/xml" }
    }
  },
  "response": {
    "status": 200,
    "headers": {
      "Content-Type": "text/xml"
    },
    "body": "<?xml version=\"1.0\"?><soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\"><soapenv:Body><CreateEntryResponse><DictID>DICT-12345</DictID><Status>ACTIVE</Status></CreateEntryResponse></soapenv:Body></soapenv:Envelope>"
  }
}
```

### 8.2. Testar gRPC Client

```bash
# Usando grpcurl
grpcurl -plaintext -d '{"key":"12345678901","type":"CPF"}' \
  localhost:50051 BridgeService/CreateEntry
```

---

## 9. Checklist de Implementação

### 9.1. Certificados mTLS

- [ ] Certificado ICP-Brasil A3 (.pfx) obtido
- [ ] Certificado extraído para cert.pem
- [ ] Chave privada extraída para key.pem
- [ ] Cadeia CA ICP-Brasil baixada (ca-chain.pem)
- [ ] Keystore Java criado (keystore.p12)
- [ ] Certificados validados com OpenSSL

### 9.2. XML Signer (Java)

- [ ] Projeto Maven criado
- [ ] Dependências (Santuario, BouncyCastle) adicionadas
- [ ] XmlSigner.java implementado
- [ ] JAR compilado (xml-signer.jar)
- [ ] Teste de assinatura XML executado

### 9.3. SOAP Client

- [ ] mTLS Transport configurado
- [ ] SOAP Client implementado
- [ ] Templates SOAP criados (create_entry.xml, create_claim.xml)
- [ ] XML Response parser implementado
- [ ] Circuit Breaker implementado

### 9.4. gRPC Server

- [ ] Proto definitions criadas
- [ ] gRPC handlers implementados
- [ ] Server configurado e rodando

### 9.5. Testing

- [ ] Wiremock configurado (mock Bacen)
- [ ] Testes de integração com Wiremock executados
- [ ] Testes de assinatura XML validados
- [ ] Testes de mTLS validados

---

## Próximos Passos

1. Integração com ambiente de homologação Bacen
2. Configurar observabilidade (OpenTelemetry)
3. Implementar retry policies
4. Configurar CI/CD
5. Testes de carga

---

**Referências**:
- [TEC-002: Bridge Specification](../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [ICP-Brasil - Cadeia de Certificados](http://acraiz.icpbrasil.gov.br/)
- [Apache Santuario](https://santuario.apache.org/)

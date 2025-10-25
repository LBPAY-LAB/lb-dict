# SEC-001: Configuração mTLS para Bacen DICT

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: RSFN Bridge - mTLS Client Configuration
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: ARCHITECT (AI Agent - Technical Architect)
**Revisor**: [Aguardando]
**Aprovador**: Security Lead, Head de Arquitetura, DPO

---

## Sumário Executivo

Este documento especifica a configuração completa de **Mutual TLS (mTLS)** para comunicação segura entre o **RSFN Bridge** e a **API DICT/SPI do Banco Central (Bacen)**, incluindo certificados ICP-Brasil A3, validação de cadeia, cipher suites e troubleshooting.

**Baseado em**:
- [TEC-002 v3.1: Bridge Specification](../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [ANA-002: Análise Repositório Bridge](../00_Analises/ANA-002_Analise_Repo_Bridge.md)
- [REG-001: Requisitos Regulatórios Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)

---

## Controle de Versão

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 1.0 | 2025-10-25 | ARCHITECT | Versão inicial - mTLS Configuration |

---

## Índice

1. [Visão Geral mTLS](#1-visão-geral-mtls)
2. [Certificados Necessários](#2-certificados-necessários)
3. [Estrutura de Diretórios](#3-estrutura-de-diretórios)
4. [Configuração Go (Bridge)](#4-configuração-go-bridge)
5. [Cipher Suites e TLS Version](#5-cipher-suites-e-tls-version)
6. [Validação de Certificados](#6-validação-de-certificados)
7. [Renovação e Rotação](#7-renovação-e-rotação)
8. [Troubleshooting](#8-troubleshooting)
9. [Monitoring](#9-monitoring)

---

## 1. Visão Geral mTLS

### 1.1. O que é Mutual TLS?

**TLS Tradicional** (1-way):
```
Client               Server
  |                    |
  |--- Hello --------->|
  |<-- Certificate ----|  (só o servidor se autentica)
  |--- Encrypted ----->|
```

**Mutual TLS** (2-way):
```
Client (Bridge)      Server (Bacen)
  |                    |
  |--- Hello --------->|
  |<-- Certificate ----|  (servidor envia cert)
  |--- Certificate --->|  (cliente TAMBÉM envia cert)
  |<-- Verify ---------|  (servidor valida cliente)
  |--- Encrypted ----->|
```

### 1.2. Por Que mTLS com Bacen?

✅ **Autenticação Mútua**: Bacen valida que requisições vêm do LBPay
✅ **Não-Repúdio**: Certificado ICP-Brasil A3 garante identidade legal
✅ **Compliance**: Exigência regulatória do Bacen (Circular BCB)
✅ **Confidencialidade**: Encryption AES-256

---

## 2. Certificados Necessários

### 2.1. Certificado Cliente (ICP-Brasil A3)

**Tipo**: Certificado Digital A3 (hardware)
**Emissor**: Autoridade Certificadora credenciada ICP-Brasil
**Subject**:
```
C=BR
O=LB Pagamentos Ltda
OU=DICT
CN=api.lbpay.com.br
serialNumber=<CNPJ-LBPay>
```

**Características**:
- **Key Type**: RSA
- **Key Size**: 2048 bits (mínimo), 4096 bits (recomendado)
- **Signature Algorithm**: SHA-256 with RSA Encryption
- **Validity**: 1-3 anos
- **Extended Key Usage**: Client Authentication
- **Storage**: HSM (Hardware Security Module) ou Token A3

**Formato de Arquivos**:
- `client-cert.pem` - Certificado em PEM format
- `client-key.pem` - Chave privada em PEM format (encrypted)
- `client-cert.p12` - Certificado + chave em PKCS#12 (backup)

### 2.2. CA Chain (Cadeia de Certificação)

```
┌───────────────────────────────┐
│   Root CA ICP-Brasil          │  ← ac-raiz-icpbrasil.pem
│   (AC Raiz ICP-Brasil v10)    │
└──────────────┬────────────────┘
               │
┌──────────────▼────────────────┐
│   Intermediate CA             │  ← ac-intermediate.pem
│   (ex: Serasa, Certisign)     │
└──────────────┬────────────────┘
               │
┌──────────────▼────────────────┐
│   Client Certificate          │  ← client-cert.pem
│   (LB Pagamentos)             │
└───────────────────────────────┘
```

**Arquivo CA Chain**:
```bash
# ca-chain.pem (concatenação)
cat ac-intermediate.pem ac-raiz-icpbrasil.pem > ca-chain.pem
```

### 2.3. Root CA Bacen (Server)

**Descrição**: Certificado raiz do Bacen para validar o servidor DICT/SPI.

**Download**:
```bash
# URL oficial Bacen (verificar URL real)
curl -o bacen-root-ca.pem https://www.bcb.gov.br/certs/dict-root-ca.pem
```

**Validação**:
```bash
# Verificar fingerprint SHA-256
openssl x509 -in bacen-root-ca.pem -fingerprint -sha256 -noout

# Deve bater com fingerprint oficial publicado pelo Bacen
```

---

## 3. Estrutura de Diretórios

### 3.1. Localização dos Certificados

```
/etc/ssl/certs/bacen/
├── client/
│   ├── client-cert.pem          # Certificado ICP-Brasil A3
│   ├── client-key.pem           # Chave privada (encrypted)
│   ├── ca-chain.pem             # Intermediate + Root ICP-Brasil
│   └── client-cert.p12          # Backup PKCS#12
│
├── server/
│   └── bacen-root-ca.pem        # Root CA do Bacen
│
└── backup/
    ├── client-cert-2024.pem     # Certificados expirados (histórico)
    └── client-cert-2025.pem
```

### 3.2. Permissões

```bash
# Root owns certificates
chown -R root:root /etc/ssl/certs/bacen/

# Client cert readable only by Bridge app
chmod 644 /etc/ssl/certs/bacen/client/client-cert.pem

# Private key MUITO restrito
chmod 400 /etc/ssl/certs/bacen/client/client-key.pem

# CA certs públicos
chmod 644 /etc/ssl/certs/bacen/client/ca-chain.pem
chmod 644 /etc/ssl/certs/bacen/server/bacen-root-ca.pem
```

### 3.3. Environment Variables

```bash
# .env (NUNCA commitar em Git!)
MTLS_CLIENT_CERT_PATH=/etc/ssl/certs/bacen/client/client-cert.pem
MTLS_CLIENT_KEY_PATH=/etc/ssl/certs/bacen/client/client-key.pem
MTLS_CLIENT_KEY_PASSWORD=<senha-chave-privada>  # Se encrypted
MTLS_CA_CHAIN_PATH=/etc/ssl/certs/bacen/client/ca-chain.pem
MTLS_BACEN_ROOT_CA_PATH=/etc/ssl/certs/bacen/server/bacen-root-ca.pem
```

---

## 4. Configuração Go (Bridge)

### 4.1. Código mTLS Client

```go
// infrastructure/mtls/client.go
package mtls

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

type MTLSConfig struct {
    ClientCertPath     string
    ClientKeyPath      string
    ClientKeyPassword  string
    CAChainPath        string
    BacenRootCAPath    string
}

// NewMTLSClient cria HTTP client com mTLS configurado
func NewMTLSClient(cfg MTLSConfig) (*http.Client, error) {
    // 1. Carregar certificado cliente + chave privada
    cert, err := loadClientCertificate(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to load client certificate: %w", err)
    }

    // 2. Carregar CA pool (Bacen root CA)
    caCertPool, err := loadCACertPool(cfg.BacenRootCAPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load CA cert pool: %w", err)
    }

    // 3. Configurar TLS
    tlsConfig := &tls.Config{
        // Client certificate
        Certificates: []tls.Certificate{cert},

        // Root CAs para validar servidor (Bacen)
        RootCAs: caCertPool,

        // TLS version mínima
        MinVersion: tls.VersionTLS12,

        // Cipher suites permitidos (ver seção 5)
        CipherSuites: getAllowedCipherSuites(),

        // IMPORTANTE: Sempre validar certificado do servidor
        InsecureSkipVerify: false,

        // Server Name Indication (SNI)
        ServerName: "dict.bcb.gov.br",  // Verificar hostname real
    }

    // 4. Criar HTTP client com transport customizado
    transport := &http.Transport{
        TLSClientConfig:     tlsConfig,
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     90,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,  // Timeout global
    }, nil
}

// loadClientCertificate carrega certificado + chave privada
func loadClientCertificate(cfg MTLSConfig) (tls.Certificate, error) {
    // Se chave está encriptada com senha
    if cfg.ClientKeyPassword != "" {
        return loadEncryptedCertificate(
            cfg.ClientCertPath,
            cfg.ClientKeyPath,
            cfg.ClientKeyPassword,
        )
    }

    // Chave não encriptada (menos seguro)
    return tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
}

// loadEncryptedCertificate carrega certificado com chave privada encrypted
func loadEncryptedCertificate(certPath, keyPath, password string) (tls.Certificate, error) {
    certPEM, err := ioutil.ReadFile(certPath)
    if err != nil {
        return tls.Certificate{}, err
    }

    keyPEM, err := ioutil.ReadFile(keyPath)
    if err != nil {
        return tls.Certificate{}, err
    }

    // Decrypt chave privada
    keyDER, err := x509.DecryptPEMBlock(keyPEM, []byte(password))
    if err != nil {
        return tls.Certificate{}, fmt.Errorf("failed to decrypt private key: %w", err)
    }

    // Parse private key
    privateKey, err := x509.ParsePKCS1PrivateKey(keyDER)
    if err != nil {
        // Tentar PKCS8
        privateKeyPKCS8, err := x509.ParsePKCS8PrivateKey(keyDER)
        if err != nil {
            return tls.Certificate{}, err
        }
        privateKey = privateKeyPKCS8.(*rsa.PrivateKey)
    }

    // Construir certificate
    return tls.Certificate{
        Certificate: [][]byte{certPEM},
        PrivateKey:  privateKey,
    }, nil
}

// loadCACertPool carrega Root CA do Bacen
func loadCACertPool(bacenRootCAPath string) (*x509.CertPool, error) {
    caCert, err := ioutil.ReadFile(bacenRootCAPath)
    if err != nil {
        return nil, err
    }

    caCertPool := x509.NewCertPool()
    if !caCertPool.AppendCertsFromPEM(caCert) {
        return nil, fmt.Errorf("failed to append CA cert to pool")
    }

    return caCertPool, nil
}
```

### 4.2. Uso no Bridge

```go
// cmd/bridge/main.go
package main

import (
    "log"
    "github.com/lbpay/rsfn-bridge/infrastructure/mtls"
)

func main() {
    // Configuração mTLS
    mtlsConfig := mtls.MTLSConfig{
        ClientCertPath:    os.Getenv("MTLS_CLIENT_CERT_PATH"),
        ClientKeyPath:     os.Getenv("MTLS_CLIENT_KEY_PATH"),
        ClientKeyPassword: os.Getenv("MTLS_CLIENT_KEY_PASSWORD"),
        BacenRootCAPath:   os.Getenv("MTLS_BACEN_ROOT_CA_PATH"),
    }

    // Criar HTTP client com mTLS
    httpClient, err := mtls.NewMTLSClient(mtlsConfig)
    if err != nil {
        log.Fatalf("Failed to create mTLS client: %v", err)
    }

    // Usar client para chamar Bacen
    resp, err := httpClient.Post(
        "https://dict.bcb.gov.br/api/v1/entries",
        "application/xml",
        soapPayload,
    )
    if err != nil {
        log.Fatalf("Failed to call Bacen: %v", err)
    }
    defer resp.Body.Close()

    // Processar resposta...
}
```

---

## 5. Cipher Suites e TLS Version

### 5.1. TLS Version

```go
// Apenas TLS 1.2 e 1.3
MinVersion: tls.VersionTLS12
MaxVersion: tls.VersionTLS13  // Se Bacen suportar
```

**Justificativa**:
- TLS 1.0/1.1: DEPRECATED (inseguro)
- TLS 1.2: Padrão atual
- TLS 1.3: Mais seguro, mas verificar suporte Bacen

### 5.2. Cipher Suites Permitidos

```go
func getAllowedCipherSuites() []uint16 {
    return []uint16{
        // ECDHE (Forward Secrecy) + AES-GCM (AEAD)
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,

        // Fallback (se Bacen não suportar ECDHE)
        tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
    }
}
```

**Ordem de Preferência**:
1. ECDHE (Elliptic Curve Diffie-Hellman Ephemeral) → Forward Secrecy
2. AES-256-GCM → Encryption forte + AEAD
3. SHA-384 → Hash forte

**❌ Cipher Suites NÃO Permitidos**:
- `TLS_RSA_WITH_3DES_EDE_CBC_SHA` (3DES fraco)
- `TLS_RSA_WITH_RC4_128_SHA` (RC4 inseguro)
- Qualquer cipher sem Forward Secrecy

---

## 6. Validação de Certificados

### 6.1. Validações Automáticas (TLS Handshake)

```go
tlsConfig := &tls.Config{
    // Validações padrão (GO faz automaticamente):
    // ✅ Certificado não expirado
    // ✅ Assinatura válida pela CA
    // ✅ Hostname match (SNI)
    // ✅ Key usage correto

    InsecureSkipVerify: false,  // NUNCA true em produção!
}
```

### 6.2. Validações Adicionais (Custom)

```go
// VerifyPeerCertificate: callback customizado
tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
    // 1. Verificar Organization
    for _, chain := range verifiedChains {
        cert := chain[0]
        if !strings.Contains(cert.Subject.Organization[0], "Banco Central do Brasil") {
            return fmt.Errorf("invalid organization: %v", cert.Subject.Organization)
        }
    }

    // 2. Verificar Extended Key Usage
    cert, _ := x509.ParseCertificate(rawCerts[0])
    validEKU := false
    for _, eku := range cert.ExtKeyUsage {
        if eku == x509.ExtKeyUsageServerAuth {
            validEKU = true
            break
        }
    }
    if !validEKU {
        return fmt.Errorf("certificate missing ServerAuth EKU")
    }

    // 3. Verificar expiration < 30 dias (alerta)
    if time.Until(cert.NotAfter) < 30*24*time.Hour {
        log.Warn("Bacen certificate expiring soon!", "not_after", cert.NotAfter)
    }

    return nil
}
```

### 6.3. Validação CRL (Certificate Revocation List)

```bash
# Download CRL do ICP-Brasil
curl -o icp-brasil.crl http://acraiz.icpbrasil.gov.br/LCRacraizv10.crl

# Verificar se certificado foi revogado
openssl crl -in icp-brasil.crl -noout -text | grep -A2 "Serial Number"
```

**Automatização** (cronjob):
```bash
# /etc/cron.daily/check-crl
#!/bin/bash
curl -o /tmp/icp-brasil.crl http://acraiz.icpbrasil.gov.br/LCRacraizv10.crl
openssl crl -in /tmp/icp-brasil.crl -noout -text | grep $(openssl x509 -in /etc/ssl/certs/bacen/client/client-cert.pem -serial -noout | cut -d= -f2)
if [ $? -eq 0 ]; then
    echo "ALERT: Client certificate was REVOKED!" | mail -s "Certificate Revoked" security@lbpay.com
fi
```

---

## 7. Renovação e Rotação

### 7.1. Processo de Renovação

**Quando renovar**:
- 30 dias antes da expiração (automático)
- Certificado comprometido (imediato)
- Mudança de razão social (imediato)

**Steps**:
1. **Solicitar novo certificado** à AC (Autoridade Certificadora)
2. **Validar novo certificado**:
   ```bash
   openssl x509 -in new-client-cert.pem -text -noout
   ```
3. **Deploy em staging** primeiro
4. **Testar conectividade** com Bacen em staging
5. **Deploy em produção** (blue-green deployment)
6. **Manter certificado antigo** por 7 dias (rollback)
7. **Arquivar certificado antigo** em `/etc/ssl/certs/bacen/backup/`

### 7.2. Rotação sem Downtime

**Estratégia**: Suportar 2 certificados simultâneos temporariamente

```go
// Carregar ambos certificados durante transição
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{
        newCert,  // Certificado novo (preferido)
        oldCert,  // Certificado antigo (fallback por 7 dias)
    },
    // ...
}
```

### 7.3. Alertas de Expiração

```go
// Monitoring: alertar 60, 30, 7 dias antes
func checkCertificateExpiration(certPath string) {
    cert, _ := loadCertificate(certPath)
    daysUntilExpiry := time.Until(cert.NotAfter).Hours() / 24

    if daysUntilExpiry < 60 {
        log.Warn("Certificate expiring in %d days", daysUntilExpiry)
        sendAlert("Certificate expiring soon", certPath, daysUntilExpiry)
    }
}
```

---

## 8. Troubleshooting

### 8.1. Erros Comuns

#### Erro: "tls: bad certificate"

**Causa**: Certificado inválido, expirado ou não confiado
**Solução**:
```bash
# Verificar validade
openssl x509 -in client-cert.pem -noout -dates

# Verificar issuer
openssl x509 -in client-cert.pem -noout -issuer

# Verificar cadeia
openssl verify -CAfile ca-chain.pem client-cert.pem
```

#### Erro: "x509: certificate signed by unknown authority"

**Causa**: CA chain incompleta ou Root CA não confiada
**Solução**:
```bash
# Adicionar Intermediate CA ao ca-chain.pem
cat ac-intermediate.pem >> ca-chain.pem

# Verificar novamente
openssl verify -CAfile ca-chain.pem client-cert.pem
# Deve retornar: client-cert.pem: OK
```

#### Erro: "tls: handshake failure"

**Causa**: Cipher suites incompatíveis ou TLS version mismatch
**Solução**:
```bash
# Testar handshake com openssl
openssl s_client -connect dict.bcb.gov.br:443 \
    -cert client-cert.pem \
    -key client-key.pem \
    -CAfile bacen-root-ca.pem \
    -tls1_2

# Ver cipher suite negociado
# Adicionar cipher suite ao código se necessário
```

### 8.2. Debug TLS Handshake

```bash
# Habilitar debug TLS em Go
export GODEBUG=x509ignoreCN=0,tlsrecord=1

# Logs vão mostrar cada step do handshake
```

---

## 9. Monitoring

### 9.1. Métricas

```prometheus
# Prometheus metrics
bridge_mtls_handshake_duration_seconds{endpoint="bacen"}
bridge_mtls_handshake_errors_total{error_type="certificate_expired"}
bridge_mtls_certificate_expiry_days{cert="client"}
bridge_mtls_connection_active{endpoint="bacen"}
```

### 9.2. Alertas (Grafana)

```yaml
# Alert: Certificado expirando
- alert: CertificateExpiringSoon
  expr: bridge_mtls_certificate_expiry_days < 30
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "mTLS certificate expiring in {{ $value }} days"

# Alert: Handshake failures
- alert: MTLSHandshakeFailures
  expr: rate(bridge_mtls_handshake_errors_total[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High rate of mTLS handshake failures"
```

---

## Próximas Revisões

**Pendências**:
- [ ] Validar hostname real da API Bacen DICT (dict.bcb.gov.br?)
- [ ] Testar cipher suites aceitos pelo Bacen
- [ ] Implementar OCSP stapling (validação online de revogação)
- [ ] Configurar HSM para armazenar chave privada

---

**Referências**:
- [TEC-002 v3.1: Bridge Specification](../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md)
- [SEC-002: ICP-Brasil Certificates](SEC-002_ICP_Brasil_Certificates.md) (pendente)
- [REG-001: Requisitos Regulatórios Bacen](../06_Regulatorio/REG-001_Requisitos_Regulatorios_Bacen.md)
- [RFC 5246: TLS 1.2](https://datatracker.ietf.org/doc/html/rfc5246)
- [RFC 8446: TLS 1.3](https://datatracker.ietf.org/doc/html/rfc8446)

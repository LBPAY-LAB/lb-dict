# SEC-006: XML Signature Security

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: ARCHITECT + Security Lead

---

## 📋 Resumo Executivo

Este documento especifica a **assinatura digital de mensagens XML SOAP** enviadas pelo Bridge ao Bacen DICT, usando certificados ICP-Brasil A3 conforme exigido pela regulação.

**Objetivo**: Garantir integridade, autenticidade e não-repúdio das mensagens SOAP trocadas com o Bacen, utilizando XML Signature (XMLDSig) com algoritmos aprovados.

---

## 🎯 Contexto Regulatório

### Por Que Assinatura XML?

**Exigência Bacen**: Todas as mensagens SOAP enviadas ao DICT Bacen devem ser assinadas digitalmente com certificado ICP-Brasil A3 (Resolução BCB nº 4.985/2021).

**Objetivos**:
1. **Integridade**: Garantir que mensagem não foi alterada em trânsito
2. **Autenticidade**: Provar que mensagem foi enviada por LB Pagamentos
3. **Não-Repúdio**: Impossibilitar negação de envio (auditoria)

---

## 🔐 Padrões e Especificações

### Standards Utilizados

| Padrão | Versão | Descrição |
|--------|--------|-----------|
| **XML Signature** | 1.0 (W3C) | Estrutura de assinatura digital XML |
| **Canonicalization** | C14N 1.0 | Normalização de XML antes de assinar |
| **Digest Algorithm** | SHA-256 | Hash da mensagem |
| **Signature Algorithm** | RSA-SHA256 | Assinatura RSA com SHA-256 |
| **Key Info** | X.509 Certificate | Inclusão do certificado na assinatura |

### Algoritmos Aprovados pelo Bacen

#### Digest (Hash)
- ✅ **SHA-256** (recomendado) - `http://www.w3.org/2001/04/xmlenc#sha256`
- ✅ SHA-384 - `http://www.w3.org/2001/04/xmlenc#sha384`
- ✅ SHA-512 - `http://www.w3.org/2001/04/xmlenc#sha512`
- ❌ SHA-1 (descontinuado, inseguro)

#### Signature
- ✅ **RSA-SHA256** (recomendado) - `http://www.w3.org/2001/04/xmldsig-more#rsa-sha256`
- ✅ RSA-SHA384 - `http://www.w3.org/2001/04/xmldsig-more#rsa-sha384`
- ✅ RSA-SHA512 - `http://www.w3.org/2001/04/xmldsig-more#rsa-sha512`

#### Canonicalization
- ✅ **Exclusive C14N** (recomendado) - `http://www.w3.org/2001/10/xml-exc-c14n#`
- ✅ Inclusive C14N - `http://www.w3.org/TR/2001/REC-xml-c14n-20010315`

---

## 📄 Estrutura XML Signature

### Template Completo

```xml
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:dict="http://www.bcb.gov.br/dict/v1">
  <soapenv:Header>
    <!-- Security header com assinatura -->
    <wsse:Security xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
                   xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
                   soapenv:mustUnderstand="1">

      <!-- Binary Security Token (certificado X.509) -->
      <wsse:BinarySecurityToken
        wsu:Id="X509Token"
        EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
        ValueType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3">
        <!-- Certificado ICP-Brasil A3 em Base64 -->
        MIIDXTCCAkWgAwIBAgIJAKL...
      </wsse:BinarySecurityToken>

      <!-- Assinatura Digital -->
      <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">

        <!-- Informações da assinatura -->
        <ds:SignedInfo>
          <!-- Algoritmo de canonicalização -->
          <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>

          <!-- Algoritmo de assinatura (RSA-SHA256) -->
          <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>

          <!-- Referência ao elemento assinado (Body) -->
          <ds:Reference URI="#Body">
            <!-- Transformações aplicadas -->
            <ds:Transforms>
              <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
            </ds:Transforms>

            <!-- Algoritmo de hash (SHA-256) -->
            <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>

            <!-- Hash do Body (calculado) -->
            <ds:DigestValue>j6lwx3rvEPO0vKtMup4NbeVu8nk=</ds:DigestValue>
          </ds:Reference>
        </ds:SignedInfo>

        <!-- Valor da assinatura digital (RSA) -->
        <ds:SignatureValue>
          MC0CFFrVLtRlk...
        </ds:SignatureValue>

        <!-- Informações da chave (certificado) -->
        <ds:KeyInfo>
          <wsse:SecurityTokenReference>
            <wsse:Reference URI="#X509Token"
              ValueType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3"/>
          </wsse:SecurityTokenReference>
        </ds:KeyInfo>
      </ds:Signature>

    </wsse:Security>
  </soapenv:Header>

  <soapenv:Body wsu:Id="Body">
    <!-- Mensagem DICT a ser assinada -->
    <dict:CreateEntryRequest>
      <KeyType>CPF</KeyType>
      <KeyValue>12345678900</KeyValue>
      <Account>
        <ISPB>00000000</ISPB>
        <AccountNumber>12345-6</AccountNumber>
      </Account>
    </dict:CreateEntryRequest>
  </soapenv:Body>
</soapenv:Envelope>
```

---

## 🛠️ Implementação em Go

### Biblioteca Recomendada

**Opção 1**: `github.com/russellhaering/goxmldsig` (XML Digital Signature for Go)
**Opção 2**: `github.com/beevik/etree` + crypto/x509 manual

**Recomendação**: `goxmldsig` (mais completo, suporta C14N e todos algoritmos)

---

### Fluxo de Assinatura

```go
// Pseudocódigo (especificação, NÃO implementar agora)
package xmlsigner

import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "io/ioutil"

    "github.com/beevik/etree"
    "github.com/russellhaering/goxmldsig"
)

// XMLSigner encapsula lógica de assinatura XML
type XMLSigner struct {
    privateKey  *rsa.PrivateKey
    certificate *x509.Certificate
}

// NewXMLSigner cria novo signer com certificado ICP-Brasil A3
func NewXMLSigner(certPath, keyPath string) (*XMLSigner, error) {
    // 1. Carregar certificado X.509
    certPEM, err := ioutil.ReadFile(certPath)
    if err != nil {
        return nil, err
    }

    certBlock, _ := pem.Decode(certPEM)
    cert, err := x509.ParseCertificate(certBlock.Bytes)
    if err != nil {
        return nil, err
    }

    // 2. Carregar chave privada RSA (do HSM via PKCS#11 em produção)
    keyPEM, err := ioutil.ReadFile(keyPath)
    if err != nil {
        return nil, err
    }

    keyBlock, _ := pem.Decode(keyPEM)
    key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
    if err != nil {
        return nil, err
    }

    return &XMLSigner{
        privateKey:  key,
        certificate: cert,
    }, nil
}

// SignSOAPMessage assina mensagem SOAP completa
func (s *XMLSigner) SignSOAPMessage(soapXML []byte) ([]byte, error) {
    // 1. Parsear XML
    doc := etree.NewDocument()
    if err := doc.ReadFromBytes(soapXML); err != nil {
        return nil, err
    }

    // 2. Encontrar elemento Body
    body := doc.FindElement("//Body")
    if body == nil {
        return nil, errors.New("SOAP Body not found")
    }

    // Adicionar atributo wsu:Id="Body" (necessário para referência)
    body.CreateAttr("wsu:Id", "Body")

    // 3. Criar contexto de assinatura
    signer, err := goxmldsig.NewSigningContext(s.privateKey)
    if err != nil {
        return nil, err
    }

    // Configurar algoritmos
    signer.Canonicalizer = goxmldsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
    signer.Hash = crypto.SHA256
    signer.SignatureMethod = goxmldsig.RSASHA256SignatureMethod

    // 4. Calcular digest do Body
    digest, err := signer.Digest(body)
    if err != nil {
        return nil, err
    }

    // 5. Criar elemento Signature
    signature, err := signer.CreateSignature(body)
    if err != nil {
        return nil, err
    }

    // 6. Adicionar certificado ao KeyInfo
    certBase64 := base64.StdEncoding.EncodeToString(s.certificate.Raw)
    keyInfo := signature.FindElement("//KeyInfo")

    securityTokenRef := keyInfo.CreateElement("wsse:SecurityTokenReference")
    binaryToken := securityTokenRef.CreateElement("wsse:BinarySecurityToken")
    binaryToken.CreateAttr("wsu:Id", "X509Token")
    binaryToken.CreateAttr("EncodingType", "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary")
    binaryToken.CreateAttr("ValueType", "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3")
    binaryToken.SetText(certBase64)

    // 7. Inserir Signature no Header/Security
    header := doc.FindElement("//Header")
    if header == nil {
        envelope := doc.FindElement("//Envelope")
        header = envelope.CreateElement("soapenv:Header")
    }

    security := header.CreateElement("wsse:Security")
    security.CreateAttr("soapenv:mustUnderstand", "1")
    security.AddChild(signature)

    // 8. Serializar XML assinado
    signedXML, err := doc.WriteToBytes()
    return signedXML, err
}

// VerifySOAPSignature valida assinatura de mensagem SOAP
func (s *XMLSigner) VerifySOAPSignature(signedXML []byte) (bool, error) {
    // 1. Parsear XML
    doc := etree.NewDocument()
    if err := doc.ReadFromBytes(signedXML); err != nil {
        return false, err
    }

    // 2. Encontrar Signature
    signature := doc.FindElement("//Signature")
    if signature == nil {
        return false, errors.New("Signature not found")
    }

    // 3. Extrair certificado do BinarySecurityToken
    tokenElement := doc.FindElement("//BinarySecurityToken")
    if tokenElement == nil {
        return false, errors.New("Certificate not found")
    }

    certBytes, err := base64.StdEncoding.DecodeString(tokenElement.Text())
    if err != nil {
        return false, err
    }

    cert, err := x509.ParseCertificate(certBytes)
    if err != nil {
        return false, err
    }

    // 4. Validar assinatura
    validator, err := goxmldsig.NewValidationContext(cert.PublicKey)
    if err != nil {
        return false, err
    }

    _, err = validator.Validate(doc, signature)
    return err == nil, err
}
```

---

## 🔄 Integração com HSM (Produção)

### Assinatura com Chave no HSM

Em produção, a chave privada **NUNCA** deve sair do HSM. Usar interface PKCS#11 para assinar.

```go
// Pseudocódigo para PKCS#11 (especificação)
package hsmsigner

import (
    "github.com/miekg/pkcs11"
)

// HSMSigner assina XML usando chave no HSM
type HSMSigner struct {
    ctx       *pkcs11.Ctx
    session   pkcs11.SessionHandle
    keyHandle pkcs11.ObjectHandle
}

// NewHSMSigner conecta ao HSM
func NewHSMSigner(libraryPath, pin, keyLabel string) (*HSMSigner, error) {
    // 1. Inicializar PKCS#11
    ctx := pkcs11.New(libraryPath)  // e.g., /usr/lib/softhsm/libsofthsm2.so
    if err := ctx.Initialize(); err != nil {
        return nil, err
    }

    // 2. Abrir sessão
    slots, err := ctx.GetSlotList(true)
    if err != nil {
        return nil, err
    }

    session, err := ctx.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
    if err != nil {
        return nil, err
    }

    // 3. Login
    if err := ctx.Login(session, pkcs11.CKU_USER, pin); err != nil {
        return nil, err
    }

    // 4. Encontrar chave privada por label
    template := []*pkcs11.Attribute{
        pkcs11.NewAttribute(pkcs11.CKA_LABEL, keyLabel),
        pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
    }

    if err := ctx.FindObjectsInit(session, template); err != nil {
        return nil, err
    }

    objects, _, err := ctx.FindObjects(session, 1)
    if err != nil {
        return nil, err
    }

    if err := ctx.FindObjectsFinal(session); err != nil {
        return nil, err
    }

    if len(objects) == 0 {
        return nil, errors.New("private key not found in HSM")
    }

    return &HSMSigner{
        ctx:       ctx,
        session:   session,
        keyHandle: objects[0],
    }, nil
}

// SignDigest assina digest SHA-256 usando HSM
func (h *HSMSigner) SignDigest(digest []byte) ([]byte, error) {
    // Configurar mecanismo RSA-SHA256
    mechanism := []*pkcs11.Mechanism{
        pkcs11.NewMechanism(pkcs11.CKM_SHA256_RSA_PKCS, nil),
    }

    // Assinar no HSM (chave privada nunca sai do HSM)
    if err := h.ctx.SignInit(h.session, mechanism, h.keyHandle); err != nil {
        return nil, err
    }

    signature, err := h.ctx.Sign(h.session, digest)
    if err != nil {
        return nil, err
    }

    return signature, nil
}
```

---

## 🧪 Validação e Testes

### Teste 1: Validar Assinatura Localmente

```go
// Pseudocódigo de teste
func TestXMLSignature(t *testing.T) {
    // 1. Carregar SOAP message de exemplo
    soapXML := []byte(`
        <soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
            <soapenv:Body wsu:Id="Body">
                <CreateEntryRequest>
                    <KeyType>CPF</KeyType>
                    <KeyValue>12345678900</KeyValue>
                </CreateEntryRequest>
            </soapenv:Body>
        </soapenv:Envelope>
    `)

    // 2. Assinar
    signer, err := NewXMLSigner("cert.pem", "key.pem")
    require.NoError(t, err)

    signedXML, err := signer.SignSOAPMessage(soapXML)
    require.NoError(t, err)

    // 3. Verificar assinatura
    valid, err := signer.VerifySOAPSignature(signedXML)
    require.NoError(t, err)
    assert.True(t, valid, "Signature should be valid")

    // 4. Validar estrutura do XML assinado
    doc := etree.NewDocument()
    doc.ReadFromBytes(signedXML)

    // Verificar presença de elementos obrigatórios
    assert.NotNil(t, doc.FindElement("//Signature"))
    assert.NotNil(t, doc.FindElement("//SignedInfo"))
    assert.NotNil(t, doc.FindElement("//SignatureValue"))
    assert.NotNil(t, doc.FindElement("//KeyInfo"))
    assert.NotNil(t, doc.FindElement("//BinarySecurityToken"))

    // Verificar algoritmos
    signatureMethod := doc.FindElement("//SignatureMethod")
    assert.Equal(t,
        "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
        signatureMethod.SelectAttr("Algorithm").Value,
    )

    digestMethod := doc.FindElement("//DigestMethod")
    assert.Equal(t,
        "http://www.w3.org/2001/04/xmlenc#sha256",
        digestMethod.SelectAttr("Algorithm").Value,
    )
}
```

### Teste 2: Validar com Bacen (Staging)

```bash
# Enviar mensagem assinada ao endpoint staging do Bacen
curl -X POST \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: CreateEntry" \
  --data @signed-message.xml \
  https://dict-staging.bcb.gov.br/api/v1/soap

# Verificar response:
# - HTTP 200: Assinatura válida
# - HTTP 400: Assinatura inválida ou certificado não ICP-Brasil
# - HTTP 500: Erro de processamento
```

---

## 🚨 Troubleshooting

### Problema 1: Assinatura Inválida (Bacen rejeita)

**Causas Possíveis**:
- Certificado não é ICP-Brasil A3
- Algoritmo de hash/assinatura não aprovado
- Canonicalization incorreta
- Namespace incorreto

**Diagnóstico**:
```bash
# 1. Validar certificado
openssl x509 -in cert.pem -text -noout | grep "Subject:"
# Verificar: serialNumber=CNPJ:XX.XXX.XXX/XXXX-XX

# 2. Validar algoritmos no XML
grep -E "(SignatureMethod|DigestMethod)" signed.xml
# Verificar:
# - SignatureMethod: rsa-sha256
# - DigestMethod: sha256

# 3. Testar assinatura localmente (sem enviar ao Bacen)
xmlsec1 --verify --pubkey-cert-pem cert.pem signed.xml
# Output esperado: "OK"
```

**Soluções**:
- Usar certificado ICP-Brasil A3 válido (ver SEC-002)
- Configurar `SignatureMethod` = `rsa-sha256`
- Configurar `DigestMethod` = `sha256`
- Usar Exclusive C14N (`http://www.w3.org/2001/10/xml-exc-c14n#`)

---

### Problema 2: Certificado Expirado/Revogado

**Erro do Bacen**: `SOAP-ENV:Client - Certificate expired`

**Diagnóstico**:
```bash
# Verificar validade do certificado
openssl x509 -in cert.pem -noout -dates
# Not Before: Jan  1 00:00:00 2024 GMT
# Not After : Dec 31 23:59:59 2025 GMT

# Verificar revogação (CRL/OCSP)
openssl ocsp -issuer ca-intermediate.pem -cert cert.pem \
  -url http://ocsp.certisign.com.br -CAfile ca-chain.pem
# Response: good (não revogado)
```

**Soluções**:
- Renovar certificado ICP-Brasil (ver SEC-002)
- Verificar se certificado não foi revogado
- Atualizar configuração do Bridge com novo certificado

---

### Problema 3: Namespace Incorreto

**Erro**: XML inválido ou assinatura não reconhecida

**Causa**: Namespaces SOAP/WS-Security incorretos

**Solução**: Usar namespaces exatos:
```xml
xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
xmlns:ds="http://www.w3.org/2000/09/xmldsig#"
```

---

## 🔒 Segurança

### Boas Práticas

1. ✅ **SEMPRE** validar assinatura de mensagens recebidas do Bacen
2. ✅ **NUNCA** aceitar mensagens não assinadas
3. ✅ **Validar certificado** do Bacen (chain até ICP-Brasil Root CA)
4. ✅ **Verificar expiração** do certificado antes de assinar
5. ✅ **Log de auditoria** de todas as assinaturas geradas/verificadas
6. ✅ **Rate limiting** em operações de assinatura (evitar DoS no HSM)

### Auditoria de Assinaturas

```go
// Pseudocódigo para logging de auditoria
type SignatureAuditLog struct {
    Timestamp     time.Time
    Operation     string  // "SIGN" ou "VERIFY"
    RequestID     string
    MessageDigest string  // SHA-256 do XML assinado
    Certificate   string  // Subject DN do certificado
    Valid         bool    // Resultado da verificação
    Error         string  // Erro (se houver)
}

func (s *XMLSigner) SignWithAudit(ctx context.Context, soapXML []byte, requestID string) ([]byte, error) {
    auditLog := &SignatureAuditLog{
        Timestamp: time.Now(),
        Operation: "SIGN",
        RequestID: requestID,
    }

    // Assinar
    signedXML, err := s.SignSOAPMessage(soapXML)
    if err != nil {
        auditLog.Error = err.Error()
        auditLog.Valid = false
        logAudit(auditLog)
        return nil, err
    }

    // Calcular digest do XML assinado
    hash := sha256.Sum256(signedXML)
    auditLog.MessageDigest = hex.EncodeToString(hash[:])
    auditLog.Certificate = s.certificate.Subject.String()
    auditLog.Valid = true

    // Gravar log de auditoria
    logAudit(auditLog)

    return signedXML, nil
}
```

---

## 📋 Checklist de Implementação

- [ ] Instalar biblioteca `goxmldsig` (ou equivalente)
- [ ] Carregar certificado ICP-Brasil A3 (ver SEC-002)
- [ ] Implementar `XMLSigner.SignSOAPMessage()`
- [ ] Implementar `XMLSigner.VerifySOAPSignature()`
- [ ] Configurar algoritmos: RSA-SHA256, SHA-256, Exclusive C14N
- [ ] Integrar com HSM via PKCS#11 (produção)
- [ ] Criar testes unitários de assinatura/verificação
- [ ] Testar com mensagens SOAP do Bacen (staging)
- [ ] Implementar logging de auditoria de assinaturas
- [ ] Configurar alertas para falhas de assinatura (> 1%)
- [ ] Documentar troubleshooting de erros comuns
- [ ] Validar performance (< 50ms por assinatura)

---

## 📚 Referências

### Documentos Internos
- [SEC-001: mTLS Configuration](SEC-001_mTLS_Configuration.md) - Configuração de transporte seguro
- [SEC-002: ICP-Brasil Certificates](SEC-002_ICP_Brasil_Certificates.md) - Gestão de certificados
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md) - SOAP Adapter
- [ANA-002: Análise Repo Bridge](../../00_Analises/ANA-002_Analise_Repo_Bridge.md) - JRE+JAR para assinatura

### Documentação Externa
- [XML Signature W3C Recommendation](https://www.w3.org/TR/xmldsig-core/)
- [WS-Security 1.1 Specification](http://docs.oasis-open.org/wss/v1.1/)
- [Exclusive XML Canonicalization](https://www.w3.org/TR/xml-exc-c14n/)
- [goxmldsig Library](https://github.com/russellhaering/goxmldsig)
- [PKCS#11 Specification](http://docs.oasis-open.org/pkcs11/pkcs11-base/v2.40/pkcs11-base-v2.40.html)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após testes com Bacen staging

---

**IMPORTANTE**: Este é um documento de **especificação técnica**. A implementação será feita pelos desenvolvedores em fase posterior, baseando-se neste documento.

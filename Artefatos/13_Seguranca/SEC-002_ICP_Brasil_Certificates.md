# SEC-002: ICP-Brasil Certificates Management

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: Security Lead + ARCHITECT

---

## 📋 Resumo Executivo

Este documento especifica a **gestão completa de certificados digitais ICP-Brasil** necessários para a comunicação segura entre o Bridge e o Bacen DICT via mTLS (Mutual TLS).

**Objetivo**: Documentar todo o ciclo de vida dos certificados ICP-Brasil A3 (solicitação, instalação, renovação, backup, revogação) para garantir conformidade regulatória e continuidade operacional.

---

## 🎯 Contexto Regulatório

### Por Que ICP-Brasil?

**Exigência Legal**: O Bacen exige certificados digitais ICP-Brasil para autenticação de instituições financeiras no DICT (Resolução BCB nº 4.985/2021).

**Níveis de Certificado**:
- **A1**: Chave privada em software (não aceito pelo Bacen)
- **A3**: Chave privada em hardware criptográfico (HSM ou Token) - **OBRIGATÓRIO**

---

## 🔐 Requisitos do Certificado

### Tipo de Certificado
- **Categoria**: e-CNPJ A3 (Pessoa Jurídica)
- **Finalidade**: Autenticação de servidor/cliente (Client Authentication)
- **Algoritmo**: RSA 2048 bits (mínimo) ou 4096 bits (recomendado)
- **Hash**: SHA-256
- **Validade**: 1 a 3 anos (recomendado: 1 ano para renovação frequente)

### Subject (DN - Distinguished Name)
```
CN=LB PAGAMENTOS LTDA
OU=DICT Bridge Service
O=LB PAGAMENTOS LTDA
L=São Paulo
ST=SP
C=BR
serialNumber=CNPJ:XX.XXX.XXX/XXXX-XX
```

### Extended Key Usage (EKU)
- **Client Authentication** (1.3.6.1.5.5.7.3.2) - **Obrigatório**
- **Server Authentication** (1.3.6.1.5.5.7.3.1) - Opcional

### Subject Alternative Names (SAN)
```
DNS:bridge.lbpay.com.br
DNS:dict-bridge-prod.lbpay.internal
```

---

## 📄 Autoridades Certificadoras Homologadas

### ACs Credenciadas ICP-Brasil

Escolher uma AC autorizada pela Infraestrutura de Chaves Públicas Brasileira:

| AC | Nível | Preço Aprox. | Suporte HSM |
|----|-------|--------------|-------------|
| **Certisign** | A3 | R$ 800-1.200/ano | ✅ Sim (nCipher, Thales) |
| **Serasa Experian** | A3 | R$ 700-1.000/ano | ✅ Sim (SafeNet, Utimaco) |
| **Valid Certificadora** | A3 | R$ 650-950/ano | ✅ Sim (nCipher) |
| **Soluti (Docusign)** | A3 | R$ 800-1.100/ano | ✅ Sim (SafeNet) |

**Recomendação**: Certisign ou Valid (maior aceitação em instituições financeiras)

---

## 🛠️ Processo de Solicitação

### Fase 1: Pré-Requisitos (2-3 dias)

#### Documentação Necessária
1. **Documentos da Empresa**:
   - CNPJ da LB Pagamentos
   - Contrato Social ou Estatuto atualizado
   - Documentos do representante legal (RG, CPF)
   - Comprovante de endereço da empresa (máximo 90 dias)

2. **Procuração** (se representante não for sócio):
   - Procuração com firma reconhecida
   - Poderes específicos para solicitar certificado digital

3. **Hardware Criptográfico**:
   - **Opção 1**: Token USB A3 (e.g., SafeNet eToken 5110, Gemalto IDPrime)
   - **Opção 2**: HSM (Hardware Security Module) - **Recomendado para produção**

#### Escolha do HSM (Produção)
**Opção Recomendada**: Cloud HSM (AWS CloudHSM ou Google Cloud HSM)

| Solução | Preço | FIPS 140-2 | HA |
|---------|-------|------------|-----|
| **AWS CloudHSM** | ~$1.60/hora/HSM | Level 3 | ✅ Multi-AZ |
| **Google Cloud HSM** | ~$1.45/hora | Level 3 | ✅ Regional |
| **Thales Luna HSM** (on-prem) | ~$15k + manutenção | Level 3 | ⚠️ Manual setup |

---

### Fase 2: Geração do CSR (Certificate Signing Request)

#### Opção A: Usando OpenSSL (para validação/staging)
```bash
# ATENÇÃO: Este é um exemplo de especificação
# NÃO executar em produção sem ajustar valores

# 1. Gerar chave privada RSA 4096 bits
openssl genrsa -out lbpay-bridge-private.key 4096

# 2. Criar arquivo de configuração CSR
cat > csr.conf <<EOF
[ req ]
default_bits       = 4096
distinguished_name = req_distinguished_name
req_extensions     = v3_req
prompt             = no

[ req_distinguished_name ]
C  = BR
ST = SP
L  = São Paulo
O  = LB PAGAMENTOS LTDA
OU = DICT Bridge Service
CN = LB PAGAMENTOS LTDA
serialNumber = CNPJ:XX.XXX.XXX/XXXX-XX

[ v3_req ]
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = bridge.lbpay.com.br
DNS.2 = dict-bridge-prod.lbpay.internal
EOF

# 3. Gerar CSR
openssl req -new -key lbpay-bridge-private.key -out lbpay-bridge.csr -config csr.conf

# 4. Verificar CSR
openssl req -in lbpay-bridge.csr -text -noout
```

#### Opção B: Usando HSM (Produção)
```bash
# Especificação para AWS CloudHSM (exemplo)

# 1. Criar chave no HSM
aws cloudhsm create-key \
  --key-spec RSA_4096 \
  --label "lbpay-bridge-dict-prod"

# 2. Gerar CSR usando chave do HSM
aws cloudhsm create-csr \
  --key-label "lbpay-bridge-dict-prod" \
  --subject "/C=BR/ST=SP/L=São Paulo/O=LB PAGAMENTOS LTDA/OU=DICT Bridge Service/CN=LB PAGAMENTOS LTDA" \
  --output-file lbpay-bridge.csr
```

---

### Fase 3: Submissão à AC (1-2 dias)

#### Passos
1. **Acesso ao Portal da AC** (e.g., Certisign, Valid)
2. **Upload do CSR** gerado
3. **Upload de Documentos**:
   - CNPJ, Contrato Social, RG/CPF representante legal
4. **Validação Presencial** (videoconferência com documento original)
5. **Aguardar Aprovação** (1-2 dias úteis)

#### Validação Presencial
**Obrigatório para A3**: Representante legal deve comparecer (presencialmente ou videoconferência) com:
- RG e CPF originais
- Documento da empresa (CNPJ)

---

### Fase 4: Instalação do Certificado (1 dia)

#### Opção A: Token USB A3
```bash
# 1. Baixar certificado emitido pela AC
# Arquivo recebido: lbpay-bridge.cer

# 2. Importar certificado no token (via software da AC)
# Interface gráfica da AC (e.g., Certisign Manager)

# 3. Validar instalação
pkcs11-tool --module /usr/lib/libeToken.so --list-objects
```

#### Opção B: Cloud HSM
```bash
# 1. Baixar certificado da AC
curl -o lbpay-bridge.cer https://ac.certisign.com.br/certs/123456

# 2. Importar certificado no CloudHSM
aws cloudhsm import-certificate \
  --key-label "lbpay-bridge-dict-prod" \
  --certificate-file lbpay-bridge.cer \
  --certificate-chain-file ca-chain.pem

# 3. Validar
aws cloudhsm describe-certificate --key-label "lbpay-bridge-dict-prod"
```

---

## 🔗 Chain of Trust (Cadeia de Certificação)

### Estrutura da Cadeia
```
Root CA (ICP-Brasil)
    ↓
Intermediate CA (AC - e.g., Certisign)
    ↓
End-Entity Certificate (LB Pagamentos e-CNPJ A3)
```

### Arquivos Necessários
1. **lbpay-bridge.key** - Chave privada (NUNCA compartilhar, manter no HSM)
2. **lbpay-bridge.cer** - Certificado da LB Pagamentos
3. **ca-intermediate.cer** - Certificado da AC intermediária
4. **ca-root-icp-brasil.cer** - Certificado raiz ICP-Brasil
5. **ca-chain.pem** - Cadeia completa (intermediate + root)

### Criando ca-chain.pem
```bash
# Concatenar certificados (ordem: intermediate → root)
cat ca-intermediate.cer ca-root-icp-brasil.cer > ca-chain.pem
```

### Baixar Certificados da ICP-Brasil
```bash
# Root CA ICP-Brasil v10
wget https://www.gov.br/iti/pt-br/assuntos/repositorio/ac-raiz/v10/ACRaizv10.crt -O ca-root-icp-brasil.cer

# Intermediate CA (depende da AC escolhida)
# Certisign:
wget https://www.certisign.com.br/repositorio/acertisignv5.crt -O ca-intermediate.cer
```

---

## 🔄 Renovação de Certificado

### Quando Renovar?
- **90 dias antes** do vencimento (recomendado)
- **30 dias antes** (mínimo)
- **Nunca** esperar expirar (causa downtime)

### Processo de Renovação
1. **Gerar novo CSR** (pode usar mesma chave privada ou gerar nova)
2. **Submeter à AC** (processo mais rápido, sem validação presencial se mesmo representante)
3. **Instalar novo certificado** no HSM
4. **Testar em staging** antes de produção
5. **Atualizar configuração do Bridge** (hot reload se suportado)
6. **Revogar certificado antigo** após validação

### Blue-Green Deployment para Renovação
```bash
# 1. Instalar novo certificado com label diferente
aws cloudhsm import-certificate \
  --key-label "lbpay-bridge-dict-prod-v2" \
  --certificate-file lbpay-bridge-new.cer

# 2. Atualizar configuração do Bridge (env var)
MTLS_CERT_LABEL="lbpay-bridge-dict-prod-v2"

# 3. Restart graceful do Bridge
kubectl rollout restart deployment/dict-bridge

# 4. Validar tráfego
# Monitorar logs de mTLS handshake

# 5. Remover certificado antigo (após 7 dias de validação)
aws cloudhsm delete-certificate --key-label "lbpay-bridge-dict-prod-v1"
```

---

## 💾 Backup e Disaster Recovery

### O Que Fazer Backup

#### 1. Chave Privada (CRÍTICO)
**Se usar HSM Cloud**:
- Backup automático da AWS/GCP (snapshots)
- Exportar backup criptografado (M of N split)

**Se usar Token USB**:
- **NUNCA** expor chave privada
- Ter **2 tokens** com mesmo certificado (hot spare)

#### 2. Certificados e CA Chain
```bash
# Criar backup criptografado
tar czf certs-backup-$(date +%Y%m%d).tar.gz \
  lbpay-bridge.cer \
  ca-chain.pem \
  ca-intermediate.cer \
  ca-root-icp-brasil.cer

# Criptografar backup
gpg --symmetric --cipher-algo AES256 certs-backup-*.tar.gz

# Armazenar em:
# - Vault (HashiCorp Vault, AWS Secrets Manager)
# - S3 bucket privado (encryption at rest)
# - Cofre físico (cópia em papel para ca-chain)
```

### Disaster Recovery Scenarios

#### Cenário 1: HSM Falha
**Solução**:
- Usar HSM replica (se CloudHSM Multi-AZ)
- Restaurar backup de HSM (M of N key shares)
- **SLA**: < 1 hora (automático se CloudHSM HA)

#### Cenário 2: Certificado Comprometido
**Ações Imediatas**:
1. **Revogar certificado** (portal da AC)
2. **Bloquear acesso** ao Bridge
3. **Gerar novo certificado** (processo acelerado, 4-8 horas)
4. **Instalar e validar**
5. **Comunicar Bacen** (obrigatório)

#### Cenário 3: Certificado Expirado (Não Renovado)
**Impacto**: ❌ Bridge NÃO consegue conectar ao Bacen (downtime total)
**Mitigação**:
- Alertas automáticos 90/60/30/7 dias antes
- Processo de renovação iniciado 90 dias antes
- Fallback: Certificado de contingência pré-aprovado

---

## 📊 Monitoramento e Alertas

### Métricas Essenciais

#### 1. Dias até Expiração
```prometheus
# Prometheus metric (implementar no Bridge)
cert_expiry_days{cert="lbpay-bridge-dict-prod"} 87

# Alert rule
alert: CertificateExpiringIn30Days
expr: cert_expiry_days < 30
severity: warning
```

#### 2. Validade da Cadeia
```bash
# Verificar chain completo
openssl verify -CAfile ca-chain.pem lbpay-bridge.cer
# Output esperado: lbpay-bridge.cer: OK
```

#### 3. Revocation Status (CRL/OCSP)
```bash
# Verificar se certificado foi revogado
openssl ocsp \
  -issuer ca-intermediate.cer \
  -cert lbpay-bridge.cer \
  -url http://ocsp.certisign.com.br \
  -CAfile ca-chain.pem
# Output esperado: Response: successful (0x0), Status: good
```

### Alertas Recomendados
| Métrica | Threshold | Severidade |
|---------|-----------|------------|
| Dias até expiração | < 90 dias | Info |
| Dias até expiração | < 30 dias | Warning |
| Dias até expiração | < 7 dias | Critical |
| Certificado revogado | Sim | Critical |
| Falha no handshake mTLS | > 5% requests | Critical |

---

## 🔒 Segurança da Chave Privada

### Regras de Ouro
1. ✅ **NUNCA** armazenar chave privada em plaintext
2. ✅ **SEMPRE** usar HSM para produção (FIPS 140-2 Level 3)
3. ✅ **NUNCA** exportar chave privada do HSM
4. ✅ **Limitar acesso** ao HSM (IAM roles, MFA)
5. ✅ **Auditar** todas as operações com chave (CloudTrail)

### Controle de Acesso (IAM - AWS CloudHSM)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudhsm:DescribeKey",
        "cloudhsm:GetPublicKey",
        "cloudhsm:Sign"
      ],
      "Resource": "arn:aws:cloudhsm:*:*:key/lbpay-bridge-dict-prod",
      "Condition": {
        "StringEquals": {
          "aws:RequestedRegion": "us-east-1"
        },
        "IpAddress": {
          "aws:SourceIp": ["10.0.0.0/8"]  # Apenas VPC privada
        }
      }
    },
    {
      "Effect": "Deny",
      "Action": "cloudhsm:ExportKey",
      "Resource": "*"
    }
  ]
}
```

---

## 🧪 Validação e Testes

### Teste 1: Validar Certificado Localmente
```bash
# 1. Verificar dados do certificado
openssl x509 -in lbpay-bridge.cer -text -noout

# Verificar:
# - Subject: CN=LB PAGAMENTOS LTDA
# - Issuer: CN=AC Certisign (ou outra AC)
# - Validity: Not Before / Not After
# - Extended Key Usage: TLS Web Client Authentication

# 2. Verificar chain
openssl verify -CAfile ca-chain.pem lbpay-bridge.cer
# Output: lbpay-bridge.cer: OK
```

### Teste 2: Testar mTLS com Bacen (Staging)
```bash
# Simular handshake mTLS
curl -v \
  --cert lbpay-bridge.cer \
  --key lbpay-bridge.key \
  --cacert ca-chain.pem \
  https://dict-staging.bcb.gov.br/health

# Verificar output:
# * TLSv1.2 (OUT), TLS handshake, Client Certificate (11):
# * SSL connection using TLSv1.2 / ECDHE-RSA-AES256-GCM-SHA384
```

### Teste 3: Validar com Bridge (Integração)
```go
// Pseudocódigo de teste (NÃO implementar agora)
func TestMTLSCertificate(t *testing.T) {
    // 1. Carregar certificado
    cert, err := tls.LoadX509KeyPair("lbpay-bridge.cer", "lbpay-bridge.key")
    require.NoError(t, err)

    // 2. Criar cliente TLS
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                Certificates: []tls.Certificate{cert},
            },
        },
    }

    // 3. Fazer request ao Bacen (staging)
    resp, err := client.Get("https://dict-staging.bcb.gov.br/api/v1/health")
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

---

## 📋 Checklist de Implementação

- [ ] Escolher AC credenciada ICP-Brasil (Certisign, Valid, Serasa)
- [ ] Provisionar HSM (AWS CloudHSM ou Google Cloud HSM)
- [ ] Reunir documentação (CNPJ, Contrato Social, RG/CPF representante)
- [ ] Gerar CSR (4096 bits RSA, SHA-256)
- [ ] Submeter CSR à AC
- [ ] Agendar validação presencial/videoconferência
- [ ] Aguardar emissão do certificado (1-2 dias)
- [ ] Baixar certificado e CA chain
- [ ] Importar certificado no HSM
- [ ] Criar ca-chain.pem (intermediate + root)
- [ ] Configurar Bridge para usar certificado (ver SEC-001)
- [ ] Testar mTLS handshake em staging
- [ ] Validar em produção (canary deployment)
- [ ] Configurar monitoramento de expiração
- [ ] Criar alertas (90/60/30/7 dias antes expiração)
- [ ] Documentar processo de renovação
- [ ] Criar backup criptografado de certificados
- [ ] Definir política de rotação (renovar 90 dias antes)

---

## 📚 Referências

### Documentos Internos
- [SEC-001: mTLS Configuration](SEC-001_mTLS_Configuration.md) - Como usar o certificado no Bridge
- [TEC-002 v3.1: Bridge Specification](../../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md) - Bridge DICT
- [REG-001: Regulatório Bacen](../../06_Regulatorio/REG-001_Regulatory_Compliance_Bacen_DICT.md) - Exigências regulatórias

### Documentação Externa
- [ICP-Brasil - Documentos da AC Raiz](https://www.gov.br/iti/pt-br/assuntos/repositorio)
- [Resolução BCB nº 4.985/2021](https://www.bcb.gov.br/estabilidadefinanceira/exibenormativo?tipo=Resolu%C3%A7%C3%A3o%20BCB&numero=4985) - Exigência de certificados
- [AWS CloudHSM Documentation](https://docs.aws.amazon.com/cloudhsm/)
- [OpenSSL Certificate Guide](https://www.openssl.org/docs/man1.1.1/man1/x509.html)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando aquisição do certificado)
**Próxima Revisão**: Após aquisição e instalação do certificado

---

**IMPORTANTE**: Este é um documento de **especificação técnica e operacional**. A aquisição e instalação do certificado deve ser feita pela equipe de Segurança e DevOps seguindo este documento.

# SEC-005: Network Security

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Versão**: 1.0
**Data**: 2025-10-25
**Status**: ✅ Especificação Completa
**Responsável**: DevOps Lead + Security Lead + Network Admin

---

## 📋 Resumo Executivo

Este documento especifica a **arquitetura de segurança de rede** para o sistema DICT, incluindo segmentação de rede, firewalls, Network Policies do Kubernetes, isolamento de tráfego, e proteção contra ataques DDoS.

**Objetivo**: Implementar defesa em profundidade (defense-in-depth) com múltiplas camadas de segurança de rede para proteger o sistema DICT contra ameaças externas e internas.

---

## 🎯 Princípios de Segurança de Rede

### 1. Zero Trust Network
- ❌ **NÃO confiar** em nenhuma rede (inclusive interna)
- ✅ **Sempre validar** identidade e autorização
- ✅ **Minimizar** superfície de ataque
- ✅ **Segmentar** redes (micro-segmentation)

### 2. Defense in Depth (Defesa em Profundidade)
```
Layer 7: Application (WAF, API Gateway rate limiting)
Layer 4-6: Transport/Session (Firewall, Network Policies)
Layer 3: Network (VPC, Subnets, Security Groups)
Layer 2: Data Link (Private VLANs)
Layer 1: Physical (Data center security)
```

### 3. Least Privilege Network Access
- Apenas portas necessárias abertas
- Tráfego default: DENY (whitelist approach)
- Comunicação entre serviços: explicitamente permitida

---

## 🏗️ Arquitetura de Rede

### Topologia Geral

```
                        Internet
                           │
                      [Cloudflare]
                       WAF + DDoS
                           │
                    ┌──────┴──────┐
                    │             │
              [Load Balancer]   [CDN]
              (Public subnet)
                    │
            ┌───────┴───────┐
            │               │
     [Ingress NGINX]   [API Gateway]
      (DMZ subnet)
            │
    ┌───────┼───────┐
    │       │       │
 [Core]  [Connect] [Bridge]
   │       │         │
   └───────┼─────────┘
           │
   [Private subnet]
           │
    ┌──────┼──────┐
    │      │      │
[PostgreSQL] [Redis] [Temporal]
           │
       [Pulsar]
           │
    [mTLS tunnel]
           │
     [Bacen DICT]
     (External)
```

---

## 🔒 Segmentação de Rede (VPC Subnets)

### AWS VPC Configuration

```yaml
VPC: dict-prod-vpc
CIDR: 10.0.0.0/16

Subnets:
  # Public subnet (internet-facing)
  - name: public-subnet-1a
    cidr: 10.0.1.0/24
    availability_zone: us-east-1a
    resources:
      - Load Balancer
      - NAT Gateway
      - Bastion Host (jump box)

  - name: public-subnet-1b
    cidr: 10.0.2.0/24
    availability_zone: us-east-1b
    resources:
      - Load Balancer (HA)

  # DMZ subnet (semi-trusted, ingress/API Gateway)
  - name: dmz-subnet-1a
    cidr: 10.0.10.0/24
    availability_zone: us-east-1a
    resources:
      - Ingress NGINX pods
      - API Gateway pods

  - name: dmz-subnet-1b
    cidr: 10.0.11.0/24
    availability_zone: us-east-1b
    resources:
      - Ingress NGINX pods (HA)

  # Application subnet (private, no internet)
  - name: app-subnet-1a
    cidr: 10.0.20.0/24
    availability_zone: us-east-1a
    resources:
      - Core DICT pods
      - Connect pods
      - Bridge pods

  - name: app-subnet-1b
    cidr: 10.0.21.0/24
    availability_zone: us-east-1b
    resources:
      - Core/Connect/Bridge pods (HA)

  # Data subnet (private, highly restricted)
  - name: data-subnet-1a
    cidr: 10.0.30.0/24
    availability_zone: us-east-1a
    resources:
      - PostgreSQL RDS
      - Redis ElastiCache
      - Temporal server

  - name: data-subnet-1b
    cidr: 10.0.31.0/24
    availability_zone: us-east-1b
    resources:
      - PostgreSQL replica
      - Redis replica
```

---

## 🛡️ Security Groups (AWS) / Firewall Rules

### Security Group: Load Balancer

```yaml
name: sg-dict-lb
description: Security group for public load balancer

inbound_rules:
  # HTTPS from internet
  - protocol: tcp
    port: 443
    source: 0.0.0.0/0
    description: HTTPS from internet

  # HTTP (redirect to HTTPS)
  - protocol: tcp
    port: 80
    source: 0.0.0.0/0
    description: HTTP redirect

outbound_rules:
  # To Ingress NGINX (DMZ)
  - protocol: tcp
    port: 80
    destination: sg-dict-ingress
    description: To Ingress NGINX
```

---

### Security Group: Ingress NGINX (DMZ)

```yaml
name: sg-dict-ingress
description: Security group for Ingress NGINX

inbound_rules:
  # From Load Balancer
  - protocol: tcp
    port: 80
    source: sg-dict-lb
    description: From Load Balancer

  # Health checks from Load Balancer
  - protocol: tcp
    port: 10254
    source: sg-dict-lb
    description: Health check

outbound_rules:
  # To Core DICT pods
  - protocol: tcp
    port: 8080
    destination: sg-dict-app
    description: To application pods
```

---

### Security Group: Application Pods (Core/Connect/Bridge)

```yaml
name: sg-dict-app
description: Security group for application pods

inbound_rules:
  # From Ingress NGINX
  - protocol: tcp
    port: 8080
    source: sg-dict-ingress
    description: HTTP from Ingress

  # gRPC between services (Core → Connect, Connect → Bridge)
  - protocol: tcp
    port: 8081
    source: sg-dict-app
    description: gRPC inter-service

  # Metrics (Prometheus scraping)
  - protocol: tcp
    port: 9090
    source: sg-monitoring
    description: Prometheus metrics

outbound_rules:
  # To PostgreSQL
  - protocol: tcp
    port: 5432
    destination: sg-dict-data
    description: PostgreSQL

  # To Redis
  - protocol: tcp
    port: 6379
    destination: sg-dict-data
    description: Redis

  # To Temporal
  - protocol: tcp
    port: 7233
    destination: sg-dict-data
    description: Temporal

  # To Pulsar
  - protocol: tcp
    port: 6650
    destination: sg-dict-data
    description: Pulsar

  # To Bacen DICT (HTTPS/mTLS)
  - protocol: tcp
    port: 443
    destination: 0.0.0.0/0
    description: Bacen DICT API
```

---

### Security Group: Data Layer (PostgreSQL, Redis, Temporal)

```yaml
name: sg-dict-data
description: Security group for databases and message brokers

inbound_rules:
  # PostgreSQL from app pods only
  - protocol: tcp
    port: 5432
    source: sg-dict-app
    description: PostgreSQL

  # Redis from app pods only
  - protocol: tcp
    port: 6379
    source: sg-dict-app
    description: Redis

  # Temporal from app pods only
  - protocol: tcp
    port: 7233
    source: sg-dict-app
    description: Temporal

  # Pulsar from app pods only
  - protocol: tcp
    port: 6650
    source: sg-dict-app
    description: Pulsar

outbound_rules:
  # PostgreSQL replication (entre replicas)
  - protocol: tcp
    port: 5432
    destination: sg-dict-data
    description: PostgreSQL replication

  # Redis replication
  - protocol: tcp
    port: 6379
    destination: sg-dict-data
    description: Redis replication
```

---

## ☸️ Kubernetes Network Policies

### Network Policy: Default Deny All

```yaml
# Deny all traffic by default (whitelist approach)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: dict-prod
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
```

---

### Network Policy: Core DICT

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: dict-core-netpol
  namespace: dict-prod
spec:
  podSelector:
    matchLabels:
      app: dict-core

  policyTypes:
    - Ingress
    - Egress

  ingress:
    # From Ingress NGINX
    - from:
      - namespaceSelector:
          matchLabels:
            name: ingress-nginx
      ports:
      - protocol: TCP
        port: 8080

    # Prometheus scraping
    - from:
      - namespaceSelector:
          matchLabels:
            name: monitoring
      ports:
      - protocol: TCP
        port: 9090

  egress:
    # To Connect (gRPC)
    - to:
      - podSelector:
          matchLabels:
            app: dict-connect
      ports:
      - protocol: TCP
        port: 8081

    # DNS (CoreDNS)
    - to:
      - namespaceSelector:
          matchLabels:
            name: kube-system
        podSelector:
          matchLabels:
            k8s-app: kube-dns
      ports:
      - protocol: UDP
        port: 53
```

---

### Network Policy: Connect

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: dict-connect-netpol
  namespace: dict-prod
spec:
  podSelector:
    matchLabels:
      app: dict-connect

  policyTypes:
    - Ingress
    - Egress

  ingress:
    # From Core DICT (gRPC)
    - from:
      - podSelector:
          matchLabels:
            app: dict-core
      ports:
      - protocol: TCP
        port: 8081

    # From Orchestration Worker (same app)
    - from:
      - podSelector:
          matchLabels:
            app: dict-connect

  egress:
    # To Bridge (gRPC)
    - to:
      - podSelector:
          matchLabels:
            app: dict-bridge
      ports:
      - protocol: TCP
        port: 8081

    # To PostgreSQL
    - to:
      - namespaceSelector:
          matchLabels:
            name: data
      ports:
      - protocol: TCP
        port: 5432

    # To Redis
    - to:
      - namespaceSelector:
          matchLabels:
            name: data
      ports:
      - protocol: TCP
        port: 6379

    # To Temporal
    - to:
      - namespaceSelector:
          matchLabels:
            name: temporal
      ports:
      - protocol: TCP
        port: 7233

    # To Pulsar
    - to:
      - namespaceSelector:
          matchLabels:
            name: pulsar
      ports:
      - protocol: TCP
        port: 6650

    # DNS
    - to:
      - namespaceSelector:
          matchLabels:
            name: kube-system
        podSelector:
          matchLabels:
            k8s-app: kube-dns
      ports:
      - protocol: UDP
        port: 53
```

---

### Network Policy: Bridge

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: dict-bridge-netpol
  namespace: dict-prod
spec:
  podSelector:
    matchLabels:
      app: dict-bridge

  policyTypes:
    - Ingress
    - Egress

  ingress:
    # From Connect (gRPC)
    - from:
      - podSelector:
          matchLabels:
            app: dict-connect
      ports:
      - protocol: TCP
        port: 8081

  egress:
    # To Bacen DICT (HTTPS/mTLS - external)
    - to:
      - namespaceSelector: {}  # Any namespace
      ports:
      - protocol: TCP
        port: 443

    # DNS
    - to:
      - namespaceSelector:
          matchLabels:
            name: kube-system
        podSelector:
          matchLabels:
            k8s-app: kube-dns
      ports:
      - protocol: UDP
        port: 53
```

---

## 🌐 WAF (Web Application Firewall)

### Cloudflare WAF

**Rules Implementadas**:

1. **Rate Limiting**
   ```
   - 100 requests/min por IP (API)
   - 1000 requests/min por IP (CDN assets)
   ```

2. **IP Reputation**
   ```
   - Bloquear IPs conhecidos por ataques (Cloudflare threat intelligence)
   - Bloquear países de alto risco (opcional)
   ```

3. **Bot Protection**
   ```
   - Challenge bots suspeitos (CAPTCHA)
   - Bloquear bad bots (scrapers, attack tools)
   ```

4. **OWASP Top 10 Protection**
   ```
   - SQL Injection
   - XSS (Cross-Site Scripting)
   - Command Injection
   - Path Traversal
   ```

5. **Custom Rules**
   ```yaml
   # Bloquear requests sem User-Agent
   - expression: http.user_agent eq ""
     action: block

   # Bloquear requests muito grandes (> 10MB)
   - expression: http.request.body.size > 10485760
     action: block

   # Permitir apenas métodos HTTP válidos
   - expression: http.request.method notin {"GET" "POST" "PUT" "PATCH" "DELETE" "OPTIONS"}
     action: block
   ```

---

## 🛑 DDoS Protection

### Layer 3/4 DDoS (Network/Transport)

**Cloudflare**:
- ✅ Absorção de tráfego DDoS (até 100+ Tbps)
- ✅ Mitigação automática de SYN flood, UDP flood, ICMP flood
- ✅ Anycast network (distribuição geográfica)

**AWS Shield Standard** (incluído):
- ✅ Proteção contra ataques comuns (SYN/ACK flood)
- ✅ Network ACLs automáticas

---

### Layer 7 DDoS (Application)

**Cloudflare Rate Limiting**:
```yaml
rate_limits:
  # API endpoints
  - match:
      request:
        url: "/api/v1/*"
    threshold: 100
    period: 60  # 100 req/min
    action: challenge  # CAPTCHA

  # Login endpoint (mais restritivo)
  - match:
      request:
        url: "/api/v1/auth/login"
    threshold: 5
    period: 60  # 5 req/min
    action: block
```

**API Gateway Rate Limiting** (adicional):
```yaml
# Kong, Traefik, ou NGINX rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;

location /api/v1/ {
    limit_req zone=api burst=20 nodelay;
}
```

---

## 🔐 TLS/SSL Configuration

### Ingress TLS Termination

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dict-ingress
  namespace: dict-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    nginx.ingress.kubernetes.io/ssl-ciphers: "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/hsts: "true"
    nginx.ingress.kubernetes.io/hsts-max-age: "31536000"
    nginx.ingress.kubernetes.io/hsts-include-subdomains: "true"
spec:
  tls:
    - hosts:
      - dict.lbpay.com.br
      - api.dict.lbpay.com.br
      secretName: dict-tls-cert
  rules:
    - host: api.dict.lbpay.com.br
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dict-core
                port:
                  number: 8080
```

---

## 🕵️ Network Monitoring

### 1. VPC Flow Logs (AWS)

```hcl
# Terraform configuration
resource "aws_flow_log" "dict_vpc" {
  vpc_id          = aws_vpc.dict_vpc.id
  traffic_type    = "ALL"  # ACCEPT, REJECT, or ALL
  iam_role_arn    = aws_iam_role.flow_logs.arn
  log_destination = aws_cloudwatch_log_group.flow_logs.arn

  tags = {
    Name = "dict-vpc-flow-logs"
  }
}
```

**Análise de Flow Logs**:
- Identificar conexões suspeitas
- Detectar port scanning
- Monitorar tráfego para IPs externos

---

### 2. Intrusion Detection (IDS)

**Falco** (Kubernetes runtime security):

```yaml
# Falco rules
- rule: Unauthorized network connection
  desc: Detect unauthorized outbound connections
  condition: >
    outbound and
    not allowed_destinations
  output: >
    Unauthorized connection
    (user=%user.name command=%proc.cmdline connection=%fd.name)
  priority: WARNING

- rule: Sensitive file access
  desc: Detect access to sensitive files
  condition: >
    open_read and
    sensitive_files
  output: >
    Sensitive file accessed
    (file=%fd.name user=%user.name)
  priority: CRITICAL
```

---

### 3. Network Metrics (Prometheus)

```prometheus
# Métricas de rede
node_network_receive_bytes_total{device="eth0"}
node_network_transmit_bytes_total{device="eth0"}

# Conexões ativas
node_netstat_Tcp_CurrEstab

# Erros de rede
rate(node_network_receive_errs_total[5m])
rate(node_network_transmit_errs_total[5m])

# Alertas
groups:
  - name: network
    rules:
      - alert: HighNetworkTraffic
        expr: rate(node_network_transmit_bytes_total[5m]) > 100000000  # > 100MB/s
        for: 5m
        labels:
          severity: warning

      - alert: NetworkErrors
        expr: rate(node_network_receive_errs_total[5m]) > 10
        for: 5m
        labels:
          severity: critical
```

---

## 🚨 Incident Response

### Breach Scenario: Port Scan Detectado

**Detecção**:
- VPC Flow Logs mostram conexões para múltiplas portas de um IP
- Falco alerta sobre comportamento suspeito

**Resposta**:
1. **Bloquear IP imediatamente** (Security Group ou WAF)
2. **Analisar logs** para identificar origem
3. **Verificar integridade** dos sistemas afetados
4. **Notificar equipe de segurança**

---

### Breach Scenario: DDoS Attack

**Detecção**:
- Cloudflare detecta spike de tráfego (10x normal)
- API Gateway rate limiting acionado

**Resposta**:
1. **Cloudflare mitigação automática** (challenge/block)
2. **Escalar recursos** (Auto Scaling Groups)
3. **Ativar "Under Attack Mode"** (Cloudflare)
4. **Monitorar disponibilidade** (SLA 99.9%)

---

## 📋 Checklist de Implementação

- [ ] Provisionar VPC com subnets públicas/privadas/DMZ/data
- [ ] Configurar Security Groups (AWS) ou Firewall rules (GCP)
- [ ] Implementar Network Policies do Kubernetes (default deny all)
- [ ] Configurar Ingress NGINX com TLS termination
- [ ] Configurar Cloudflare WAF (rate limiting, bot protection)
- [ ] Habilitar AWS Shield Standard
- [ ] Configurar VPC Flow Logs
- [ ] Instalar Falco (Kubernetes IDS)
- [ ] Configurar monitoramento de rede (Prometheus)
- [ ] Configurar alertas (network errors, high traffic)
- [ ] Documentar topologia de rede (diagramas)
- [ ] Criar runbook de incident response
- [ ] Realizar penetration testing (security audit)
- [ ] Validar egress firewall (apenas destinos permitidos)

---

## 📚 Referências

### Documentos Internos
- [SEC-001: mTLS Configuration](SEC-001_mTLS_Configuration.md)
- [SEC-004: API Authentication](SEC-004_API_Authentication.md)
- [DevOps Pipelines](../../15_DevOps/Pipelines/)
- [Diagramas de Arquitetura](../../02_Arquitetura/Diagramas/)

### Documentação Externa
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [AWS VPC Security Best Practices](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-best-practices.html)
- [Cloudflare WAF](https://www.cloudflare.com/waf/)
- [Falco Runtime Security](https://falco.org/docs/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

**Versão**: 1.0
**Status**: ✅ Especificação Completa (Aguardando implementação)
**Próxima Revisão**: Após setup de infraestrutura

---

**IMPORTANTE**: Este é um documento de **especificação técnica e de infraestrutura**. A implementação será feita pela equipe de DevOps/Network Admin em fase posterior, baseando-se neste documento.

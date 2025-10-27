#!/bin/bash
#
# generate-dev-certs.sh
#
# Generates self-signed certificates for mTLS in development mode
# For production, replace these with ICP-Brasil A3 certificates
#
# Usage: ./scripts/generate-dev-certs.sh
#

set -e

CERT_DIR="./certs/dev"
DAYS_VALID=365

echo "🔐 Generating self-signed certificates for mTLS (dev mode)..."

# Create certs directory
mkdir -p "$CERT_DIR"

# 1. Generate CA (Certificate Authority)
echo "📝 Step 1: Generating CA private key and certificate..."
openssl genrsa -out "$CERT_DIR/ca.key" 4096

openssl req -new -x509 -days "$DAYS_VALID" -key "$CERT_DIR/ca.key" -out "$CERT_DIR/ca.crt" \
  -subj "/C=BR/ST=SP/L=SaoPaulo/O=LBPay Lab/OU=Development/CN=LBPay CA"

# 2. Generate Server Certificate (Bridge)
echo "📝 Step 2: Generating server certificate (Bridge)..."
openssl genrsa -out "$CERT_DIR/server.key" 4096

openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
  -subj "/C=BR/ST=SP/L=SaoPaulo/O=LBPay Lab/OU=Development/CN=localhost"

openssl x509 -req -days "$DAYS_VALID" -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/server.crt"

rm "$CERT_DIR/server.csr"

# 3. Generate Client Certificate (Bridge -> Bacen)
echo "📝 Step 3: Generating client certificate (Bridge -> Bacen)..."
openssl genrsa -out "$CERT_DIR/client.key" 4096

openssl req -new -key "$CERT_DIR/client.key" -out "$CERT_DIR/client.csr" \
  -subj "/C=BR/ST=SP/L=SaoPaulo/O=LBPay Lab/OU=Development/CN=bridge-client"

openssl x509 -req -days "$DAYS_VALID" -in "$CERT_DIR/client.csr" \
  -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/client.crt"

rm "$CERT_DIR/client.csr"

# 4. Generate Bacen Simulator Certificate (for testing)
echo "📝 Step 4: Generating Bacen simulator certificate..."
openssl genrsa -out "$CERT_DIR/bacen.key" 4096

openssl req -new -key "$CERT_DIR/bacen.key" -out "$CERT_DIR/bacen.csr" \
  -subj "/C=BR/ST=DF/L=Brasilia/O=Banco Central/OU=DICT/CN=bacen-simulator"

openssl x509 -req -days "$DAYS_VALID" -in "$CERT_DIR/bacen.csr" \
  -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/bacen.crt"

rm "$CERT_DIR/bacen.csr"
rm "$CERT_DIR/ca.srl"

# 5. Set permissions
chmod 600 "$CERT_DIR"/*.key
chmod 644 "$CERT_DIR"/*.crt

echo ""
echo "✅ Certificates generated successfully in $CERT_DIR/"
echo ""
echo "📂 Generated files:"
echo "  - ca.crt / ca.key         (Certificate Authority)"
echo "  - server.crt / server.key (Bridge server)"
echo "  - client.crt / client.key (Bridge -> Bacen client)"
echo "  - bacen.crt / bacen.key   (Bacen simulator)"
echo ""
echo "⚠️  IMPORTANT: These are DEVELOPMENT certificates only!"
echo "   For PRODUCTION, use ICP-Brasil A3 certificates from a trusted CA."
echo ""
echo "🔍 To verify certificates:"
echo "   openssl x509 -in $CERT_DIR/server.crt -text -noout"
echo ""
echo "🚀 Update your .env file with:"
echo "   MTLS_ENABLED=true"
echo "   MTLS_CA_CERT_PATH=$CERT_DIR/ca.crt"
echo "   MTLS_CLIENT_CERT_PATH=$CERT_DIR/client.crt"
echo "   MTLS_CLIENT_KEY_PATH=$CERT_DIR/client.key"
echo ""
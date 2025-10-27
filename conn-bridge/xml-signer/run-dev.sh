#!/bin/bash
# Quick start script for XML Signer service (Development mode)

set -e

echo "=================================="
echo "XML Signer Service - Quick Start"
echo "=================================="
echo ""

# Check Java
if ! command -v java &> /dev/null; then
    echo "ERROR: Java is not installed. Please install Java 17 or higher."
    exit 1
fi

JAVA_VERSION=$(java -version 2>&1 | awk -F '"' '/version/ {print $2}' | cut -d'.' -f1)
if [ "$JAVA_VERSION" -lt 11 ]; then
    echo "WARNING: Java $JAVA_VERSION detected. Java 17+ is recommended."
fi

# Check Maven
if ! command -v mvn &> /dev/null; then
    echo "ERROR: Maven is not installed. Please install Maven 3.8 or higher."
    exit 1
fi

echo "1. Building application..."
mvn clean package -DskipTests

echo ""
echo "2. Generating test certificate..."
if [ ! -f "test-cert.p12" ]; then
    java -cp target/xml-signer-1.0.0.jar \
        com.lbpay.xmlsigner.util.CertificateGenerator \
        test-cert.p12 changeit test "LBPay Dev Certificate" 365
    echo "   ✓ Test certificate created: test-cert.p12"
else
    echo "   ✓ Test certificate already exists: test-cert.p12"
fi

echo ""
echo "3. Starting service in development mode..."
echo "   Port: 8080"
echo "   Mode: DEV (accepts self-signed certificates)"
echo ""
echo "Press Ctrl+C to stop"
echo ""

export DEV_MODE=true
java -jar target/xml-signer-1.0.0.jar

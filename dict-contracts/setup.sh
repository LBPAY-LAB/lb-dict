#!/bin/bash

# Script de setup para dict-contracts
# Autor: api-specialist
# Data: 2025-10-26

set -e

echo "=========================================="
echo "DICT Contracts - Setup Script"
echo "=========================================="
echo ""

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado. Por favor, instale Go primeiro."
    echo "   https://golang.org/doc/install"
    exit 1
fi

echo "✅ Go encontrado: $(go version)"
echo ""

# Verificar se protoc está instalado
if ! command -v protoc &> /dev/null; then
    echo "⚠️  protoc não está instalado."
    echo ""
    echo "Instalando protoc..."

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install protobuf
        else
            echo "❌ Homebrew não encontrado. Instale manualmente:"
            echo "   https://github.com/protocolbuffers/protobuf/releases"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        sudo apt-get update
        sudo apt-get install -y protobuf-compiler
    else
        echo "❌ Sistema operacional não suportado automaticamente."
        echo "   Instale protoc manualmente:"
        echo "   https://github.com/protocolbuffers/protobuf/releases"
        exit 1
    fi
fi

echo "✅ protoc encontrado: $(protoc --version)"
echo ""

# Instalar plugins Go
echo "Instalando plugins Go..."
echo ""

echo "1. Instalando protoc-gen-go..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

echo "2. Instalando protoc-gen-go-grpc..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo ""
echo "✅ Plugins instalados com sucesso!"
echo ""

# Verificar se os plugins estão no PATH
GOPATH=$(go env GOPATH)
if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
    echo "⚠️  ATENÇÃO: $GOPATH/bin não está no seu PATH"
    echo ""
    echo "Adicione ao seu ~/.bashrc ou ~/.zshrc:"
    echo "  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
    echo ""
fi

# Criar diretório gen se não existir
mkdir -p gen

echo "=========================================="
echo "Setup concluído com sucesso!"
echo "=========================================="
echo ""
echo "Próximos passos:"
echo "  1. Gerar código: make proto-gen"
echo "  2. Validar proto: make proto-lint"
echo ""
echo "Para ver todos os comandos: make help"
echo ""

.PHONY: help proto-gen proto-lint proto-breaking proto-clean proto-deps test test-unit coverage

# Default target
help:
	@echo "DICT Contracts - Makefile"
	@echo ""
	@echo "Comandos disponíveis:"
	@echo "  make proto-gen       - Gerar código Go a partir dos proto files"
	@echo "  make proto-lint      - Validar proto files com buf lint"
	@echo "  make proto-breaking  - Verificar breaking changes"
	@echo "  make proto-clean     - Limpar código gerado"
	@echo "  make proto-deps      - Instalar dependências (protoc plugins)"
	@echo "  make test            - Executar testes"
	@echo "  make test-unit       - Executar apenas testes unitários"
	@echo "  make coverage        - Gerar relatório de cobertura"

# Gerar código Go a partir dos proto files
proto-gen:
	@echo "Gerando código Go..."
	@mkdir -p gen/proto/common/v1 gen/proto/core/v1 gen/proto/bridge/v1
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--proto_path=. \
		proto/*.proto
	@mv proto/*.pb.go gen/proto/ 2>/dev/null || true
	@mv gen/proto/common.pb.go gen/proto/common/v1/ 2>/dev/null || true
	@mv gen/proto/core_dict*.pb.go gen/proto/core/v1/ 2>/dev/null || true
	@mv gen/proto/bridge*.pb.go gen/proto/bridge/v1/ 2>/dev/null || true
	@echo "Código gerado com sucesso em gen/"
	@echo "Executando go mod tidy..."
	@go mod tidy
	@echo "Build verificado com sucesso!"

# Lint dos arquivos proto
proto-lint:
	@echo "Executando buf lint..."
	buf lint

# Verificar breaking changes
proto-breaking:
	@echo "Verificando breaking changes..."
	buf breaking --against '.git#branch=main'

# Limpar código gerado
proto-clean:
	@echo "Limpando código gerado..."
	rm -rf gen/*
	@echo "Código gerado removido"

# Instalar dependências
proto-deps:
	@echo "Instalando dependências..."
	@echo "1. Instalando protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "2. Instalando protoc-gen-go-grpc..."
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo ""
	@echo "Dependências instaladas!"
	@echo "Certifique-se de que $(go env GOPATH)/bin está no seu PATH"

# Test targets
test:
	@echo "Executando testes..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Testes concluídos"

test-unit:
	@echo "Executando testes unitários..."
	go test -v -short ./...
	@echo "Testes unitários concluídos"

coverage:
	@echo "Gerando relatório de cobertura..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Cobertura total: " $$3}'

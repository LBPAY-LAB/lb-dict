#!/bin/bash
# Atualizar todos os links para nova estrutura

find . -name "*.md" -type f | while read file; do
    sed -i '' \
        -e 's|06_Especificacoes_Tecnicas|11_Especificacoes_Tecnicas|g' \
        -e 's|07_Seguranca|13_Seguranca|g' \
        -e 's|04_Regulatorio|06_Regulatorio|g' \
        -e 's|03_Requisitos|05_Requisitos|g' \
        -e 's|06_Integracao|12_Integracao|g' \
        "$file"
done

echo "Links atualizados em todos os arquivos .md"

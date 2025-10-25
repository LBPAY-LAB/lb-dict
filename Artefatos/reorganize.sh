#!/bin/bash
# Reorganização completa da numeração

# Renomear para temporários (evitar conflitos)
mv "03_Requisitos" "temp_Requisitos_A"
mv "04_Regulatorio" "temp_Regulatorio"
mv "04_Processos" "temp_Processos"
mv "05_Requisitos" "temp_Requisitos_B"
mv "05_Implementacao" "temp_Implementacao"
mv "06_Integracao" "temp_Integracao"

# Agora renumerar para posições finais corretas
mv "temp_Requisitos_A" "05_Requisitos"
mv "temp_Regulatorio" "06_Regulatorio"
mv "temp_Processos" "07_Processos"
mv "05_Frontend" "08_Frontend"
mv "temp_Implementacao" "09_Implementacao"
mv "temp_Requisitos_B" "10_Requisitos_User_Stories"  # Diferenciar
mv "06_Especificacoes_Tecnicas" "11_Especificacoes_Tecnicas"
mv "temp_Integracao" "12_Integracao"
mv "07_Seguranca" "13_Seguranca"
mv "08_Testes" "14_Testes"
mv "09_DevOps" "15_DevOps"
mv "10_Compliance" "16_Compliance"
mv "11_Gestao" "17_Gestao"

echo "Renumeração completa!"

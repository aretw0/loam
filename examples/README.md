# Exemplos Loam

Este diretório contém exemplos de uso e provas de conceito do projeto Loam.

## Projetos de Exemplo

- **[Basic App](./basic-app)**: Um exemplo simples e completo de como usar o pacote principal (`github.com/aretw0/loam`) para criar, salvar e ler notas em uma aplicação Go.
  - Demonstra a inicialização do Vault.
  - Mostra como usar a substituição de módulo (`replace` no `go.mod`) para desenvolver usando a versão local do Loam.

## Spikes (Pesquisa)

A pasta **[spikes](./spikes)** contém experimentos e provas de conceito (PoCs) isoladas criadas durante o design e pesquisa do Loam. Estes códigos servem como registro histórico e demonstração técnica de conceitos específicos (como lock de arquivos, concorrência git, etc), mas **não** devem ser tomados como exemplos de boas práticas ou uso da API atual.

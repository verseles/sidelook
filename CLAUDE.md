# Regras do Projeto Sidelook

## Workflow de Desenvolvimento

Ao concluir qualquer tarefa neste projeto, siga rigorosamente esta sequência:

1. **Testes**: Execute `go test ./...` para garantir que todos os testes passam
2. **Lint** (se disponível): Execute `make lint` para verificar qualidade do código
3. **Build**: Execute `make build` para compilar o binário para o sistema operacional atual
4. **Verificação manual**: Permita que o usuário teste o binário gerado antes de prosseguir
5. **Commit**: Somente após confirmação, faça o commit das alterações

Esta sequência garante que o usuário sempre tenha um binário funcional e testado disponível para validação manual antes de qualquer alteração ser commitada.

## Comandos Importantes

- `make build` - Compila binário para o SO atual
- `make build-all` - Compila para todas as plataformas (Linux, macOS, Windows)
- `make test` - Executa todos os testes
- `make lint` - Verifica qualidade do código (requer golangci-lint)
- `make run` - Executa em modo desenvolvimento
- `go test ./...` - Executa testes (alternativa ao make test)

## Estrutura do Projeto

- `cmd/sidelook/` - Ponto de entrada da aplicação
- `internal/watcher/` - Monitoramento de diretório
- `internal/server/` - Servidor HTTP + WebSocket
- `internal/assets/` - HTML/CSS/JS embutidos
- `internal/browser/` - Abertura automática do navegador
- `internal/updater/` - Sistema de auto-update
- `pkg/` - Pacotes reutilizáveis

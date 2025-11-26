# sidelook

Visualizador de imagens em tempo real via navegador.

## Funcionalidades

- 🖼 Monitora um diretório por novas imagens
- 🌐 Serve as imagens via HTTP local
- ⚡ Atualização em tempo real via WebSocket
- 🎬 Modo slideshow com N imagens mais recentes
- 🔄 Detecção automática de imagens deletadas/movidas
- 🖥 Abre navegador automaticamente
- 🔄 Auto-update integrado
- 🎯 Fullscreen ao clicar (ou tecla F)

## Instalação

Baixe na [página de releases](https://github.com/verseles/sidelook/releases/latest):

- Linux: `sidelook-linux`
- macOS: `sidelook-macos` (Universal)
- Windows: `sidelook-windows.exe`

```bash
chmod +x sidelook-linux
sudo mv sidelook-linux /usr/local/bin/sidelook
```

## Uso

```bash
sidelook                      # Diretório atual
sidelook ~/Downloads          # Pasta específica
sidelook -p 3000              # Porta específica
sidelook -s 4                 # Slideshow com 4 imagens mais recentes
sidelook -s 4 -t 5            # Slideshow mudando a cada 5 segundos
sidelook --slideshow 10 --time 3   # Forma longa dos comandos
sidelook --update             # Atualizar
sidelook --version            # Versão
```

### Opções

- `-p, --port` - Porta HTTP (padrão: 8080, tenta sequencialmente se ocupada)
- `-s, --slideshow` - Número de imagens no slideshow (0 = desabilitado)
- `-t, --time` - Intervalo em segundos entre imagens no slideshow (padrão: 3)
- `--update` - Verificar e instalar atualizações
- `--version` - Mostrar versão

## Modos de Operação

### Modo Padrão
Exibe apenas a imagem mais recente do diretório. Atualiza automaticamente quando:
- Nova imagem é adicionada
- Imagem atual é deletada ou movida (mostra a próxima mais recente)

### Modo Slideshow
Ativado com `-s N`, exibe as N imagens mais recentes em rotação automática:
- Transição suave entre imagens
- Intervalo configurável com `-t SEGUNDOS`
- Lista atualizada automaticamente quando novas imagens chegam

## Formatos Suportados

JPG, JPEG, PNG, GIF, WebP, SVG, BMP, TIFF, TIF

## Desenvolvimento

```bash
go mod tidy
go run ./cmd/sidelook .
go test ./...
make build
```

### Workflow

Consulte [CLAUDE.md](CLAUDE.md) para regras de desenvolvimento do projeto.

## Licença

MIT

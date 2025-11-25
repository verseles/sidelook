# Plano Completo para o CLI "sidelook" em Go
## Instruções para IA Generativa de Código

> **IMPORTANTE**: Este documento contém instruções sequenciais. Execute na ordem apresentada. Não pule etapas. Ao encontrar problemas, resolva antes de prosseguir.

---

## 0. Pré-requisitos e Verificações Iniciais

### 0.1 Antes de Começar
```bash
# Verificar se Go está instalado
go version
# Esperado: go version go1.21+ (ou superior)

# Verificar diretório de trabalho
pwd
```

### 0.2 Criar Estrutura Base
```bash
# Criar diretório do projeto
mkdir -p sidelook
cd sidelook

# Inicializar módulo Go
go mod init github.com/insign/sidelook
```

---

## 1. Estrutura Completa do Projeto

```
sidelook/
├── cmd/
│   └── sidelook/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/
│   │   └── cli.go               # Parser de argumentos
│   ├── server/
│   │   ├── server.go            # HTTP server + WebSocket
│   │   └── handlers.go          # Request handlers
│   ├── watcher/
│   │   └── watcher.go           # Directory watcher
│   ├── updater/
│   │   └── updater.go           # Auto-update via GitHub
│   ├── browser/
│   │   └── browser.go           # Abertura cross-platform
│   ├── assets/
│   │   └── html.go              # HTML/CSS/JS embarcados
│   └── version/
│       └── version.go           # Versão (set via ldflags)
├── pkg/
│   └── semver/
│       └── semver.go            # Comparação de versões
├── test/
│   └── integration_test.go      # Testes de integração
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── CHANGELOG.md
├── LICENSE
└── .github/
    └── workflows/
        └── release.yml
```

**NOTA sobre estrutura Go:**
- `cmd/` - Entry points de executáveis
- `internal/` - Código privado do projeto (não importável externamente)
- `pkg/` - Código público reutilizável (opcional)

---

## 2. Ordem de Implementação (SEGUIR ESTA SEQUÊNCIA)

### Fase 1: Fundação
1. `go.mod` (já criado)
2. `internal/version/version.go`
3. `pkg/semver/semver.go`
4. `pkg/semver/semver_test.go`

### Fase 2: Core Features
5. `internal/cli/cli.go`
6. `internal/watcher/watcher.go`
7. `internal/watcher/watcher_test.go`
8. `internal/assets/html.go`
9. `internal/server/handlers.go`
10. `internal/server/server.go`

### Fase 3: Features Auxiliares
11. `internal/browser/browser.go`
12. `internal/updater/updater.go`
13. `internal/updater/updater_test.go`

### Fase 4: Integração
14. `cmd/sidelook/main.go`
15. Testar: `go run ./cmd/sidelook`

### Fase 5: Build e Documentação
16. `Makefile`
17. `README.md`
18. `CHANGELOG.md`
19. `LICENSE`
20. `.github/workflows/release.yml`

---

## 3. Implementações Detalhadas

### 3.1 go.mod

```go
module github.com/insign/sidelook

go 1.21

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/gorilla/websocket v1.5.1
)
```

**Após criar, executar:**
```bash
go mod tidy
```

**NOTAS:**
- `fsnotify` - Watcher de filesystem cross-platform (usa inotify/kqueue/ReadDirectoryChangesW)
- `gorilla/websocket` - Implementação robusta de WebSocket
- Go stdlib já tem tudo para HTTP server, CLI args, JSON, etc.

---

### 3.2 internal/version/version.go

```go
// internal/version/version.go
package version

// Estas variáveis são definidas em tempo de compilação via ldflags
// Exemplo: go build -ldflags "-X github.com/insign/sidelook/internal/version.Version=1.0.0"
var (
	// Version é a versão semântica do aplicativo
	Version = "dev"
	
	// Commit é o hash do commit git
	Commit = "unknown"
	
	// BuildDate é a data de compilação
	BuildDate = "unknown"
)

// Info retorna string formatada com informações de versão
func Info() string {
	return Version
}

// Full retorna informações completas de versão
func Full() string {
	return Version + " (" + Commit + ") built " + BuildDate
}
```

---

### 3.3 pkg/semver/semver.go

```go
// pkg/semver/semver.go
package semver

import (
	"strconv"
	"strings"
)

// Comparison representa o resultado da comparação de versões
type Comparison int

const (
	// Older indica que a versão local é mais antiga
	Older Comparison = -1
	// Equal indica que as versões são iguais
	Equal Comparison = 0
	// Newer indica que a versão local é mais recente
	Newer Comparison = 1
)

// Normalize remove prefixos e sufixos de uma string de versão
// Exemplo: "v1.2.3-beta" -> "1.2.3"
func Normalize(version string) string {
	v := strings.TrimSpace(strings.ToLower(version))
	
	// Remover prefixo 'v'
	v = strings.TrimPrefix(v, "v")
	
	// Remover sufixo pré-release (-beta, -alpha, etc)
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	
	// Remover metadata de build (+build)
	if idx := strings.Index(v, "+"); idx != -1 {
		v = v[:idx]
	}
	
	return v
}

// Parse converte uma string de versão em [major, minor, patch]
func Parse(version string) [3]int {
	normalized := Normalize(version)
	parts := strings.Split(normalized, ".")
	
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		if n, err := strconv.Atoi(parts[i]); err == nil {
			result[i] = n
		}
	}
	
	return result
}

// Compare compara duas versões
// Retorna Older se local < remote, Equal se iguais, Newer se local > remote
func Compare(local, remote string) Comparison {
	localParts := Parse(local)
	remoteParts := Parse(remote)
	
	for i := 0; i < 3; i++ {
		if localParts[i] < remoteParts[i] {
			return Older
		}
		if localParts[i] > remoteParts[i] {
			return Newer
		}
	}
	
	return Equal
}

// HasUpdate verifica se há atualização disponível
func HasUpdate(local, remote string) bool {
	return Compare(local, remote) == Older
}
```

---

### 3.4 pkg/semver/semver_test.go

```go
// pkg/semver/semver_test.go
package semver

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"v1.2.3-beta", "1.2.3"},
		{"1.2.3+build", "1.2.3"},
		{"V1.2.3", "1.2.3"},
		{"  v1.2.3  ", "1.2.3"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"1.2", [3]int{1, 2, 0}},
		{"1", [3]int{1, 0, 0}},
		{"v1.2.3", [3]int{1, 2, 3}},
		{"", [3]int{0, 0, 0}},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Parse(tt.input)
			if result != tt.expected {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		local    string
		remote   string
		expected Comparison
	}{
		{"1.0.0", "1.0.0", Equal},
		{"1.0.0", "2.0.0", Older},
		{"2.0.0", "1.0.0", Newer},
		{"1.1.0", "1.2.0", Older},
		{"1.0.0", "1.0.1", Older},
		{"v1.0.0", "1.0.0", Equal},
		{"1.0.0-beta", "1.0.0", Equal},
	}
	
	for _, tt := range tests {
		name := tt.local + "_vs_" + tt.remote
		t.Run(name, func(t *testing.T) {
			result := Compare(tt.local, tt.remote)
			if result != tt.expected {
				t.Errorf("Compare(%q, %q) = %v, want %v", tt.local, tt.remote, result, tt.expected)
			}
		})
	}
}

func TestHasUpdate(t *testing.T) {
	if !HasUpdate("1.0.0", "1.0.1") {
		t.Error("HasUpdate(1.0.0, 1.0.1) should be true")
	}
	if HasUpdate("1.0.0", "1.0.0") {
		t.Error("HasUpdate(1.0.0, 1.0.0) should be false")
	}
	if HasUpdate("2.0.0", "1.0.0") {
		t.Error("HasUpdate(2.0.0, 1.0.0) should be false")
	}
}
```

**Executar testes:**
```bash
go test ./pkg/semver/...
```

---

### 3.5 internal/cli/cli.go

```go
// internal/cli/cli.go
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/insign/sidelook/internal/version"
)

// Config contém a configuração parseada dos argumentos CLI
type Config struct {
	// Directory é o diretório a monitorar
	Directory string
	
	// Port é a porta especificada (0 = auto)
	Port int
	
	// Update indica se deve executar atualização
	Update bool
	
	// ShowVersion indica se deve exibir versão
	ShowVersion bool
	
	// ShowHelp indica se deve exibir ajuda
	ShowHelp bool
}

// Parse faz o parse dos argumentos de linha de comando
func Parse(args []string) (*Config, error) {
	cfg := &Config{}
	
	fs := flag.NewFlagSet("sidelook", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	
	fs.IntVar(&cfg.Port, "p", 0, "Porta do servidor HTTP (padrão: 8080)")
	fs.IntVar(&cfg.Port, "port", 0, "Porta do servidor HTTP (padrão: 8080)")
	fs.BoolVar(&cfg.Update, "u", false, "Atualizar para a versão mais recente")
	fs.BoolVar(&cfg.Update, "update", false, "Atualizar para a versão mais recente")
	fs.BoolVar(&cfg.ShowVersion, "v", false, "Exibir versão atual")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Exibir versão atual")
	fs.BoolVar(&cfg.ShowHelp, "h", false, "Exibir esta ajuda")
	fs.BoolVar(&cfg.ShowHelp, "help", false, "Exibir esta ajuda")
	
	// Custom usage
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, Usage())
	}
	
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	
	// Diretório é o primeiro argumento posicional
	if fs.NArg() > 0 {
		cfg.Directory = fs.Arg(0)
	} else {
		cfg.Directory = "."
	}
	
	// Validar porta
	if cfg.Port != 0 && (cfg.Port < 1 || cfg.Port > 65535) {
		return nil, fmt.Errorf("porta inválida: %d. Use um número entre 1 e 65535", cfg.Port)
	}
	
	return cfg, nil
}

// Usage retorna o texto de ajuda
func Usage() string {
	return fmt.Sprintf(`sidelook %s - Visualizador de imagens em tempo real

Uso: sidelook [opções] [diretório]

Opções:
  -p, --port <número>   Porta do servidor HTTP (padrão: 8080)
  -u, --update          Atualizar para a versão mais recente
  -v, --version         Exibir versão atual
  -h, --help            Exibir esta ajuda

Exemplos:
  sidelook                    # Monitora diretório atual
  sidelook ~/Downloads        # Monitora pasta Downloads
  sidelook -p 3000            # Usa porta 3000
  sidelook --update           # Atualiza para versão mais recente

`, version.Version)
}
```

---

### 3.6 internal/watcher/watcher.go

```go
// internal/watcher/watcher.go
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// SupportedExtensions são as extensões de imagem suportadas
var SupportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
}

// IsImageFile verifica se um arquivo é uma imagem suportada
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return SupportedExtensions[ext]
}

// ImageInfo contém informações sobre uma imagem
type ImageInfo struct {
	Path    string
	ModTime time.Time
}

// ImageWatcher monitora um diretório por novas imagens
type ImageWatcher struct {
	dir     string
	watcher *fsnotify.Watcher
	
	mu           sync.RWMutex
	currentImage *ImageInfo
	
	// OnNewImage é chamado quando uma nova imagem é detectada
	OnNewImage func(path string)
	
	done chan struct{}
}

// New cria um novo ImageWatcher
func New(dir string) (*ImageWatcher, error) {
	// Verificar se diretório existe
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}
	
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	
	return &ImageWatcher{
		dir:     dir,
		watcher: w,
		done:    make(chan struct{}),
	}, nil
}

// ScanExisting faz scan inicial e retorna a imagem mais recente
func (iw *ImageWatcher) ScanExisting() (count int, mostRecent *ImageInfo, err error) {
	entries, err := os.ReadDir(iw.dir)
	if err != nil {
		return 0, nil, err
	}
	
	var latestInfo *ImageInfo
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(iw.dir, entry.Name())
		if !IsImageFile(path) {
			continue
		}
		
		count++
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if latestInfo == nil || info.ModTime().After(latestInfo.ModTime) {
			latestInfo = &ImageInfo{
				Path:    path,
				ModTime: info.ModTime(),
			}
		}
	}
	
	iw.mu.Lock()
	iw.currentImage = latestInfo
	iw.mu.Unlock()
	
	return count, latestInfo, nil
}

// CurrentImage retorna a imagem atual
func (iw *ImageWatcher) CurrentImage() *ImageInfo {
	iw.mu.RLock()
	defer iw.mu.RUnlock()
	return iw.currentImage
}

// CurrentImageRelative retorna o caminho relativo da imagem atual
func (iw *ImageWatcher) CurrentImageRelative() string {
	img := iw.CurrentImage()
	if img == nil {
		return ""
	}
	rel, err := filepath.Rel(iw.dir, img.Path)
	if err != nil {
		return filepath.Base(img.Path)
	}
	return rel
}

// Start inicia o monitoramento
func (iw *ImageWatcher) Start() error {
	if err := iw.watcher.Add(iw.dir); err != nil {
		return err
	}
	
	go iw.loop()
	return nil
}

func (iw *ImageWatcher) loop() {
	for {
		select {
		case event, ok := <-iw.watcher.Events:
			if !ok {
				return
			}
			iw.handleEvent(event)
			
		case err, ok := <-iw.watcher.Errors:
			if !ok {
				return
			}
			// Log error mas continua
			_ = err
			
		case <-iw.done:
			return
		}
	}
}

func (iw *ImageWatcher) handleEvent(event fsnotify.Event) {
	// Interessados em Create e Write
	if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
		return
	}
	
	path := event.Name
	if !IsImageFile(path) {
		return
	}
	
	// Verificar se arquivo existe (pode ter sido deletado rapidamente)
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	
	// Atualizar imagem atual
	newImage := &ImageInfo{
		Path:    path,
		ModTime: info.ModTime(),
	}
	
	iw.mu.Lock()
	iw.currentImage = newImage
	iw.mu.Unlock()
	
	// Notificar callback
	if iw.OnNewImage != nil {
		rel, err := filepath.Rel(iw.dir, path)
		if err != nil {
			rel = filepath.Base(path)
		}
		iw.OnNewImage(rel)
	}
}

// Stop para o monitoramento
func (iw *ImageWatcher) Stop() error {
	close(iw.done)
	return iw.watcher.Close()
}

// Dir retorna o diretório monitorado
func (iw *ImageWatcher) Dir() string {
	return iw.dir
}
```

---

### 3.7 internal/watcher/watcher_test.go

```go
// internal/watcher/watcher_test.go
package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"foto.jpg", true},
		{"foto.jpeg", true},
		{"foto.PNG", true}, // case insensitive
		{"animacao.gif", true},
		{"imagem.webp", true},
		{"vetor.svg", true},
		{"bitmap.bmp", true},
		{"scan.tiff", true},
		{"scan.tif", true},
		{"arquivo.txt", false},
		{"doc.pdf", false},
		{"arquivo", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsImageFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsImageFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestImageWatcher_New(t *testing.T) {
	// Diretório válido
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	w, err := New(tmpDir)
	if err != nil {
		t.Errorf("New() error = %v, want nil", err)
	}
	if w != nil {
		w.Stop()
	}
	
	// Diretório inválido
	_, err = New("/caminho/que/nao/existe")
	if err == nil {
		t.Error("New() com diretório inválido deveria retornar erro")
	}
}

func TestImageWatcher_ScanExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Criar algumas imagens
	img1 := filepath.Join(tmpDir, "img1.png")
	if err := os.WriteFile(img1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}
	
	time.Sleep(100 * time.Millisecond)
	
	img2 := filepath.Join(tmpDir, "img2.jpg")
	if err := os.WriteFile(img2, []byte{0xFF, 0xD8, 0xFF}, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Criar arquivo não-imagem
	txt := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txt, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	
	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	
	count, mostRecent, err := w.ScanExisting()
	if err != nil {
		t.Errorf("ScanExisting() error = %v", err)
	}
	
	if count != 2 {
		t.Errorf("ScanExisting() count = %d, want 2", count)
	}
	
	if mostRecent == nil {
		t.Fatal("ScanExisting() mostRecent = nil, want non-nil")
	}
	
	if mostRecent.Path != img2 {
		t.Errorf("ScanExisting() mostRecent.Path = %q, want %q", mostRecent.Path, img2)
	}
}

func TestImageWatcher_DetectNewImage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	
	detected := make(chan string, 1)
	w.OnNewImage = func(path string) {
		detected <- path
	}
	
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	
	// Dar tempo para watcher inicializar
	time.Sleep(100 * time.Millisecond)
	
	// Criar nova imagem
	newImg := filepath.Join(tmpDir, "nova.png")
	if err := os.WriteFile(newImg, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Aguardar detecção
	select {
	case path := <-detected:
		if path != "nova.png" {
			t.Errorf("OnNewImage path = %q, want %q", path, "nova.png")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout aguardando detecção de nova imagem")
	}
}
```

---

### 3.8 internal/assets/html.go

```go
// internal/assets/html.go
package assets

import "fmt"

// GenerateHTML gera o HTML completo da página do visualizador
func GenerateHTML(initialImage string) string {
	imageDisplay := `<div id="waiting">Aguardando primeira imagem...</div>`
	if initialImage != "" {
		imageDisplay = fmt.Sprintf(`<img id="viewer" src="/image/%s" alt="Imagem">`, initialImage)
	}
	
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>sidelook</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    html, body {
      width: 100%%;
      height: 100%%;
      background: #000;
      overflow: hidden;
    }
    
    #container {
      width: 100%%;
      height: 100%%;
      display: flex;
      justify-content: center;
      align-items: center;
      cursor: pointer;
    }
    
    #viewer {
      max-width: 100%%;
      max-height: 100%%;
      object-fit: contain;
      opacity: 1;
      transform: scale(1);
      transition: opacity 200ms ease-out, transform 200ms ease-out;
    }
    
    #viewer.fade-out {
      opacity: 0;
      transform: scale(0.98);
    }
    
    #waiting {
      color: #666;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 1.5rem;
      text-align: center;
      padding: 2rem;
    }
    
    #status {
      position: fixed;
      bottom: 10px;
      right: 10px;
      padding: 5px 10px;
      border-radius: 4px;
      font-family: monospace;
      font-size: 12px;
      opacity: 0.7;
      transition: opacity 0.3s;
    }
    
    #status:hover {
      opacity: 1;
    }
    
    #status.connected {
      background: #1a472a;
      color: #4ade80;
    }
    
    #status.disconnected {
      background: #4a1a1a;
      color: #f87171;
    }
    
    :fullscreen #container,
    :-webkit-full-screen #container {
      background: #000;
    }
  </style>
</head>
<body>
  <div id="container" onclick="toggleFullscreen()">
    %s
  </div>
  <div id="status" class="disconnected">Desconectado</div>
  
  <script>
    const container = document.getElementById('container');
    const status = document.getElementById('status');
    let ws;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 10;
    const reconnectDelay = 2000;
    
    function connect() {
      const wsUrl = 'ws://' + window.location.host + '/ws';
      ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        console.log('WebSocket conectado');
        status.textContent = 'Conectado';
        status.className = 'connected';
        reconnectAttempts = 0;
      };
      
      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'new_image') {
          updateImage(data.path);
        }
      };
      
      ws.onclose = () => {
        console.log('WebSocket desconectado');
        status.textContent = 'Desconectado';
        status.className = 'disconnected';
        scheduleReconnect();
      };
      
      ws.onerror = (error) => {
        console.error('WebSocket erro:', error);
      };
    }
    
    function scheduleReconnect() {
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        console.log('Tentando reconectar... (' + reconnectAttempts + '/' + maxReconnectAttempts + ')');
        setTimeout(connect, reconnectDelay);
      }
    }
    
    function updateImage(imagePath) {
      const current = document.getElementById('viewer');
      const waiting = document.getElementById('waiting');
      
      if (waiting) {
        waiting.remove();
      }
      
      if (current) {
        current.classList.add('fade-out');
        
        setTimeout(() => {
          const newImg = document.createElement('img');
          newImg.id = 'viewer';
          newImg.src = '/image/' + imagePath + '?t=' + Date.now();
          newImg.alt = 'Imagem';
          newImg.classList.add('fade-out');
          
          newImg.onload = () => {
            current.remove();
            container.appendChild(newImg);
            void newImg.offsetWidth;
            newImg.classList.remove('fade-out');
          };
          
          newImg.onerror = () => {
            console.error('Erro ao carregar imagem:', imagePath);
          };
        }, 200);
      } else {
        const newImg = document.createElement('img');
        newImg.id = 'viewer';
        newImg.src = '/image/' + imagePath + '?t=' + Date.now();
        newImg.alt = 'Imagem';
        container.appendChild(newImg);
      }
    }
    
    function toggleFullscreen() {
      if (!document.fullscreenElement) {
        container.requestFullscreen().catch(err => {
          console.log('Erro ao entrar em fullscreen:', err);
        });
      } else {
        document.exitFullscreen();
      }
    }
    
    document.addEventListener('keydown', (e) => {
      if (e.key === 'f' || e.key === 'F') {
        toggleFullscreen();
      }
    });
    
    connect();
  </script>
</body>
</html>
`, imageDisplay)
}
```

---

### 3.9 internal/server/handlers.go

```go
// internal/server/handlers.go
package server

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/insign/sidelook/internal/assets"
	"github.com/insign/sidelook/internal/watcher"
)

// registerRoutes registra os handlers HTTP
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/image/", s.handleImage)
}

// handleIndex serve a página HTML principal
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	initialImage := s.watcher.CurrentImageRelative()
	html := assets.GenerateHTML(initialImage)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleImage serve arquivos de imagem
func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	// Extrair caminho da imagem (remover /image/ prefix)
	imagePath := strings.TrimPrefix(r.URL.Path, "/image/")
	if imagePath == "" {
		http.NotFound(w, r)
		return
	}
	
	// Construir caminho completo
	fullPath := filepath.Join(s.watcher.Dir(), imagePath)
	
	// Segurança: verificar se o caminho está dentro do diretório monitorado
	absDir, err := filepath.Abs(s.watcher.Dir())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(absPath, absDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	// Verificar se é uma imagem válida
	if !watcher.IsImageFile(fullPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	// Verificar se arquivo existe
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	
	// Determinar content type
	ext := filepath.Ext(fullPath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	http.ServeFile(w, r, fullPath)
}

// wsMessage é a estrutura de mensagem WebSocket
type wsMessage struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

// broadcastNewImage envia notificação de nova imagem para todos os clientes
func (s *Server) broadcastNewImage(path string) {
	msg := wsMessage{
		Type: "new_image",
		Path: path,
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	
	for client := range s.clients {
		go func(c *wsClient) {
			c.send <- data
		}(client)
	}
}
```

---

### 3.10 internal/server/server.go

```go
// internal/server/server.go
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/insign/sidelook/internal/watcher"
)

// Server é o servidor HTTP com suporte a WebSocket
type Server struct {
	watcher   *watcher.ImageWatcher
	server    *http.Server
	mux       *http.ServeMux
	port      int
	upgrader  websocket.Upgrader
	
	clients   map[*wsClient]bool
	clientsMu sync.RWMutex
}

// wsClient representa um cliente WebSocket conectado
type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

// New cria um novo servidor
func New(w *watcher.ImageWatcher, preferredPort int) *Server {
	s := &Server{
		watcher: w,
		mux:     http.NewServeMux(),
		clients: make(map[*wsClient]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Aceitar qualquer origem (localhost)
			},
		},
	}
	
	if preferredPort == 0 {
		preferredPort = 8080
	}
	s.port = preferredPort
	
	s.registerRoutes()
	
	// Configurar callback do watcher
	w.OnNewImage = s.broadcastNewImage
	
	return s
}

// Start inicia o servidor, tentando portas sequenciais se necessário
func (s *Server) Start() error {
	const maxAttempts = 100
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		addr := fmt.Sprintf("127.0.0.1:%d", s.port)
		
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			s.port++
			continue
		}
		
		s.server = &http.Server{
			Handler:      s.mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		
		go s.server.Serve(listener)
		return nil
	}
	
	return fmt.Errorf("não foi possível iniciar servidor: portas %d-%d indisponíveis",
		s.port-maxAttempts+1, s.port)
}

// Port retorna a porta em que o servidor está rodando
func (s *Server) Port() int {
	return s.port
}

// URL retorna a URL completa do servidor
func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

// Stop para o servidor graciosamente
func (s *Server) Stop() error {
	// Fechar todos os WebSockets
	s.clientsMu.Lock()
	for client := range s.clients {
		close(client.send)
		client.conn.Close()
	}
	s.clients = make(map[*wsClient]bool)
	s.clientsMu.Unlock()
	
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleWebSocket gerencia conexões WebSocket
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	
	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 256),
	}
	
	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()
	
	go s.writePump(client)
	go s.readPump(client)
}

// writePump envia mensagens para o cliente WebSocket
func (s *Server) writePump(client *wsClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump lê mensagens do cliente WebSocket (principalmente para detectar desconexão)
func (s *Server) readPump(client *wsClient) {
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, client)
		s.clientsMu.Unlock()
		client.conn.Close()
	}()
	
	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			break
		}
	}
}
```

---

### 3.11 internal/browser/browser.go

```go
// internal/browser/browser.go
package browser

import (
	"os/exec"
	"runtime"
)

// Open abre a URL no navegador padrão do sistema
// Retorna erro apenas se o comando falhar ao executar
func Open(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		// SO não suportado, retornar nil (falha silenciosa)
		return nil
	}
	
	// Executar de forma não-bloqueante
	return cmd.Start()
}
```

---

### 3.12 internal/updater/updater.go

```go
// internal/updater/updater.go
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/insign/sidelook/internal/version"
	"github.com/insign/sidelook/pkg/semver"
)

const (
	apiURL  = "https://api.github.com/repos/insign/sidelook/releases/latest"
	timeout = 10 * time.Second
)

// ReleaseInfo contém informações sobre uma release do GitHub
type ReleaseInfo struct {
	Version     string
	DownloadURL string
	PublishedAt time.Time
}

// githubRelease representa a resposta da API do GitHub
type githubRelease struct {
	TagName     string        `json:"tag_name"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckLatestRelease verifica a release mais recente no GitHub
func CheckLatestRelease() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: timeout}
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "sidelook/"+version.Version)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}
	
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	
	// Encontrar asset para o SO atual
	assetName := getAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	publishedAt, _ := time.Parse(time.RFC3339, release.PublishedAt)
	
	return &ReleaseInfo{
		Version:     release.TagName,
		DownloadURL: downloadURL,
		PublishedAt: publishedAt,
	}, nil
}

// getAssetName retorna o nome do asset para o SO atual
func getAssetName() string {
	switch runtime.GOOS {
	case "linux":
		return "sidelook-linux"
	case "darwin":
		return "sidelook-macos"
	case "windows":
		return "sidelook-windows.exe"
	default:
		return "sidelook"
	}
}

// CheckResult é o resultado da verificação de atualização
type CheckResult struct {
	HasUpdate      bool
	CurrentVersion string
	LatestVersion  string
}

// CheckInBackground verifica atualizações em background
// Retorna um channel que receberá o resultado
func CheckInBackground() <-chan *CheckResult {
	ch := make(chan *CheckResult, 1)
	
	go func() {
		defer close(ch)
		
		release, err := CheckLatestRelease()
		if err != nil {
			return
		}
		
		result := &CheckResult{
			CurrentVersion: version.Version,
			LatestVersion:  release.Version,
			HasUpdate:      semver.HasUpdate(version.Version, release.Version),
		}
		
		ch <- result
	}()
	
	return ch
}

// PerformUpdate executa a atualização completa
func PerformUpdate() error {
	fmt.Println("Verificando atualizações...")
	
	release, err := CheckLatestRelease()
	if err != nil {
		return fmt.Errorf("não foi possível verificar atualizações: %w", err)
	}
	
	if !semver.HasUpdate(version.Version, release.Version) {
		fmt.Printf("Você já está na versão mais recente (%s)\n", version.Version)
		return nil
	}
	
	if release.DownloadURL == "" {
		return fmt.Errorf("não foi encontrado download para %s", runtime.GOOS)
	}
	
	fmt.Printf("Baixando versão %s...\n", release.Version)
	
	// Baixar para arquivo temporário
	resp, err := http.Get(release.DownloadURL)
	if err != nil {
		return fmt.Errorf("falha no download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download falhou com status %d", resp.StatusCode)
	}
	
	// Criar arquivo temporário
	tempFile, err := os.CreateTemp("", "sidelook_update_*")
	if err != nil {
		return fmt.Errorf("não foi possível criar arquivo temporário: %w", err)
	}
	tempPath := tempFile.Name()
	
	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("falha ao salvar download: %w", err)
	}
	
	// Obter caminho do executável atual
	currentExe, err := os.Executable()
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("não foi possível determinar executável atual: %w", err)
	}
	currentExe, _ = filepath.EvalSymlinks(currentExe)
	
	fmt.Println("Instalando atualização...")
	
	if runtime.GOOS == "windows" {
		// Windows: não consegue sobrescrever executável em uso
		fmt.Println("\nNo Windows, a atualização precisa ser feita manualmente.")
		fmt.Printf("1. Feche este programa\n")
		fmt.Printf("2. Substitua %s pelo arquivo baixado em:\n   %s\n", currentExe, tempPath)
		return nil
	}
	
	// Unix: substituição atômica
	
	// Dar permissão de execução
	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("falha ao definir permissões: %w", err)
	}
	
	// Backup do atual (opcional)
	backupPath := currentExe + ".old"
	os.Remove(backupPath) // Ignorar erro
	os.Rename(currentExe, backupPath) // Ignorar erro no backup
	
	// Mover novo para posição do atual
	if err := os.Rename(tempPath, currentExe); err != nil {
		// Tentar restaurar backup
		os.Rename(backupPath, currentExe)
		return fmt.Errorf("falha ao instalar atualização: %w", err)
	}
	
	// Limpar backup
	os.Remove(backupPath)
	
	fmt.Printf("Atualizado para versão %s!\n", release.Version)
	return nil
}
```

---

### 3.13 cmd/sidelook/main.go

```go
// cmd/sidelook/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/insign/sidelook/internal/browser"
	"github.com/insign/sidelook/internal/cli"
	"github.com/insign/sidelook/internal/server"
	"github.com/insign/sidelook/internal/updater"
	"github.com/insign/sidelook/internal/version"
	"github.com/insign/sidelook/internal/watcher"
)

// Cores ANSI
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func main() {
	// Parse argumentos
	config, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ %s%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	
	// --help
	if config.ShowHelp {
		fmt.Print(cli.Usage())
		return
	}
	
	// --version
	if config.ShowVersion {
		fmt.Println(version.Info())
		return
	}
	
	// --update
	if config.Update {
		if err := updater.PerformUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗ %s%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}
		return
	}
	
	// Modo normal: iniciar servidor
	if err := runServer(config); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ %s%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}

func runServer(config *cli.Config) error {
	// Criar watcher
	w, err := watcher.New(config.Directory)
	if err != nil {
		return fmt.Errorf("diretório inválido: %s", config.Directory)
	}
	
	// Verificar atualizações em background
	updateCh := updater.CheckInBackground()
	
	// Scan inicial
	count, _, err := w.ScanExisting()
	if err != nil {
		return fmt.Errorf("erro ao escanear diretório: %w", err)
	}
	
	if count > 0 {
		fmt.Printf("%sℹ %d imagem(ns) encontrada(s)%s\n", colorBlue, count, colorReset)
	} else {
		fmt.Printf("%sℹ Nenhuma imagem encontrada. Aguardando...%s\n", colorBlue, colorReset)
	}
	
	// Iniciar watcher
	if err := w.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar monitoramento: %w", err)
	}
	defer w.Stop()
	
	// Iniciar servidor
	srv := server.New(w, config.Port)
	if err := srv.Start(); err != nil {
		return err
	}
	defer srv.Stop()
	
	// Exibir URL
	fmt.Println()
	fmt.Printf("%s%s🖼  sidelook rodando%s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%s   %s%s\n", colorDim, srv.URL(), colorReset)
	fmt.Println()
	
	// Abrir navegador
	if err := browser.Open(srv.URL()); err != nil {
		fmt.Printf("%s⚠ Não foi possível abrir o navegador automaticamente%s\n", colorYellow, colorReset)
		fmt.Printf("  Acesse manualmente: %s\n", srv.URL())
	}
	
	// Verificar resultado da checagem de atualização
	go func() {
		if result := <-updateCh; result != nil && result.HasUpdate {
			printUpdateAvailable(result.CurrentVersion, result.LatestVersion)
		}
	}()
	
	// Aguardar sinal de interrupção
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	
	fmt.Printf("\n%sℹ Encerrando...%s\n", colorBlue, colorReset)
	return nil
}

func printUpdateAvailable(current, latest string) {
	fmt.Println()
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s  Nova versão disponível: %s → %s%s\n", colorBold, colorCyan, current, latest, colorReset)
	fmt.Printf("%s  Execute: sidelook --update%s\n", colorCyan, colorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorCyan, colorReset)
	fmt.Println()
}
```

---

### 3.14 Makefile

```makefile
# Makefile para sidelook

# Variáveis
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-s -w \
	-X github.com/insign/sidelook/internal/version.Version=$(VERSION) \
	-X github.com/insign/sidelook/internal/version.Commit=$(COMMIT) \
	-X github.com/insign/sidelook/internal/version.BuildDate=$(BUILD_DATE)"

BINARY := sidelook
CMD_PATH := ./cmd/sidelook

# Alvos
.PHONY: all build build-all test lint clean run install

all: test build

# Build para o SO atual
build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD_PATH)

# Build para todas as plataformas
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux $(CMD_PATH)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-macos-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-macos-arm64 $(CMD_PATH)

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows.exe $(CMD_PATH)

# Testes
test:
	go test -v -race -cover ./...

# Lint (requer golangci-lint instalado)
lint:
	golangci-lint run

# Limpar artefatos
clean:
	rm -f $(BINARY) $(BINARY)-linux $(BINARY)-macos-* $(BINARY)-windows.exe
	go clean

# Executar em modo desenvolvimento
run:
	go run $(CMD_PATH) .

# Instalar no sistema
install: build
	cp $(BINARY) $(GOPATH)/bin/

# Formatar código
fmt:
	go fmt ./...

# Verificar dependências
tidy:
	go mod tidy
	go mod verify
```

---

## 4. CI/CD - .github/workflows/release.yml

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Verificar formatação
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Código não formatado:"
            gofmt -d .
            exit 1
          fi
      
      - name: Executar testes
        run: go test -v -race -cover ./...

  build-linux:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          COMMIT=$(git rev-parse --short HEAD)
          BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          
          go build -ldflags "-s -w \
            -X github.com/insign/sidelook/internal/version.Version=$VERSION \
            -X github.com/insign/sidelook/internal/version.Commit=$COMMIT \
            -X github.com/insign/sidelook/internal/version.BuildDate=$BUILD_DATE" \
            -o sidelook-linux ./cmd/sidelook
      
      - uses: actions/upload-artifact@v4
        with:
          name: sidelook-linux
          path: sidelook-linux

  build-macos:
    needs: test
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build (Intel + ARM)
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          COMMIT=$(git rev-parse --short HEAD)
          BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          LDFLAGS="-s -w \
            -X github.com/insign/sidelook/internal/version.Version=$VERSION \
            -X github.com/insign/sidelook/internal/version.Commit=$COMMIT \
            -X github.com/insign/sidelook/internal/version.BuildDate=$BUILD_DATE"
          
          GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o sidelook-macos-amd64 ./cmd/sidelook
          GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o sidelook-macos-arm64 ./cmd/sidelook
          lipo -create -output sidelook-macos sidelook-macos-amd64 sidelook-macos-arm64
      
      - uses: actions/upload-artifact@v4
        with:
          name: sidelook-macos
          path: sidelook-macos

  build-windows:
    needs: test
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build
        shell: bash
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          COMMIT=$(git rev-parse --short HEAD)
          BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          
          go build -ldflags "-s -w \
            -X github.com/insign/sidelook/internal/version.Version=$VERSION \
            -X github.com/insign/sidelook/internal/version.Commit=$COMMIT \
            -X github.com/insign/sidelook/internal/version.BuildDate=$BUILD_DATE" \
            -o sidelook-windows.exe ./cmd/sidelook
      
      - uses: actions/upload-artifact@v4
        with:
          name: sidelook-windows
          path: sidelook-windows.exe

  release:
    needs: [build-linux, build-macos, build-windows]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts
      
      - name: Preparar assets
        run: |
          mkdir -p release-assets
          cp artifacts/sidelook-linux/sidelook-linux release-assets/
          cp artifacts/sidelook-macos/sidelook-macos release-assets/
          cp artifacts/sidelook-windows/sidelook-windows.exe release-assets/
          chmod +x release-assets/sidelook-linux release-assets/sidelook-macos
      
      - name: Gerar Release Notes
        id: notes
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          
          if [ -f CHANGELOG.md ]; then
            NOTES=$(sed -n "/^## \[$VERSION\]/,/^## \[/p" CHANGELOG.md | sed '1d;$d' | head -50)
          fi
          
          [ -z "$NOTES" ] && NOTES="Release $VERSION"
          echo "$NOTES" > release_notes.txt
      
      - name: Criar Release
        uses: softprops/action-gh-release@v1
        with:
          name: ${{ steps.notes.outputs.version }}
          body_path: release_notes.txt
          files: |
            release-assets/sidelook-linux
            release-assets/sidelook-macos
            release-assets/sidelook-windows.exe
          draft: false
          prerelease: false
```

---

## 5. Documentação

### README.md

```markdown
# sidelook

Visualizador de imagens em tempo real via navegador.

## Funcionalidades

- 🖼 Monitora um diretório por novas imagens
- 🌐 Serve as imagens via HTTP local
- ⚡ Atualização em tempo real via WebSocket
- 🖥 Abre navegador automaticamente
- 🔄 Auto-update integrado
- 🎯 Fullscreen ao clicar (ou tecla F)

## Instalação

Baixe na [página de releases](https://github.com/insign/sidelook/releases/latest):

- Linux: `sidelook-linux`
- macOS: `sidelook-macos` (Universal)
- Windows: `sidelook-windows.exe`

```bash
chmod +x sidelook-linux
sudo mv sidelook-linux /usr/local/bin/sidelook
```

## Uso

```bash
sidelook                    # Diretório atual
sidelook ~/Downloads        # Pasta específica
sidelook -p 3000            # Porta específica
sidelook --update           # Atualizar
sidelook --version          # Versão
```

## Formatos

JPG, PNG, GIF, WebP, SVG, BMP, TIFF

## Desenvolvimento

```bash
go mod tidy
go run ./cmd/sidelook .
go test ./...
make build
```

## Licença

MIT
```

---

## 6. Checklist de Conclusão

- [ ] `go.mod` criado e `go mod tidy` executado
- [ ] `internal/version/version.go`
- [ ] `pkg/semver/semver.go` + testes
- [ ] `internal/cli/cli.go`
- [ ] `internal/watcher/watcher.go` + testes
- [ ] `internal/assets/html.go`
- [ ] `internal/server/handlers.go`
- [ ] `internal/server/server.go`
- [ ] `internal/browser/browser.go`
- [ ] `internal/updater/updater.go`
- [ ] `cmd/sidelook/main.go`
- [ ] Testes passando: `go test ./...`
- [ ] Formatação: `gofmt -l .` (sem output)
- [ ] `Makefile`
- [ ] README.md, CHANGELOG.md, LICENSE
- [ ] `.github/workflows/release.yml`
- [ ] Teste manual funcionando

---

## 7. Comandos Finais

```bash
go mod tidy
gofmt -w .
go test ./...
make build
./sidelook --version
./sidelook /tmp
```

---

**FIM DO DOCUMENTO**

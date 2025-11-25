# Plano Completo para o CLI "sidelook" em Dart
## Instruções para IA Generativa de Código

> **IMPORTANTE**: Este documento contém instruções sequenciais. Execute na ordem apresentada. Não pule etapas. Ao encontrar problemas, resolva antes de prosseguir.

---

## 0. Pré-requisitos e Verificações Iniciais

### 0.1 Antes de Começar
```bash
# Verificar se Dart está instalado
dart --version
# Esperado: Dart SDK version >= 3.0.0

# Verificar diretório de trabalho
pwd
# Criar o projeto no diretório atual ou especificado
```

### 0.2 Criar Estrutura Base
```bash
# Criar diretório do projeto
mkdir -p sidelook
cd sidelook

# Inicializar projeto Dart
dart create -t console .
# NOTA: Isso criará alguns arquivos. Vamos sobrescrever os necessários.
```

---

## 1. Estrutura Completa do Projeto

```
sidelook/
├── bin/
│   └── sidelook.dart           # Entry point - APENAS chama main() de lib
├── lib/
│   ├── sidelook.dart           # Exporta APIs públicas (se necessário)
│   └── src/
│       ├── version.g.dart      # GERADO - Não editar manualmente
│       ├── cli.dart            # Parser de argumentos com package:args
│       ├── server.dart         # HttpServer + WebSocket
│       ├── watcher.dart        # Directory watcher + filtro de imagens
│       ├── updater.dart        # GitHub API + download + substituição
│       ├── browser.dart        # Abertura cross-platform de navegador
│       ├── html_assets.dart    # Strings com HTML/CSS/JS
│       └── utils/
│           ├── logger.dart     # Formatação de output colorido
│           └── version_compare.dart  # Comparação semântica de versões
├── tool/
│   └── generate_version.dart   # Script para gerar version.g.dart
├── test/
│   ├── watcher_test.dart
│   ├── server_test.dart
│   ├── updater_test.dart
│   ├── version_compare_test.dart
│   └── cli_test.dart
├── pubspec.yaml
├── analysis_options.yaml
├── dart_test.yaml
├── README.md
├── CHANGELOG.md
├── LICENSE
└── .github/
    └── workflows/
        └── release.yml
```

---

## 2. Ordem de Implementação (SEGUIR ESTA SEQUÊNCIA)

### Fase 1: Fundação (Arquivos de Configuração)
1. `pubspec.yaml`
2. `analysis_options.yaml`
3. `dart_test.yaml`
4. `tool/generate_version.dart`
5. Gerar `lib/src/version.g.dart` executando o script

### Fase 2: Utilitários Base
6. `lib/src/utils/logger.dart`
7. `lib/src/utils/version_compare.dart`
8. Testes: `test/version_compare_test.dart`

### Fase 3: Core Features
9. `lib/src/cli.dart`
10. `lib/src/watcher.dart`
11. Testes: `test/watcher_test.dart`
12. `lib/src/html_assets.dart`
13. `lib/src/server.dart`
14. Testes: `test/server_test.dart`

### Fase 4: Features Auxiliares
15. `lib/src/browser.dart`
16. `lib/src/updater.dart`
17. Testes: `test/updater_test.dart`

### Fase 5: Integração
18. `bin/sidelook.dart` (main)
19. `test/cli_test.dart`
20. Testar manualmente: `dart run bin/sidelook.dart`

### Fase 6: Documentação e CI/CD
21. `README.md`
22. `CHANGELOG.md`
23. `LICENSE`
24. `.github/workflows/release.yml`

---

## 3. Implementações Detalhadas

### 3.1 pubspec.yaml

```yaml
name: sidelook
version: 0.1.0
description: CLI para visualizar imagens em tempo real via navegador
repository: https://github.com/insign/sidelook

environment:
  sdk: ^3.0.0

dependencies:
  watcher: ^1.1.0
  args: ^2.4.0
  http: ^1.2.0
  path: ^1.9.0
  yaml: ^3.1.0

dev_dependencies:
  test: ^1.25.0
  lints: ^4.0.0
  mocktail: ^1.0.0
```

**NOTAS IMPORTANTES:**
- `sdk: ^3.0.0` garante null safety
- `mocktail` é preferido sobre `mockito` para Dart puro (sem code generation)
- Não adicione `shelf` - vamos usar `dart:io` HttpServer diretamente (menos dependências)

---

### 3.2 analysis_options.yaml

```yaml
include: package:lints/recommended.yaml

analyzer:
  exclude:
    - "**/*.g.dart"
  language:
    strict-casts: true
    strict-inference: true
    strict-raw-types: true

linter:
  rules:
    - always_declare_return_types
    - avoid_print
    - prefer_single_quotes
    - sort_constructors_first
    - unawaited_futures
```

**NOTA:** `avoid_print` forçará uso do logger customizado.

---

### 3.3 dart_test.yaml

```yaml
concurrency: 4

file_reporters:
  json: reports/tests.json

override_platforms:
  vm:
    settings:
      verbosity: all
```

---

### 3.4 tool/generate_version.dart

```dart
// tool/generate_version.dart
// Executar com: dart run tool/generate_version.dart

import 'dart:io';
import 'package:yaml/yaml.dart';
import 'package:path/path.dart' as p;

Future<void> main() async {
  // Encontrar pubspec.yaml (pode estar em .. se rodando de tool/)
  final scriptDir = p.dirname(Platform.script.toFilePath());
  final projectRoot = p.dirname(scriptDir);
  
  var pubspecPath = p.join(projectRoot, 'pubspec.yaml');
  if (!File(pubspecPath).existsSync()) {
    // Fallback: talvez esteja rodando do root
    pubspecPath = 'pubspec.yaml';
  }
  
  final pubspecFile = File(pubspecPath);
  if (!pubspecFile.existsSync()) {
    stderr.writeln('ERRO: pubspec.yaml não encontrado');
    exit(1);
  }
  
  final content = await pubspecFile.readAsString();
  final yaml = loadYaml(content) as YamlMap;
  
  final name = yaml['name'] as String? ?? 'sidelook';
  final version = yaml['version'] as String? ?? '0.0.0';
  final description = yaml['description'] as String? ?? '';
  final repository = yaml['repository'] as String? ?? '';
  
  final output = '''
// GERADO AUTOMATICAMENTE - NÃO EDITAR
// Gerado por: dart run tool/generate_version.dart

/// Nome do pacote
const String packageName = '$name';

/// Versão atual do pacote (semver)
const String packageVersion = '$version';

/// Descrição do pacote
const String packageDescription = '$description';

/// URL do repositório GitHub
const String packageRepository = '$repository';
''';
  
  // Garantir que diretório existe
  final outputDir = Directory(p.join(projectRoot, 'lib', 'src'));
  if (!outputDir.existsSync()) {
    await outputDir.create(recursive: true);
  }
  
  final outputFile = File(p.join(outputDir.path, 'version.g.dart'));
  await outputFile.writeAsString(output);
  
  print('✓ Gerado: ${outputFile.path}');
  print('  Nome: $name');
  print('  Versão: $version');
}
```

**EXECUTAR IMEDIATAMENTE APÓS CRIAR:**
```bash
dart run tool/generate_version.dart
```

---

### 3.5 lib/src/utils/logger.dart

```dart
// lib/src/utils/logger.dart

import 'dart:io';

/// ANSI color codes para terminal
class _AnsiColors {
  static const reset = '\x1B[0m';
  static const red = '\x1B[31m';
  static const green = '\x1B[32m';
  static const yellow = '\x1B[33m';
  static const blue = '\x1B[34m';
  static const magenta = '\x1B[35m';
  static const cyan = '\x1B[36m';
  static const white = '\x1B[37m';
  static const bold = '\x1B[1m';
  static const dim = '\x1B[2m';
}

/// Logger centralizado para output formatado
class Logger {
  /// Detecta se o terminal suporta cores
  static bool get supportsAnsi {
    if (Platform.isWindows) {
      // Windows 10+ suporta ANSI, mas é complexo detectar
      // Assumir que suporta se stdout é TTY
      return stdout.hasTerminal;
    }
    return stdout.hasTerminal && 
           Platform.environment['TERM'] != 'dumb';
  }
  
  static String _colorize(String text, String color) {
    if (!supportsAnsi) return text;
    return '$color$text${_AnsiColors.reset}';
  }
  
  /// Log informativo (azul)
  static void info(String message) {
    stdout.writeln(_colorize('ℹ $message', _AnsiColors.blue));
  }
  
  /// Log de sucesso (verde)
  static void success(String message) {
    stdout.writeln(_colorize('✓ $message', _AnsiColors.green));
  }
  
  /// Log de aviso (amarelo)
  static void warn(String message) {
    stdout.writeln(_colorize('⚠ $message', _AnsiColors.yellow));
  }
  
  /// Log de erro (vermelho)
  static void error(String message) {
    stderr.writeln(_colorize('✗ $message', _AnsiColors.red));
  }
  
  /// Log de atualização disponível (cyan + negrito)
  static void updateAvailable(String currentVersion, String newVersion) {
    final msg = '''

${_colorize('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━', _AnsiColors.cyan)}
${_colorize('  Nova versão disponível: $currentVersion → $newVersion', '${_AnsiColors.bold}${_AnsiColors.cyan}')}
${_colorize('  Execute: sidelook --update', _AnsiColors.cyan)}
${_colorize('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━', _AnsiColors.cyan)}
''';
    stdout.write(msg);
  }
  
  /// Log de servidor iniciado
  static void serverStarted(String url) {
    stdout.writeln('');
    stdout.writeln(_colorize('🖼  sidelook rodando', '${_AnsiColors.bold}${_AnsiColors.green}'));
    stdout.writeln(_colorize('   $url', _AnsiColors.dim));
    stdout.writeln('');
  }
  
  /// Log de nova imagem detectada
  static void newImage(String filename) {
    stdout.writeln(_colorize('→ Nova imagem: $filename', _AnsiColors.magenta));
  }
  
  /// Log simples sem formatação
  static void plain(String message) {
    stdout.writeln(message);
  }
}
```

---

### 3.6 lib/src/utils/version_compare.dart

```dart
// lib/src/utils/version_compare.dart

/// Resultado da comparação de versões
enum VersionComparison {
  /// Versão local é mais antiga (atualização disponível)
  older,
  /// Versões são iguais
  equal,
  /// Versão local é mais recente (dev/pre-release?)
  newer,
}

/// Compara versões semânticas (SemVer simplificado)
/// 
/// Suporta formatos:
/// - "1.0.0"
/// - "v1.0.0" (ignora prefixo 'v')
/// - "1.0.0-beta" (ignora sufixos pre-release para simplificar)
class VersionCompare {
  
  /// Limpa a string de versão removendo prefixos e sufixos
  static String normalize(String version) {
    var v = version.trim().toLowerCase();
    // Remover prefixo 'v'
    if (v.startsWith('v')) {
      v = v.substring(1);
    }
    // Remover sufixos como -beta, -alpha, +build
    final dashIndex = v.indexOf('-');
    if (dashIndex != -1) {
      v = v.substring(0, dashIndex);
    }
    final plusIndex = v.indexOf('+');
    if (plusIndex != -1) {
      v = v.substring(0, plusIndex);
    }
    return v;
  }
  
  /// Converte versão em lista de inteiros [major, minor, patch]
  static List<int> parse(String version) {
    final normalized = normalize(version);
    final parts = normalized.split('.');
    
    final result = <int>[];
    for (var i = 0; i < 3; i++) {
      if (i < parts.length) {
        result.add(int.tryParse(parts[i]) ?? 0);
      } else {
        result.add(0);
      }
    }
    return result;
  }
  
  /// Compara duas versões
  /// 
  /// Retorna:
  /// - [VersionComparison.older] se local < remote
  /// - [VersionComparison.equal] se local == remote
  /// - [VersionComparison.newer] se local > remote
  static VersionComparison compare(String local, String remote) {
    final localParts = parse(local);
    final remoteParts = parse(remote);
    
    for (var i = 0; i < 3; i++) {
      if (localParts[i] < remoteParts[i]) {
        return VersionComparison.older;
      }
      if (localParts[i] > remoteParts[i]) {
        return VersionComparison.newer;
      }
    }
    return VersionComparison.equal;
  }
  
  /// Verifica se há atualização disponível
  static bool hasUpdate(String local, String remote) {
    return compare(local, remote) == VersionComparison.older;
  }
}
```

---

### 3.7 lib/src/cli.dart

```dart
// lib/src/cli.dart

import 'package:args/args.dart';

/// Configuração parseada dos argumentos CLI
class CliConfig {
  /// Diretório a monitorar
  final String directory;
  
  /// Porta especificada (null = auto)
  final int? port;
  
  /// Executar atualização
  final bool update;
  
  /// Exibir versão
  final bool showVersion;
  
  /// Exibir ajuda
  final bool showHelp;
  
  const CliConfig({
    required this.directory,
    this.port,
    this.update = false,
    this.showVersion = false,
    this.showHelp = false,
  });
}

/// Parser de argumentos de linha de comando
class CliParser {
  late final ArgParser _parser;
  
  CliParser() {
    _parser = ArgParser()
      ..addOption(
        'port',
        abbr: 'p',
        help: 'Porta do servidor HTTP (padrão: 8080)',
        valueHelp: 'número',
      )
      ..addFlag(
        'update',
        abbr: 'u',
        negatable: false,
        help: 'Atualizar para a versão mais recente',
      )
      ..addFlag(
        'version',
        abbr: 'v',
        negatable: false,
        help: 'Exibir versão atual',
      )
      ..addFlag(
        'help',
        abbr: 'h',
        negatable: false,
        help: 'Exibir esta ajuda',
      );
  }
  
  /// Gera texto de ajuda
  String get usage => '''
sidelook - Visualizador de imagens em tempo real

Uso: sidelook [opções] [diretório]

Opções:
${_parser.usage}

Exemplos:
  sidelook                    # Monitora diretório atual
  sidelook ~/Downloads        # Monitora pasta Downloads
  sidelook -p 3000            # Usa porta 3000
  sidelook --update           # Atualiza para versão mais recente
''';
  
  /// Faz parse dos argumentos
  /// 
  /// Lança [FormatException] se argumentos inválidos
  CliConfig parse(List<String> args) {
    final ArgResults results;
    try {
      results = _parser.parse(args);
    } on FormatException catch (e) {
      throw FormatException('Erro ao processar argumentos: ${e.message}');
    }
    
    // Diretório é o primeiro argumento posicional, ou '.' se não especificado
    final directory = results.rest.isEmpty ? '.' : results.rest.first;
    
    // Parse da porta
    int? port;
    final portStr = results.option('port');
    if (portStr != null) {
      port = int.tryParse(portStr);
      if (port == null || port < 1 || port > 65535) {
        throw FormatException(
          'Porta inválida: "$portStr". Use um número entre 1 e 65535.',
        );
      }
    }
    
    return CliConfig(
      directory: directory,
      port: port,
      update: results.flag('update'),
      showVersion: results.flag('version'),
      showHelp: results.flag('help'),
    );
  }
}
```

---

### 3.8 lib/src/watcher.dart

```dart
// lib/src/watcher.dart

import 'dart:async';
import 'dart:io';
import 'package:path/path.dart' as p;
import 'package:watcher/watcher.dart';

/// Extensões de imagem suportadas (lowercase)
const supportedImageExtensions = {
  '.jpg',
  '.jpeg',
  '.png',
  '.gif',
  '.webp',
  '.svg',
  '.bmp',
  '.tiff',
  '.tif',
};

/// Verifica se um arquivo é uma imagem suportada
bool isImageFile(String path) {
  final ext = p.extension(path).toLowerCase();
  return supportedImageExtensions.contains(ext);
}

/// Resultado do scan inicial de imagens
class ImageScanResult {
  /// Imagem mais recente encontrada (null se nenhuma)
  final File? mostRecent;
  
  /// Total de imagens encontradas
  final int count;
  
  const ImageScanResult({this.mostRecent, required this.count});
}

/// Monitor de diretório para imagens
class ImageWatcher {
  final String directoryPath;
  final Directory _directory;
  
  DirectoryWatcher? _watcher;
  StreamSubscription<WatchEvent>? _subscription;
  
  /// Stream de novas imagens detectadas
  final _imageController = StreamController<File>.broadcast();
  Stream<File> get onNewImage => _imageController.stream;
  
  /// Imagem atual sendo exibida
  File? _currentImage;
  File? get currentImage => _currentImage;
  
  ImageWatcher(this.directoryPath)
      : _directory = Directory(directoryPath);
  
  /// Verifica se o diretório existe e é válido
  Future<bool> validate() async {
    if (!await _directory.exists()) {
      return false;
    }
    // Verificar se é um diretório (não um arquivo)
    final stat = await _directory.stat();
    return stat.type == FileSystemEntityType.directory;
  }
  
  /// Faz scan inicial e retorna a imagem mais recente
  Future<ImageScanResult> scanExisting() async {
    File? mostRecent;
    DateTime? mostRecentTime;
    var count = 0;
    
    await for (final entity in _directory.list()) {
      if (entity is File && isImageFile(entity.path)) {
        count++;
        try {
          final modified = await entity.lastModified();
          if (mostRecentTime == null || modified.isAfter(mostRecentTime)) {
            mostRecentTime = modified;
            mostRecent = entity;
          }
        } catch (_) {
          // Ignorar arquivos que não conseguimos ler
        }
      }
    }
    
    _currentImage = mostRecent;
    return ImageScanResult(mostRecent: mostRecent, count: count);
  }
  
  /// Inicia o monitoramento
  Future<void> start() async {
    _watcher = DirectoryWatcher(directoryPath);
    _subscription = _watcher!.events.listen(_handleEvent);
  }
  
  void _handleEvent(WatchEvent event) {
    // Interessados apenas em criação e modificação
    if (event.type == ChangeType.REMOVE) return;
    
    final path = event.path;
    if (!isImageFile(path)) return;
    
    final file = File(path);
    
    // Verificar se o arquivo existe (pode ter sido deletado rapidamente)
    if (!file.existsSync()) return;
    
    // Atualizar imagem atual e notificar
    _currentImage = file;
    _imageController.add(file);
  }
  
  /// Para o monitoramento e libera recursos
  Future<void> stop() async {
    await _subscription?.cancel();
    await _imageController.close();
  }
  
  /// Obtém o nome do arquivo relativo ao diretório monitorado
  String getRelativePath(File file) {
    return p.relative(file.path, from: directoryPath);
  }
}
```

**ATENÇÃO - PONTO CRÍTICO:**
O `package:watcher` pode ter comportamentos diferentes em cada SO:
- Linux: usa `inotify` (eficiente)
- macOS: usa `FSEvents` (eficiente)
- Windows: usa polling (pode ter delay)

Isso é transparente para o código, mas pode afetar testes.

---

### 3.9 lib/src/html_assets.dart

```dart
// lib/src/html_assets.dart

/// Gera o HTML completo da página do visualizador
/// 
/// [wsPort] - Porta para conexão WebSocket
/// [initialImage] - Caminho da imagem inicial (null = aguardando)
String generateHtml({
  required int wsPort,
  String? initialImage,
}) {
  final imageDisplay = initialImage != null
      ? '<img id="viewer" src="/image/$initialImage" alt="Imagem">'
      : '<div id="waiting">Aguardando primeira imagem...</div>';
  
  return '''
<!DOCTYPE html>
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
      width: 100%;
      height: 100%;
      background: #000;
      overflow: hidden;
    }
    
    #container {
      width: 100%;
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
      cursor: pointer;
    }
    
    #viewer {
      max-width: 100%;
      max-height: 100%;
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
    
    /* Fullscreen styles */
    :fullscreen #container,
    :-webkit-full-screen #container {
      background: #000;
    }
  </style>
</head>
<body>
  <div id="container" onclick="toggleFullscreen()">
    $imageDisplay
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
      // Usar mesmo host que a página
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
      
      // Se estava no estado "aguardando", remover mensagem
      if (waiting) {
        waiting.remove();
      }
      
      if (current) {
        // Fade out
        current.classList.add('fade-out');
        
        setTimeout(() => {
          // Criar nova imagem
          const newImg = document.createElement('img');
          newImg.id = 'viewer';
          // Adicionar timestamp para evitar cache
          newImg.src = '/image/' + imagePath + '?t=' + Date.now();
          newImg.alt = 'Imagem';
          newImg.classList.add('fade-out');
          
          newImg.onload = () => {
            current.remove();
            container.appendChild(newImg);
            // Trigger reflow para animação funcionar
            void newImg.offsetWidth;
            newImg.classList.remove('fade-out');
          };
          
          newImg.onerror = () => {
            console.error('Erro ao carregar imagem:', imagePath);
          };
        }, 200);
      } else {
        // Primeira imagem
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
    
    // Atalho de teclado: F para fullscreen, Escape para sair
    document.addEventListener('keydown', (e) => {
      if (e.key === 'f' || e.key === 'F') {
        toggleFullscreen();
      }
    });
    
    // Iniciar conexão
    connect();
  </script>
</body>
</html>
''';
}
```

---

### 3.10 lib/src/server.dart

```dart
// lib/src/server.dart

import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:path/path.dart' as p;

import 'html_assets.dart';
import 'watcher.dart';
import 'utils/logger.dart';

/// Servidor HTTP com suporte a WebSocket
class ImageServer {
  final int startPort;
  final ImageWatcher watcher;
  
  HttpServer? _server;
  int? _actualPort;
  final List<WebSocket> _clients = [];
  
  /// Porta em que o servidor está rodando
  int? get port => _actualPort;
  
  /// URL completa do servidor
  String get url => 'http://localhost:$_actualPort';
  
  ImageServer({
    required this.startPort,
    required this.watcher,
  });
  
  /// Inicia o servidor, tentando portas sequenciais se necessário
  Future<void> start() async {
    var port = startPort;
    const maxAttempts = 100;
    
    for (var attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        _server = await HttpServer.bind(
          InternetAddress.loopbackIPv4,
          port,
        );
        _actualPort = port;
        break;
      } on SocketException catch (e) {
        if (e.osError?.errorCode == 98 || // Linux: Address already in use
            e.osError?.errorCode == 48 || // macOS: Address already in use
            e.osError?.errorCode == 10048) { // Windows: WSAEADDRINUSE
          port++;
          continue;
        }
        rethrow;
      }
    }
    
    if (_server == null) {
      throw StateError(
        'Não foi possível iniciar servidor. Portas $startPort-${startPort + maxAttempts - 1} indisponíveis.',
      );
    }
    
    // Escutar requisições
    _server!.listen(_handleRequest);
    
    // Escutar novas imagens do watcher
    watcher.onNewImage.listen(_broadcastNewImage);
  }
  
  Future<void> _handleRequest(HttpRequest request) async {
    final path = request.uri.path;
    
    try {
      if (path == '/' || path.isEmpty) {
        await _serveHtml(request);
      } else if (path == '/ws') {
        await _handleWebSocket(request);
      } else if (path.startsWith('/image/')) {
        await _serveImage(request);
      } else {
        request.response.statusCode = HttpStatus.notFound;
        await request.response.close();
      }
    } catch (e) {
      Logger.error('Erro ao processar requisição: $e');
      try {
        request.response.statusCode = HttpStatus.internalServerError;
        await request.response.close();
      } catch (_) {}
    }
  }
  
  Future<void> _serveHtml(HttpRequest request) async {
    final currentImage = watcher.currentImage;
    final imagePath = currentImage != null
        ? watcher.getRelativePath(currentImage)
        : null;
    
    final html = generateHtml(
      wsPort: _actualPort!,
      initialImage: imagePath,
    );
    
    request.response
      ..statusCode = HttpStatus.ok
      ..headers.contentType = ContentType.html
      ..write(html);
    await request.response.close();
  }
  
  Future<void> _handleWebSocket(HttpRequest request) async {
    final socket = await WebSocketTransformer.upgrade(request);
    _clients.add(socket);
    
    socket.done.then((_) {
      _clients.remove(socket);
    });
  }
  
  Future<void> _serveImage(HttpRequest request) async {
    // Extrair caminho da imagem (remover /image/ prefix)
    var imagePath = request.uri.path.substring('/image/'.length);
    // Decodificar URL encoding
    imagePath = Uri.decodeComponent(imagePath);
    
    // Construir caminho completo
    final fullPath = p.join(watcher.directoryPath, imagePath);
    final file = File(fullPath);
    
    if (!await file.exists()) {
      request.response.statusCode = HttpStatus.notFound;
      await request.response.close();
      return;
    }
    
    // Verificar se é uma imagem válida
    if (!isImageFile(fullPath)) {
      request.response.statusCode = HttpStatus.forbidden;
      await request.response.close();
      return;
    }
    
    // Determinar content type
    final ext = p.extension(fullPath).toLowerCase();
    final contentType = _getContentType(ext);
    
    request.response
      ..statusCode = HttpStatus.ok
      ..headers.contentType = contentType
      ..headers.add('Cache-Control', 'no-cache, no-store, must-revalidate');
    
    await file.openRead().pipe(request.response);
  }
  
  ContentType _getContentType(String extension) {
    switch (extension) {
      case '.jpg':
      case '.jpeg':
        return ContentType('image', 'jpeg');
      case '.png':
        return ContentType('image', 'png');
      case '.gif':
        return ContentType('image', 'gif');
      case '.webp':
        return ContentType('image', 'webp');
      case '.svg':
        return ContentType('image', 'svg+xml');
      case '.bmp':
        return ContentType('image', 'bmp');
      case '.tiff':
      case '.tif':
        return ContentType('image', 'tiff');
      default:
        return ContentType('application', 'octet-stream');
    }
  }
  
  void _broadcastNewImage(File image) {
    final relativePath = watcher.getRelativePath(image);
    final message = jsonEncode({
      'type': 'new_image',
      'path': relativePath,
    });
    
    Logger.newImage(relativePath);
    
    // Enviar para todos os clientes conectados
    for (final client in _clients.toList()) {
      try {
        client.add(message);
      } catch (e) {
        // Cliente provavelmente desconectado
        _clients.remove(client);
      }
    }
  }
  
  /// Para o servidor
  Future<void> stop() async {
    // Fechar todos os WebSockets
    for (final client in _clients) {
      await client.close();
    }
    _clients.clear();
    
    await _server?.close();
  }
}
```

---

### 3.11 lib/src/browser.dart

```dart
// lib/src/browser.dart

import 'dart:io';

import 'utils/logger.dart';

/// Abre URL no navegador padrão do sistema
class BrowserLauncher {
  /// Abre a URL no navegador padrão
  /// 
  /// Retorna true se conseguiu executar o comando (não garante que abriu)
  static Future<bool> open(String url) async {
    try {
      final String command;
      final List<String> args;
      
      if (Platform.isLinux) {
        command = 'xdg-open';
        args = [url];
      } else if (Platform.isMacOS) {
        command = 'open';
        args = [url];
      } else if (Platform.isWindows) {
        command = 'cmd';
        args = ['/c', 'start', '', url];
      } else {
        Logger.warn('Sistema operacional não suportado para abertura automática de navegador.');
        Logger.info('Acesse manualmente: $url');
        return false;
      }
      
      // Executar de forma não-bloqueante
      final result = await Process.run(
        command,
        args,
        runInShell: Platform.isWindows,
      );
      
      if (result.exitCode != 0) {
        Logger.warn('Não foi possível abrir o navegador automaticamente.');
        Logger.info('Acesse manualmente: $url');
        return false;
      }
      
      return true;
    } catch (e) {
      Logger.warn('Erro ao abrir navegador: $e');
      Logger.info('Acesse manualmente: $url');
      return false;
    }
  }
}
```

---

### 3.12 lib/src/updater.dart

```dart
// lib/src/updater.dart

import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:path/path.dart' as p;

import 'version.g.dart';
import 'utils/logger.dart';
import 'utils/version_compare.dart';

/// Informações sobre uma release do GitHub
class ReleaseInfo {
  final String version;
  final String? downloadUrl;
  final DateTime? publishedAt;
  
  const ReleaseInfo({
    required this.version,
    this.downloadUrl,
    this.publishedAt,
  });
}

/// Gerenciador de atualizações via GitHub Releases
class Updater {
  static const _apiUrl = 'https://api.github.com/repos/insign/sidelook/releases/latest';
  static const _timeout = Duration(seconds: 10);
  
  /// Obtém informações da release mais recente
  static Future<ReleaseInfo?> checkLatestRelease() async {
    try {
      final response = await http.get(
        Uri.parse(_apiUrl),
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'User-Agent': 'sidelook/$packageVersion',
        },
      ).timeout(_timeout);
      
      if (response.statusCode != 200) {
        return null;
      }
      
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final tagName = data['tag_name'] as String?;
      if (tagName == null) return null;
      
      // Encontrar asset para o SO atual
      final assets = data['assets'] as List<dynamic>?;
      String? downloadUrl;
      
      if (assets != null) {
        final assetName = _getAssetName();
        for (final asset in assets) {
          if (asset['name'] == assetName) {
            downloadUrl = asset['browser_download_url'] as String?;
            break;
          }
        }
      }
      
      return ReleaseInfo(
        version: tagName,
        downloadUrl: downloadUrl,
        publishedAt: DateTime.tryParse(data['published_at'] as String? ?? ''),
      );
    } catch (e) {
      // Silenciosamente falhar - não queremos interromper o usuário
      return null;
    }
  }
  
  /// Retorna o nome do asset para o SO atual
  static String _getAssetName() {
    if (Platform.isLinux) return 'sidelook-linux';
    if (Platform.isMacOS) return 'sidelook-macos';
    if (Platform.isWindows) return 'sidelook-windows.exe';
    return 'sidelook';
  }
  
  /// Verifica atualizações em background e notifica se houver
  /// 
  /// NÃO bloqueia - retorna imediatamente
  static void checkInBackground() {
    // Executar em um Future separado (não await)
    Future(() async {
      final release = await checkLatestRelease();
      if (release == null) return;
      
      if (VersionCompare.hasUpdate(packageVersion, release.version)) {
        Logger.updateAvailable(packageVersion, release.version);
      }
    });
  }
  
  /// Executa a atualização completa
  /// 
  /// Retorna true se atualização foi bem-sucedida
  static Future<bool> performUpdate() async {
    Logger.info('Verificando atualizações...');
    
    final release = await checkLatestRelease();
    if (release == null) {
      Logger.error('Não foi possível verificar atualizações. Verifique sua conexão.');
      return false;
    }
    
    if (!VersionCompare.hasUpdate(packageVersion, release.version)) {
      Logger.success('Você já está na versão mais recente ($packageVersion)');
      return true;
    }
    
    if (release.downloadUrl == null) {
      Logger.error('Não foi encontrado download para ${Platform.operatingSystem}');
      return false;
    }
    
    Logger.info('Baixando versão ${release.version}...');
    
    try {
      // Baixar para arquivo temporário
      final response = await http.get(Uri.parse(release.downloadUrl!)).timeout(
        const Duration(minutes: 5),
      );
      
      if (response.statusCode != 200) {
        Logger.error('Falha no download (HTTP ${response.statusCode})');
        return false;
      }
      
      // Salvar em arquivo temporário
      final tempDir = Directory.systemTemp;
      final tempFile = File(p.join(tempDir.path, 'sidelook_update_${DateTime.now().millisecondsSinceEpoch}'));
      await tempFile.writeAsBytes(response.bodyBytes);
      
      // Obter caminho do executável atual
      final currentExe = File(Platform.resolvedExecutable);
      final currentPath = currentExe.path;
      
      Logger.info('Instalando atualização...');
      
      if (Platform.isWindows) {
        // Windows: não consegue sobrescrever executável em uso
        // Criar script batch para fazer a substituição após sair
        final batchContent = '''
@echo off
timeout /t 1 /nobreak >nul
del "$currentPath"
move "${tempFile.path}" "$currentPath"
echo Atualização concluída!
''';
        final batchFile = File(p.join(tempDir.path, 'sidelook_update.bat'));
        await batchFile.writeAsString(batchContent);
        
        Logger.warn('No Windows, execute manualmente: ${batchFile.path}');
        Logger.info('Ou baixe a nova versão em: ${release.downloadUrl}');
        return true;
      } else {
        // Unix: substituição atômica via rename
        final backupPath = '$currentPath.old';
        
        // Dar permissão de execução ao novo arquivo
        await Process.run('chmod', ['+x', tempFile.path]);
        
        // Backup do atual (opcional, pode falhar)
        try {
          if (await File(backupPath).exists()) {
            await File(backupPath).delete();
          }
          await currentExe.rename(backupPath);
        } catch (_) {
          // Ignorar erro no backup
        }
        
        // Mover novo para posição do atual
        await tempFile.rename(currentPath);
        
        // Limpar backup
        try {
          await File(backupPath).delete();
        } catch (_) {}
        
        Logger.success('Atualizado para versão ${release.version}!');
        return true;
      }
    } catch (e) {
      Logger.error('Erro durante atualização: $e');
      return false;
    }
  }
}
```

**ATENÇÃO - PONTOS CRÍTICOS:**
1. A API do GitHub tem rate limit (60 req/hora para não autenticados)
2. No Windows, não é possível deletar/renomear um executável em uso
3. Em Unix, `rename()` é atômico se no mesmo filesystem

---

### 3.13 bin/sidelook.dart (Entry Point)

```dart
// bin/sidelook.dart

import 'dart:io';

import 'package:sidelook/src/cli.dart';
import 'package:sidelook/src/browser.dart';
import 'package:sidelook/src/server.dart';
import 'package:sidelook/src/updater.dart';
import 'package:sidelook/src/version.g.dart';
import 'package:sidelook/src/watcher.dart';
import 'package:sidelook/src/utils/logger.dart';

Future<void> main(List<String> args) async {
  final parser = CliParser();
  
  // Parse de argumentos
  final CliConfig config;
  try {
    config = parser.parse(args);
  } on FormatException catch (e) {
    Logger.error(e.message);
    exit(1);
  }
  
  // --help
  if (config.showHelp) {
    Logger.plain(parser.usage);
    exit(0);
  }
  
  // --version
  if (config.showVersion) {
    Logger.plain('$packageName $packageVersion');
    exit(0);
  }
  
  // --update
  if (config.update) {
    final success = await Updater.performUpdate();
    exit(success ? 0 : 1);
  }
  
  // Modo normal: iniciar servidor
  await _runServer(config);
}

Future<void> _runServer(CliConfig config) async {
  // Verificar diretório
  final watcher = ImageWatcher(config.directory);
  if (!await watcher.validate()) {
    Logger.error('Diretório não encontrado: ${config.directory}');
    exit(1);
  }
  
  // Verificar atualizações em background (não bloqueia)
  Updater.checkInBackground();
  
  // Scan inicial de imagens
  final scanResult = await watcher.scanExisting();
  if (scanResult.count > 0) {
    Logger.info('${scanResult.count} imagem(ns) encontrada(s)');
  } else {
    Logger.info('Nenhuma imagem encontrada. Aguardando...');
  }
  
  // Iniciar watcher
  await watcher.start();
  
  // Iniciar servidor
  final server = ImageServer(
    startPort: config.port ?? 8080,
    watcher: watcher,
  );
  
  try {
    await server.start();
  } catch (e) {
    Logger.error('Falha ao iniciar servidor: $e');
    await watcher.stop();
    exit(1);
  }
  
  Logger.serverStarted(server.url);
  
  // Abrir navegador
  await BrowserLauncher.open(server.url);
  
  // Configurar handler para SIGINT (Ctrl+C)
  ProcessSignal.sigint.watch().listen((_) async {
    Logger.info('Encerrando...');
    await server.stop();
    await watcher.stop();
    exit(0);
  });
  
  // Manter processo rodando
  // O servidor já mantém o event loop vivo, mas podemos adicionar
  // um Completer se necessário
}
```

---

### 3.14 lib/sidelook.dart (Exports)

```dart
// lib/sidelook.dart
// Exporta APIs públicas (principalmente para testes)

export 'src/version.g.dart';
export 'src/cli.dart';
export 'src/watcher.dart';
export 'src/server.dart';
export 'src/browser.dart';
export 'src/updater.dart';
export 'src/utils/logger.dart';
export 'src/utils/version_compare.dart';
```

---

## 4. Testes

### 4.1 test/version_compare_test.dart

```dart
import 'package:test/test.dart';
import 'package:sidelook/src/utils/version_compare.dart';

void main() {
  group('VersionCompare', () {
    group('normalize', () {
      test('remove prefixo v', () {
        expect(VersionCompare.normalize('v1.2.3'), equals('1.2.3'));
      });
      
      test('remove sufixo -beta', () {
        expect(VersionCompare.normalize('1.2.3-beta'), equals('1.2.3'));
      });
      
      test('remove sufixo +build', () {
        expect(VersionCompare.normalize('1.2.3+456'), equals('1.2.3'));
      });
      
      test('mantém versão limpa', () {
        expect(VersionCompare.normalize('1.2.3'), equals('1.2.3'));
      });
    });
    
    group('parse', () {
      test('versão completa', () {
        expect(VersionCompare.parse('1.2.3'), equals([1, 2, 3]));
      });
      
      test('versão sem patch', () {
        expect(VersionCompare.parse('1.2'), equals([1, 2, 0]));
      });
      
      test('versão só major', () {
        expect(VersionCompare.parse('1'), equals([1, 0, 0]));
      });
    });
    
    group('compare', () {
      test('versão igual', () {
        expect(
          VersionCompare.compare('1.0.0', '1.0.0'),
          equals(VersionComparison.equal),
        );
      });
      
      test('local mais antiga (major)', () {
        expect(
          VersionCompare.compare('1.0.0', '2.0.0'),
          equals(VersionComparison.older),
        );
      });
      
      test('local mais antiga (minor)', () {
        expect(
          VersionCompare.compare('1.1.0', '1.2.0'),
          equals(VersionComparison.older),
        );
      });
      
      test('local mais antiga (patch)', () {
        expect(
          VersionCompare.compare('1.0.0', '1.0.1'),
          equals(VersionComparison.older),
        );
      });
      
      test('local mais recente', () {
        expect(
          VersionCompare.compare('2.0.0', '1.0.0'),
          equals(VersionComparison.newer),
        );
      });
      
      test('ignora prefixo v', () {
        expect(
          VersionCompare.compare('v1.0.0', '1.0.0'),
          equals(VersionComparison.equal),
        );
      });
    });
    
    group('hasUpdate', () {
      test('retorna true quando há atualização', () {
        expect(VersionCompare.hasUpdate('1.0.0', '1.0.1'), isTrue);
      });
      
      test('retorna false quando versões iguais', () {
        expect(VersionCompare.hasUpdate('1.0.0', '1.0.0'), isFalse);
      });
      
      test('retorna false quando local é mais recente', () {
        expect(VersionCompare.hasUpdate('2.0.0', '1.0.0'), isFalse);
      });
    });
  });
}
```

### 4.2 test/watcher_test.dart

```dart
import 'dart:io';
import 'package:test/test.dart';
import 'package:sidelook/src/watcher.dart';
import 'package:path/path.dart' as p;

void main() {
  group('isImageFile', () {
    test('aceita jpg', () => expect(isImageFile('foto.jpg'), isTrue));
    test('aceita jpeg', () => expect(isImageFile('foto.jpeg'), isTrue));
    test('aceita png', () => expect(isImageFile('foto.PNG'), isTrue)); // case insensitive
    test('aceita gif', () => expect(isImageFile('animacao.gif'), isTrue));
    test('aceita webp', () => expect(isImageFile('imagem.webp'), isTrue));
    test('aceita svg', () => expect(isImageFile('vetor.svg'), isTrue));
    test('aceita bmp', () => expect(isImageFile('bitmap.bmp'), isTrue));
    test('aceita tiff', () => expect(isImageFile('scan.tiff'), isTrue));
    test('aceita tif', () => expect(isImageFile('scan.tif'), isTrue));
    
    test('rejeita txt', () => expect(isImageFile('arquivo.txt'), isFalse));
    test('rejeita pdf', () => expect(isImageFile('doc.pdf'), isFalse));
    test('rejeita sem extensão', () => expect(isImageFile('arquivo'), isFalse));
  });
  
  group('ImageWatcher', () {
    late Directory tempDir;
    
    setUp(() async {
      tempDir = await Directory.systemTemp.createTemp('sidelook_test_');
    });
    
    tearDown(() async {
      await tempDir.delete(recursive: true);
    });
    
    test('validate retorna true para diretório válido', () async {
      final watcher = ImageWatcher(tempDir.path);
      expect(await watcher.validate(), isTrue);
    });
    
    test('validate retorna false para diretório inexistente', () async {
      final watcher = ImageWatcher('/caminho/que/nao/existe');
      expect(await watcher.validate(), isFalse);
    });
    
    test('scanExisting encontra imagem mais recente', () async {
      // Criar algumas imagens com diferentes timestamps
      final img1 = File(p.join(tempDir.path, 'img1.png'));
      await img1.writeAsBytes([0x89, 0x50, 0x4E, 0x47]); // PNG magic bytes
      await Future.delayed(Duration(milliseconds: 100));
      
      final img2 = File(p.join(tempDir.path, 'img2.jpg'));
      await img2.writeAsBytes([0xFF, 0xD8, 0xFF]); // JPEG magic bytes
      
      final watcher = ImageWatcher(tempDir.path);
      final result = await watcher.scanExisting();
      
      expect(result.count, equals(2));
      expect(result.mostRecent?.path, equals(img2.path));
    });
    
    test('scanExisting retorna null se não há imagens', () async {
      // Criar arquivo não-imagem
      final txt = File(p.join(tempDir.path, 'readme.txt'));
      await txt.writeAsString('Hello');
      
      final watcher = ImageWatcher(tempDir.path);
      final result = await watcher.scanExisting();
      
      expect(result.count, equals(0));
      expect(result.mostRecent, isNull);
    });
    
    test('detecta nova imagem', () async {
      final watcher = ImageWatcher(tempDir.path);
      await watcher.start();
      
      // Criar listener antes de adicionar arquivo
      final completer = Completer<File>();
      watcher.onNewImage.first.then(completer.complete);
      
      // Dar tempo para watcher inicializar
      await Future.delayed(Duration(milliseconds: 100));
      
      // Criar nova imagem
      final newImg = File(p.join(tempDir.path, 'nova.png'));
      await newImg.writeAsBytes([0x89, 0x50, 0x4E, 0x47]);
      
      // Aguardar detecção (com timeout)
      final detected = await completer.future.timeout(
        Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('Imagem não detectada'),
      );
      
      expect(detected.path, equals(newImg.path));
      
      await watcher.stop();
    }, timeout: Timeout(Duration(seconds: 10)));
  });
}
```

### 4.3 test/cli_test.dart

```dart
import 'package:test/test.dart';
import 'package:sidelook/src/cli.dart';

void main() {
  late CliParser parser;
  
  setUp(() {
    parser = CliParser();
  });
  
  group('CliParser', () {
    test('parse sem argumentos usa diretório atual', () {
      final config = parser.parse([]);
      expect(config.directory, equals('.'));
      expect(config.port, isNull);
      expect(config.update, isFalse);
      expect(config.showVersion, isFalse);
      expect(config.showHelp, isFalse);
    });
    
    test('parse com diretório', () {
      final config = parser.parse(['/home/user/imagens']);
      expect(config.directory, equals('/home/user/imagens'));
    });
    
    test('parse --port com valor válido', () {
      final config = parser.parse(['-p', '3000']);
      expect(config.port, equals(3000));
    });
    
    test('parse -p abreviado', () {
      final config = parser.parse(['-p', '8000']);
      expect(config.port, equals(8000));
    });
    
    test('parse --port inválido lança exceção', () {
      expect(
        () => parser.parse(['--port', 'abc']),
        throwsA(isA<FormatException>()),
      );
    });
    
    test('parse --port fora do range lança exceção', () {
      expect(
        () => parser.parse(['--port', '99999']),
        throwsA(isA<FormatException>()),
      );
    });
    
    test('parse --update', () {
      final config = parser.parse(['--update']);
      expect(config.update, isTrue);
    });
    
    test('parse -u abreviado', () {
      final config = parser.parse(['-u']);
      expect(config.update, isTrue);
    });
    
    test('parse --version', () {
      final config = parser.parse(['--version']);
      expect(config.showVersion, isTrue);
    });
    
    test('parse --help', () {
      final config = parser.parse(['--help']);
      expect(config.showHelp, isTrue);
    });
    
    test('parse combinação de argumentos', () {
      final config = parser.parse(['-p', '9000', '/tmp/imagens']);
      expect(config.port, equals(9000));
      expect(config.directory, equals('/tmp/imagens'));
    });
    
    test('usage contém informações essenciais', () {
      final usage = parser.usage;
      expect(usage, contains('sidelook'));
      expect(usage, contains('--port'));
      expect(usage, contains('--update'));
      expect(usage, contains('--version'));
      expect(usage, contains('--help'));
    });
  });
}
```

---

## 5. CI/CD

### 5.1 .github/workflows/release.yml

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  # Job 1: Análise e Testes
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: dart-lang/setup-dart@v1
        with:
          sdk: stable
      
      - name: Instalar dependências
        run: dart pub get
      
      - name: Gerar version.g.dart
        run: dart run tool/generate_version.dart
      
      - name: Verificar formatação
        run: dart format --set-exit-if-changed .
      
      - name: Analisar código
        run: dart analyze --fatal-infos
      
      - name: Executar testes
        run: dart test --concurrency=4

  # Job 2: Build para cada plataforma
  build-linux:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: dart-lang/setup-dart@v1
        with:
          sdk: stable
      
      - run: dart pub get
      - run: dart run tool/generate_version.dart
      
      - name: Compilar executável
        run: dart compile exe bin/sidelook.dart -o sidelook-linux
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: sidelook-linux
          path: sidelook-linux

  build-macos:
    needs: test
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: dart-lang/setup-dart@v1
        with:
          sdk: stable
      
      - run: dart pub get
      - run: dart run tool/generate_version.dart
      
      - name: Compilar executável
        run: dart compile exe bin/sidelook.dart -o sidelook-macos
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: sidelook-macos
          path: sidelook-macos

  build-windows:
    needs: test
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: dart-lang/setup-dart@v1
        with:
          sdk: stable
      
      - run: dart pub get
      - run: dart run tool/generate_version.dart
      
      - name: Compilar executável
        run: dart compile exe bin/sidelook.dart -o sidelook-windows.exe
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: sidelook-windows
          path: sidelook-windows.exe

  # Job 3: Criar Release
  release:
    needs: [build-linux, build-macos, build-windows]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download todos os artifacts
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
        id: release_notes
        run: |
          # Extrair versão da tag
          VERSION=${GITHUB_REF#refs/tags/}
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          
          # Tentar extrair do CHANGELOG, ou usar mensagem padrão
          if [ -f CHANGELOG.md ]; then
            # Pegar seção da versão atual
            NOTES=$(sed -n "/^## \[$VERSION\]/,/^## \[/p" CHANGELOG.md | sed '1d;$d' | head -50)
          fi
          
          if [ -z "$NOTES" ]; then
            NOTES="Release $VERSION"
          fi
          
          # Salvar em arquivo para multiline
          echo "$NOTES" > release_notes.txt
      
      - name: Criar Release
        uses: softprops/action-gh-release@v1
        with:
          name: ${{ steps.release_notes.outputs.version }}
          body_path: release_notes.txt
          files: |
            release-assets/sidelook-linux
            release-assets/sidelook-macos
            release-assets/sidelook-windows.exe
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 6. Documentação

### 6.1 README.md

```markdown
# sidelook

Visualizador de imagens em tempo real via navegador.

## Funcionalidades

- 🖼 Monitora um diretório por novas imagens
- 🌐 Serve as imagens via HTTP local
- ⚡ Atualização em tempo real via WebSocket (sem refresh)
- 🖥 Abre navegador automaticamente
- 🔄 Auto-update integrado
- 🎯 Fullscreen ao clicar na imagem

## Instalação

### Download Direto

Baixe o executável para seu sistema operacional na [página de releases](https://github.com/insign/sidelook/releases/latest):

- **Linux**: `sidelook-linux`
- **macOS**: `sidelook-macos`
- **Windows**: `sidelook-windows.exe`

Depois de baixar, dê permissão de execução (Linux/macOS):

```bash
chmod +x sidelook-linux
sudo mv sidelook-linux /usr/local/bin/sidelook
```

## Uso

```bash
# Monitorar diretório atual
sidelook

# Monitorar pasta específica
sidelook ~/Downloads

# Especificar porta
sidelook -p 3000

# Atualizar para versão mais recente
sidelook --update

# Ver versão
sidelook --version

# Ajuda
sidelook --help
```

## Formatos Suportados

- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif`)
- WebP (`.webp`)
- SVG (`.svg`)
- BMP (`.bmp`)
- TIFF (`.tiff`, `.tif`)

## Desenvolvimento

```bash
# Clonar repositório
git clone https://github.com/insign/sidelook.git
cd sidelook

# Instalar dependências
dart pub get

# Gerar arquivo de versão
dart run tool/generate_version.dart

# Rodar em modo desenvolvimento
dart run bin/sidelook.dart

# Executar testes
dart test

# Compilar executável
dart compile exe bin/sidelook.dart -o sidelook
```

## Licença

MIT
```

### 6.2 CHANGELOG.md

```markdown
# Changelog

Todas as mudanças notáveis deste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [Unreleased]

## [0.1.0] - YYYY-MM-DD

### Added
- Monitoramento de diretório para novas imagens
- Servidor HTTP local com WebSocket
- Interface web com fundo preto e centralização
- Transição suave entre imagens (fade + scale)
- Fullscreen ao clicar (ou tecla F)
- Auto-detecção de porta disponível
- Abertura automática do navegador
- Verificação de atualizações em background
- Comando --update para auto-atualização
- Suporte a múltiplos formatos de imagem
- Builds para Linux, macOS e Windows
```

### 6.3 LICENSE

```
MIT License

Copyright (c) 2024 insign

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## 7. Comandos de Verificação Final

Após implementar tudo, execute estes comandos para garantir que está funcionando:

```bash
# 1. Gerar versão
dart run tool/generate_version.dart

# 2. Verificar formatação
dart format .

# 3. Analisar código
dart analyze

# 4. Executar testes
dart test

# 5. Testar manualmente
dart run bin/sidelook.dart /tmp

# 6. Compilar executável (opcional)
dart compile exe bin/sidelook.dart -o sidelook

# 7. Testar executável
./sidelook --version
./sidelook --help
./sidelook /tmp
```

---

## 8. Checklist de Conclusão

- [ ] `pubspec.yaml` criado e dependências instaladas (`dart pub get`)
- [ ] `analysis_options.yaml` criado
- [ ] `dart_test.yaml` criado
- [ ] `tool/generate_version.dart` criado e executado
- [ ] `lib/src/version.g.dart` gerado automaticamente
- [ ] `lib/src/utils/logger.dart` implementado
- [ ] `lib/src/utils/version_compare.dart` implementado
- [ ] `lib/src/cli.dart` implementado
- [ ] `lib/src/watcher.dart` implementado
- [ ] `lib/src/html_assets.dart` implementado
- [ ] `lib/src/server.dart` implementado
- [ ] `lib/src/browser.dart` implementado
- [ ] `lib/src/updater.dart` implementado
- [ ] `lib/sidelook.dart` (exports) criado
- [ ] `bin/sidelook.dart` (main) implementado
- [ ] Todos os testes passando (`dart test`)
- [ ] `dart analyze` sem erros
- [ ] `dart format` aplicado
- [ ] README.md escrito
- [ ] CHANGELOG.md iniciado
- [ ] LICENSE adicionado
- [ ] `.github/workflows/release.yml` configurado
- [ ] Teste manual funcionando

---

## 9. Troubleshooting Comum

### Erro: "Could not find a file named 'pubspec.yaml'"
- Verifique se está no diretório correto do projeto
- Execute `pwd` para confirmar

### Erro: "Target URI doesn't exist" para version.g.dart
- Execute `dart run tool/generate_version.dart` primeiro

### Erro de porta em uso
- O servidor tentará automaticamente as próximas portas
- Use `-p NUMERO` para especificar outra porta

### WebSocket não conecta
- Verifique se não há firewall bloqueando
- Tente acessar manualmente `http://localhost:PORTA`

### Imagens não atualizam
- Verifique se o arquivo tem extensão suportada
- Confirme que está salvando no diretório correto
- No Windows, pode haver delay devido ao polling

### Testes falhando com timeout
- Aumente o timeout nos testes que dependem de filesystem events
- No Windows, eventos podem demorar mais

---

**FIM DO DOCUMENTO**

Siga este plano na ordem apresentada. Cada seção depende das anteriores.
Teste frequentemente. Não pule etapas.

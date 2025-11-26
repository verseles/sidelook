# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Detecção automática de deleção/movimentação de imagem com troca para a próxima mais recente
- Suporte a eventos RENAME e REMOVE para detectar quando imagens são deletadas ou movidas para lixeira
- Transição suave ao trocar de imagem após deleção

### Fixed
- Correção de duplicação de imagens ao receber nova imagem via WebSocket
- Imagem quebrada quando a atual é deletada agora atualiza automaticamente

### Initial
- Initial release of sidelook.

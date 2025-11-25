import 'dart:io';
import 'package:test/test.dart';
import 'package:sidelook/src/server.dart';
import 'package:sidelook/src/watcher.dart';

void main() {
  group('ImageServer', () {
    late Directory tempDir;
    late ImageWatcher watcher;
    late ImageServer server;

    setUp(() async {
      tempDir = await Directory.systemTemp.createTemp('sidelook_test_');
      watcher = ImageWatcher(tempDir.path);
      server = ImageServer(startPort: 8080, watcher: watcher);
    });

    tearDown(() async {
      await server.stop();
      await tempDir.delete(recursive: true);
    });

    test('start e stop sem erro', () async {
      await server.start();
      expect(server.port, isNotNull);
      expect(server.port, greaterThanOrEqualTo(8080));

      // Stop não deve lançar exceção mesmo sem clientes conectados
      await server.stop();
    });

    test('stop com WebSocket conectado não lança exceção', () async {
      await server.start();

      // Conectar um WebSocket
      await WebSocket.connect('ws://localhost:${server.port}/ws');

      // Aguardar conexão se estabelecer
      await Future<void>.delayed(Duration(milliseconds: 100));

      // Stop não deve lançar exceção de modificação concorrente
      // Este era o bug: "Concurrent modification during iteration"
      await expectLater(server.stop(), completes);
    });
  });
}

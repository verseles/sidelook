import 'dart:async';
import 'dart:io';
import 'package:test/test.dart';
import 'package:sidelook/src/watcher.dart';
import 'package:path/path.dart' as p;

/// Helper para indicar que um Future não será awaited intencionalmente
void unawaited(Future<void> future) {}

void main() {
  group('isImageFile', () {
    test('aceita jpg', () => expect(isImageFile('foto.jpg'), isTrue));
    test('aceita jpeg', () => expect(isImageFile('foto.jpeg'), isTrue));
    test('aceita png',
        () => expect(isImageFile('foto.PNG'), isTrue)); // case insensitive
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
      await Future<void>.delayed(Duration(seconds: 1));

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
      unawaited(watcher.onNewImage.first.then(completer.complete));

      // Dar tempo para watcher inicializar
      await Future<void>.delayed(Duration(milliseconds: 100));

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

    test('detecta remoção da imagem atual e atualiza para próxima', () async {
      final tempDir = await Directory.systemTemp.createTemp('sidelook_test_');
      addTearDown(() => tempDir.delete(recursive: true));

      // Criar duas imagens com delay para garantir ordem de modificação
      final img1 = File(p.join(tempDir.path, 'primeira.jpg'));
      await img1.writeAsBytes([0xFF, 0xD8, 0xFF, 0xE0]);
      await Future<void>.delayed(Duration(seconds: 1));

      final img2 = File(p.join(tempDir.path, 'segunda.jpg'));
      await img2.writeAsBytes([0xFF, 0xD8, 0xFF, 0xE0]);

      final watcher = ImageWatcher(tempDir.path);
      final result = await watcher.scanExisting();

      // img2 deve ser a mais recente
      expect(result.mostRecent?.path, equals(img2.path));
      expect(watcher.currentImage?.path, equals(img2.path));

      await watcher.start();

      // Configurar listener para capturar a próxima imagem
      final completer = Completer<File>();
      unawaited(watcher.onNewImage.first.then(completer.complete));

      // Dar tempo para watcher inicializar
      await Future<void>.delayed(Duration(milliseconds: 100));

      // Remover a imagem atual (img2)
      await img2.delete();

      // Aguardar que o watcher detecte e notifique a próxima imagem
      final nextImage = await completer.future.timeout(
        Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('Próxima imagem não detectada'),
      );

      // Deve notificar img1 como a nova imagem mais recente
      expect(nextImage.path, equals(img1.path));
      expect(watcher.currentImage?.path, equals(img1.path));

      await watcher.stop();
    }, timeout: Timeout(Duration(seconds: 10)));

    test('detecta remoção quando não há mais imagens', () async {
      final tempDir = await Directory.systemTemp.createTemp('sidelook_test_');
      addTearDown(() => tempDir.delete(recursive: true));

      // Criar apenas uma imagem
      final img = File(p.join(tempDir.path, 'unica.jpg'));
      await img.writeAsBytes([0xFF, 0xD8, 0xFF, 0xE0]);

      final watcher = ImageWatcher(tempDir.path);
      await watcher.scanExisting();
      expect(watcher.currentImage?.path, equals(img.path));

      await watcher.start();
      await Future<void>.delayed(Duration(milliseconds: 100));

      // Remover a única imagem
      await img.delete();

      // Aguardar processamento
      await Future<void>.delayed(Duration(seconds: 2));

      // currentImage deve ser null
      expect(watcher.currentImage, isNull);

      await watcher.stop();
    }, timeout: Timeout(Duration(seconds: 10)));

    test('ignora remoção de imagem que não é a atual', () async {
      final tempDir = await Directory.systemTemp.createTemp('sidelook_test_');
      addTearDown(() => tempDir.delete(recursive: true));

      // Criar duas imagens
      final img1 = File(p.join(tempDir.path, 'primeira.jpg'));
      await img1.writeAsBytes([0xFF, 0xD8, 0xFF, 0xE0]);
      await Future<void>.delayed(Duration(seconds: 1));

      final img2 = File(p.join(tempDir.path, 'segunda.jpg'));
      await img2.writeAsBytes([0xFF, 0xD8, 0xFF, 0xE0]);

      final watcher = ImageWatcher(tempDir.path);
      await watcher.scanExisting();

      // img2 é a atual
      expect(watcher.currentImage?.path, equals(img2.path));

      await watcher.start();
      await Future<void>.delayed(Duration(milliseconds: 100));

      // Remover img1 (não é a atual)
      await img1.delete();

      // Aguardar
      await Future<void>.delayed(Duration(seconds: 1));

      // currentImage ainda deve ser img2
      expect(watcher.currentImage?.path, equals(img2.path));

      await watcher.stop();
    }, timeout: Timeout(Duration(seconds: 10)));
  });
}

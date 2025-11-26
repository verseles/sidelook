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

func TestImageWatcher_DetectImageDeletion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Criar duas imagens
	img1 := filepath.Join(tmpDir, "img1.png")
	if err := os.WriteFile(img1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	img2 := filepath.Join(tmpDir, "img2.jpg")
	if err := os.WriteFile(img2, []byte{0xFF, 0xD8, 0xFF}, 0644); err != nil {
		t.Fatal(err)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Scan inicial - img2 deve ser a mais recente
	_, _, err = w.ScanExisting()
	if err != nil {
		t.Fatal(err)
	}

	deleted := make(chan string, 1)
	w.OnImageDeleted = func(path string) {
		deleted <- path
	}

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Dar tempo para watcher inicializar
	time.Sleep(100 * time.Millisecond)

	// Deletar a imagem atual (img2)
	if err := os.Remove(img2); err != nil {
		t.Fatal(err)
	}

	// Aguardar notificação de deleção
	select {
	case path := <-deleted:
		if path != "img1.png" {
			t.Errorf("OnImageDeleted path = %q, want %q (próxima mais recente)", path, "img1.png")
		}

		// Verificar se a imagem atual foi atualizada
		current := w.CurrentImageRelative()
		if current != "img1.png" {
			t.Errorf("CurrentImageRelative() = %q, want %q", current, "img1.png")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout aguardando detecção de deleção de imagem")
	}
}

func TestImageWatcher_DetectImageRename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Criar diretório de destino para rename
	dstDir, err := os.MkdirTemp("", "sidelook_dst_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)

	// Criar duas imagens
	img1 := filepath.Join(tmpDir, "img1.png")
	if err := os.WriteFile(img1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	img2 := filepath.Join(tmpDir, "img2.jpg")
	if err := os.WriteFile(img2, []byte{0xFF, 0xD8, 0xFF}, 0644); err != nil {
		t.Fatal(err)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Scan inicial - img2 deve ser a mais recente
	_, _, err = w.ScanExisting()
	if err != nil {
		t.Fatal(err)
	}

	deleted := make(chan string, 1)
	w.OnImageDeleted = func(path string) {
		deleted <- path
	}

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Dar tempo para watcher inicializar
	time.Sleep(100 * time.Millisecond)

	// Mover a imagem atual (img2) para fora do diretório (simula RENAME)
	dstPath := filepath.Join(dstDir, "img2.jpg")
	if err := os.Rename(img2, dstPath); err != nil {
		t.Fatal(err)
	}

	// Aguardar notificação de deleção
	select {
	case path := <-deleted:
		if path != "img1.png" {
			t.Errorf("OnImageDeleted path = %q, want %q (próxima mais recente)", path, "img1.png")
		}

		// Verificar se a imagem atual foi atualizada
		current := w.CurrentImageRelative()
		if current != "img1.png" {
			t.Errorf("CurrentImageRelative() = %q, want %q", current, "img1.png")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout aguardando detecção de rename/movimentação de imagem")
	}
}

func TestNewWithSlideshowCount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	w, err := NewWithSlideshowCount(tmpDir, 5)
	if err != nil {
		t.Errorf("NewWithSlideshowCount() error = %v, want nil", err)
	}
	if w != nil {
		defer w.Stop()
		if w.maxRecent != 5 {
			t.Errorf("maxRecent = %d, want 5", w.maxRecent)
		}
	}
}

func TestRecentImages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Criar 6 imagens com diferentes timestamps
	images := []string{"img1.png", "img2.png", "img3.png", "img4.png", "img5.png", "img6.png"}
	for i, name := range images {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
			t.Fatal(err)
		}
		if i < len(images)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Criar watcher com limite de 4 imagens
	w, err := NewWithSlideshowCount(tmpDir, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Scan inicial
	count, _, err := w.ScanExisting()
	if err != nil {
		t.Fatal(err)
	}

	if count != 6 {
		t.Errorf("ScanExisting() count = %d, want 6", count)
	}

	// Verificar se mantém apenas 4 mais recentes
	recent := w.RecentImages()
	if len(recent) != 4 {
		t.Errorf("RecentImages() length = %d, want 4", len(recent))
	}

	// Verificar se está ordenado (mais recente primeiro)
	recentPaths := w.RecentImagesRelative()
	if len(recentPaths) != 4 {
		t.Fatal("RecentImagesRelative() length != 4")
	}

	// img6 deve ser a mais recente (primeiro)
	if recentPaths[0] != "img6.png" {
		t.Errorf("RecentImagesRelative()[0] = %q, want img6.png", recentPaths[0])
	}
	// img3 deve ser a mais antiga das 4 (último)
	if recentPaths[3] != "img3.png" {
		t.Errorf("RecentImagesRelative()[3] = %q, want img3.png", recentPaths[3])
	}
}

func TestSlideshowUpdatesOnNewImage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Criar 2 imagens iniciais
	img1 := filepath.Join(tmpDir, "img1.png")
	if err := os.WriteFile(img1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	img2 := filepath.Join(tmpDir, "img2.png")
	if err := os.WriteFile(img2, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}

	// Criar watcher com limite de 3
	w, err := NewWithSlideshowCount(tmpDir, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	_, _, err = w.ScanExisting()
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Dar tempo para watcher inicializar
	time.Sleep(100 * time.Millisecond)

	// Verificar estado inicial: 2 imagens
	before := w.RecentImagesRelative()
	if len(before) != 2 {
		t.Fatalf("Before: RecentImagesRelative() length = %d, want 2", len(before))
	}

	// Adicionar nova imagem
	img3 := filepath.Join(tmpDir, "img3.png")
	if err := os.WriteFile(img3, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}

	// Aguardar detecção
	time.Sleep(200 * time.Millisecond)

	// Verificar que lista foi atualizada para 3 imagens
	after := w.RecentImagesRelative()
	if len(after) != 3 {
		t.Errorf("After: RecentImagesRelative() length = %d, want 3", len(after))
	}

	// img3 deve ser a mais recente (primeiro)
	if after[0] != "img3.png" {
		t.Errorf("After: RecentImagesRelative()[0] = %q, want img3.png", after[0])
	}
}

func TestImageWatcher_DetectLastImageDeletion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sidelook_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Criar apenas uma imagem
	img1 := filepath.Join(tmpDir, "img1.png")
	if err := os.WriteFile(img1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatal(err)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Scan inicial
	_, _, err = w.ScanExisting()
	if err != nil {
		t.Fatal(err)
	}

	deleted := make(chan string, 1)
	w.OnImageDeleted = func(path string) {
		deleted <- path
	}

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Dar tempo para watcher inicializar
	time.Sleep(100 * time.Millisecond)

	// Deletar a única imagem
	if err := os.Remove(img1); err != nil {
		t.Fatal(err)
	}

	// Aguardar notificação de deleção
	select {
	case path := <-deleted:
		if path != "" {
			t.Errorf("OnImageDeleted path = %q, want empty string (sem mais imagens)", path)
		}

		// Verificar se não há mais imagem atual
		current := w.CurrentImageRelative()
		if current != "" {
			t.Errorf("CurrentImageRelative() = %q, want empty string", current)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout aguardando detecção de deleção de última imagem")
	}
}

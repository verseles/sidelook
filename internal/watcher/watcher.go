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
	recentImages []*ImageInfo // N imagens mais recentes ordenadas (mais recente primeiro)
	maxRecent    int          // Número máximo de imagens recentes a manter

	// OnNewImage é chamado quando uma nova imagem é detectada
	OnNewImage func(path string)

	// OnImageDeleted é chamado quando a imagem atual é deletada
	OnImageDeleted func(path string)

	done chan struct{}
}

// New cria um novo ImageWatcher
func New(dir string) (*ImageWatcher, error) {
	return NewWithSlideshowCount(dir, 0)
}

// NewWithSlideshowCount cria um novo ImageWatcher com suporte a slideshow
func NewWithSlideshowCount(dir string, slideshowCount int) (*ImageWatcher, error) {
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
		dir:          dir,
		watcher:      w,
		done:         make(chan struct{}),
		maxRecent:    slideshowCount,
		recentImages: make([]*ImageInfo, 0, slideshowCount),
	}, nil
}

// ScanExisting faz scan inicial e retorna a imagem mais recente
func (iw *ImageWatcher) ScanExisting() (count int, mostRecent *ImageInfo, err error) {
	entries, err := os.ReadDir(iw.dir)
	if err != nil {
		return 0, nil, err
	}

	var allImages []*ImageInfo

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

		allImages = append(allImages, &ImageInfo{
			Path:    path,
			ModTime: info.ModTime(),
		})
	}

	// Ordenar por ModTime (mais recente primeiro)
	for i := 0; i < len(allImages); i++ {
		for j := i + 1; j < len(allImages); j++ {
			if allImages[j].ModTime.After(allImages[i].ModTime) {
				allImages[i], allImages[j] = allImages[j], allImages[i]
			}
		}
	}

	// Pegar as N mais recentes
	iw.mu.Lock()
	if len(allImages) > 0 {
		iw.currentImage = allImages[0]

		if iw.maxRecent > 0 {
			maxIdx := iw.maxRecent
			if maxIdx > len(allImages) {
				maxIdx = len(allImages)
			}
			iw.recentImages = allImages[:maxIdx]
		}
	}
	iw.mu.Unlock()

	if len(allImages) > 0 {
		return count, allImages[0], nil
	}
	return count, nil, nil
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

// RecentImages retorna as N imagens mais recentes
func (iw *ImageWatcher) RecentImages() []*ImageInfo {
	iw.mu.RLock()
	defer iw.mu.RUnlock()

	// Retornar cópia para evitar problemas de concorrência
	result := make([]*ImageInfo, len(iw.recentImages))
	copy(result, iw.recentImages)
	return result
}

// RecentImagesRelative retorna os caminhos relativos das N imagens mais recentes
func (iw *ImageWatcher) RecentImagesRelative() []string {
	images := iw.RecentImages()
	paths := make([]string, len(images))

	for i, img := range images {
		rel, err := filepath.Rel(iw.dir, img.Path)
		if err != nil {
			paths[i] = filepath.Base(img.Path)
		} else {
			paths[i] = rel
		}
	}

	return paths
}

// findMostRecentImage procura a imagem mais recente no diretório
func (iw *ImageWatcher) findMostRecentImage() *ImageInfo {
	entries, err := os.ReadDir(iw.dir)
	if err != nil {
		return nil
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

	return latestInfo
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
	path := event.Name

	// Tratar deleção (Remove ou Rename para fora do diretório)
	if event.Op&fsnotify.Remove != 0 || event.Op&fsnotify.Rename != 0 {
		if !IsImageFile(path) {
			return
		}

		// Verificar se arquivo ainda existe (RENAME pode ser renomear dentro do mesmo diretório)
		if event.Op&fsnotify.Rename != 0 {
			if _, err := os.Stat(path); err == nil {
				// Arquivo ainda existe, não é uma deleção
				return
			}
		}

		// Verificar se a imagem deletada é a atual
		iw.mu.RLock()
		currentPath := ""
		if iw.currentImage != nil {
			currentPath = iw.currentImage.Path
		}
		iw.mu.RUnlock()

		if currentPath == path {
			// Encontrar próxima imagem mais recente
			nextImage := iw.findMostRecentImage()

			iw.mu.Lock()
			iw.currentImage = nextImage
			iw.mu.Unlock()

			// Notificar callback de deleção
			if iw.OnImageDeleted != nil {
				var relPath string
				if nextImage != nil {
					rel, err := filepath.Rel(iw.dir, nextImage.Path)
					if err != nil {
						relPath = filepath.Base(nextImage.Path)
					} else {
						relPath = rel
					}
				}
				iw.OnImageDeleted(relPath)
			}
		}
		return
	}

	// Tratar criação e modificação
	if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
		return
	}

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

	// Atualizar lista de imagens recentes se slideshow está ativado
	if iw.maxRecent > 0 {
		// Adicionar nova imagem no início
		iw.recentImages = append([]*ImageInfo{newImage}, iw.recentImages...)

		// Manter apenas maxRecent imagens
		if len(iw.recentImages) > iw.maxRecent {
			iw.recentImages = iw.recentImages[:iw.maxRecent]
		}
	}
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

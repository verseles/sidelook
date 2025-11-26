// internal/cli/cli.go
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/verseles/sidelook/internal/version"
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

	// SlideshowCount é o número de imagens no slideshow (0 = desabilitado)
	SlideshowCount int

	// SlideshowInterval é o intervalo em segundos entre transições (padrão: 3)
	SlideshowInterval int
}

// Parse faz o parse dos argumentos de linha de comando
func Parse(args []string) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("sidelook", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.IntVar(&cfg.Port, "p", 0, "Porta do servidor HTTP (padrão: 8080)")
	fs.IntVar(&cfg.Port, "port", 0, "Porta do servidor HTTP (padrão: 8080)")
	fs.IntVar(&cfg.SlideshowCount, "s", 0, "Número de imagens no slideshow (0 = desabilitado)")
	fs.IntVar(&cfg.SlideshowCount, "slideshow", 0, "Número de imagens no slideshow (0 = desabilitado)")
	fs.IntVar(&cfg.SlideshowInterval, "t", 3, "Intervalo em segundos entre imagens (padrão: 3)")
	fs.IntVar(&cfg.SlideshowInterval, "time", 3, "Intervalo em segundos entre imagens (padrão: 3)")
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

	// Validar slideshow
	if cfg.SlideshowCount < 0 {
		return nil, fmt.Errorf("número de imagens no slideshow inválido: %d. Use um número >= 0", cfg.SlideshowCount)
	}
	if cfg.SlideshowInterval < 1 {
		return nil, fmt.Errorf("intervalo de slideshow inválido: %d. Use um número >= 1", cfg.SlideshowInterval)
	}

	return cfg, nil
}

// Usage retorna o texto de ajuda
func Usage() string {
	return fmt.Sprintf(`sidelook %s - Visualizador de imagens em tempo real

Uso: sidelook [opções] [diretório]

Opções:
  -p, --port <número>       Porta do servidor HTTP (padrão: 8080)
  -s, --slideshow <número>  Número de imagens no slideshow (0 = desabilitado)
  -t, --time <segundos>     Intervalo entre imagens no slideshow (padrão: 3)
  -u, --update              Atualizar para a versão mais recente
  -v, --version             Exibir versão atual
  -h, --help                Exibir esta ajuda

Exemplos:
  sidelook                       # Monitora diretório atual (última imagem)
  sidelook ~/Downloads           # Monitora pasta Downloads
  sidelook -p 3000               # Usa porta 3000
  sidelook -s 4                  # Slideshow com 4 últimas imagens (3s cada)
  sidelook -s 4 -t 2             # Slideshow com 4 imagens (2s cada)
  sidelook --slideshow 10 --time 5  # Slideshow com 10 imagens (5s cada)
  sidelook --update              # Atualiza para versão mais recente

`, version.Version)
}

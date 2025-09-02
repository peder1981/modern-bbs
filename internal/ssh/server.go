package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"modern-bbs/internal/database"
	"modern-bbs/pkg/tui"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"
)

// Server representa o servidor SSH do BBS.
type Server struct {
	Addr   string
	config *ssh.ServerConfig
}

// NewServer cria e configura uma nova instância do servidor SSH.
func NewServer(addr string) (*Server, error) {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			user, passwordHash, err := database.GetUserByUsername(c.User())
			if err != nil {
				log.Printf("Erro ao buscar usuário '%s': %v", c.User(), err)
				return nil, fmt.Errorf("erro interno do servidor")
			}

			if user == nil || !database.CheckPasswordHash(string(pass), passwordHash) {
				log.Printf("Falha na autenticação para o usuário: %s", c.User())
				return nil, fmt.Errorf("usuário ou senha inválidos")
			}

			log.Printf("Usuário '%s' autenticado com sucesso.", c.User())
			return nil, nil // Autenticação bem-sucedida
		},
	}

	signer, err := getOrCreateHostKey("host_key")
	if err != nil {
		return nil, fmt.Errorf("falha ao obter ou criar a chave do host: %w", err)
	}
	config.AddHostKey(signer)

	return &Server{
		Addr:   addr,
		config: config,
	}, nil
}

// ListenAndServe inicia o listener e serve as conexões SSH.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("falha ao escutar em %s: %w", s.Addr, err)
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("Falha ao aceitar conexão: %v", err)
			continue
		}

		go s.handleConnection(nConn)
	}
}

func (s *Server) handleConnection(nConn net.Conn) {
	defer nConn.Close()
	sshConn, newChannels, reqs, err := ssh.NewServerConn(nConn, s.config)
	if err != nil {
		// Este erro ocorre se a autenticação falhar, o que já é logado no callback.
		// log.Printf("Falha no handshake SSH para %s: %v", nConn.RemoteAddr(), err)
		return
	}
	log.Printf("Login bem-sucedido para %s (%s)", sshConn.User(), sshConn.RemoteAddr())

	// Descarte de requisições globais que não nos interessam.
	go ssh.DiscardRequests(reqs)

	// Aceita e lida com novos canais (sessões).
	for newChannel := range newChannels {
		go s.handleChannel(sshConn, newChannel)
	}
}

func (s *Server) handleChannel(sshConn *ssh.ServerConn, newChannel ssh.NewChannel) {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "tipo de canal desconhecido")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Não foi possível aceitar o canal: %v", err)
		return
	}
	defer channel.Close()

	// Lida com requisições na sessão (pty-req, shell, etc.)
	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "pty-req":
				// O cliente está solicitando um PTY. Aceitamos.
				req.Reply(true, nil)
			case "shell":
				// O cliente está solicitando um shell. Aceitamos.
				req.Reply(true, nil)
			default:
				req.Reply(false, nil)
			}
		}
	}(requests)

	// Busca os dados completos do usuário para obter o papel (role).
	user, _, err := database.GetUserByUsername(sshConn.User())
	if err != nil || user == nil {
		log.Printf("Erro crítico: não foi possível obter dados do usuário '%s' após a autenticação: %v", sshConn.User(), err)
		return
	}

	// Inicia a aplicação TUI com Bubble Tea.
	m := tui.InitialModel(user.Username, user.Role)
	p := tea.NewProgram(m, tea.WithInput(channel), tea.WithOutput(channel))

	if _, err := p.Run(); err != nil {
		log.Printf("Erro ao executar o programa TUI para %s: %v", sshConn.User(), err)
	}

	log.Printf("Sessão TUI encerrada para %s", sshConn.User())
}

// getOrCreateHostKey carrega a chave privada do host de um arquivo ou cria uma nova.
func getOrCreateHostKey(path string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Nenhuma chave de host encontrada. Gerando uma nova...")
			return createHostKey(path)
		}
		return nil, fmt.Errorf("falha ao ler a chave do host: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer o parse da chave do host: %w", err)
	}

	log.Println("Chave de host carregada com sucesso.")
	return signer, nil
}

// createHostKey gera uma nova chave privada RSA e a salva no caminho especificado.
func createHostKey(path string) (ssh.Signer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar a chave privada: %w", err)
	}

	// Codifica a chave em formato PEM.
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyPEM := pem.EncodeToMemory(pemBlock)

	// Salva a chave no arquivo.
	if err := os.WriteFile(path, privateKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("falha ao salvar a chave do host: %w", err)
	}

	log.Printf("Nova chave de host salva em %s", path)

	// Retorna um signer da chave recém-criada.
	return ssh.NewSignerFromKey(privateKey)
}

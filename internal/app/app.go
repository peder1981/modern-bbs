package app

import (
	"log"
	"modern-bbs/internal/database"
	"modern-bbs/internal/ssh"
	"os"
)

// Run inicia a aplicação principal do BBS.
func Run() {
	// Configuração da aplicação a partir de variáveis de ambiente ou valores padrão.
	dbPath := getEnv("BBS_DB_PATH", "bbs.db")
	port := getEnv("BBS_PORT", "7778")
	addr := ":" + port

	// Inicializa o banco de dados.
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Erro ao inicializar o banco de dados em '%s': %v", dbPath, err)
	}

	// Cria e inicia o servidor SSH.
	server, err := ssh.NewServer(addr)
	if err != nil {
		log.Fatalf("Erro ao criar o servidor SSH: %v", err)
	}

	log.Printf("Servidor BBS escutando em %s...", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

// getEnv busca uma variável de ambiente ou retorna um valor padrão.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"modern-bbs/internal/database"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	dbPath := getEnv("BBS_DB_PATH", "bbs.db")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Erro ao inicializar o banco de dados em '%s': %v", dbPath, err)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "adduser":
		handleAddUser()
	case "addforum":
		handleAddForum()
	case "setrole":
		handleSetRole()
	case "deleteuser":
		handleDeleteUser()
	case "resetpassword":
		handleResetPassword()
	case "editforum":
		handleEditForum()
	case "deleteforum":
		handleDeleteForum()
	case "deletetopic":
		handleDeleteTopic()
	case "deletepost":
		handleDeletePost()
	default:
		fmt.Printf("Comando desconhecido: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Uso: bbs-admin <comando> [argumentos]")
	fmt.Println("Comandos:")
	fmt.Println("  adduser    - Adiciona um novo usuário")
	fmt.Println("  addforum   - Adiciona um novo fórum")
	fmt.Println("  setrole    - Define o papel de um usuário (user, moderator, admin)")
	fmt.Println("  deleteuser - Deleta um usuário")
	fmt.Println("  resetpassword - Reseta a senha de um usuário")
	fmt.Println("  editforum     - Edita um fórum existente")
	fmt.Println("  deleteforum   - Deleta um fórum")
	fmt.Println("  deletetopic   - Deleta um tópico")
	fmt.Println("  deletepost    - Deleta um post")
}

func handleAddUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o nome do usuário: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Digite a senha: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Falha ao ler a senha: %v", err)
	}
	password := string(bytePassword)
	fmt.Println()

	_, err = database.CreateUser(username, password)
	if err != nil {
		log.Fatalf("Erro ao criar usuário: %v", err)
	}

	fmt.Printf("Usuário '%s' criado com sucesso!\n", username)
}

func handleAddForum() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o nome do fórum: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Digite a descrição do fórum (opcional): ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	forum, err := database.CreateForum(name, description)
	if err != nil {
		log.Fatalf("Erro ao criar fórum: %v", err)
	}

	fmt.Printf("Fórum '%s' (ID: %d) criado com sucesso!\n", forum.Name, forum.ID)
}

func handleSetRole() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o nome do usuário: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Digite o novo papel (user, moderator, admin): ")
	role, _ := reader.ReadString('\n')
	role = strings.TrimSpace(role)

	if err := database.SetUserRole(username, role); err != nil {
		log.Fatalf("Erro ao definir o papel do usuário: %v", err)
	}

	fmt.Printf("O papel do usuário '%s' foi definido como '%s' com sucesso!\n", username, role)
}

func handleDeleteUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o nome do usuário a ser deletado: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if err := database.DeleteUser(username); err != nil {
		log.Fatalf("Erro ao deletar usuário: %v", err)
	}

	fmt.Printf("Usuário '%s' deletado com sucesso!\n", username)
}

func handleResetPassword() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o nome do usuário: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Digite a nova senha: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Falha ao ler a senha: %v", err)
	}
	password := string(bytePassword)
	fmt.Println()

	if err := database.AdminResetPassword(username, password); err != nil {
		log.Fatalf("Erro ao resetar a senha: %v", err)
	}

	fmt.Printf("Senha do usuário '%s' resetada com sucesso!\n", username)
}

func handleEditForum() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o ID do fórum a ser editado: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Fatalf("ID do fórum inválido: %v", err)
	}

	fmt.Print("Digite o novo nome do fórum: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Digite a nova descrição do fórum (opcional): ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	if err := database.UpdateForum(id, name, description); err != nil {
		log.Fatalf("Erro ao editar fórum: %v", err)
	}

	fmt.Printf("Fórum ID %d atualizado com sucesso!\n", id)
}

func handleDeleteForum() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o ID do fórum a ser deletado: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Fatalf("ID do fórum inválido: %v", err)
	}

	if err := database.DeleteForum(id); err != nil {
		log.Fatalf("Erro ao deletar fórum: %v", err)
	}

	fmt.Printf("Fórum ID %d deletado com sucesso!\n", id)
}

func handleDeleteTopic() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o ID do tópico a ser deletado: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatalf("ID do tópico inválido: %v", err)
	}

	if err := database.DeleteTopic(id); err != nil {
		log.Fatalf("Erro ao deletar tópico: %v", err)
	}

	fmt.Printf("Tópico ID %d deletado com sucesso!\n", id)
}

func handleDeletePost() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Digite o ID do post a ser deletado: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatalf("ID do post inválido: %v", err)
	}

	if err := database.DeletePost(id); err != nil {
		log.Fatalf("Erro ao deletar post: %v", err)
	}

	fmt.Printf("Post ID %d deletado com sucesso!\n", id)
}

# Modern BBS

Modern BBS é um sistema de Bulletin Board System (BBS) baseado em texto, acessível via SSH. Ele oferece uma experiência de fórum clássica com uma interface de linha de comando moderna e interativa.

## Arquitetura

O projeto é escrito em Go e utiliza uma arquitetura modular para separar as responsabilidades:

- **`cmd/`**: Contém os pontos de entrada da aplicação.
  - **`bbs-server`**: O servidor principal do BBS que escuta por conexões SSH.
  - **`bbs-admin`**: Uma ferramenta de linha de comando para tarefas administrativas, como criar usuários e fóruns.
- **`internal/`**: Contém a lógica de negócio principal da aplicação.
  - **`app`**: Orquestra a inicialização do servidor e do banco de dados.
  - **`ssh`**: Gerencia as conexões SSH, autenticação de usuários e o ciclo de vida das sessões.
  - **`database`**: Lida com toda a interação com o banco de dados SQLite, incluindo a definição do esquema e as operações CRUD para usuários, fóruns, tópicos e posts.
- **`pkg/`**: Contém pacotes reutilizáveis.
  - **`tui`**: Implementa a Interface de Usuário de Texto (TUI) usando a biblioteca Bubble Tea. É responsável por renderizar todas as telas com as quais o usuário interage.

## Como Colocar em Operação

### Pré-requisitos

- Go (versão 1.18 ou superior)

### 1. Instalar Dependências

Navegue até o diretório raiz do projeto e execute o comando a seguir para baixar as dependências:

```bash
go mod tidy
```

### 2. Compilar os Binários

Compile o servidor e a ferramenta de administração:

```bash
go build -o bbs-server ./cmd/bbs-server
go build -o bbs-admin ./cmd/bbs-admin
```

Isso criará dois executáveis: `bbs-server` e `bbs-admin`.

### 3. Executar o Servidor

Para iniciar o servidor BBS, execute o seguinte comando:

```bash
./bbs-server
```

Por padrão, o servidor irá:
- Criar (se não existir) um banco de dados chamado `bbs.db`.
- Criar (se não existir) uma chave de host SSH chamada `host_key`.
- Escutar por conexões na porta `7778`.

Você pode customizar o comportamento usando variáveis de ambiente:
- `BBS_DB_PATH`: Caminho para o arquivo do banco de dados (ex: `BBS_DB_PATH=/var/data/prod.db`).
- `BBS_PORT`: Porta para o servidor SSH (ex: `BBS_PORT=2222`).

### 4. Acessar o BBS

Com o servidor em execução, você pode se conectar a ele usando um cliente SSH:

```bash
ssh <username>@localhost -p 7778
```

Na primeira execução, alguns usuários padrão são criados:
- **Usuário**: `admin`, **Senha**: `adminpass`
- **Usuário**: `mod`, **Senha**: `modpass`
- **Usuário**: `user`, **Senha**: `userpass`

### 5. Usar a Ferramenta de Administração

A ferramenta `bbs-admin` permite gerenciar o BBS via linha de comando.

**Uso:**
```bash
./bbs-admin <comando>
```

**Comandos disponíveis:**
- `adduser`: Adiciona um novo usuário de forma interativa.
- `addforum`: Adiciona um novo fórum.
- `setrole`: Define o papel de um usuário (`user`, `moderator`, `admin`).

## Interação com a TUI

A interface do BBS é controlada pelos seguintes atalhos:
- **Navegação**: `↑`/`k` (para cima) e `↓`/`j` (para baixo).
- **Seleção**: `enter`.
- **Voltar**: `esc`.
- **Criar Novo (Tópico/Post)**: `n`.
- **Sair**: `q` ou `ctrl+c`.

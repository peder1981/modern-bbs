# Changelog

## [Unreleased]

### Added
- Implementado o menu de administração (`pkg/tui/admin.go`) com uma lista de opções ("Gerenciamento de Usuários", "Gerenciamento de Fóruns") usando o componente `list.Model` da biblioteca `bubbles`.
- Adicionado um menu de ações na tela de gerenciamento de usuários (`pkg/tui/user_management.go`) com as opções: "Alterar Papel", "Deletar Usuário" e "Resetar Senha".
- Implementada a lógica para deletar usuários e resetar senhas através da TUI, conectando com as funções `database.DeleteUser` e `database.AdminResetPassword`.
- Adicionada a funcionalidade completa de gerenciamento de fóruns na TUI, incluindo criação, edição e deleção com confirmação.
- Implementados poderes de moderação na TUI, permitindo que administradores e moderadores deletem tópicos e posts com confirmação.

### Changed
- A navegação de retorno (`navigateBackMsg`) agora utiliza o histórico de `breadcrumbs` para voltar à tela anterior, em vez de sempre retornar ao menu principal.
- A tela de gerenciamento de usuários foi refatorada para um fluxo de múltiplos passos, melhorando a usabilidade e escalabilidade.

### Fixed
- Corrigido um erro de compilação causado pela re-declaração da `struct usersLoadedMsg` em `pkg/tui/user_management.go`. A declaração duplicada foi removida, centralizando a definição em `pkg/tui/model.go`.
- Corrigidos múltiplos erros de compilação em `pkg/tui/topics.go` e `pkg/tui/posts.go` relacionados a declarações de `structs` duplicadas e lógica de recarregamento de dados incorreta.
- Refatorado o carregamento de dados nos modelos de tópicos e posts para ser assíncrono, melhorando a responsividade da interface.
- O modelo de visualização de posts (`postsModel`) foi refatorado de um `viewport` para uma lista com cursor, permitindo a seleção e deleção de posts individuais.

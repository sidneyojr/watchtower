
# Watchtower — Fork, Setup no Debian e Plano de Ação

> Documento gerado em 2026-01-23 02:58:05 (horário local)

## 1) Contexto
- Projeto-alvo: **watchtower** — ferramenta para automatizar atualização de imagens base de containers Docker.
- Repositório de referência (upstream original): https://github.com/containrrr/watchtower  
  \- Observação: o projeto **original foi arquivado em 2025-12-17** e está *read-only*. Ver README do upstream.  
- Objetivo do fork: estabilizar base, padronizar fluxo de branches, atualizar linguagem/dependências e **só depois** propor melhorias/novas funcionalidades.

## 2) Histórico do que foi feito (passo a passo)

### 2.1 Fork do repositório
- Foi realizado um **fork** do repositório *upstream* apenas a partir da `main`.

### 2.2 Renomeação do branch padrão
- No GitHub, o branch padrão foi **renomeado de `main` para `master`**.  
  Motivo: alinhar com preferência do mantenedor do fork e facilitar estratégias que já utilizam `master` como linha estável.

### 2.3 Clone local
```bash
# Clonar o fork do Sidney
git clone https://github.com/sidneyojr/watchtower.git
cd watchtower

# Conferir branches remotos
git branch -a
```

### 2.4 Criação do branch de integração
- Criado o branch **`develop`** a partir de `master`:
```bash
git checkout master
git pull origin master
git checkout -b develop
git push -u origin develop
```

### 2.5 Ambiente: Debian 13
- Sistema: **Debian 13** (máquina de desenvolvimento).

### 2.6 Dependências instaladas (via `apt`)
Para **compilar** a partir do código (Go):
```bash
sudo apt update
sudo apt install -y   git   golang-go   build-essential   pkg-config   ca-certificates
```

> Observação: O projeto é em **Go** e utiliza **Go modules** (`go.mod`). Em versões recentes havia referência ao stdlib Go 1.20.x; se o `golang-go` local estiver muito antigo, considerar instalar uma versão mais nova do Go (1.21+/1.22+) antes de compilar.

(\*Opcional) Para **rodar via Docker** (sem compilar localmente):
```bash
sudo apt update
sudo apt install -y   docker.io   docker-buildx-plugin   docker-compose-plugin
sudo usermod -aG docker "$USER"  # logout/login após isso
```

## 3) Estado do upstream e alternativas mantidas
- Upstream original (`containrrr/watchtower`) foi **arquivado**; segue útil como referência, mas não recebe atualizações.
- Comunidade tem usado forks mantidos, como:
  - `nicholas-fedor/watchtower` — fork ativo e com documentação própria.
  - `beatkind/watchtower` — outro fork com atividade.

> Decisão acordada: **avaliar merge/rebase a partir do fork `nicholas-fedor`** antes de implantar melhorias/funcionalidades no nosso fork.

## 4) Stack técnica e execução
- **Linguagem:** Go
- **Gerenciamento de dependências:** Go modules (`go.mod`, `go.sum`)
- **Distribuição/uso comum:** imagem Docker (montando `/var/run/docker.sock`). Exemplo de execução:
```bash
docker run -d   --name watchtower   -v /var/run/docker.sock:/var/run/docker.sock   containrrr/watchtower
```

## 5) Estratégia de branches (acordada)
- `master` → linha **estável** (liberações)
- `develop` → **integração** contínua
- `feature/*` → novas funcionalidades
- `refactor/*` → refatorações e melhorias internas
- `chore/*` → tooling, lint, deps, CI

## 6) Plano de ação proposto (antes das melhorias)

### 6.1 Auditoria técnica do fork atual
- Revisar estrutura do projeto, `go.mod`, versões de libs (Docker SDK, Cobra, Prometheus client, etc.).
- Garantir build local reproducível e execução do binário.
- Adicionar `Makefile` com alvos: `make build`, `make test`, `make lint`, `make release`.
- Integrar **golangci-lint** e **pre-commit** (opcional) para padronização.

### 6.2 Sincronização com fork mantido (Nicholas Fedor)
- Criar branch `sync/nicholas-fedor-<data>`.
- Avaliar abordagem:
  1) **Merge** do repositório `nicholas-fedor/watchtower` no nosso `develop`, resolvendo conflitos; ou
  2) **Rebase**/cherry-picks seletivos de commits relevantes (quando o delta for muito grande, começar por módulos/dirs específicos: `cmd/`, `internal/`, `pkg/`).
- Rodar testes e validar execução.
- Atualizar docs/README se houver mudanças de flags/compatibilidade.

### 6.3 Atualização de linguagem e dependências
- Definir **versão-alvo do Go** (sugestão inicial: 1.21+ ou 1.22+).
- Atualizar dependências no `go.mod` e aplicar *breaking changes* (se houver APIs alteradas do Docker SDK, etc.).
- Revisar *Dockerfile* (se presente) para imagem base alinhada e tamanho reduzido (multi-stage build).

### 6.4 CI/CD
- Configurar **GitHub Actions**:
  - *Jobs*: build + testes (Linux/amd64 e arm64), lint, *matrix* para versões de Go.
  - Pipeline de **release**: geração de binários (amd64/arm64) e imagem Docker `ghcr.io/sidneyojr/watchtower`.
- Assinar imagens (cosign) e publicar *checksums*.

### 6.5 Qualidade, segurança e manutenção
- Ativar varreduras básicas (dependabot ou Renovate para Go modules).
- Verificar *secrets* acidentalmente versionados.
- Adicionar *CODEOWNERS*, *Contributing*, *PR template* e *Issue templates*.

### 6.6 Roadmap de funcionalidades (pós-sincronização)
- Só após a sincronização com o fork mantido e atualização de stack:
  - Melhorias de UX/flags (ex.: modos *monitor-only*, *schedule* mais flexível via cron e TZ, exclusões por regex/labels).
  - Integrações de **notificações** (via Shoutrrr) e métricas.
  - Documentação completa (pt-BR/EN) e exemplos de *docker-compose*.

## 7) Checklist executável
- [x] Fork realizado
- [x] Renomear `main` → `master` (GitHub UI)
- [x] `git clone` do fork localmente
- [x] Criar `develop` a partir de `master`
- [x] Instalar dependências do Debian (Go + toolchain)
- [x] Validar `go build` local
- [x] Instalar toolchain moderna sem sudo (Go 1.26.6 em `~/.local/go`, golangci-lint, mockery, goreleaser v2 em `~/go/bin`)
- [x] **Sincronizar com `nicholas-fedor/watchtower`** (merge `--allow-unrelated-histories -X theirs upstream/main`, commit `b2884f4b`)
- [x] **Rebrand do módulo** para `github.com/sidneyojr/watchtower` (commit `f113266b`; `tools/tplprev` vira módulo separado `github.com/sidneyojr/tplprev`)
- [x] **CI/CD GHCR-only**: workflows, goreleaser (`stable.yaml`/`nightly.yaml`), Dockerfiles, mkdocs e issue templates apontam para `ghcr.io/sidneyojr/watchtower`; Docker Hub removido; `.circleci` removido (commits `91aa1a52`)
- [x] **Hook Conventional Commits** versionado como `.githooks/commit-msg` + `git config core.hooksPath .githooks` (commit `cbe1f99d`)
- [x] **Docs rebranded** para sidneyojr (README, docs/, exemplos, SECURITY, swagger, fixtures) (commit `4e5c3664`)
- [ ] Merge `develop` → `master` e push dos dois branches (Fase 5)
- [ ] Validar `make build` / `make lint` / `make vet` / `make test` (Fase 5)
- [ ] Configurar proteção de branch em `master` (opcional, pós-push)
- [ ] Testes, validação e release inicial
- [ ] Só então: melhorias/novas features

## 8) Apêndice — Comandos úteis
```bash
# Atualizar pacotes
sudo apt update && sudo apt upgrade -y

# Compilar o projeto
cd /caminho/para/watchtower
go mod download
go build ./...

# Gerar binário
go build -o bin/watchtower .

# Executar container (uso clássico)
docker run -d   --name watchtower   -v /var/run/docker.sock:/var/run/docker.sock   containrrr/watchtower
```

---

### Referências
- Repositório original (arquivado) e README: https://github.com/containrrr/watchtower  
- README com *Quick Start* (Docker run): https://github.com/containrrr/watchtower/blob/main/README.md  
- Discussão/comunidade apontando forks mantidos (ex.: `nicholas-fedor`): https://community.getchannels.com/t/psa-recommended-new-fork-for-watchtower/45270


# project-example — пример облачной разработки с codexctl

Этот репозиторий демонстрирует полный цикл облачной разработки
с помощью `codexctl`, Kubernetes и GitHub Actions:

- инфраструктура: PostgreSQL, Redis, Jaeger, ingress‑контроллер, Codex‑pod;
- сервисы:
  - `django_backend` — Django‑админка и миграции схемы БД;
  - `chat_backend` — Go‑сервис с HTTP API простого чата;
  - `web_frontend` — фронтенд на Vue3 + Pinia;
- автоматизация:
  - деплой стейджинга по push в `main`;
  - планирование задач по метке `[ai-plan]`;
  - разработка с агентом по метке `[ai-dev]`;
  - восстановление стейджинга агентом по метке `[ai-repair]` (если что‑то сломалось);
  - review/fix по ревью PR.

## MCP‑серверы (yaml-mcp-server)

В `services.yaml` описаны два MCP‑сервера, которые подключаются к Codex:

- `github_secrets_postgres_k8s_mcp` — approval‑gateway для операций с GitHub secrets и созданием БД PostgreSQL в Kubernetes.
- `github_review_mcp` — детерминированная работа с review‑комментариями и вопросами в PR (list/reply/resolve).

Оба сервера используют единый образ `yaml-mcp-server`, но разные встроенные конфиги:

- `configs/github_secrets_postgres_k8s.yaml`
- `configs/github_review.yaml`

Вся конфигурация MCP и tool‑описания находятся в `services.yaml` в секции `codex.mcp.servers`.

### Переменные и секреты для GitHub

Значения задаются в GitHub (Secrets/Variables) и подставляются в манифесты через `services.yaml`.

**mcp-secrets-postgres-k8s**
- Secrets: `YAML_MCP_SECRETS_GH_PAT`
- Variables (опционально): `YAML_MCP_SECRETS_GITHUB_REPO`, `YAML_MCP_SECRETS_APPROVER_URL`, `YAML_MCP_SECRETS_LANG`, `YAML_MCP_SECRETS_LOG_LEVEL`, `YAML_MCP_SECRETS_POSTGRES_POD_SELECTOR`

**mcp-github-review**
- Secrets: `YAML_MCP_REVIEW_GH_PAT`
- Variables: `YAML_MCP_REVIEW_GH_USERNAME` (обязательно)
- Variables (опционально): `YAML_MCP_REVIEW_GITHUB_REPO`, `YAML_MCP_REVIEW_LANG`, `YAML_MCP_REVIEW_LOG_LEVEL`

## 1. Подготовка кластера (Ubuntu 24.04)

Примечание: в этом проекте self‑hosted runner работает только внутри Kubernetes.
Команды ниже нужны для подготовки кластера и базового runner‑образа/runner‑пода,
а не для запуска runner на хосте.

Если хотите более «взрослый» кластер на K3s с Calico/Longhorn, бэкапами и
registry per namespace, используйте готовые скрипты в `bootstrap/`.
Для single‑node сервера они работают в режиме NodePort/Ingress и
используют один внешний IP (MetalLB можно отключить).
Они автоматизируют:
- установку K3s (без flannel, с отключённым встроенным netpol и local‑storage);
- установку Calico и разметку IP‑пулов по namespace;
- установку Longhorn + политики бэкапов через recurring jobs;
- установку cert‑manager и (опционально) MetalLB L2;
- деплой registry в каждом namespace;
- настройку `/etc/hosts` на хосте, чтобы `registry.<ns>.svc.cluster.local` резолвился контейнерным runtime.

Как пользоваться:
1) Откройте `bootstrap/config.env` и заполните:
   - `PROJECTS` — список проектов через пробел (до 4);
   - `AI_DEV_SLOTS` — количество ai‑слотов;
   - `ENABLE_METALLB` — `false` для single‑node с одним IP;
   - `METALLB_L2_ADDRESSES` — диапазоны L2‑адресов (только если MetalLB включён);
   - `LH_S3_*` — параметры S3 для бэкапов Longhorn;
   - `VELERO_*` — параметры S3 для Velero (backup k8s‑объектов);
   - `DEFAULT_SC` — StorageClass по умолчанию (`prod|staging|dev`);
   - `ENABLE_HOST_FIREWALL` и `*_ALLOW_CIDR` — правила хост‑фаервола.
2) Запустите:

```bash
sudo bash bootstrap/bootstrap.sh
```

Если позже измените `PROJECTS` или `AI_DEV_SLOTS`, выполните:

```bash
sudo bash bootstrap/scripts/10_generate_k3s_registries.sh
sudo systemctl restart k3s
sudo bash bootstrap/scripts/61_update_registry_hosts.sh
```

### 1.1. Базовые пакеты

```bash
sudo apt-get update
sudo apt-get install -y git curl jq build-essential ca-certificates software-properties-common
```

### 1.2. Установка microk8s

```bash
sudo apt install snapd
sudo snap install microk8s --classic
sudo usermod -aG microk8s "$USER"
newgrp microk8s
```

Включаем необходимые аддоны:

```bash
microk8s enable dns storage ingress registry rbac
```

Проверяем, что кластер готов:

```bash
microk8s status --wait-ready
microk8s kubectl get nodes
```

Сохраняем kubeconfig для админского доступа (опционально):

```bash
mkdir -p ~/.kube
microk8s config | sudo tee ~/.kube/microk8s.config >/dev/null
sudo chown -R "$USER":"$USER" ~/.kube
ln -sfn ~/.kube/microk8s.config ~/.kube/config
```

### 1.3. Kaniko и registry в кластере

Сборка образов выполняется Kaniko в CI, локальный Docker не нужен.
В `services.yaml` развёрнут in‑cluster registry (Deployment + Service + PVC)
в каждом namespace (ai‑staging и ai‑слоты), по умолчанию доступный как:

```
registry.<namespace>.svc.cluster.local:5000
```

Что требуется на runner:

- бинарник `kaniko` (по умолчанию `/kaniko/executor`, либо задайте `CODEXCTL_KANIKO_EXECUTOR`);
- переменная `CODEXCTL_REGISTRY_HOST` (опционально, для переопределения; по умолчанию `registry.<namespace>.svc.cluster.local:5000`).

Если registry без TLS, задайте в окружении CI:

```
CODEXCTL_KANIKO_INSECURE=true
CODEXCTL_KANIKO_SKIP_TLS_VERIFY=true
CODEXCTL_KANIKO_SKIP_TLS_VERIFY_PULL=true
```


### 1.4. Установка Golang 1.25+

Через snap:

```bash
sudo snap install go --classic
```

Если `codexctl` ставится через `go install` внутри runner‑pod, убедитесь, что
`go`/`gofmt` доступны в `PATH`:

Вариант A (проще и надежнее):

```bash
sudo ln -sf /snap/bin/go /usr/local/bin/go
sudo ln -sf /snap/bin/gofmt /usr/local/bin/gofmt
```

Либо вручную:

```bash
cd /tmp
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
```

Добавить Go в PATH

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 1.5. Установка kubectl

Для работы `codexctl` с Kubernetes нужен бинарник `kubectl`. Установим его
в `/usr/local/bin/kubectl` (эта директория добавлена в `PATH` в воркфлоу).

```bash
KUBECTL_VERSION=v1.34.1   # или нужная вам версия
curl -fsSL -o kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl

kubectl version --client --output=yaml || true
```

Также для `codexctl` требуются утилиты:
- `bash` (обычно уже установлен);
- `git`;
- `gh` (GitHub CLI);
- `kubectl`;
- `kaniko` (executor для сборки образов, по умолчанию `/kaniko/executor`
  или путь из `CODEXCTL_KANIKO_EXECUTOR`).

Проверка доступности утилит:

```bash
KANIKO_EXECUTOR="${CODEXCTL_KANIKO_EXECUTOR:-/kaniko/executor}"
for t in kubectl bash git gh; do
  if command -v "$t" >/dev/null 2>&1; then
    echo "OK  $t -> $(command -v "$t")"
  else
    echo "MISS $t"
  fi
done
if [ -x "$KANIKO_EXECUTOR" ]; then
  echo "OK  kaniko -> $KANIKO_EXECUTOR"
else
  echo "MISS kaniko ($KANIKO_EXECUTOR)"
fi
```

Установка `git` и `gh` (Ubuntu 24):

```bash
sudo apt-get update
sudo apt-get install -y git gh
```

Установка последней версии `codexctl`:

```bash
go install github.com/codex-k8s/codexctl/cmd/codexctl@latest
```

Добавить Go bin в PATH (чтобы `codexctl` был доступен в сессии):

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

Если runner‑образ не содержит `codexctl`, его можно ставить в workflow через
`go install` или собрать кастомный runner‑образ с предустановленным `codexctl`.

## 2. Структура проекта

Основные директории:

- `services/django_backend` — Django‑проект с моделями `ChatUser` и `Message`, админкой и миграциями;
- `services/chat_backend` — Go‑сервис с REST API чата (`/api/*`);
- `services/web_frontend` — SPA на Vue3 + Pinia;
- `deploy/` — Kubernetes‑манифесты инфраструктуры:
  - `namespace.yaml`, `configmap.yaml`, `secret.yaml`;
  - `postgres.service.yaml`, `redis.service.yaml`;
  - `jaeger.yaml`, `dns.configmap.yaml`;
  - `cluster-issuer.yaml`, `ingress-nginx.controller.yaml`;
  - `codex/*` — Pod Codex, ingress для dev‑слотов и RBAC для service account `codex-sa`;
  - `runner/*` — ARC runner’ы, RBAC и Dockerfile runner‑образа;
- `services.yaml` — конфигурация `codexctl`;
- `.github/workflows/*.yml` — CI/CD и AI‑воркфлоу;
- `docs/*.md` — документация по архитектуре, моделям, деплою и т.д.

## 3. Запуск self‑hosted GitHub Runner в Kubernetes

Рекомендуемый путь — GitHub Actions Runner Controller (ARC) в кластере.
В этом проекте поддерживается только in‑cluster запуск: runner’ы работают
исключительно в pod’ах Kubernetes.

Для ARC нужен Helm. Если кластер разворачивали через `bootstrap/`, Helm уже установлен.
Иначе установите вручную:

```bash
curl -fsSL -o /tmp/get-helm-3 https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 /tmp/get-helm-3
sudo /tmp/get-helm-3
```

Образы runner’ов в этой конфигурации хранятся во внутреннем registry
в namespace `actions-runner-system`:
`registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest`.
Этот namespace и registry создаются bootstrap‑скриптами.

Перед установкой ARC соберите и запушьте образ runner’а в этот registry через Kaniko:

```bash
/kaniko/executor \
  --context "$(pwd)" \
  --dockerfile deploy/runner/Dockerfile \
  --destination registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest \
  --insecure --skip-tls-verify --skip-tls-verify-pull
```

Важно: это первичная сборка runner‑образа, поэтому `kaniko` должен быть
доступен **вне** runner‑пода (его ещё нет). Варианты:
- поставить kaniko executor на сервер и использовать команду выше;
- временно собрать образ через Docker на сервере;
- запустить одноразовый kaniko‑pod в кластере.

Пример одноразового kaniko‑пода (публичный репозиторий):

```bash
kubectl -n actions-runner-system run kaniko-build --rm -it --restart=Never \
  --image=gcr.io/kaniko-project/executor:debug \
  -- /kaniko/executor \
  --context=git://github.com/codex-k8s/project-example.git#refs/heads/main \
  --dockerfile=deploy/runner/Dockerfile \
  --destination=registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest \
  --insecure --skip-tls-verify --skip-tls-verify-pull
```

Для приватного репозитория (HTTPS + токен):

```bash
kubectl -n actions-runner-system create secret generic kaniko-git \
  --from-literal=GIT_TOKEN="YOUR_TOKEN"

kubectl -n actions-runner-system run kaniko-build --rm -it --restart=Never \
  --image=gcr.io/kaniko-project/executor:debug \
  --env=GIT_TOKEN=$(kubectl -n actions-runner-system get secret kaniko-git -o jsonpath='{.data.GIT_TOKEN}' | base64 -d) \
  -- /kaniko/executor \
  --context=git://github.com/codex-k8s/project-example.git#refs/heads/main \
  --dockerfile=deploy/runner/Dockerfile \
  --destination=registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest \
  --insecure --skip-tls-verify --skip-tls-verify-pull \
  --build-arg=GIT_TOKEN
```

Либо используйте заранее смонтированный том с кодом.
Схема установки (высокоуровнево):

- установить ARC в кластер (Helm‑чарт);
- создать RunnerScaleSet/RunnerDeployment в нужном namespace;
- добавить label’ы, которые используются в `runs-on` (например, `ai-staging` и `ai`);
- использовать runner‑образ с установленными `kubectl`, `gh`, `git`, `bash`, `kaniko`
  (и при необходимости `go`/`codexctl`, либо ставить `codexctl` в workflow);
- выдать serviceAccount права на namespace, с которыми работает `codexctl`.

В этом репозитории workflows используют:

```
runs-on: [self-hosted, ai]         # AI-dev слоты
runs-on: [self-hosted, ai-staging] # deploy/repair/cleanup ai-staging
```

Если хотите изолировать исполнение по namespace, создайте отдельные runner‑деплойменты
с разными label и обновите `runs-on` в соответствующих воркфлоу.

### 3.1. Пример ARC (две группы runner’ов)

Ниже — укороченный пример Helm values для двух RunnerScaleSet. Установите chart
два раза с разными values (пример для `ai` и `ai-staging`).

```yaml
# values-ai.yaml
githubConfigUrl: https://github.com/codex-k8s/project-example
githubConfigSecret: gha-runner-secret
runnerScaleSetName: ai
runnerLabels: ["ai"]
template:
  spec:
    serviceAccountName: gha-runner-ai
    containers:
      - name: runner
        image: registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest
```

```yaml
# values-ai-staging.yaml
githubConfigUrl: https://github.com/codex-k8s/project-example
githubConfigSecret: gha-runner-secret
runnerScaleSetName: ai-staging
runnerLabels: ["ai-staging"]
template:
  spec:
    serviceAccountName: gha-runner-ai-staging
    containers:
      - name: runner
        image: registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest
```

Установка:

```bash
helm upgrade --install gha-runner-ai \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  -f values-ai.yaml

helm upgrade --install gha-runner-ai-staging \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  -f values-ai-staging.yaml
```

Примечание: `serviceAccountName` должен иметь права на нужные namespace’ы.

### 3.2. Манифесты runner’ов, RBAC и связь с GitHub

Готовые манифесты лежат в `deploy/runner/`:

- `runner-scale-set-ai.yaml`, `runner-scale-set-ai-staging.yaml` — RunnerScaleSet для ARC;
- `github-token-secret.yaml` / `github-app-secret.yaml` — секреты для связи с GitHub;
- `rbac-ai-base.yaml` — ServiceAccount + ClusterRole для runner’ов `ai`;
- `rbac-ai-slots.yaml` — RoleBinding для конкретного слота (применяется `codexctl` после создания слота);
- `rbac-ai-staging.yaml` — права для runner’ов `ai-staging` в `project-example-ai-staging`;
- `Dockerfile` — образ runner’а с `kubectl`, `gh`, `git`, `kaniko`, `go`, `codexctl`.

Все манифесты рассчитаны на namespace `actions-runner-system`.

Связь с GitHub:

1. Вариант A (PAT, проще): создайте PAT с правами на репозиторий и self‑hosted runner’ы
   и примените `deploy/runner/github-token-secret.yaml` (заменив токен).
2. Вариант B (GitHub App, рекомендуется): создайте GitHub App с правами на Actions/Administration,
   установите её в репозиторий и примените `deploy/runner/github-app-secret.yaml`.

Перед применением замените `ORG/REPO` в `runner-scale-set-*.yaml`
на свои значения и, при необходимости, поправьте namespace’ы под ваш проект.

Минимальный чек‑лист:

- `githubConfigUrl` должен указывать на репозиторий (`https://github.com/codex-k8s/project-example`) или организацию;
- `githubConfigSecret` в RunnerScaleSet должен совпадать с именем секрета (`gha-runner-secret`);
- PAT/APP должен иметь права на управление self‑hosted runner’ами в целевом репозитории.

Установка манифестов (пример):

```bash
kubectl apply -f deploy/runner/namespace.yaml
# установить ARC (controller + CRD) в actions-runner-system
kubectl apply -f deploy/runner/github-token-secret.yaml
kubectl apply -f deploy/runner/rbac-ai-base.yaml
kubectl apply -f deploy/runner/rbac-ai-staging.yaml
kubectl apply -f deploy/runner/runner-scale-set-ai.yaml
kubectl apply -f deploy/runner/runner-scale-set-ai-staging.yaml
```

Выбор образа runner’а:

- соберите образ из `deploy/runner/Dockerfile` и запушьте в
  `registry.actions-runner-system.svc.cluster.local:5000/codex-runner:latest`;
- укажите этот образ в `spec.template.spec.containers[0].image` в `runner-scale-set-*.yaml`.

RBAC и удалённые namespace’ы:

- RoleBinding создаётся в namespace; если namespace не существует — apply упадёт;
- `rbac-ai-slots.yaml` применяется `codexctl` автоматически после создания слота,
  поэтому ничего заранее создавать не нужно.
  Это делается через `environments.ai.slotBootstrapInfra` в `services.yaml`.

По умолчанию в `services.yaml`:

- `registry: registry.<namespace>.svc.cluster.local:5000`;
- домены:
  - `baseDomain.dev` по умолчанию `dev.example-domain.ru`;
  - `baseDomain.ai-staging` по умолчанию `ai-staging.example-domain.ru`;
  - `baseDomain.ai` по умолчанию совпадает с `ai-staging`.

Эти домены можно переопределить через переменные окружения:

- `CODEXCTL_BASE_DOMAIN_DEV` — домен для dev‑окружения;
- `CODEXCTL_BASE_DOMAIN_AI_STAGING` — домен для ai‑staging;
- `CODEXCTL_BASE_DOMAIN_AI` — домен для AI‑слотов (если не задан, берётся `CODEXCTL_BASE_DOMAIN_AI_STAGING`).

Рекомендуется задать их как Repository Variables в GitHub и/или
как переменные среды при запуске `codexctl`.

Таймаут ожидания деплоя:
- `codex.timeouts.deployWait` в `services.yaml` управляет временем ожидания `kubectl wait` после `codexctl apply/ci ensure-ready`.
- Значение по умолчанию — `10m` (если не задано в `services.yaml` и не переопределено флагом `--wait-timeout`).

## 4. Переменные и секреты в GitHub

### 5.1. Repository Variables (`Settings → Secrets and variables → Actions → Variables`)

Рекомендуемые переменные:

- `CODEXCTL_CODE_ROOT_BASE` — базовый путь для исходников dev‑AI слотов и ai‑staging‑копии репозитория (пример: `/workspace/codex/envs`):
  - dev‑AI слоты: `${CODEXCTL_CODE_ROOT_BASE}/<slot>/src`;
  - ai-staging: `${CODEXCTL_CODE_ROOT_BASE}/ai-staging/src`;
  например `/workspace/codex/envs/`;
- `CODEXCTL_BASE_DOMAIN_DEV` — домен для dev‑окружения;
- `CODEXCTL_BASE_DOMAIN_AI_STAGING` — домен для ai‑staging;
- `CODEXCTL_BASE_DOMAIN_AI` — домен для AI‑слотов (если не задан, берётся `CODEXCTL_BASE_DOMAIN_AI_STAGING`);
- `CODEXCTL_WORKSPACE_MOUNT` — точка монтирования рабочей PVC (обычно `/workspace`);
- `CODEXCTL_WORKSPACE_PVC` — имя PVC для исходников (например, `project-example-workspace`);
- `CODEXCTL_DATA_PVC` — имя PVC для Postgres/Redis (например, `project-example-data`);
- `CODEXCTL_REGISTRY_PVC` — имя PVC для registry (например, `project-example-registry`);
- `CODEXCTL_REGISTRY_HOST` — адрес registry в кластере (обычно не задаётся; по умолчанию `registry.<namespace>.svc.cluster.local:5000`). Примеры: `registry.project-example-ai-staging.svc.cluster.local:5000`, `registry.project-example-dev-1.svc.cluster.local:5000`.
- `CODEXCTL_SYNC_IMAGE` — образ для синхронизации исходников (например, `busybox:1.37.0`);
- `CODEXCTL_STORAGE_CLASS_WORKSPACE` — StorageClass для workspace PVC;
- `CODEXCTL_STORAGE_CLASS_DATA` — StorageClass для data PVC;
- `CODEXCTL_STORAGE_CLASS_REGISTRY` — StorageClass для registry PVC;
- `CODEXCTL_KANIKO_EXECUTOR` — путь к kaniko executor (по умолчанию `/kaniko/executor`);
- `CODEXCTL_DEV_SLOTS_MAX` — максимальное количество dev‑AI слотов (например, `2`).
- `CODEXCTL_ALLOWED_USERS` — список GitHub‑логинов, которым разрешено запускать AI‑воркфлоу (например, `user1,user2`), добавляется как Repository Variable.
- `CODEXCTL_GH_USERNAME` — GitHub‑логин бота, добавляется как Repository Variable.
- `CODEXCTL_GH_EMAIL` — email бота для git‑коммитов (например, `codex-bot@example.com`), добавляется как Repository Variable.
- `CODEXCTL_VERSION` — версия `codexctl` для установки в воркфлоу (например, `v0.3.1`), если не задана — используется `latest`.
- `LETSENCRYPT_EMAIL` — email для регистрации ACME аккаунта в Let’s Encrypt (например, `admin@example-domain.ru`).

### 5.2. Repository Secrets

Минимальный набор:

- `CODEXCTL_GH_PAT` — GitHub Personal Access Token бота (с правами на создание и редактирование PR/Issue и комментариев к ним. С правами на создание веток и пуш коммитов в них.);
- `OPENAI_API_KEY` — ключ для OpenAI;
- `CONTEXT7_API_KEY` — ключ для Context7;

Все эти переменные используются в:

- `deploy/secret.yaml`;
- `services.yaml` (hook’и и apply);
- GitHub Actions (`ai_staging_deploy.yml`, `ai_*` воркфлоу).

## 5. Первый деплой стейджинга

После настройки runner и секретов:

1. Убедитесь, что изменения закоммичены и пушнуты в ветку `main`.
2. В GitHub во вкладке Actions появится workflow
   **“AI Staging deploy 🚀”** (`.github/workflows/ai_staging_deploy.yml`).
3. При следующем push в `main`:
   - Kaniko соберёт и отзеркалит образы в кластерный registry (`CODEXCTL_ENV=ai-staging`, `CODEXCTL_MIRROR_IMAGES=true`,
     `CODEXCTL_BUILD_IMAGES=true`, далее `codexctl ci images`);
   - исходники будут синхронизированы в `${CODEXCTL_CODE_ROOT_BASE}/ai-staging/src` внутри PVC и примонтированы в ai-staging‑подах;
   - `codexctl ci apply` применит инфраструктуру и сервисы (`CODEXCTL_ENV=ai-staging`, `CODEXCTL_PREFLIGHT=true`, `CODEXCTL_WAIT=true`);
   - в кластере появится неймспейс `project-example-ai-staging`.

Проверка:

```bash
microk8s kubectl get pods -n project-example-ai-staging
microk8s kubectl get ingress -n project-example-ai-staging
```

Если DNS и TLS настроены, фронтенд будет доступен по `https://ai-staging.example-domain.ru/` (по вашему домену).

Для локального теста можно:

```bash
microk8s kubectl port-forward -n project-example-ai-staging svc/web-frontend 8080:80
```

и открыть `http://localhost:8080`.

## 6. Флоу планирования задач с агентом

1. Создайте Issue в репозитории и опишите задачу/подпроект.
2. Повесьте на Issue метку `[ai-plan]`.
3. Запустится workflow `.github/workflows/ai_plan_issue.yml`:
   - создаст или переиспользует AI‑slot (namespace `project-example-dev-<slot>`);
   - развернёт инфраструктуру и сервисы через `codexctl ci ensure-ready`;
   - запустит планирующего агента `prompt run --kind plan_issue` (язык через `CODEXCTL_LANG=ru`).
4. Агент оставит в Issue комментарий с планом работ, предложенной архитектурой
   и структурой подзадач.

Чтобы пересобрать или дополнительно уточнить план:

- оставьте комментарий с текстом, содержащим `[ai-plan]`;
- workflow `ai_plan_review.yml` найдёт корневой планирующий Issue
  и запустит агент `plan_review` (короткий или полный режим в зависимости от пересоздания окружения).

## 7. Флоу разработки с агентом

1. Для конкретной задачи (Issue) повесьте метку `[ai-dev]`.
2. Workflow `ai_dev_issue.yml`:
   - выделит/найдёт слот через `codexctl ci ensure-slot`;
   - развернёт окружение в этом слоте (`codexctl ci ensure-ready`, `CODEXCTL_PREPARE_IMAGES=true`, `CODEXCTL_APPLY=true`);
   - создаст/переключится на ветку `codex/issue-<номер>`;
   - запустит dev‑агента `prompt run --kind dev_issue` (язык через `CODEXCTL_LANG=ru`);
   - после завершения работы агента закоммитит изменения и запушит ветку.
3. Если для ветки уже есть PR, workflow попытается найти его и
   оставит комментарий с ссылками на окружение.

Дальше вы можете:

- просмотреть дифф в PR;
- дополнительно поправить код вручную;
- дать агенту новые инструкции и перезапустить `[ai-dev]`.

## 8. Флоу восстановления ai‑staging

1. Создайте Issue с описанием проблемы ai‑staging.
2. Повесьте метку `[ai-repair]`.
3. Запустится workflow `ai_repair_issue.yml`:
   - выделит слот `ai-repair`;
   - синхронизирует исходники в `${CODEXCTL_CODE_ROOT_BASE}/ai-staging/src`;
   - поднимет Pod `codex` в namespace `project-example-ai-staging` (полный RBAC в namespace);
   - запустит агента `prompt run --kind ai-repair_issue` (язык через `CODEXCTL_LANG=ru`).
4. Для PR с правками ai-staging‑ремонта ревью запускается через `ai_repair_pr_review.yml` (использует outputs `codexctl_new_env` и `codexctl_env_ready` для выбора continuation/resume).

Очистка `ai-repair` удаляет только ресурсы Codex/RBAC в namespace и не трогает сам namespace.

## 9. Флоу review/fix для PR

Для уже открытого PR:

1. Попросите ревьюера сделать обычный code review.
2. Если ревьюер ставит состояние `changes requested`,
   сработает `ai_pr_review.yml`:
   - поднимет (или переиспользует) AI‑slot для этого PR;
   - запустит агента `prompt run --kind dev_review` (язык через `CODEXCTL_LANG=ru`);
   - агент применит изменения в слоте;
   - команда `codexctl pr review-apply` перенесёт изменения в PR‑ветку
     (commit + push) и добавит комментарий.

## 10. Что дальше

- Обзор архитектуры — `docs/architecture_project.md`.
- Описание моделей БД — `docs/models.md`.
- Правила миграций и фикстур — `docs/migrations_and_fixtures.md`.
- Описание Go‑сервисов — `docs/go_services.md`.
- Наблюдаемость и логи — `docs/observability.md`.
- Общие договорённости по библиотекам — `docs/libs.md`.

Этот репозиторий задуман как учебный пример:
вы можете форкнуть его, адаптировать `services.yaml`,
Kubernetes‑манифесты и документацию под свои сервисы и домены
и получить готовый skeleton для облачной разработки с Codex‑агентом.

## 11. Безопасность

### Базовая настройка брандмауэра

Рекомендуется **сначала** ограничить входящие порты на уровне хостинга
(Cloud Firewall / Security Group / Firewall в панели провайдера).
Оставьте открытыми на входящие соединения только:

- `22/tcp` — SSH (лучше ограничить по своему IP);
- `80/tcp` и `443/tcp` — входной трафик на ingress;
- дополнительные порты — **только если они реально нужны** (например, метрики или админ‑инструменты).

После этого включите и настройте брандмауэр (UFW) на VPS,
ограничив доступ к тем же портам.

На исходящие соединения обычно можно не ограничивать.

### Работа агента

По умолчанию агент запускается с повышенными привилегиями, если вы используете [config_default.toml](https://github.com/codex-k8s/codexctl/blob/29561461741b8bbad654e3bf34645619a3d6f4bb/internal/prompt/templates/config_default.toml)
из репозитория `codexctl`. Это позволяет агенту выполнять широкий спектр задач,
но может представлять риск безопасности.

Также агент в процессе работы имеет доступ к Kubernetes‑кластеру в соответствии
с политиками RBAC, настроенными в `deploy/codex/rbac.yaml`.

Внимательно следите за тем, какие задачи вы поручаете агенту,
особенно если репозиторий открыт для внешних контрибьюторов.

### Запуск workflow только для доверенных пользователей

В переменную `CODEXCTL_ALLOWED_USERS` рекомендуется добавить список доверенных
GitHub‑логинов, которым разрешено запускать AI‑воркфлоу. Это поможет предотвратить
неавторизованный запуск агентов и потенциальные риски безопасности в этом процессе.

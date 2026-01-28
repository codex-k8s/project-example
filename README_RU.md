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
  - review/fix по ревью PR.

## 1. Подготовка VPS (Ubuntu 24.04)

### 1.1. Базовые пакеты

```bash
sudo apt-get update
sudo apt-get install -y git curl jq build-essential ca-certificates software-properties-common
```

### 1.2. Создание пользователя runner (если есть только root)

Если на VPS сейчас есть только пользователь `root`, создайте отдельного
пользователя `runner`, под которым будут работать GitHub Runner и все dev‑флоу:

```bash
adduser runner
usermod -aG sudo runner
su runner
```

Далее рекомендуется подключаться к серверу по SSH уже под пользователем `runner`
и выполнять большинство команд от его имени (через `sudo`, где требуется).

### 1.3. Установка microk8s

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

Сохраняем kubeconfig для стейджинга (под runner’ом) и сделаем его по умолчанию:

```bash
mkdir -p /home/runner/.kube
microk8s config | sudo tee /home/runner/.kube/microk8s.config >/dev/null
sudo chown -R runner:runner /home/runner/.kube
ln -sfn /home/runner/.kube/microk8s.config /home/runner/.kube/config
```

### 1.4. Установка Docker и insecure‑registry

```bash
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo usermod -aG docker "$USER"
sudo usermod -aG docker runner
```

Добавляем `localhost:32000` (встроенный registry microk8s) в insecure‑registries:

```bash
sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "insecure-registries": [
    "localhost:32000"
  ]
}
EOF

sudo systemctl restart docker
```

После установки Docker **обязательно авторизуйтесь на Docker Hub** под пользователем,
от имени которого запускается runner (обычно `runner`), чтобы избежать лимитов
на анонимные `pull`:

```bash
sudo -iu runner
docker login
```

Логин/пароль берутся от вашего Docker Hub аккаунта. Без этого при работе
`codexctl images mirror`/`build` можно упереться в сообщение вида:

> You have reached your unauthenticated pull rate limit.


### 1.5. Установка Golang 1.25+

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

### 1.6. Установка kubectl

Для работы `codexctl` с Kubernetes нужен бинарник `kubectl`. Установим его
в `/usr/local/bin/kubectl` (эта директория добавлена в `PATH` в воркфлоу).

```bash
KUBECTL_VERSION=v1.34.1   # или нужная вам версия
curl -fsSL -o kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl

kubectl version --client --output=yaml || true
```

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
  - `codex/*` — Pod Codex и ingress для dev‑слотов;
- `services.yaml` — конфигурация `codexctl`;
- `.github/workflows/*.yml` — CI/CD и AI‑воркфлоу;
- `docs/*.md` — документация по архитектуре, моделям, деплою и т.д.

## 3. Установка self‑hosted GitHub Runner

1. В интерфейсе GitHub репозитория:
   - Settings → Actions → Runners → New self‑hosted runner;
   - выберите Linux x64 и выполните предложенные команды на VPS
     (создание каталога, скачивание архива, запуск `config.sh`).
2. Запустите runner как сервис (рекомендуется):

> ниже пример для версии runner’а 2.329.0
```bash
cd ~/
mkdir actions-runner && cd actions-runner

curl -o actions-runner-linux-x64-2.329.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.329.0/actions-runner-linux-x64-2.329.0.tar.gz
echo "194f1e1e4bd02f80b7e9633fc546084d8d4e19f3928a324d512ea53430102e1d  actions-runner-linux-x64-2.329.0.tar.gz" | shasum -a 256 -c
tar xzf ./actions-runner-linux-x64-2.329.0.tar.gz

./config.sh --url https://github.com/codex-k8s/project-example --token YOUR_RUNNER_TOKEN
// нажимайте Enter для имени и выбора типа runner’а

sudo ./svc.sh install
sudo ./svc.sh start
```

Убедитесь, что runner работает от пользователя, который:

- входит в группы `microk8s` и `docker`;
- видит kubeconfig по пути `/home/runner/.kube/microk8s.config`.

## 4. Подготовка директорий

```bash
cd ~/
mkdir -p ~/codex/envs ~/codex/data
```

По умолчанию в `services.yaml`:

- `registry: localhost:32000`;
- `environments.staging.kubeconfig: "/home/runner/.kube/microk8s.config"`;
- домены:
  - `baseDomain.dev` по умолчанию `dev.example-domain.ru`;
  - `baseDomain.staging` по умолчанию `staging.example-domain.ru`;
  - `baseDomain.ai` по умолчанию совпадает с `staging`.

Эти домены можно переопределить через переменные окружения:

- `BASE_DOMAIN_DEV` — домен для dev‑окружения;
- `BASE_DOMAIN_STAGING` — домен для стейджинга;
- `BASE_DOMAIN_AI` — домен для AI‑слотов (если не задан, берётся `BASE_DOMAIN_STAGING`).

Рекомендуется задать их как Repository Variables в GitHub и/или
как переменные среды при запуске `codexctl`.

Таймаут ожидания деплоя:
- `codex.timeouts.deployWait` в `services.yaml` управляет временем ожидания `kubectl wait` после `codexctl apply/ci ensure-ready`.
- Значение по умолчанию — `10m` (если не задано в `services.yaml` и не переопределено флагом `--wait-timeout`).

## 5. Переменные и секреты в GitHub

### 5.1. Repository Variables (`Settings → Secrets and variables → Actions → Variables`)

Рекомендуемые переменные:

- `CODE_ROOT_BASE` — базовый каталог для исходников dev‑AI слотов,
  например `/home/runner/codex/envs/`;
- `DATA_ROOT` — каталог с данными БД/Redis,
  например `/home/runner/codex/data/`;
- `DEV_SLOTS_MAX` — максимальное количество dev‑AI слотов (например, `2`).
- `AI_ALLOWED_USERS` — список GitHub‑логинов, которым разрешено запускать AI‑воркфлоу (например, `user1,user2`), добавляется как Repository Variable.
- `CODEX_GH_USERNAME` — GitHub‑логин бота, добавляется как Repository Variable.
- `LETSENCRYPT_EMAIL` — email для регистрации ACME аккаунта в Let’s Encrypt (например, `admin@example-domain.ru`).

### 5.2. Repository Secrets

Минимальный набор:

- `CODEX_GH_PAT` — GitHub Personal Access Token бота (с правами на создание и редактирование PR/Issue и комментариев к ним. С правами на создание веток и пуш коммитов в них.);
- `OPENAI_API_KEY` — ключ для OpenAI;
- `CONTEXT7_API_KEY` — ключ для Context7;
- `POSTGRES_USER` — логин БД (по умолчанию `chat`);
- `POSTGRES_PASSWORD` — пароль БД;
- `REDIS_PASSWORD` — пароль Redis;
- `SECRET_KEY` — секрет Django (`python -c "import secrets; print(secrets.token_urlsafe(50))"`).

Все эти переменные используются в:

- `deploy/secret.yaml`;
- `services.yaml` (hook’и и apply);
- GitHub Actions (`staging_deploy_main.yml`, `ai_*` воркфлоу).

## 6. Первый деплой стейджинга

После настройки runner и секретов:

1. Убедитесь, что изменения закоммичены и пушнуты в ветку `main`.
2. В GitHub во вкладке Actions появится workflow
   **“Staging deploy 🚀”** (`.github/workflows/staging_deploy_main.yml`).
3. При следующем push в `main`:
   - соберутся и отзеркалятся необходимые образы (`codexctl ci images --env staging --mirror --build`);
   - `codexctl ci apply --env staging --preflight --wait` применит инфраструктуру и сервисы;
   - в кластере появится неймспейс `project-example-staging`.

Проверка:

```bash
microk8s kubectl get pods -n project-example-staging
microk8s kubectl get ingress -n project-example-staging
```

Если DNS и TLS настроены, фронтенд будет доступен по `https://staging.example-domain.ru/` (по вашему домену).

Для локального теста можно:

```bash
microk8s kubectl port-forward -n project-example-staging svc/web-frontend 8080:80
```

и открыть `http://localhost:8080`.

## 7. Флоу планирования задач с агентом

1. Создайте Issue в репозитории и опишите задачу/подпроект.
2. Повесьте на Issue метку `[ai-plan]`.
3. Запустится workflow `.github/workflows/ai_plan_issue.yml`:
   - создаст или переиспользует AI‑slot (namespace `project-example-dev-<slot>`);
   - развернёт инфраструктуру и сервисы через `codexctl ci ensure-ready`;
   - запустит планирующего агента `prompt run --kind plan_issue --lang ru`.
4. Агент оставит в Issue комментарий с планом работ, предложенной архитектурой
   и структурой подзадач.

Чтобы пересобрать или дополнительно уточнить план:

- оставьте комментарий с текстом, содержащим `[ai-plan]`;
- workflow `ai_plan_review.yml` найдёт корневой планирующий Issue
  и запустит агент `plan_review` (короткий или полный режим в зависимости от пересоздания окружения).

## 8. Флоу разработки с агентом

1. Для конкретной задачи (Issue) повесьте метку `[ai-dev]`.
2. Workflow `ai_dev_issue.yml`:
   - выделит/найдёт слот через `codexctl ci ensure-slot`;
   - развернёт окружение в этом слоте (`codexctl ci ensure-ready --prepare-images --apply`);
   - создаст/переключится на ветку `codex/issue-<номер>`;
   - запустит dev‑агента `prompt run --kind dev_issue --lang ru`;
   - после завершения работы агента закоммитит изменения и запушит ветку.
3. Если для ветки уже есть PR, workflow попытается найти его и
   оставит комментарий с ссылками на окружение.

Дальше вы можете:

- просмотреть дифф в PR;
- дополнительно поправить код вручную;
- дать агенту новые инструкции и перезапустить `[ai-dev]`.

## 9. Флоу review/fix для PR

Для уже открытого PR:

1. Попросите ревьюера сделать обычный code review.
2. Если ревьюер ставит состояние `changes requested`,
   сработает `ai_pr_review.yml`:
   - поднимет (или переиспользует) AI‑slot для этого PR;
   - запустит агента `prompt run --kind dev_review --lang ru`;
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

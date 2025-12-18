ARG PYTHON_IMAGE_VERSION=3.13-bookworm
FROM python:${PYTHON_IMAGE_VERSION} AS project_example_django_backend

ARG SERVICE_VERSION
ENV SERVICE_VERSION=${SERVICE_VERSION}

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    TZ=UTC

WORKDIR /backend

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends curl wget jq build-essential && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt /backend/requirements.txt
RUN pip install --upgrade pip && pip install -r requirements.txt

COPY . /backend

RUN chmod +x /backend/entrypoint.sh && \
    SECRET_KEY=dummy python manage.py collectstatic --noinput

CMD ["/backend/entrypoint.sh"]


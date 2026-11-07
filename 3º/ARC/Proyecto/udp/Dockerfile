<<<<<<< HEAD
# CPython estable, código fuente y checksum oficiales; no requiere imagen «t» implícita.
FROM debian:bookworm-slim AS build
ARG PYTHON_VERSION=3.14.10
ARG PYTHON_SHA256=de414a57119abef52784d8993e341d53aed635fc2b05c2607d9f98a2c5b72032
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates curl pkg-config xz-utils \
    libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev \
    libffi-dev liblzma-dev libncurses-dev uuid-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /build
RUN curl --fail --location --retry 3 \
      "https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tar.xz" -o python.tar.xz \
    && echo "${PYTHON_SHA256}  python.tar.xz" | sha256sum -c - \
    && tar -xJf python.tar.xz --strip-components=1 \
    && ./configure --prefix=/opt/python --disable-gil --with-ensurepip=no \
    && make -j4 && make install

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libssl3 zlib1g libbz2-1.0 libreadline8 libsqlite3-0 \
    libffi8 liblzma5 libncursesw6 libuuid1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /opt/python /opt/python
ENV PATH="/opt/python/bin:${PATH}" PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
RUN python3.14t -c 'import asyncio,json,socket,sys,sysconfig; assert sysconfig.get_config_var("Py_GIL_DISABLED") == 1; assert not sys._is_gil_enabled()'

FROM runtime AS application
WORKDIR /app
COPY *.py ./
COPY tests/ ./tests/

FROM application AS generator
ENV PYTHON_GIL=1
ENTRYPOINT ["python3.14t", "-u", "generator_profile.py"]

FROM application AS server
EXPOSE 8888/udp
STOPSIGNAL SIGTERM
ENTRYPOINT ["python3.14t", "-u", "server_parallel.py"]
=======
FROM python:3.12-slim

# Prevent Python from writing .pyc files and keep stdout/stderr unbuffered for real-time logs
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install basic networking utilities for diagnosis and troubleshooting
RUN apt-get update && apt-get install -y --no-install-recommends \
    iproute2 \
    net-tools \
    iputils-ping \
    && rm -rf /var/lib/apt/lists/*

# Copy application files
COPY . .

# Expose default UDP simulation port
EXPOSE 8888/udp

# Bind to 0.0.0.0 so external machines can reach the container through the host
ENTRYPOINT ["python3", "server.py"]
CMD ["--host", "0.0.0.0", "--port", "8888"]
>>>>>>> d0a63e964b1bc5d624bd67e84f926dc5c6416ddd

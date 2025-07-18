# deploying the server
major part of this app is that it is easy to deploy. so there's two ways.

# using docker-compose
docker-compose will combine all the services needed to run the app into one container. this is especially helpful for local development but
can also be used in cloud environments.

here's how to start locally:
## 0. clone the repo
## 1. deps
- [docker](https://www.docker.com/get-started) and [docker-compose](https://docs.docker.com/compose/install/) installed (easiest is just installing the docker desktop app)

## 2. configuration
- specify values in the `docker-compose.yml` file:
```yaml
services:
    keylock:
        build:
            context: .
            dockerfile: Dockerfile
        container_name: keylock
        depends_on:
            - redis
            - postgres
            - vault
            - vault-init
        volumes:
            - ./keylock.toml:/keylock.toml:ro
        ports:
            - "8755:8755"
        command: ["/keylock"]
        environment:
            - VAULT_ADDR=http://vault:8200
        restart: unless-stopped

    redis:
        image: redis:8
        container_name: redis
        ports:
            - "6379:6379"
        restart: unless-stopped

    postgres:
        image: postgres:14.15
        container_name: postgres
        environment:
            POSTGRES_DB: keylock
            POSTGRES_USER: <specify a postgres user>
            POSTGRES_PASSWORD: <specify a postgres password>
        ports:
            - "5432:5432"
        volumes:
            - pgdata:/var/lib/postgresql/data
        restart: unless-stopped

    vault:
        image: hashicorp/vault:1.20.0
        container_name: vault
        ports:
            - "8200:8200"
        cap_add:
            - IPC_LOCK
        environment:
            VAULT_DEV_ROOT_TOKEN_ID: <specify a vault root token>
            VAULT_DEV_LISTEN_ADDRESS: "0.0.0.0:8200"
        volumes:
            - vault_data:/vault/data
        restart: unless-stopped

    vault-init:
        image: hashicorp/vault:1.20.0
        depends_on:
            - vault
        entrypoint: ["/bin/sh", "-c", "/init/vault-init.sh"]
        volumes:
            - ./scripts/vault-init.sh:/init/vault-init.sh:ro
        env_file:
            - .env.vault-init
        environment:
            VAULT_ADDR: http://vault:8200
            VAULT_TOKEN: <use the one you specified in VAULT_DEV_ROOT_TOKEN_ID>

volumes:
    pgdata:
    vault_data:
```
- create a keylock.toml file in the main directory and specify the vault token.
```toml
addr = ":8755" # this is the address the server will listen on
debug = true # debug mode, set to false in production

[redis]
network = "tcp" # network type to connect to redis, use this value.
hostport = "redis:6379" # host and port of the redis server, use this value.
db = 0 # redis database to use, use this value.
timeout = 30 # timeout for redis connections in seconds

[postgres]
host = "postgres" # postgres host, use this value.
port = 5432 # port of the postgres server, use this value.
ssl = false # whether to use ssl for postgres connections.
database = "keylock" # database name, use this value.

[vault]
address = "http://vault:8200" # address of the vault server, use this value.
timeout = 10 # timeout for vault connections in seconds.
retry_wait_min = 1500 # minimum wait time between retries in milliseconds.
retry_max = 3 # maximum number of retries.
token = "<use the one you specified in docker-compose.yml>" # vault token.
```

- create a .env.vault-init in the main directory. specify the fields.
```env
PSQL_USER=<use the one you specified in docker-compose.yml>
PSQL_PASS=<use the one you specified in docker-compose.yml>
ENC_KEY=<specify a 64 character long hex string (32 bytes)>
```

- ensure `scripts/vault-init.sh` can be executed.
```bash
chmod +x scripts/vault-init.sh
```
you may need to use sudo

## 3. run
- run the following commands in the main directory:
```bash
docker-compose down -v
docker-compose up
```

# using docker
docker can be used to run the keylock app in a single container however the other services will need to be run separately (postgres, redis, hashicorp vault).

ill publish instructions when i use this method.

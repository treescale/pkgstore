# About

[TreeScale.com](http://treescale.com) pkgstore is a simple NPM and Pypi registry server, which also acts as a proxy to the generic public registries. It is built for easy maintainability and performance.

pkgstore is built with an extendable structure that allows adding more storage backends or databases to keep the package metadata information. Currently, by default, the storage backend is an AWS S3 bucket or Minio Bucket if you have a self-hosted environment.

The database is a simple SQLite file, which is configurable from the environment variable of `DATABASE_URL`, and it acts as a database type selector based on the given database URL prefix, like if you have a `postgresql://...` then the database instance will act with a PostgreSQL driver. Otherwise, it will fall back to SQLite.
You can see how it's done in [`docker-compose.yaml` file](https://github.com/treescale/pkgstore/blob/main/docker-compose.yaml#L10-L10)

## Running Locally with Go

This is a standard Golang Gin project and all the dependencies are inside `go.mod` and `go.sum` files. So you can run it with the following steps:

#### 1. Install Go Dependencies

Command below is going to download Go dependencies and put them inside `vendor` folder.

```bash
go mod download
go mod vendor
```

#### 2. Run the Project

Finally, after having the UI built and Go dependencies installed, we can run the project with the following command:

```bash
go run ./cmd/server
```

OR, we can just build the final binary and run it:

```bash
go build -o pkgstore ./cmd/server

./pkgstore
```

## Running with Docker

We have a `docker-compose.yaml` file that you can use to run the project with Docker. It will run the following services:

- `pkgstore`: The main pkgstore service
- `minio`: A self-hosted S3 compatible storage service
- `postgres`: A PostgreSQL database service

**Note:** for Docker-Compose based configuration we are using PostgreSQL database, which makes it easier to run the project with Docker. But you can change the database URL to any other database type, like MySQL, SQLite, etc.

```bash
docker-compose build
docker-compose up
```

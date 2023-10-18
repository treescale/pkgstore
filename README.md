# About

![Alin.io Package Store UI](https://i.imgur.com/UVyOgq8.png)

[Alin.io](http://Alin.io) pkgstore is a simple NPM and Pypi registry server, which also acts as a proxy to the generic public registries. It is built for easy maintainability and performance.

pkgstore is built with an extendable structure that allows adding more storage backends or databases to keep the package metadata information. Currently, by default, the storage backend is an AWS S3 bucket or Minio Bucket if you have a self-hosted environment.

The database is a simple SQLite file, which is configurable from the environment variable of `DATABASE_URL`, and it acts as a database type selector based on the given database URL prefix, like if you have a `postgresql://...` then the database instance will act with a PostgreSQL driver. Otherwise, it will fall back to SQLite.
You can see how it's done in [`docker-compose.yaml` file](https://github.com/alin-io/pkgstore/blob/79af6bbff49be70c394277473655b7fd5618bced/docker-compose.yaml#L10-L10)

![Alin.io Package Store UI](https://i.imgur.com/aY365Pa.png)

## Running Locally with Go

This is a standard Golang Gin project and all the dependencies are inside `go.mod` and `go.sum` files. So you can run it with the following steps:

#### 1. Install Go Dependencies

Command below is going to download Go dependencies and put them inside `vendor` folder.

```bash
go mod download
go mod vendor
```

#### 2. Build the UI Project

We have a React.js based UI inside `ui` folder to list the available packages and their versions. You can build the UI project with the following command:

```bash
cd ui
npm install
npm run build
```
This will build the `Vite React TypeScript` project and will copy all the bundles to the `cmd/server/ui` folder, which will then get bundled with the Go binary.
So that at the end, we should have a single binary with the UI inside it.

#### 3. Run the Project

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

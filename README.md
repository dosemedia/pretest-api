# Hasura Base

This is a starter project that uses Go as a companion server for Hasura.

For a node.js based version see : [https://github.com/aaronblondeau/hasura-base](https://github.com/aaronblondeau/hasura-base)

Authentication is handled by Hasura and JWTs.

## Getting Started

1.  Install dependencies

```
go get
```

Also install the hasura cli : https://hasura.io/docs/latest/hasura-cli/install-hasura-cli/

2. Copy .env.example to .env and update values (note "TODO" values).  The S3 secret and key will be updated in step 4 below.

3. Start the docker containers

Before starting the containers, switch the postgres image to something like "postgis/postgis:15-3.3" in docker-compose.yml if you need PostGIS support.

```
docker compose up -d
```

4. Use the minio UI (http://localhost:9090/) to create a 'user-public' bucket as well as to create an api access key and secret. Update S3_ACCESS_KEY and S3_SECRET_KEY in .env file.

5. Run hasura migrations and apply metadata

```powershell
setx HASURA_GRAPHQL_ADMIN_SECRET "mydevsecret"
```

```bash
export HASURA_GRAPHQL_ADMIN_SECRET=mydevsecret
```

```
hasura migrate apply --project ./hasura
hasura metadata apply --project ./hasura
```

6. Update prisma:

```
go run github.com/steebchen/prisma-client-go db pull
go run github.com/steebchen/prisma-client-go generate
```

7. Start the hasura console

```powershell
setx HASURA_GRAPHQL_ADMIN_SECRET "mydevsecret"
hasur

```bash
export HASURA_GRAPHQL_ADMIN_SECRET=mydevsecret
hasura console --project ./hasura
```

8. Start the golang server

```
go run main.go
```

9. Use the hasura console to create additonal tables, actions, events, relationships, and permissions.

Other admin tools are available at (see .env file for passwords):
Minio UI : http://localhost:9090/

10. When done, stop the docker containers

```
docker compose down
```

## Email Templates

Emails templates are managed with [maizzle](https://maizzle.com/).

To develop email templates make sure you have Node.js installed and then:

```
cd emails
npm install
npm run dev
```

Once the dev server has started, go to http://localhost:3050/.  Updates to templates will live reload!

To build email templates

```
cd emails
npm run build
```

Note that email templates are embedded in the go executable so they must be generated before building or running "go run main.go".

## Troubleshooting

https://goprisma.org/docs/getting-started/quickstart

## TODO

How to set go prisma db conn string from env var?
Update env vars (crew specific ones added, search for os.Getenv)
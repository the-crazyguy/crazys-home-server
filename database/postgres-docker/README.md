# Using PostgreSQL in docker

1. Set up the container
```
docker compose up -d
```
2. Accessing the container
```
docker exec -it my-postgres bash
```
3. Opening the created database
```
psql -U crzy -d crzy-fs-db
```
where `-U` denotes the username and `-d` the database name

## Queries

1. Add a user
```SQL
INSERT INTO "users" ("id", "username", "password", "created_at")
VALUES (gen_random_uuid(), 'test_1', 'test123', now());
```

2. Add Trust

User A trusts user B
```SQL
INSERT INTO user_trusts (user_id, trusted_user_id)
VALUES ('UUID', 'UUID');
```

## References

- [Postgres dockerhub page](https://hub.docker.com/_/postgres)
- [Postgres in docker video](https://www.youtube.com/watch?v=4p7x6x2kq3g)

docker pull postgres
docker run --name skywire-postgres-db -e POSTGRES_PASSWORD=supersecretpass -p 5432:5432 -d postgres
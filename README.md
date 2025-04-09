# Authorization Service

This service handles authorization using SpiceDB with CockroachDB as the persistent storage. It implements a permission system where users can manage their own profiles and grant manager access to other users.

## Prerequisites

- Go 1.23.3 or later
- Docker and Docker Compose (for development)
- Make (optional, for using Makefile commands)

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/Gambitier/spicedb-authz-poc.git
cd spicedb-authz-poc
```

2. Start the development environment:
```bash
docker-compose up -d
```

This will start:
- SpiceDB (port 50051)
- CockroachDB (port 26257)
- Authorization Service (ports 8085, 8086, 9093)

3. Apply the SpiceDB schema:
```bash
docker-compose exec spicedb spicedb migrate head
```

4. Run the service locally:
```bash
go mod download
go run cmd/server/main.go
```

## Production Deployment

For production deployment, you'll need to:

1. Set up a CockroachDB cluster
2. Configure SpiceDB with the CockroachDB connection
3. Deploy the authorization service

### CockroachDB Setup

1. Create a CockroachDB cluster
2. Create the database:
```sql
CREATE DATABASE authz;
```

### SpiceDB Setup

1. Deploy SpiceDB with CockroachDB configuration:
```yaml
datastoreEngine: cockroachdb
datastoreConnUri: "postgresql://root@cockroachdb:26257/authz?sslmode=disable"
```

2. Apply the schema:
```bash
spicedb migrate head
```

## API Endpoints

### HTTP API (Port 8085)

- `POST /v1/check-permission`
  - Check if a user has permission on a profile
  - Body: `{"user_id": "string", "profile_id": "string", "permission": "string"}`

- `POST /v1/set-manager`
  - Set a manager for a profile
  - Body: `{"owner_id": "string", "profile_id": "string", "manager_id": "string"}`

### gRPC API (Port 8086)

See `proto/authorization.proto` for detailed service definitions.

## Monitoring

- Prometheus metrics available on port 9093
- Health check endpoint: `/health`

## Configuration

Configuration can be provided through:
- Environment variables
- `default.yaml` file
- Command line flags

See `default.yaml` for available configuration options.

## Development Guidelines

1. Follow Go best practices and coding standards
2. Write tests for new features
3. Update documentation when making changes
4. Use conventional commits for commit messages

## Troubleshooting

1. Check service logs:
```bash
docker-compose logs -f auth-service
```

2. Check SpiceDB logs:
```bash
docker-compose logs -f spicedb
```

3. Check CockroachDB logs:
```bash
docker-compose logs -f cockroachdb
```

## License

Proprietary - All rights reserved 
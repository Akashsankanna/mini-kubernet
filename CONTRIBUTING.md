# Contributing

## Development Setup

1. **Clone repository**
```bash
git clone https://github.com/mini-kubernet/mini-kubernet.git
cd mini-kubernet
```

2. **Create feature branch**
```bash
git checkout -b feature/your-feature
```

3. **Set up environment**
```bash
cp .env.example .env
# Edit .env with your values
```

4. **Start local environment**
```bash
docker-compose up -d
```

## Code Style

### Go
- Use `gofmt` for formatting
- Follow `golint` recommendations
- Use meaningful variable names
- Comment exported functions

```bash
go fmt ./...
golint ./...
go vet ./...
```

### JavaScript/React
- Use ESLint configuration
- Follow Prettier for formatting
- Use functional components
- Props validation with PropTypes

```bash
npm run lint
npm run format
```

## Testing

### Run All Tests
```bash
make test
```

### Backend Tests
```bash
cd backend/auth-service && go test -v ./...
```

### Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Git Workflow

1. **Commit messages**
```
<type>: <subject>

<body>

<footer>
```

Types: feat, fix, docs, style, refactor, perf, test, chore

Example:
```
feat: add canary deployment support

Implement canary deployment with automatic rollback on error rate threshold

Closes #123
```

2. **Pull Request**
- Describe changes
- Reference related issues
- Ensure tests pass
- Wait for review

3. **Merge**
- Squash commits if needed
- Delete branch after merge
- Update changelog

## Code Review Checklist

- [ ] Code follows style guide
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Commits are clean
- [ ] Performance impact assessed
- [ ] Security implications reviewed

## Deployment Testing

### Local Testing
```bash
docker-compose up -d
# Run integration tests
```

### Staging Deployment
```bash
kubectl config use-context staging
make deploy-k8s
# Verify functionality
```

### Production Canary
```bash
bash scripts/canary-deploy.sh
# Monitor metrics
```

## Performance Guidelines

- API response time: < 100ms (p95)
- Database queries: < 50ms
- Cache hit rate: > 90%
- Error rate: < 0.1%

## Security Guidelines

- Never commit secrets to repository
- Use `.env.example` for defaults
- Sanitize user inputs
- Validate all API inputs
- Use prepared statements for SQL
- Enable HTTPS everywhere
- Implement rate limiting

## Documentation

- Update README for feature changes
- Add API documentation for new endpoints
- Document configuration options
- Include examples
- Update architecture diagrams if needed

## Release Process

1. **Version Bump**
```bash
# Update version in relevant files
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

2. **Changelog**
- Document all changes
- List breaking changes
- Include upgrade instructions

3. **Announce**
- GitHub release
- Documentation update
- Email notification

## Support

- Issues: https://github.com/mini-kubernet/issues
- Discussions: https://github.com/mini-kubernet/discussions
- Email: support@mini-kubernet.dev

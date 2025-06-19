# Contributing to Shopmon

Thank you for your interest in contributing to Shopmon! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/shopmon.git
   cd shopmon
   ```
3. Add the upstream repository as a remote:
   ```bash
   git remote add upstream https://github.com/original-owner/shopmon.git
   ```

## Development Setup

### Prerequisites

- Docker and Docker Compose (for running the demo environment)
- Node.js 24+ (for some build tools)

### Initial Setup

1. Install dependencies:
   ```bash
   make setup
   make up
   ```

2. Set up the database:
   ```bash
   make load-fixtures
   ```

This will create a few users and a shop with sample data to help you get started.

3. Start the development environment:
   ```bash
   make dev
   ```

This will start:
- API server on http://localhost:5789
- Frontend dev server on http://localhost:3000

### Environment Variables

Copy the example environment files and configure them:

```bash
# API configuration
cp api/.env.example api/.env
```

See the Environment Configuration section in [CLAUDE.md](./CLAUDE.md) for detailed variable descriptions.

## Making Changes

### Branch Naming

Create a new branch for your feature or fix:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Development Workflow

1. Make your changes in the appropriate directory:
   - API code: `/api/src/`
   - Frontend code: `/frontend/src/`
   - Database migrations: `/api/drizzle/`

2. Run code quality checks:
   ```bash
   bun run biome:check
   ```

3. Fix any issues:
   ```bash
   bun run biome:fix
   ```

### Database Changes

If your changes require database modifications:

1. Generate a new migration:
   ```bash
   cd api
   bun run db:generate
   ```

2. Apply the migration:
   ```bash
   bun run db:migrate
   ```

## Code Style

We use [Biome](https://biomejs.dev/) for code formatting and linting. The configuration is in `biome.json`.

### Key Guidelines

- Use TypeScript for all new code
- Follow the existing patterns in the codebase
- Keep functions small and focused
- Add types for all function parameters and return values
- Use meaningful variable and function names
- Comment complex logic

### Frontend Specific

- Use Composition API for Vue components
- Keep components small and reusable
- Use Pinia stores for state management
- Follow the existing CSS structure with PostCSS

### API Specific

- Use tRPC procedures for API endpoints
- Add proper input validation with Zod
- Handle errors appropriately
- Keep database queries efficient

## Testing

### Manual Testing

1. Test your changes locally in the development environment
2. Test with the demo environment:
   ```bash
   make up
   ```

### Automated Tests

Currently, the project doesn't have automated tests. If you'd like to contribute tests, that would be greatly appreciated!

## Submitting Changes

1. Commit your changes with clear, descriptive messages:
   ```bash
   git commit -m "feat: add shop health dashboard"
   # or
   git commit -m "fix: resolve pagination issue in shop list"
   ```

2. Push to your fork:
   ```bash
   git push origin your-branch-name
   ```

3. Create a Pull Request on GitHub:
   - Provide a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Describe testing performed

### Commit Message Format

We follow conventional commits:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Screenshots (if applicable)
6. Your environment details:
   - OS and version
   - Browser (for frontend issues)
   - Bun version
   - Any relevant logs

### Feature Requests

For feature requests, please include:

1. A clear description of the feature
2. Use cases and benefits
3. Any implementation ideas (optional)
4. Mockups or examples (if applicable)

## Questions?

If you have questions about contributing, feel free to:

1. Check the existing issues and pull requests
2. Open a discussion issue

Thank you for contributing to Shopmon! 🎉

# Scope V1

## In Scope

- Small to medium backend Go features
- Backend Go bug fixes
- Simple DB schema changes and migrations
- Simple external integrations (single provider, straightforward request/response mapping)

## Out of Scope

- Large refactors spanning many bounded contexts
- Distributed systems redesign
- Complex infra/platform changes
- Large cross-team architecture proposals
- Autonomous merge/deploy actions

## Repository Assumptions

- Go service with handler/service/repository layering
- Migration mechanism exists
- Unit test framework exists
- Error handling conventions exist
- Repository structure is stable enough for pattern reuse

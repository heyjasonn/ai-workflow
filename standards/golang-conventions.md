# Golang Conventions

- Keep handlers thin: parse, validate, and delegate.
- Keep business logic in service layer.
- Keep data access in repository layer.
- Propagate `context.Context` through all boundaries.
- Use explicit error wrapping with actionable messages.
- Add interfaces only at boundaries requiring substitution/testing.
- Avoid hidden global state for request processing.

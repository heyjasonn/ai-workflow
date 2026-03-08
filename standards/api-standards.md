# API Standards

- Use stable request/response naming and versioned routes.
- Keep backward compatibility by default (additive changes first).
- Define structured error contract (`code`, `message`, optional `details`).
- Document pagination/filter semantics consistently.
- Validate input at boundary and return deterministic error responses.

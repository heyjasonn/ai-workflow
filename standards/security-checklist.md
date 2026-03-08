# Security Checklist

- Validate authentication and authorization per endpoint/use case.
- Validate/sanitize external inputs.
- Avoid secret leakage in logs/errors.
- Guard against SQL injection and unsafe query composition.
- Minimize and redact PII in logging/telemetry.
- Enforce timeout and retry limits for external integrations.

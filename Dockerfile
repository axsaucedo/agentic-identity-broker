# Agentic Identity Broker - Production Docker Image
# Multi-stage optimized container with Alpine base
# This Dockerfile packages pre-built backend and frontend artifacts

ARG TARGETARCH=amd64

# Runtime stage: Alpine base with pre-built artifacts
FROM alpine:latest

# Build arguments for multi-architecture support
ARG TARGETARCH

# Image metadata labels
LABEL org.opencontainers.image.title="Agentic Identity Broker"
LABEL org.opencontainers.image.description="Production Docker image for Agentic Identity Broker"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.maintainer="Agentic Identity Broker Contributors"
LABEL org.opencontainers.image.source="https://github.com/anthropics/agentic-identity-broker"

# Create non-root user for security (uid=1000, gid=1000)
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy pre-built backend binary from ./bin/
# Assumes binary is already compiled and available at ./bin/identity-broker
COPY --chown=1000:1000 ./bin/identity-broker /app/identity-broker

# Copy pre-built frontend assets from ./web/dist/
# Assumes frontend is already built and available at ./web/dist/consent/
COPY --chown=1000:1000 ./web/dist/consent /app/web/dist/consent

# Ensure binary is executable
RUN chmod +x /app/identity-broker

# Expose backend service port
EXPOSE 8000

# Switch to non-root user for container execution (security best practice)
USER 1000:1000

# Application entrypoint
ENTRYPOINT ["/app/identity-broker"]

# Default to empty CMD; arguments can be passed at runtime
CMD []

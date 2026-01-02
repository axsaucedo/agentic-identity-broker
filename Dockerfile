# Agentic Identity Broker - Production Docker Image
# This Dockerfile packages pre-built backend and frontend artifacts
FROM alpine:latest

# Build arguments for multi-architecture support via docker buildx
# When using: docker buildx build --platform linux/amd64,linux/arm64 ...
# TARGETARCH is automatically set to: amd64, arm64, etc.
ARG TARGETARCH=amd64

# Image metadata labels
LABEL org.opencontainers.image.title="Agentic Identity Broker"
LABEL org.opencontainers.image.description="Production Docker image for Agentic Identity Broker"
LABEL org.opencontainers.image.version="1.0.0"
# LABEL org.opencontainers.image.maintainer="Agentic Identity Broker Contributors" # Uncomment and set maintainer once repo is public
# LABEL org.opencontainers.image.source="<repository-url>"  # Uncomment and set repository URL once repo is public

# Create non-root user for security (uid=1000, gid=1000)
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy pre-built backend binary from architecture-specific directory
# Supports multi-architecture builds via docker buildx
# TARGETARCH automatically set to: amd64, arm64, etc.
# Binary should be built with: just build-linux-amd64 or just build-linux-arm64
COPY --chown=1000:1000 ./build/linux/${TARGETARCH}/identity-broker /app/identity-broker

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

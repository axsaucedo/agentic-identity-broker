# Feature Specification: Docker Setup

**Feature Branch**: `010-docker-setup`
**Created**: 2025-12-30
**Status**: Draft
**Input**: User description: "docker setup"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Build Complete Application Image (Priority: P1)

A developer or CI/CD pipeline needs to package the Go backend application and React frontend into a single Docker image for deployment to container registries and runtime environments. The Dockerfile should be structured to support future multiarch builds by using TARGETARCH argument.

**Why this priority**: Docker containerization is essential for modern deployment. A single image containing both backend and frontend enables simpler deployment and orchestration.

**Independent Test**: Can be fully tested by running `docker build` and verifying the resulting container image starts successfully, serves both backend API and frontend UI.

**Acceptance Scenarios**:

1. **Given** a Dockerfile exists in the repository root, **When** running `docker build -t identity-broker .`, **Then** the build succeeds and produces a runnable container image with both backend and frontend.
2. **Given** the Docker image is built, **When** running the container with `docker run`, **Then** the application starts successfully with backend on port 8000 and frontend on port 3000.
3. **Given** the container is running, **When** accessing the frontend at `http://localhost:3000`, **Then** it loads successfully and can communicate with the backend API.

---

### User Story 2 - Extend Justfile with Docker Commands (Priority: P1)

Developers want to integrate Docker operations into the existing justfile workflow so that building and running containers uses the same established task automation pattern.

**Why this priority**: Consistency in developer workflow reduces cognitive load. Docker commands should be discoverable through `just --list` and follow the established justfile pattern.

**Independent Test**: Can be fully tested by running `just docker-build` and `just docker-run` commands and verifying they execute the Docker operations successfully.

**Acceptance Scenarios**:

1. **Given** justfile is updated with Docker tasks, **When** running `just docker-build`, **Then** the Docker image is built successfully with both backend and frontend.
2. **Given** the Docker image exists, **When** running `just docker-run`, **Then** the container starts with both services accessible.
3. **Given** a developer runs `just --list`, **Then** Docker-related tasks are visible with descriptions.

---

### Edge Cases

- What happens if Docker is not installed? Commands should fail gracefully with a helpful error message.
- How are environment variables passed to the running container? Port and configuration should be customizable.
- What happens if ports 8000 or 3000 are already in use? The command should allow specifying alternative ports.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a Dockerfile that packages both Go backend and React frontend into a single container image.
- **FR-002**: System MUST use a multi-stage Dockerfile build to minimize final image size (separate stages for building backend, building frontend, and runtime).
- **FR-003**: System MUST declare `ARG TARGETARCH` in the Dockerfile runtime stage to support multiarch-compatible binary selection.
- **FR-004**: System MUST copy the application binary from `build/linux/${TARGETARCH}/` directory path in the Dockerfile, enabling pre-built binaries for different architectures.
- **FR-005**: System MUST build the React frontend from source and include the built frontend assets in the image.
- **FR-006**: System MUST configure the backend to serve frontend static files from the appropriate directory.
- **FR-007**: System MUST include a `.dockerignore` file to exclude unnecessary files from the build context.
- **FR-008**: System MUST add `docker-build` justfile task that builds a Docker image with appropriate tag.
- **FR-009**: System MUST add `docker-run` justfile task that runs the built Docker image with ports 8000 and 3000 exposed.
- **FR-010**: System MUST use an appropriate lightweight base image for the runtime stage.
- **FR-011**: System MUST expose ports 8000 (backend) and 3000 (frontend UI) in the Dockerfile EXPOSE instruction.
- **FR-012**: System MUST set appropriate ENTRYPOINT or CMD in Dockerfile to start the application.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Dockerfile builds successfully without errors, compiling both backend and frontend.
- **SC-002**: Resulting Docker image runs successfully with backend responding to HTTP requests on port 8000.
- **SC-003**: Frontend is accessible and loads successfully on port 3000 within the running container.
- **SC-004**: Frontend can communicate with backend API endpoints successfully.
- **SC-005**: `just docker-build` command builds the image successfully with both backend and frontend.
- **SC-006**: `just docker-run` command starts the container with both backend and frontend services ready.
- **SC-007**: Final Docker image is optimized through multi-stage builds.
- **SC-008**: Application handles graceful shutdown of both backend and frontend services.
- **SC-009**: Dockerfile declares `ARG TARGETARCH` for future multiarch build compatibility.
- **SC-010**: Binary copy path in Dockerfile uses `build/linux/${TARGETARCH}/` pattern for architecture-specific selection.

## Assumptions

- Docker is installed on developer and CI/CD machines.
- Go 1.24.0 is the target version for backend compilation.
- Node.js is available in the build environment for frontend compilation.
- Application backend listens on port 8000.
- Frontend is built to static assets (via Vite or similar) and served by the backend.
- PostgreSQL and other external services are configured via environment variables.
- Binaries can be pre-built and placed in `build/linux/amd64/` and `build/linux/arm64/` directories for docker builds to consume.

## Out of Scope

- Docker Compose configuration.
- Kubernetes deployment.
- CI/CD pipeline integration.
- Container registry operations.
- Actual multiarch build automation (that can be added later as a separate feature).

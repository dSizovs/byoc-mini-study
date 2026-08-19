# BYOC Mini Project (for my own studying purposes)

A minimal **Bring Your Own Compute (BYOC)** prototype. Zero trust data plane execution, pull based gRPC scheduling over mTLS and sandboxed execution using Google's gVisor.

## Architecture

* Raw data never leaves the client VM.
* The worker connects outbound to the control plane.
* Full cryptographic authentication and SAN validation between Control Plane and Worker.
* Workloads execute inside isolated `gVisor` (`runsc`) containers with `--network=none`.

## Components

* `/control-plane`: Go server running gRPC on port `50051`.
* `/worker`: Python agent executing tasks locally via Docker + gVisor.
* `/certs`: OpenSSL script generating Root CA, server, and client mTLS certificates.

## Quickstart

1. **Generate Certificates:**
   ```bash
   ./certs/gen_certs.sh
2. **Start Control Plane:**
   ```bash
   cd control-plane
   go run main.go
3. **Run Worker Agent:**
   ```bash
   .cd worker
   python3 worker.py

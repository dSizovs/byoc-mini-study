# BYOC Mini Project (for my own studying purposes)

I set up two isolated virtual machines where VM 1 knew how to convert text to uppercase but didn't know the word was `secret`, while VM 2 held the actual word `secret` and converted it to `SECRET` locally without VM 1 ever seeing the text.

Why it's cool: VM 1 provided the instructions (the script to run the transformation e.g. a->A), VM 2 provided the execution power and secret data.
So VM 1 didn't see the text, while VM 2 carried out the work without coming up with the instructions itself.

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

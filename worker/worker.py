import grpc
import subprocess
import byoc_pb2
import byoc_pb2_grpc

CONTROL_PLANE_IP = "192.168.2.7:50051"

def run_worker():
    # Load mTLS Certificates
    with open('/home/ubuntu/byoc_certs/ca.crt', 'rb') as f:
        root_ca = f.read()
    with open('/home/ubuntu/byoc_certs/worker.key', 'rb') as f:
        private_key = f.read()
    with open('/home/ubuntu/byoc_certs/worker.crt', 'rb') as f:
        cert_chain = f.read()

    # Create secure mTLS channel credentials
    credentials = grpc.ssl_channel_credentials(
        root_certificates=root_ca,
        private_key=private_key,
        certificate_chain=cert_chain
    )

    print(f"Connecting to Control Plane at {CONTROL_PLANE_IP} over mTLS...")
    
    # Establish encrypted channel with Server Name Override matching server certificate
    options = (('grpc.ssl_target_name_override', 'control-plane'),)
    with grpc.secure_channel(CONTROL_PLANE_IP, credentials, options) as channel:
        stub = byoc_pb2_grpc.ControlPlaneStub(channel)

        # 1. Poll for Task
        print("Polling Control Plane for tasks...")
        response = stub.FetchTask(byoc_pb2.TaskRequest(worker_id="worker-zurich-01"))

        if not response.has_task:
            print("No pending tasks.")
            return

        print(f"Received Task [{response.task_id}]")
        print(f"Instruction Received: {response.command}")

        # 2. Execute command locally inside gVisor container
        docker_cmd = [
            "sudo", "docker", "run", "--rm",
            "--runtime=runsc",
            "--network=none",
            "-v", "/home/ubuntu/byoc_data:/data",
            "python:3.11-slim",
            "bash", "-c", response.command
        ]

        print("Executing task inside local gVisor sandbox...")
        res = subprocess.run(docker_cmd, capture_output=True, text=True)

        # 3. Report execution status back to Control Plane
        success = (res.returncode == 0)
        status = "Execution successful inside gVisor sandbox." if success else res.stderr

        stub.SubmitResult(byoc_pb2.ResultRequest(
            task_id=response.task_id,
            worker_id="worker-zurich-01",
            success=success,
            status_message=status
        ))
        print("Reported execution completion back to Control Plane.")

if __name__ == "__main__":
    run_worker()

set -e

mkdir -p out && cd out

# 1. Root CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=BYOC-Root-CA"

# 2. Server Cert
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=control-plane"

cat <<EOT > server_ext.cnf
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
IP.1 = 192.168.2.7
DNS.1 = control-plane
EOT

openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -extfile server_ext.cnf

# 3. Worker Cert
openssl genrsa -out worker.key 2048
openssl req -new -key worker.key -out worker.csr -subj "/CN=byoc-worker"

cat <<EOT > worker_ext.cnf
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOT

openssl x509 -req -days 365 -in worker.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out worker.crt -extfile worker_ext.cnf
echo "Certificates generated successfully in ./certs/out/"

default:
    @just --list

fresh: stop start

start: create-cluster install-operator install-rbac install-job install-stream create-secret create-mock-stream

stop:
    kind delete cluster

create-cluster:
    kind create cluster

install-operator:
    helm install arcane oci://ghcr.io/sneaksanddata/helm/arcane-operator \
      --version v0.0.14  \
      --namespace arcane \
      --create-namespace

install-stream:
    helm install arcane-stream-microsoft-sql-server oci://ghcr.io/sneaksanddata/helm/arcane-stream-microsoft-sql-server \
        --namespace arcane \
        --version v1.0.8

install-rbac:
    kubectl apply -f integration_tests/manifests/rbac.yaml

install-job:
    kubectl apply -f integration_tests/manifests/job_template.yaml

create-secret:
    kubectl apply -f integration_tests/manifests/mock_connection_string.yaml

create-mock-stream:
    kubectl apply -f integration_tests/manifests/mock_stream.yaml

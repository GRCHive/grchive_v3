load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
)

def load_docker_container_deps():
    container_pull(
        name = "nginx",
        registry = "index.docker.io",
        repository = "library/nginx",
        tag = "1.19.1"
    )

    container_pull(
        name = "wordpress",
        registry = "index.docker.io",
        repository = "library/wordpress",
        tag = "5.4-php7.4-apache",
    )


docker_build('bar', '.',
    dockerfile='Dockerfile')
k8s_yaml('kubernetes/bar-deployment.yml')
k8s_resource('bar', port_forwards=8080)
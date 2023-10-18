docker_build('bar', '.',
    dockerfile='BarDockerfile')
docker_build('foo', '.',
    dockerfile='FooDockerfile')
k8s_yaml('kubernetes/bar-deployment.yml')
k8s_yaml('kubernetes/foo-deployment.yml')
k8s_yaml('kubernetes/foo-service.yml')
k8s_resource('bar', port_forwards=8080)

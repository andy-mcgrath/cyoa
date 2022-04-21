# -*- mode: Python -*-

# For more on Extensions, see: https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts([
  'rancher-desktop',
  'minikube',
  'dockerdesktop',
  'kind',
])

# Records the current time, then kicks off a server update.
# Normally, you would let Tilt do deploys automatically, but this
# shows you how to set up a custom workflow that measures it.
# local_resource(
#     'deploy',
#     'python record-start-time.py',
# )

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webapp ./'

local_resource(
  'go-compile',
  compile_cmd,
  deps=['./main.go'],
)

docker_build_with_restart(
  'go-webstory',
  '.',
  entrypoint=['/app/webapp'],
  dockerfile='deployment/Dockerfile',
  only=[
    './web',
    './webapp'
  ],
  live_update=[
    sync('./web', '/app/web'),
    sync('./webapp', '/app/webapp'),
  ],
)

k8s_yaml('deployment/kubernetes.yaml')
k8s_resource(
  'go-webstory',
  port_forwards=8080,
  resource_deps=['go-compile'],
)
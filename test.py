#!/usr/bin/env python
from subprocess import call
from sys import argv

with open("test-manifests/service-deployment.json", "r") as myfile:
    deployment = myfile.read().replace('\n', ' ').replace('\r', '')

with open("test-manifests/request-params.json", "r") as myfile:
    params = myfile.read().replace('\n', ' ').replace('\r', '')

with open("test-manifests/plan.json", "r") as myfile:
    plan = myfile.read().replace('\n', ' ').replace('\r', '')

call(['go', 'run', 'cmd/service-adapter/main.go', argv[1], deployment, plan, params, '---', '{}'])

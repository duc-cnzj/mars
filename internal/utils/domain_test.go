package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPreOccupiedLenByValuesYaml(t *testing.T) {
	var tests = []struct {
		yaml  string
		wants int
	}{
		{
			yaml:  "- host: aa<.Host1>",
			wants: 2,
		},
		{
			yaml:  "- host: <.Host23>",
			wants: 0,
		},
		{
			yaml:  "- host: 12345<.Host3>",
			wants: 5,
		},
		{
			yaml:  "- host: 123 45<.Host2>",
			wants: 2,
		},
		{
			yaml:  "-    a_-45<.Host10>",
			wants: 5,
		},
		{
			yaml:  " 2345e<.Host1>",
			wants: 5,
		},
		{
			yaml: `
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

image:
  repository: duccnzj/demo
  pullPolicy: IfNotPresent
  # Overrides the image tag whose default is the chart appVersion.
  tag: "v1"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  # Specifies whether a service account should be created
  create: false
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name: ""

podAnnotations: {}

podSecurityContext: {}
  # fsGroup: 2000

securityContext: {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
  # runAsUser: 1000

service:
  type: NodePort

ingress:
  enabled: true
  className: ""
  annotations: 
    kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  hosts:
    - host: 123<.Host1>
      paths:
        - path: /
          pathType: Prefix
  tls: 
   - secretName: <.TlsSecret1>
     hosts:
       - qwer<.Host1>

resources: 
  # We usually recommend not to specify default resources and to leave this as a conscious
  # choice for the user. This also increases chances charts run on environments with little
  # resources, such as Minikube. If you do want to specify resources, uncomment the following
  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80
  # targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}

`,
			wants: 4,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.wants, GetPreOccupiedLenByValuesYaml(test.yaml))
	}
}

func Test_max(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 1, max(1, -2))
}

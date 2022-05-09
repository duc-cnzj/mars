package socket

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeOrderedSetString(t *testing.T) {
	assert.IsType(t, (*timeOrderedSetString)(nil), NewTimeOrderedSetString(time.Now))
}

func Test_getPodSelectorsInDeploymentAndStatefulSetByManifest(t *testing.T) {
	manifest := getPodSelectorsInDeploymentAndStatefulSetByManifest([]string{`
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    meta.helm.sh/release-name: mars
  generation: 56
  labels:
    app.kubernetes.io/name: mars
  name: mars
  namespace: default
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: mars
      app.kubernetes.io/name: mars
`,
		`apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    meta.helm.sh/release-name: mars
  generation: 56
  labels:
    app.kubernetes.io/name: mars
  name: mars
  namespace: default
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: abc
      app.kubernetes.io/name: abc
`,
		`
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app.kubernetes.io/component: primary
    app.kubernetes.io/instance: mars-db
  name: mars-db-mysql-primary
  namespace: default
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: primary
      app.kubernetes.io/instance: mars-db
      app.kubernetes.io/name: mysql
`,
		`
W0509 17:36:48.835823   98185 helpers.go:555] --dry-run is deprecated and can be replaced with --dry-run=client.
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  labels:
    run: nginx
  name: nginx
spec:
  containers:
  - image: nginx
    name: nginx
    resources: {}
  dnsPolicy: ClusterFirst
  restartPolicy: Always
status: {}
`,
	})
	assert.Equal(t, []string{"app.kubernetes.io/instance=mars,app.kubernetes.io/name=mars", "app.kubernetes.io/instance=abc,app.kubernetes.io/name=abc", "app.kubernetes.io/component=primary,app.kubernetes.io/instance=mars-db,app.kubernetes.io/name=mysql"}, manifest)
}

func Test_getPreOccupiedLenByValuesYaml(t *testing.T) {
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
		assert.Equal(t, test.wants, getPreOccupiedLenByValuesYaml(test.yaml))
	}
}

func Test_imageUsedPipelineVars(t *testing.T) {
	var tests = []struct {
		pipelineVars pipelineVars
		wants        bool
		manifest     string
	}{
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants:    false,
			manifest: "image: xxx:v1",
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: true,
			manifest: `
image: xxx:v1
image: xxx:v2
image: xxx:p
image: xxx:c
image: xxx:b
`,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wants, imageUsedPipelineVars(test.pipelineVars, test.manifest))
	}
}

func Test_matchDockerImage(t *testing.T) {
	var tests = []struct {
		pipelineVars pipelineVars
		wants        string
		manifest     string
	}{
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants:    "xxx:v1",
			manifest: "image: xxx:v1",
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: "xxx:v1 xxx:v2",
			manifest: `
image: xxx:v1
image: xxx:v2
`,
		},
		{
			pipelineVars: pipelineVars{
				Pipeline: "p",
				Commit:   "c",
				Branch:   "b",
			},
			wants: "xxx:p xxx:c xxx:b",
			manifest: `
image: xxx:v1
image: xxx:v2
image: xxx:p
image: xxx:c
image: xxx:b
`,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wants, matchDockerImage(test.pipelineVars, test.manifest))
	}
}

func Test_max(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 1, max(1, -2))
}

func Test_orderedItemList(t *testing.T) {
	var o = orderedItemList{
		{
			t:    time.Now().Add(1 * time.Second),
			data: "ccc",
		},
		{
			t:    time.Now(),
			data: "bbb",
		},
		{
			t:    time.Now().Add(-1 * time.Second),
			data: "aaa",
		},
	}
	sort.Sort(o)
	assert.Equal(t, "aaa", o[0].data)
	assert.Equal(t, "bbb", o[1].data)
	assert.Equal(t, "ccc", o[2].data)
}

func Test_timeOrderedSetString_add(t *testing.T) {
	var called int
	fn := func() time.Time {
		called++
		parse, _ := time.Parse("2006-01-02 00:00:00", "2022-05-09 00:00:00")

		return parse.Add(-10 * time.Duration(called) * time.Second)
	}
	o := NewTimeOrderedSetString(fn)
	o.add("a")
	o.add("b")
	s := o.sortedItems()
	assert.Equal(t, "b", s[0])
	assert.Equal(t, "a", s[1])
}

func Test_timeOrderedSetString_has(t *testing.T) {
	o := NewTimeOrderedSetString(time.Now)
	o.add("a")
	assert.True(t, o.has("a"))
	assert.False(t, o.has("c"))
}

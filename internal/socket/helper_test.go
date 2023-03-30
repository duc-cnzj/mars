package socket

import (
	"sort"
	"testing"
	"time"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeOrderedSetString(t *testing.T) {
	assert.IsType(t, (*timeOrderedSetString)(nil), newTimeOrderedSetString(time.Now))
}

func Test_getPodSelectorsInDeploymentAndStatefulSetByManifest(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{
			in: dedent.Dedent(`
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
				`),
			out: "app.kubernetes.io/instance=mars,app.kubernetes.io/name=mars",
		},
		{
			in: dedent.Dedent(`
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
				      app.kubernetes.io/instance: abc
				      app.kubernetes.io/name: abc
				`),
			out: "app.kubernetes.io/instance=abc,app.kubernetes.io/name=abc",
		},
		{
			in: dedent.Dedent(`
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
				`),
			out: "app.kubernetes.io/component=primary,app.kubernetes.io/instance=mars-db,app.kubernetes.io/name=mysql",
		},
		{
			in: dedent.Dedent(`
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
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: Job
				metadata:
				  name: pi
				spec:
				  template:
				    metadata:
				      labels:
				        app: jobRunner-one
				    spec:
				      containers:
				      - name: pi
				        image: perl:5.34.0
				        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
				      restartPolicy: Never
				  backoffLimit: 4
				`),
			out: "app=jobRunner-one",
		},
		{
			in: dedent.Dedent(`
				apiVersion: apps/v1
				kind: DaemonSet
				metadata:
				  name: fluentd-elasticsearch
				spec:
				  selector:
				    matchLabels:
				      name: fluentd-elasticsearch
				  template:
				    metadata:
				      labels:
				        name: fluentd-elasticsearch
				    spec:
				      containers:
				      - name: fluentd-elasticsearch
				        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
				        volumeMounts:
				        - name: varlog
				          mountPath: /var/log
				`),
			out: "name=fluentd-elasticsearch",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob",
		},
		{
			in: dedent.Dedent(`
				apiVersion: batch/v1beta1
				kind: CronJob
				metadata:
				  name: hello
				spec:
				  schedule: "* * * * *"
				  jobTemplate:
				    spec:
				      template:
				        metadata:
				          labels:
				            app: cronjob-v1beta1
				        spec:
				          containers:
				          - name: hello
				            image: busybox:1.28
				            imagePullPolicy: IfNotPresent
				            command:
				            - /bin/sh
				            - -c
				            - date; echo Hello from the Kubernetes cluster
				          restartPolicy: OnFailure
				`),
			out: "app=cronjob-v1beta1",
		},
	}

	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			labels := getPodSelectorsByManifest([]string{tt.in})
			if len(labels) > 0 {
				assert.Equal(t, tt.out, labels[0])
			} else {
				assert.Equal(t, tt.out, "")
			}
		})
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
image: xxx:v1
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
	o := newTimeOrderedSetString(fn)
	o.add("a")
	o.add("b")
	o.add("b")
	s := o.sortedItems()
	assert.Equal(t, "b", s[0])
	assert.Equal(t, "a", s[1])
}

func Test_timeOrderedSetString_has(t *testing.T) {
	o := newTimeOrderedSetString(time.Now)
	o.add("a")
	assert.True(t, o.has("a"))
	assert.False(t, o.has("c"))
}

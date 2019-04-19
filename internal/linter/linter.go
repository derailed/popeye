package linter

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Issue indicates a potential linter issue.
	Issue interface {
		Severity() Level
		SetSeverity(Level)
		Description() string
		HasSubIssues() bool
		SubIssues() Issues
	}

	// Issues a collection of linter issues.
	Issues map[string][]Issue

	// Linter describes a lint resource.
	Linter struct {
		Loader

		log    *zerolog.Logger
		issues Issues
	}

	// Spinach represents a Popeye configuration object.
	Spinach interface {
		PodCPULimit() float64
		PodMEMLimit() float64
		NodeCPULimit() float64
		NodeMEMLimit() float64
		RestartsLimit() int

		CPUResourceLimits() config.Allocations
		MEMResourceLimits() config.Allocations

		Sections() []string
		LinterLevel() int
		ExcludedNS(ns string) bool
		ExcludedNode(n string) bool
	}

	// Fetcher fetches Kubernetes resources from the apiserver.
	Fetcher interface {
		ActiveCluster() string
		ActiveNamespace() string

		ClusterHasMetrics() (bool, error)
		FetchNodesMetrics() ([]mv1beta1.NodeMetrics, error)
		FetchPodsMetrics(ns string) ([]mv1beta1.PodMetrics, error)

		FetchNodes() (*v1.NodeList, error)
		FetchNamespaces() (*v1.NamespaceList, error)
		FetchPods() (*v1.PodList, error)
		FetchPodsByLabels(string) (*v1.PodList, error)
		FetchConfigMaps() (*v1.ConfigMapList, error)
		FetchSecrets() (*v1.SecretList, error)
		FetchServiceAccounts() (*v1.ServiceAccountList, error)
		FetchEndpoints() (*v1.EndpointsList, error)
		FetchServices() (*v1.ServiceList, error)
		FetchClusterRoleBindings() (*rbacv1.ClusterRoleBindingList, error)
		FetchRoleBindings() (*rbacv1.RoleBindingList, error)
		FetchPersistentVolumes() (*v1.PersistentVolumeList, error)
		FetchPersistentVolumeClaims() (*v1.PersistentVolumeClaimList, error)
		FetchHorizontalPodAutoscalers() (*autoscalingv1.HorizontalPodAutoscalerList, error)
		FetchDeployments() (*appsv1.DeploymentList, error)
		FetchStatefulSets() (*appsv1.StatefulSetList, error)
	}

	// Lister list Kubernetes resource based on configuration scopes.
	Lister interface {
		ListNodesMetrics([]v1.Node, []mv1beta1.NodeMetrics, k8s.NodesMetrics)
		ListPodsMetrics([]mv1beta1.PodMetrics, k8s.PodsMetrics)

		ListServices() (map[string]v1.Service, error)
		ListNodes() ([]v1.Node, error)
		GetEndpoints(fqn string) (*v1.Endpoints, error)
		PodsNamespaces(used []string)
		GetPod(map[string]string) (*v1.Pod, error)
		ListPods() (map[string]v1.Pod, error)
		ListPodsByLabels(string) (map[string]v1.Pod, error)
		ListAllPods() (map[string]v1.Pod, error)
		ListNamespaces() (map[string]v1.Namespace, error)
		ListRoleBindings() (map[string]rbacv1.RoleBinding, error)
		ListAllRoleBindings() (map[string]rbacv1.RoleBinding, error)
		ListAllClusterRoleBindings() (map[string]rbacv1.ClusterRoleBinding, error)
		ListConfigMaps() (map[string]v1.ConfigMap, error)
		ListSecrets() (map[string]v1.Secret, error)
		ListServiceAccounts() (map[string]v1.ServiceAccount, error)
		ListPersistentVolumes() (map[string]v1.PersistentVolume, error)
		ListPersistentVolumeClaims() (map[string]v1.PersistentVolumeClaim, error)
		ListAllPersistentVolumeClaims() (map[string]v1.PersistentVolumeClaim, error)
		ListHorizontalPodAutoscalers() (map[string]autoscalingv1.HorizontalPodAutoscaler, error)
		ListAllHorizontalPodAutoscalers() (map[string]autoscalingv1.HorizontalPodAutoscaler, error)
		ListDeployments() (map[string]appsv1.Deployment, error)
		ListAllDeployments() (map[string]appsv1.Deployment, error)
		ListStatefulSets() (map[string]appsv1.StatefulSet, error)
		ListAllStatefulSets() (map[string]appsv1.StatefulSet, error)
	}

	// Loader loads prefiltered Kubernetes resources.
	Loader interface {
		Spinach
		Fetcher
		Lister
	}
)

// MaxSeverity scans the issues and reports the highest severity.
func (i Issues) MaxSeverity(res string) Level {
	max := OkLevel
	for _, issue := range i[res] {
		if issue.Severity() > max {
			max = issue.Severity()
		}
	}
	return max
}

// NewLinter returns a new linter.
func NewLinter(l Loader, log *zerolog.Logger) *Linter {
	return &Linter{Loader: l, log: log, issues: Issues{}}
}

// NoIssues return true if not lint errors were detected. False otherwize
func (l *Linter) NoIssues(res string) bool {
	return len(l.issues[res]) == 0
}

// Issues returns a collection of linter issues.
func (l *Linter) Issues() Issues {
	return l.issues
}

// MaxSeverity return the highest severity level.
func (l *Linter) MaxSeverity(res string) Level {
	return l.issues.MaxSeverity(res)
}

func (l *Linter) initIssues(res string) {
	l.issues[res] = []Issue{}
}

func (l *Linter) addContainerIssues(res string, issues Issues) {
	var err Issue = NewError(OkLevel, "")
	newErr := true
	if v, ok := l.issues[res]; ok {
		if len(v) == 1 {
			newErr = false
			err = v[0]
		}
	}

	maxLevel := OkLevel
	for k, v := range issues {
		for _, i := range v {
			if i.Severity() > maxLevel {
				maxLevel = i.Severity()
			}

			if v, ok := err.SubIssues()[k]; ok {
				err.SubIssues()[k] = append(v, i)
			} else {
				err.SubIssues()[k] = []Issue{i}
			}
		}
	}

	if newErr {
		err.SetSeverity(maxLevel)
		l.issues[res] = append(l.issues[res], err)
	}
}

func (l *Linter) addErrors(res string, errs ...error) {
	for _, e := range errs {
		l.addIssue(res, ErrorLevel, e.Error())
	}
}

func (l *Linter) addError(res string, err error) {
	l.addIssue(res, ErrorLevel, err.Error())
}

func (l *Linter) addIssue(res string, level Level, msg string) {
	l.addIssues(res, NewError(level, msg))
}

func (l *Linter) addIssuef(res string, level Level, format string, args ...interface{}) {
	l.addIssues(res, NewErrorf(level, format, args...))
}

func (l *Linter) addIssues(res string, issues ...Issue) {
	l.issues[res] = append(l.issues[res], issues...)
}

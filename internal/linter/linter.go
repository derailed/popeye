package linter

import (
	"fmt"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Issue indicates a potential linter issue.
	Issue interface {
		Severity() Level
		Description() string
	}

	// Issues a collection of linter issues.
	Issues map[string][]Issue

	// Linter describes a lint resource.
	Linter struct {
		Loader

		log    *zerolog.Logger
		issues Issues
	}

	// Loader loads prefiltered Kubernetes resources.
	Loader interface {
		Konfig
		Lister
	}

	// Konfig represents a configuration object.
	Konfig interface {
		PodCPULimit() float64
		PodMEMLimit() float64
		NodeCPULimit() float64
		NodeMEMLimit() float64
		RestartsLimit() int

		Sections() []string
		LinterLevel() int
		ActiveCluster() string
		ActiveNamespace() string
		ExcludedNS(ns string) bool
		ExcludedNode(n string) bool
	}

	// Fetcher fetches Kubernetes resources from the apiserver.
	Fetcher interface {
		ClusterHasMetrics() (bool, error)
		FetchNodesMetrics() ([]mv1beta1.NodeMetrics, error)
		FetchPodsMetrics(ns string) ([]mv1beta1.PodMetrics, error)

		FetchNOs() (*v1.NodeList, error)
		FetchNSs() (*v1.NamespaceList, error)
		FetchPOs() (*v1.PodList, error)
		FetchCMs() (*v1.ConfigMapList, error)
		FetchSECs() (*v1.SecretList, error)
		FetchSAs() (*v1.ServiceAccountList, error)
		FetchEPs() (*v1.EndpointsList, error)
		FetchSVCs() (*v1.ServiceList, error)
		FetchCRBs() (*rbacv1.ClusterRoleBindingList, error)
		FetchRBs() (*rbacv1.RoleBindingList, error)
	}

	// Lister list Kubernetes resource based on configuration scopes.
	Lister interface {
		Fetcher

		ListNodesMetrics([]v1.Node, []mv1beta1.NodeMetrics, NodesMetrics)
		ListPodsMetrics([]mv1beta1.PodMetrics, PodsMetrics)

		ListServices() (map[string]v1.Service, error)
		ListNodes() ([]v1.Node, error)
		GetEndpoints(fqn string) (*v1.Endpoints, error)
		PodsNamespaces(used []string)
		GetPod(map[string]string) (*v1.Pod, error)
		ListPods() (map[string]v1.Pod, error)
		ListAllPods() (map[string]v1.Pod, error)
		ListNS() (map[string]v1.Namespace, error)
		ListRBs() (map[string]rbacv1.RoleBinding, error)
		ListAllRBs() (map[string]rbacv1.RoleBinding, error)
		ListAllCRBs() (map[string]rbacv1.ClusterRoleBinding, error)
		ListCMs() (map[string]v1.ConfigMap, error)
		ListSecs() (map[string]v1.Secret, error)
		ListSAs() (map[string]v1.ServiceAccount, error)
	}
)

// NewLinter returns a new linter.
func NewLinter(l Loader, log *zerolog.Logger) *Linter {
	return &Linter{Loader: l, log: log, issues: Issues{}}
}

// MaxSeverity scans the lint messages and return the highest severity.
func (l *Linter) MaxSeverity(res string) Level {
	max := OkLevel
	for _, issue := range l.issues[res] {
		if issue.Severity() > max {
			max = issue.Severity()
		}
	}
	return max
}

// NoIssues return true if not lint errors were detected. False otherwize
func (l *Linter) NoIssues(res string) bool {
	return len(l.issues[res]) == 0
}

// Issues returns a collection of linter issues.
func (l *Linter) Issues() Issues {
	return l.issues
}

func (l *Linter) initIssues(res string) {
	l.issues[res] = []Issue{}
}

func (l *Linter) addIssuesMap(res string, issues Issues) {
	for k, v := range issues {
		for _, i := range v {
			l.issues[res] = append(l.issues[res], Error{
				severity:    i.Severity(),
				description: fmt.Sprintf("%s%s%s", k, Delimiter, i.Description()),
			})
		}
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

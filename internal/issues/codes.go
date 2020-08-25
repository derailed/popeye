package issues

import (
	"github.com/derailed/popeye/pkg/config"
	"gopkg.in/yaml.v2"
)

type (
	// Codes represents a collection of sanitizer codes.
	Codes struct {
		Glossary config.Glossary `yaml:"codes"`
	}
)

// LoadCodes retrieves sanitizers codes from yaml file.
func LoadCodes() (*Codes, error) {
	var cc Codes
	if err := yaml.Unmarshal([]byte(codes), &cc); err != nil {
		return &cc, err
	}

	return &cc, nil
}

// Refine overrides code severity based on user input.
func (c *Codes) Refine(gloss config.Glossary) {
	for k, v := range gloss {
		c, ok := c.Glossary[k]
		if !ok {
			continue
		}
		if validSeverity(v.Severity) {
			c.Severity = v.Severity
		}
	}
}

// Helpers...

func validSeverity(l config.Level) bool {
	return l > 0 && l < 4
}

const codes = `
codes:
  # -------------------------------------------------------------------------
  # Container
  100:
    message:  Untagged docker image in use
    severity: 3
  101:
    message:  Image tagged "latest" in use
    severity: 2
  102:
    message:  No probes defined
    severity: 2
  103:
    message:  No liveness probe
    severity: 2
  104:
    message:  No readiness probe
    severity: 2
  105:
    message:  "%s probe uses a port#, prefer a named port"
    severity: 1
  106:
    message:  No resources requests/limits defined
    severity: 2
  107:
    message:  No resource limits defined
    severity: 2
  108:
    message:  "Unnamed port %d"
    severity: 1
  109:
    message:  CPU Current/Request (%s/%s) reached user %d%% threshold (%d%%)
    severity: 2
  110:
    message:  Memory Current/Request (%s/%s) reached user %d%% threshold (%d%%)
    severity: 2
  111:
    message:  CPU Current/Limit (%s/%s) reached user %d%% threshold (%d%%)
    severity: 3
  112:
    message:  Memory Current/Limit (%s/%s) reached user %d%% threshold (%d%%)
    severity: 3

  # -------------------------------------------------------------------------
  # Pod
  200:
    message:  Pod is terminating [%d/%d]
    severity: 2
  201:
    message:  Pod is terminating [%d/%d] %s
    severity: 2
  202:
    message:  Pod is waiting [%d/%d]
    severity: 3
  203:
    message:  Pod is waiting [%d/%d] %s
    severity: 3
  204:
    message:  Pod is not ready [%d/%d]
    severity: 3
  205:
    message:  Pod was restarted (%d) %s
    severity: 2
  206:
    message:  No PodDisruptionBudget defined
    severity: 1
  207:
    message:  Pod is in an unhappy phase (%s)
    severity: 3

  # -------------------------------------------------------------------------
  # Security
  300:
    message:  Using "default" ServiceAccount
    severity: 2
  301:
    message:  Connects to API Server? ServiceAccount token is mounted
    severity: 2
  302:
    message:  Pod could be running as root user. Check SecurityContext/Image
    severity: 2
  303:
    message: Do you mean it? ServiceAccount is automounting APIServer credentials
    severity: 2
  304:
    message: References a secret "%s" which does not exist
    severity: 3
  305:
    message: References a docker-image "%s" pull secret which does not exist
    severity: 3
  306:
    message: Container could be running as root user. Check SecurityContext/Image
    severity: 2

  # -------------------------------------------------------------------------
  # General
  400:
    message:  Used? Unable to locate resource reference
    severity: 1
  401:
    message:  Key "%s" used? Unable to locate key reference
    severity: 1
  402:
    message: No metric-server detected %v
    severity: 1
  403:
    message:  Deprecated %s API group "%s". Use "%s" instead
    severity: 2
  404:
    message:  Deprecation check failed. %v
    severity: 1
  405:
    message:  Is this a jurassic cluster? Might want to upgrade K8s a bit
    severity: 2
  406:
    message:  K8s version OK
    severity: 0

  # -------------------------------------------------------------------------
  # Deployment + StatefulSet
  500:
    message:  Zero scale detected
    severity: 2
  501:
    message:  "Unhealthy %d desired but have %d available"
    severity: 3
  503:
    message:  "At current load, CPU under allocated. Current:%s vs Requested:%s (%s)"
    severity: 2
  504:
    message:  "At current load, CPU over allocated. Current:%s vs Requested:%s (%s)"
    severity: 2
  505:
    message:  "At current load, Memory under allocated. Current:%s vs Requested:%s (%s)"
    severity: 2
  506:
    message:  "At current load, Memory over allocated. Current:%s vs Requested:%s (%s)"
    severity: 2
  507:
    message: "Deployment references ServiceAccount %q which does not exist"
    severity: 3

  # -------------------------------------------------------------------------
  # HPA
  600:
    message:  HPA %s references a Deployment %s which does not exist
    severity: 3
  601:
    message:  HPA %s references a StatefulSet %s which does not exist
    severity: 3
  602:
    message:  Replicas (%d/%d) at burst will match/exceed cluster CPU(%s) capacity by %s
    severity: 2
  603:
    message:  Replicas (%d/%d) at burst will match/exceed cluster memory(%s) capacity by %s
    severity: 2
  604:
    message:  If ALL HPAs triggered, %s will match/exceed cluster CPU(%s) capacity by %s
    severity: 2
  605:
    message:  If ALL HPAs triggered, %s will match/exceed cluster memory(%s) capacity by %s
    severity: 2

  # -------------------------------------------------------------------------
  # Node
  700:
    message:  Found taint "%s" but no pod can tolerate
    severity: 2
  701:
    message:  Node is in an unknown condition
    severity: 3
  702:
    message:  Node is not in ready state
    severity: 3
  703:
    message:  Out of disk space
    severity: 3
  704:
    message:  Insuficient memory
    severity: 2
  705:
    message:  Insuficient disk space
    severity: 2
  706:
    message:  Insuficient PIDS on Node
    severity: 3
  707:
    message:  No network configured on node
    severity: 3
  708:
    message:  No node metrics available
    severity: 1
  709:
    message:  CPU threshold (%d%%) reached %d%%
    severity: 2
  710:
    message:  Memory threshold (%d%%) reached %d%%
    severity: 2
  711:
    message: Scheduling disabled
    severity: 2

  # -------------------------------------------------------------------------
  # Namespace
  800:
    message:  Namespace is inactive
    severity: 3

  # PodDisruptionBudget
  900:
    message:  Used? No pods match selector
    severity: 2
  901:
    message:  MinAvailable (%d) is greater than the number of pods(%d) currently running
    severity: 2

  # -------------------------------------------------------------------------
  # PV/PVC
  1000:
    message:  Available
    severity: 1
  1001:
    message:  Pending volume detected
    severity: 3
  1002:
    message:  Lost volume detected
    severity: 3
  1003:
    message:  Pending claim detected
    severity: 3
  1004:
    message:  Lost claim detected
    severity: 3

  # -------------------------------------------------------------------------
  # Service
  1100:
    message:  No pods match service selector
    severity: 3
  1101:
    message:  Skip ports check. No explicit ports detected on pod %s
    severity: 1
  1102:
    message:  "Use of target port #%s for service port %s. Prefer named port"
    severity: 1
  1103:
    message:  Type Loadbalancer detected. Could be expensive
    severity: 1
  1104:
    message:  Do you mean it? Type NodePort detected
    severity: 1
  1105:
    message:  No associated endpoints
    severity: 3
  1106:
    message:  "No target ports match service port %s"
    severity: 3
  1107:
    message: "LoadBalancer detected but service sets ExternalTrafficPolicy: Cluster"
    severity: 1
  # -------------------------------------------------------------------------
  # ReplicaSet
  1120:
    message:  Unhealthy ReplicaSet %d desired but have %d ready
    severity: 3

  # -------------------------------------------------------------------------
  # NetworkPolicies
  1200:
    message:  No pods match %s pod selector
    severity: 2
  1201:
    message:  No namespaces match %s namespace selector
    severity: 2

  # -------------------------------------------------------------------------
  # RBAC

  1300:
    message:  References a %s (%s) which does not exist
    severity: 2
`

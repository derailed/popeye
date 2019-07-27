package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListIngresses list all included Ingresses.
func ListIngresses(c *k8s.Client, cfg *config.Config) (map[string]*extv1beta1.Ingress, error) {
	ings, err := listAllIngresses(c)
	if err != nil {
		return map[string]*extv1beta1.Ingress{}, err
	}

	res := make(map[string]*extv1beta1.Ingress, len(ings))
	for fqn, ing := range ings {
		if includeNS(c, cfg, ing.Namespace) && !cfg.ShouldExclude("ingress", fqn) {
			res[fqn] = ing
		}
	}

	return res, nil
}

// ListAllIngresses fetch all Ingresses on the cluster.
func listAllIngresses(c *k8s.Client) (map[string]*extv1beta1.Ingress, error) {
	ll, err := fetchIngresses(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	ings := make(map[string]*extv1beta1.Ingress, len(ll.Items))
	for i := range ll.Items {
		ings[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return ings, nil
}

// FetchIngresses retrieves all Ingresses on the cluster.
func fetchIngresses(c *k8s.Client) (*extv1beta1.IngressList, error) {
	return c.DialOrDie().ExtensionsV1beta1().Ingresses(c.ActiveNamespace()).List(metav1.ListOptions{})
}

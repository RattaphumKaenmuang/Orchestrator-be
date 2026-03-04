package decision

import "orchestrator/internal/model"

// Decide returns "k3s" if the instance needs a GPU, otherwise "openstack".
func Decide(instance *model.Instance) string {
	if instance.GPU > 0 {
		return "k3s"
	}
	return "openstack"
}

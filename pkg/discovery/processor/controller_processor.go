package processor

import (
	"strings"

	"github.com/golang/glog"

	"github.com/davecgh/go-spew/spew"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/turbonomic/kubeturbo/pkg/cluster"
	"github.com/turbonomic/kubeturbo/pkg/discovery/repository"
	discoveryutil "github.com/turbonomic/kubeturbo/pkg/discovery/util"
	"github.com/turbonomic/kubeturbo/pkg/util"
)

// Query and cache predefined controllers
var (
	supportedControllers = []schema.GroupVersionResource{
		{
			Group:    util.K8sAPIReplicationControllerGV.Group,
			Version:  util.K8sAPIReplicationControllerGV.Version,
			Resource: util.ReplicationControllerResName,
		},
		{
			Group:    util.K8sAPIReplicasetGV.Group,
			Version:  util.K8sAPIReplicasetGV.Version,
			Resource: util.ReplicaSetResName,
		},
		{
			Group:    util.K8sAPIDeploymentGV.Group,
			Version:  util.K8sAPIDeploymentGV.Version,
			Resource: util.DeploymentResName,
		},
		{
			Group:    util.OpenShiftAPIDeploymentConfigGV.Group,
			Version:  util.OpenShiftAPIDeploymentConfigGV.Version,
			Resource: util.DeploymentConfigResName,
		},
		{
			Group:    util.K8sAPIStatefulsetGV.Group,
			Version:  util.K8sAPIStatefulsetGV.Version,
			Resource: util.StatefulSetResName,
		},
		{
			Group:    util.K8sAPIDaemonsetGV.Group,
			Version:  util.K8sAPIDaemonsetGV.Version,
			Resource: util.DaemonSetResName,
		},
		{
			Group:    util.K8sAPIJobGV.Group,
			Version:  util.K8sAPIJobGV.Version,
			Resource: util.JobResName,
		},
		{
			Group:    util.K8sAPICronJobGV.Group,
			Version:  util.K8sAPICronJobGV.Version,
			Resource: util.CronJobResName,
		},
	}
)

type ControllerProcessor struct {
	ClusterInfoScraper cluster.ClusterScraperInterface
	KubeCluster        *repository.KubeCluster
}

func NewControllerProcessor(clusterInfoScraper cluster.ClusterScraperInterface,
	kubeCluster *repository.KubeCluster) *ControllerProcessor {
	return &ControllerProcessor{
		ClusterInfoScraper: clusterInfoScraper,
		KubeCluster:        kubeCluster,
	}
}

func (cp *ControllerProcessor) ProcessControllers() {
	cp.cacheAllControllers()
}

func (cp *ControllerProcessor) cacheAllControllers() {
	scs := spew.ConfigState{
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		Indent:                  "  ",
		SortKeys:                true,
	}
	controllerMap := make(map[string]*repository.K8sController)
	for _, controller := range supportedControllers {
		list, err := cp.ClusterInfoScraper.GetResources(controller)
		if err != nil {
			if apierrors.IsNotFound(err) && strings.Contains(err.Error(), "the server could not find the requested resource") {
				glog.V(3).Infof("Resource %v not found ", controller.Resource)
			} else {
				glog.Errorf("Failed to list workload controller for %v", controller.Resource)
			}
			continue
		}
		for _, item := range list.Items {
			uid := string(item.GetUID())
			kind := item.GetKind()
			name := item.GetName()
			namespace := item.GetNamespace()
			containerNames, err := discoveryutil.GetContainerNames(&item)
			if err != nil {
				glog.Warningf("Could not find containers in %s %s/%s: %s", kind, namespace, name, err)
			}
			// insert into the map
			k8sController := repository.
				NewK8sController(kind, name, namespace, uid).
				WithLabels(item.GetLabels()).
				WithAnnotations(item.GetAnnotations()).
				WithOwnerReferences(item.GetOwnerReferences()).
				WithContainerNames(containerNames)
			replicas, found, err := unstructured.NestedInt64(item.Object, "spec", "replicas")
			if err != nil {
				glog.Warningf("The spec.replicas of %s %s/%s is not an integer.", kind, namespace, name)
			} else if found {
				k8sController.WithReplicas(replicas)
			}
			if kind == util.KindDaemonSet {
				// For daemonset controller, set the replicas as the number of nodes in the cluster
				k8sController.WithReplicas(int64(len(cp.KubeCluster.Nodes)))
			}
			controllerMap[uid] = k8sController
			glog.V(3).Infof("Discovered %s %s/%s %s.", kind, namespace, name, uid)
			glog.V(4).Infof("%+v", scs.Sdump(k8sController))
		}
	}
	cp.KubeCluster.ControllerMap = controllerMap
}

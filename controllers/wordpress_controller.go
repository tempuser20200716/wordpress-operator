package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1 "github.com/example-inc/memcached-operator/api/v1"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=wordpresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

func (r *WordpressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("wordpress", req.NamespacedName)

	// Fetch the Wordpress instance
	wordpress := &cachev1.Wordpress{}
	err := r.Get(ctx, req.NamespacedName, wordpress)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Wordpress resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Wordpress")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name, Namespace: wordpress.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForWordpress(wordpress)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Check if the mysql deployment already exists, if not create a new one
	found2 := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name + "-mysql", Namespace: wordpress.Namespace}, found2)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMySQL(wordpress)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name+"-mysql")
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name+"-mysql")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	found3 := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name, Namespace: wordpress.Namespace}, found3)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.serviceForWordpress(wordpress)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Check if the mysql service already exists, if not create a new one
	found4 := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name + "-mysql", Namespace: wordpress.Namespace}, found4)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc2 := r.serviceForMySQL(wordpress)
		log.Info("Creating a new Service", "Service.Namespace", svc2.Namespace, "Service.Name", svc2.Name+"-mysql")
		err = r.Create(ctx, svc2)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc2.Namespace, "Service.Name", svc2.Name+"-mysql")
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Check if the PVC already exists, if not create a new one
	found5 := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name + "-pvc", Namespace: wordpress.Namespace}, found5)
	if err != nil && errors.IsNotFound(err) {
		// Define a new PVC
		pvc := r.wordpressClaim(wordpress, "fast")
		log.Info("Creating a new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := wordpress.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Wordpress status with the pod names
	// List the pods for this wordpress's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(wordpress.Namespace),
		client.MatchingLabels(labelsForWordpress(wordpress.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Wordpress.Namespace", wordpress.Namespace, "Wordpress.Name", wordpress.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, wordpress.Status.Nodes) {
		wordpress.Status.Nodes = podNames
		err := r.Status().Update(ctx, wordpress)
		if err != nil {
			log.Error(err, "Failed to update Wordpress status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForWordpress returns a wordpress Deployment object
func (r *WordpressReconciler) deploymentForWordpress(m *cachev1.Wordpress) *appsv1.Deployment {
	ls := labelsForWordpress(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "wordpress",
							Image: "wordpress:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "wordpress",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 31735,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      m.Name,
									MountPath: "/var/www/html",
								}},
						}},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: m.Name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: m.Name + "-pvc",
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Wordpress instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// deploymentForMySQL returns a mysql Deployment object
func (r *WordpressReconciler) deploymentForMySQL(m *cachev1.Wordpress) *appsv1.Deployment {
	ls := labelsForWordpress(m.Name + "-mysql")
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-mysql",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "mysql:latest",
							Name:  "mysql",
							//Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mysql",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3306,
								}},
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: m.Spec.SQLRootPassword,
								}},
						}},
				},
			},
		},
	}
	// Set Wordpress instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// serviceForWordpress returns a wordpress Service object
func (r *WordpressReconciler) serviceForWordpress(m *cachev1.Wordpress) *corev1.Service {
	ls := labelsForWordpress(m.Name)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     "LoadBalancer",
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Name:       "wordpress",
				Port:       31735,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}

	// Set Wordpress instance as the owner and controller
	ctrl.SetControllerReference(m, svc, r.Scheme)
	return svc
}

// serviceForMySQL returns a mysql Service object
func (r *WordpressReconciler) serviceForMySQL(m *cachev1.Wordpress) *corev1.Service {
	ls := labelsForWordpress(m.Name + "-mysql")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-mysql",
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Name: "mysql",
				Port: 3306,
			}},
		},
	}

	// Set Wordpress instance as the owner and controller
	ctrl.SetControllerReference(m, svc, r.Scheme)
	return svc
}

func (r *WordpressReconciler) wordpressClaim(m *cachev1.Wordpress, storageclassName string) *v1.PersistentVolumeClaim {
	ls := labelsForWordpress(m.Name)

	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.Name + "-pvc",
			Namespace:       m.Namespace,
			ResourceVersion: "0",
			SelfLink:        "/api/v1/namespaces/" + m.Namespace + "/persistentvolumeclaims/" + m.Name,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
			//VolumeName:       m.Name + "-volume",
			StorageClassName: &storageclassName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimPending,
		},
	}
	return claim
}

func newClaim(name, claimUID, provisioner, volumeName, storageclassName string, annotations map[string]string) *v1.PersistentVolumeClaim {
	claim := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       v1.NamespaceDefault,
			UID:             types.UID(claimUID),
			ResourceVersion: "0",
			SelfLink:        "/api/v1/namespaces/" + v1.NamespaceDefault + "/persistentvolumeclaims/" + name,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce, v1.ReadOnlyMany},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Mi"),
				},
			},
			VolumeName:       volumeName,
			StorageClassName: &storageclassName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimPending,
		},
	}
	for k, v := range annotations {
		claim.Annotations[k] = v
	}
	return claim
}

// labelsForWordpress returns the labels for selecting the resources
// belonging to the given wordpress CR name.
func labelsForWordpress(name string) map[string]string {
	return map[string]string{"app": "wordpress", "wordpress_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1.Wordpress{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

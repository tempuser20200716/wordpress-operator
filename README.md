# wordpress-operator

Some of the links in the assignment page did not resolve, so I started with

https://sdk.operatorframework.io/docs/golang/quickstart/
and
https://opensource.com/article/20/3/kubernetes-operator-sdk

Throughout my work on the project, I also referred to various resources including blog posts, GoDocs, Kubernetes documentation and various Github repositories.

Initially, I set up my environment on a Windows 10 PC with minikube, using Ubuntu WSL and VSCode. However, after working in that environment for a while, I moved to a CentOS7 VM with Kubernetes 1.18.5 due to issues building the Operator in the WSL environment. I also had some issues using the tools required by the Operator SDK with Go 1.13, so I built the project using Go 1.14.4.

The attached tar.gz contains the wordpress-operator project that will deploy an operator that will accept the following configuration file (located at config/samples/v1_wordpress.yaml):
```
apiVersion: example.com/v1
kind: Wordpress
metadata:
  name: wordpress-sample
spec:
  # Add fields here
  size: 1
  sqlRootPassword: plaintextpassword
  ```
From this, the operator will create two deployments, two services and two pods using the desired containers.
I currently have one container partially configured to create and bind to a PVC provided by the StorageOS Operator (using the StorageClass name ‘fast’), however, there does seem to be an issue with the binding such that the pod and the PVC remain in a ‘Pending’ state. The relevant logs (viewed with kubernetes logs wordpress-operator-controller-manager- manager) show some errors, but I was unable to troubleshoot these issues in the time I specified for getting back to you. Also, while the provided password is passed to the MySQL pod via an environment variable, I did not have time to store that value in a created secret.

For using the provided code, the commands needed will be:

`make docker-build IMG=quay.io/example/wordpress-operator:v0.0.1`

to build the wordpress-operator container

`make deploy IMG=quay.io/example/wordpress-operator:v0.0.1`

to deploy the operator

`kubectl apply -f config/samples/v1_wordpress.yaml`

to deploy the custom resource

Output should be similar to:
```
    [tjm@devbox wordpress-operator]$ kubectl get deployment,service,pvc,secret
NAME                                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/wordpress-operator-controller-manager   1/1     1            1           55m
deployment.apps/wordpress-sample                        0/1     1            0           55m
deployment.apps/wordpress-sample-mysql                  1/1     1            1           55m

NAME                                                            TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)           AGE
service/kubernetes                                              ClusterIP      10.96.0.1        <none>        443/TCP           14h
service/wordpress-operator-controller-manager-metrics-service   ClusterIP      10.110.49.252    <none>        8443/TCP          55m
service/wordpress-sample                                        LoadBalancer   10.101.186.11    <pending>     31735:30817/TCP   54m
service/wordpress-sample-mysql                                  ClusterIP      10.97.254.132    <none>        3306/TCP          54m

NAME                                         STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/wordpress-sample-pvc   Pending                                      fast           54m

NAME                         TYPE                                  DATA   AGE
secret/default-token-tgdhk   kubernetes.io/service-account-token   3      14h
```
Cleanup can be done as follows:

`kubectl delete -f config/samples/v1_wordpress.yaml`

to delete the CR deployment

`kubectl delete deployments,service -l control-plane=controller-manager`

to uninstall the operator

`kubectl delete role,rolebinding --all`

to clean up the remaining RBAC

With more time, I would, first of course finish the desired behavior. After that, the code could certainly use more commenting, cleanup and an overall better implementation of how the resources are deployed. I also found many more resources for using a more ‘template focused’ implementation, but my understanding was that the wordpress-operator should primarily utilize the Go libraries rather than executing YAML templates, so I tried to stick with that approach as much as possible.

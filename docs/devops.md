# Set Up DevOps Environment

DevOps is recommended to use for this project. Please follow the instructions below to set up your environment. We use Jenkins with Blue Ocean plugin and deploy it on Kubernetes, also continuously deploy KubeSphere on the Kubernetes cluster.  

----

- [Create Kubernetes Cluster](#create-kubernetes-cluster)
- [Deploy Jenkins](#deploy-jenkins)
- [Configure Jenkins](#configure-jenkins)
- [Create a Pipeline](#create-a-pipeline)

## Create Kubernetes Cluster

We are using [Kubernetes on QingCloud](https://appcenter.qingcloud.com/apps/app-u0llx5j8) to create a kubernetes production environment by one click. Please follow the [instructions](https://appcenter-docs.qingcloud.com/user-guide/apps/docs/kubernetes/) to create your own cluster. Access the Kubernetes client using one of the following options.
  - **Open VPN**<a id="openvpn"></a>: Go to the left navigation tree of the [QingCloud console](https://console.qingcloud.com), choose _Networks & CDN_, then _VPC Networks_; on the content of the VPC page, choose _Management Configuration_, _VPN Service_, then you will find _Open VPN_ service. Here is the [screenshot](images/openvpn.png) of the page.
  - **Port Forwarding**<a id="port-forwarding"></a>: same as Open VPN, but choose _Port Forwarding_ on the content of VPC page instead of VPN Service; and add a rule to forward source port to the port of ssh port of the kubernetes client, for instance, forward 10007 to 22 of the kubernetes client with the private IP being 192.168.100.7. After that, you need to open the firewall to allow the port 10007 accessible from outside. Please click the _Security Group_ ID on the same page of the VPC, and add the downstream rule for the firewall.
  - **VNC**: If you don't want to access the client node remotely, just go to the kubernetes cluster detailed page on the [QingCloud console](https://console.qingcloud.com), and click the windows icon aside of the client ID shown as the [screenshot](images/kubernets.png) (user/password: root/k8s). The way is not recommended, however you can check kubernetes quickly using VNC since you don't configure anything. 

## Deploy Jenkins

1. Copy the [yaml file](../devops/kubernetes/jenkins-qingcloud.yaml) to the kubernetes client, and deploy
   ```
   # kubectl apply -f jenkins-qingcloud.yaml
   ```

2. Access Jenkins console by opening http://\<ip\>:9200 where ip depends on how you expose the Jenkins service to outside explained below. (You can find your way to access Jenkins console such as ingress, cloud LB etc.) On the kubernetes client
   ```
   # iptables -t nat -A PREROUTING -p tcp -i eth0 --dport 9200 -j DNAT --to-destination "$(kubectl get svc -n jenkins --selector=app=jenkins -o jsonpath='{.items..spec.clusterIP}')":9200
   # iptables -t nat -A POSTROUTING -p tcp --dport 9200 -j MASQUERADE
   # sysctl -w net.ipv4.conf.eth0.route_localnet=1
   ```

3. Now the request to the kubernetes client port 9200 will be forwarded to the Jenkins service. 
   
   - If you use [Open VPN](#openvpn) to access the kubernetes client, then open http://\<kubernetes client private ip\>:9200 to access Jenkins console. 
   - If you use [Port Forwarding](#port-forwarding) to access the client, then forward the VPC port 9200 to the kubernetes client port 9200. Now open http://\<VPC EIP\>:9200 to access Jenkins console.

## Configure Jenkins
   > You can refer [jenkins.io](https://jenkins.io/doc/tutorials/using-jenkins-to-build-a-java-maven-project/) about how to configure Jenkins and create a pipeline.

1. Unlock Jenkins

   - Get the Adminstrator password from the log on the kubernetes client
     ```
     # kubectl logs "$(kubectl get pods -n jenkins --selector=app=jenkins -o jsonpath='{.items..metadata.name}')" -c jenkins -n jenkins
     ```
   - Go to Jenkins console, paste the password and continue. Install suggested plugins, then create the first admin user and save & finish.

2. Configure Jenkins
   
   We will deploy KubeSphere application into the same Kubernetes cluster as the one that the Jenkins is running on. So we need configure the Jenkins pod to access the Kubernetes cluster, and log in docker registry given that during the [Jenkins pipeline](#create-a-pipeline) we push KubeSphere image into a registry which you can change on your own. 
  
   On the Kubernetes client, execute the following to log in Jenkins container.
  
     ```
     # kubectl exec -it "$(kubectl get pods -n jenkins --selector=app=jenkins -o jsonpath='{.items..metadata.name}')" -c jenkins -n jenkins -- /bin/bash
     ```
  
     After logging in the Jenkins container, then run the following to log in docker registry and prepare folder to hold kubectl configuration.

     ```
     bash-4.3# docker login -u xxx -p xxxx
     bash-4.3# mkdir /root/.kube
     bash-4.3# exit
     ```
  
     Once back again to the Kubernetes client, run the following to copy the tool kubectl and its configuration from the client to the Jenkins container.

     ```
     # kubectl cp /usr/bin/kubectl jenkins/"$(kubectl get pods -n jenkins --selector=app=jenkins -o jsonpath='{.items..metadata.name}')":/usr/bin/kubectl -c jenkins
     # kubectl cp /root/.kube/config jenkins/"$(kubectl get pods -n jenkins --selector=app=jenkins -o jsonpath='{.items..metadata.name}')":/root/.kube/config -c jenkins
     ```  

## Create a pipeline
  - Fork KubeSphere from github for your development. You need to change the docker repository to your own in the files [kubesphere.yaml](devops/kubernetes/kubesphere.yaml), [build-images.sh](devops/scripts/build-images.sh), [push-images.sh](devops/scripts/push-images.sh) and [clean.sh](devops/scripts/clean.sh).
  - On the Jenkins panel, click _Open Blue Ocean_ and start to create a new pipeline. Choose _GitHub_, paste your access key of GitHub, select the repository you want to create a CI/CD pipeline. We already created the pipeline Jenkinsfile on the upstream repository which includes compiling KubeSphere, building images, push images, deploying the application, verifying the application and cleaning up.
  - It is better to configure one more thing. On the Jenkins panel, go to the configuration of KubeSphere, check _Periodically if not otherwise run_ under _Scan Repository Triggers_ and select the interval at your will. 
  - If your repository is an upstream, you can select _Discover pull requests from forks_ under _Behaviors_ so that the pipeline will work for PR before merged.
  - Now it is good to go. Whenever you commit a change to your forked repository, the pipeline will work during the Jenkins trigger interval. 

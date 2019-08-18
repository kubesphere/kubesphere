# 安装kubesphere的常见问题
内容主要是关于安装过程中遇到的问题及解决方法，也很欢迎大家遇到的问题或者已解决了问题在这里留下痕迹。


* 1、遇到ansible指令不能用，Command "python setup.py egg_info" failed with error code 1 in/tmp/pip.build.344f90/ansible
```
方法1：pip install setuptools==33.1.1
方法2：下载新版setuptools，
wget https://files.pythonhosted.org/packages/e5/53/92a8ac9d252ec170d9197dcf988f07e02305a06078d7e83a41ba4e3ed65b/setuptools-33.1.1-py2.py3-none-any.whl
pip install setuptools-33.1.1-py2.py3-none-any.whl
```
* 2、dial tcp 20.233.0.1: 443: connect: no route to host"], "stdout": "", "stdout_lines": []}
检查下是不是机器防火墙还开着 建议关掉 systemctl stop firewalld

* 3、FAELED - RETRYING: KubeSphere Waiting for ks-console (30 retries left)一直卡在这里。
```
检查cpu和内存，目前单机版部署至少8核16G
当free -m 时buff/cache占用内存资源多，执行echo 3 > /proc/sys/vm/drop_caches 指令释放下内存
```
* 4、kubectl get pod -n kubesphere-system出现大量的Pending状态时
可以kubectl describe看下这几个没起来的pod，应该是内存不足

* 5、kubesphere v2.0.2 离线安装失败，提示Failed to create 'IPPool' resource: resource already exists: IPPool(default-pool)" 
```
机器目前需要干净的机器，如果已安装了k8s，则可以参考如下链接安装kubesphere：https://github.com/kubesphere/ks-installer
```
* 6、centos7.4 2.0.2离线安装后，reboot系统重启服务不恢复
```
kube-system下的 coredns 没有起来，kubectl describe看下coredns的错误日志;
coredns 无法连接 apiserver，可以执行 ipvsadm -Ln 看下 6443 端口是否对应主节点地址
```
* 7、kubesphere能否独立使用
https://github.com/kubesphere/ks-installer

* 8、centos7.6  file:///kubeinstaller/yum_repo/iso/repodata/repomd.xml: [Errno 14] curl#37 - "Couldn't open file /kubeinstaller/yum_repo/iso/repodata/repomd.xml"
Trying other mirror.
```
centos7.6系统还不支持离线下载部署。可以用centos7.4或7.5或在线部署。
如果环境可以联网，建议使用在线安装，支持centos7.6
链接：https://pan.baidu.com/s/1HBlBaD69mx6z050sSgj0Kg
提取码：eqe6
把该iso放到离线环境的kubesphere-all-offline-advanced-2.0.2/Repos/下，umount /kubeinstaller/yum_repo/iso重新跑
```
* 9、离线安装失败转为在线安装
刚才执行离线安装修改的yum源，先还原一下 ，/etc/yum.repo.d 有备份

* 10、Could not find a version that satisfies the requirement ansible==2.7.6 (from -r /home/sunhaizhou/all-in-one/kubesphere-all-offline-advanced-2.0.2/scripts/os/requirements.txt (line 1)) (from versions: )
Error: Install the pip packages failed!
```
方法1：pip install ansible==2.7.6
方法2：可以手动umount一下/kubeinstaller/pip_repo/pip27/iso 或者重启机器 应该可以解决该问题
```
* 11、fatal: [ks-allinone]: FAILED! => {"ansible_facts": {"pkg_mgr": "yum"}, "changed": false, "msg": "The Python 2 bindings for rpm are needed for this module. If you require Python 3 support use the dnf Ansible module instead.. The Python 2 yum module is needed for this module. If you require Python 3 support use the dnf Ansible module instead."}
将机器的python3改成python2
* 12、KubeSphere 2.0.1 安装失败 ：no matches for kind \"S2iBuilderTemplate\
```
kubectl get pvc --all-namespaces 看下pvc是否处于pending状态
如果pvc pending，地址和路径也都配置正确的话，可以试下本地机器上是否可以挂载；如果是自建的nfs，可以看下nfs的配置参数。
先执行uninstall, 然后再install。
```
* 13、fatal: [hippo04kf]: FAILED! => {"changed": true, "msg": "non-zero return code", "rc": 5, "stderr": "Warning: Permanently added '99.13.XX.XX' (ECDSA) to the list of known hosts.\r\nPermission denied, please try again(publickey,gssapi-keyex,gssapi-with-mic).\r\n", "stderr_lines": ["Warning: Permanently added '99.13.XX.XX' (ECDSA) to the list of known hosts."
```
只贴了all的格式，ansible_ssh_pass改成ansible_become_pass，最后执行脚本时，需要转成root用户执行。
[all]
master ansible_connection=local ip=192.168.0.5 ansible_user=tester ansible_become_pass=@@@@@@
node1 ansible_host=192.168.0.6 ip=192.168.0.6 ansible_user=tester ansible_become_pass=@@@@@@
```
* 14 、ks-devops/jenkins unable to parse quantity's suffix,error found in #10 byte of
```
有特殊字符无法解析 如果编辑过jenkins的相关配置的话可以检查下是不是写入了特殊字符
```
* 15、kubesphere-all-offline-advanced-2.0.0 离线安装提示 Failed connect to 192.168.10.137:5080; Connection refused
```
docker ps -a|grep nginx     nginx是否正常启动，且5080端口已监听。
df -hT|grep -v docker|grep -v kubelet    看pip和iso是否都已挂载。
以上都有问题的话，机器先停止运行，然后执行uninstall脚本，再运行install脚本。
```
* 16、failed ansibelundefinedvariable dict object has no attribute address
```
第一行这样配试下，ip换成自己的，如果有网的话建议联网安装
ks-allinone ansible_connection=local ip=192.168.0.2
[local-registry]
ks-allinone
[kube-master]
ks-allinone
[etcd]
ks-allinone
[kube-node]
ks-allinone
[k8s-cluster:children]
kube-node
kube-master
```
* 17、FAILED! => {"reason": "'delegate_to' is not a valid attribute for a TaskInclude
```
不要手动安装ansible，如果安装过程中出现了ERROR: Command "python setup.py egg_info" failed with error code 1 in /tmp/pip-install-voTIRk/ansible/
./multi-node.sh: line 41: ansible-playbook: command not found
pip install --upgrade setuptools==30.1.0先执行这个再安装
```
* 18、centos7.6 FAILED - RETRYING: KubeSphere| Installing JQ (YUM) (5 retries left),没有jq这个package，我下了jq的二进制包放在/usr/bin目录下，jq可以使用。
```
可以在kubesphere/roles/prepare/nodes/tasks/main.yaml中注释掉安装jq的相关tasks
```
* 19、FAILED - RETRYING: ks-alerting | Waiting for alerting-db-init (2 retries left).fatal: [ks-allinone]: FAILED! => {"attempts": 3, "changed": true, "cmd": "/usr/local/bin/kubectl -n kubesphere-alerting-system get pod | grep alerting-db-init | awk '{print $3}'", "delta":  "stdout": "Init:0/1"
```
检查配置是否满足要求8核cpu、16G。检查存储相关配置。
```
* 20、2.0.2 版本安装后配置docker 私有库问题，配置了一个 docker 私有库，修改完 /etc/docker/daemon.json 文件，然后重新启动 docker报错。
```
kubesphere 在安装时指定了 DOCKER_OPTS= --insecure-registry ，所以不能再加载 /etc/docker/daemon.json 的 insecure-registries 设置，
如果要添加的，/etc/systemd/system/docker.service.d文件中添加，重启。
```
* 21、用的自己搭建的nfs服务器，发现创建的pvc全是pending导致了相关的pod启动不起来。
```
可以参考 *(rw,insecure,sync,no_subtree_check,no_root_squash)这个配置下nfs，然后先试下主机上是否可以挂载nfs
```
* 22、ubuntu系统，非root用户安装时，在kubernetes-apps/network_plugin/calico: start calico resources get http://localhost:8080/api?timeout=32s: dial tcp 127.0.0.1: 8080: connect: connection refused
采用root用户安装脚本。
* 23、FAILED - RETRYING: container_download | Download containers if pull is required or told to always pull (all nodes) (4 retries left)
由于到dockerhub网络不是特别好造成的。
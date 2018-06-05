Name: kubernetes-cni
Version: OVERRIDE_THIS
Release: 00
License: ASL 2.0
Summary: Container Cluster Manager - CNI plugins

URL: https://kubernetes.io

%description
Binaries required to provision container networking.

%prep
mkdir -p ./bin
tar -C ./bin -xz -f {cni-plugins-amd64-v0.6.0.tgz}

%install

install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}/opt/cni
mv bin/ %{buildroot}/opt/cni/

%files
/opt/cni
%{_sysconfdir}/cni/net.d/

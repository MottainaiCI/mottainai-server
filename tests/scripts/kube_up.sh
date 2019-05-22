#!/bin/bash
set -ex

sudo ACCEPT_LICENSE=* equo i git
curl -L https://git.io/get_helm.sh | sudo bash
set +e
curl -sfL https://get.k3s.io | sudo sh -
sudo k3s server &
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
sleep 30
sudo k3s kubectl get node
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
helm init --service-account tiller --upgrade
sleep 30

set -e
helm repo add mottainai 'https://raw.githubusercontent.com/MottainaiCI/helm-repo/master/'
helm repo update


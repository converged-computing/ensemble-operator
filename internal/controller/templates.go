package controller

var (
	endPost = `# Finalize the view so we can use Python (not default for a sidecar)
cp -R /mnt/flux/software /opt/software
git clone https://github.com/converged-computing/flux-metrics-api /opt/metrics-api
cd /opt/metrics-api
source /mnt/flux/flux-view.sh
# Note that this is the python in the view
/mnt/flux/view/bin/python3.11 -m ensurepip
/mnt/flux/view/bin/python3.11 -m pip install .
export FLUX_URI=$fluxsocket
unset LD_LIBRARY_PATH PYTHONPATH
flux-metrics-api start --port 5000 --namespace default --host 0.0.0.0
else
 sleep infinity
fi
`
	rockyLinuxPostTemplate = `
if [[ ${JOB_COMPLETION_INDEX} -eq 0 ]]; then
dnf update && dnf install -y git
` + endPost

	ubuntuPostTemplate = `
if [[ ${JOB_COMPLETION_INDEX} -eq 0 ]]; then
apt-get update && apt-get install -y git
` + endPost
)

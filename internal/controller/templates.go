package controller

// This is the same for ubuntu and rocky
var (
	postCommand = `if [[ ${JOB_COMPLETION_INDEX} -eq 0 ]]; then
# Finalize the view so we can use Python (not default for a sidecar)
cp -R /mnt/flux/software /opt/software
ls
source /mnt/flux/flux-view.sh
# Note that this is the python in the view
/mnt/flux/view/bin/python3.11 -m ensurepip
/mnt/flux/view/bin/python3.11 -m pip install .
export FLUX_URI=$fluxsocket
unset LD_LIBRARY_PATH PYTHONPATH
# Note these are default
ensemble-api start --port %s --workers %d
else
 sleep infinity
fi
`
)

from .workload_demand import WorkloadDemand


def get_algorithm(name, options=None):
    """
    Get a named backend.
    """
    options = options or {}
    name = name.lower()
    if name == "workload-demand":
        return WorkloadDemand(**options)
    raise ValueError(f"Algorithm {name} is not known")

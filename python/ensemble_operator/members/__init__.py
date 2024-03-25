from .minicluster import MiniClusterMember


def get_member(name, options=None):
    """
    Get a named member type
    """
    options = options or {}
    name = name.lower()
    if name == "minicluster":
        return MiniClusterMember(**options)
    raise ValueError(f"Member type {name} is not known")

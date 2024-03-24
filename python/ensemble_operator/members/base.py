class MemberBase:
    """
    The MemberBase is an abstract base to show functions defined.
    """

    def __init__(self, **options):
        """
        Create a new member type (e.g., minicluster)
        """
        # Set options as attributes
        for key, value in options.items():
            setattr(self, key, value)

    def count_inactive(self, *args, **kwargs):
        pass

    def status(self, *args, **kwargs):
        """
        Ask the member for a status
        """
        raise NotImplementedError

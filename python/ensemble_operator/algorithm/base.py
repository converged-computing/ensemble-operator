# An algorithm is a named class that can receive the payload from
# a StatusRequest, and optionally act on the queue (with the action)
# depending on the algorithm in question.


class AlgorithmBase:
    """
    The AlgorithmBase is an abstract base to show functions defined.
    """

    def submit(self, *args, **kwargs):
        """
        The submit action will submit one or more jobs.
        """
        raise NotImplementedError

    def run(self, *args, **kwargs):
        """
        Receive the payload from the Ensemble Operator and take action.
        """
        raise NotImplementedError

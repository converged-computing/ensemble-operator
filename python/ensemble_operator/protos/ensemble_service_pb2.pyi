from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class StatusRequest(_message.Message):
    __slots__ = ("name", "secret")
    NAME_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    name: str
    secret: str
    def __init__(self, name: _Optional[str] = ..., secret: _Optional[str] = ...) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("payload", "status")
    class ResultType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        UNSPECIFIED: _ClassVar[StatusResponse.ResultType]
        SUCCESS: _ClassVar[StatusResponse.ResultType]
        ERROR: _ClassVar[StatusResponse.ResultType]
        DENIED: _ClassVar[StatusResponse.ResultType]
        EXISTS: _ClassVar[StatusResponse.ResultType]
    UNSPECIFIED: StatusResponse.ResultType
    SUCCESS: StatusResponse.ResultType
    ERROR: StatusResponse.ResultType
    DENIED: StatusResponse.ResultType
    EXISTS: StatusResponse.ResultType
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    payload: str
    status: StatusResponse.ResultType
    def __init__(self, payload: _Optional[str] = ..., status: _Optional[_Union[StatusResponse.ResultType, str]] = ...) -> None: ...

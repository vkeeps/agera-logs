// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.28.2
// source: proto/log.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogRequest struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Schema            string                 `protobuf:"bytes,1,opt,name=schema,proto3" json:"schema,omitempty"`
	Module            string                 `protobuf:"bytes,2,opt,name=module,proto3" json:"module,omitempty"`
	Output            string                 `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
	Detail            string                 `protobuf:"bytes,4,opt,name=detail,proto3" json:"detail,omitempty"`
	ErrorInfo         string                 `protobuf:"bytes,5,opt,name=error_info,json=errorInfo,proto3" json:"error_info,omitempty"`
	Service           string                 `protobuf:"bytes,6,opt,name=service,proto3" json:"service,omitempty"`
	OperatorID        string                 `protobuf:"bytes,7,opt,name=operatorID,proto3" json:"operatorID,omitempty"`
	Operator          string                 `protobuf:"bytes,8,opt,name=operator,proto3" json:"operator,omitempty"`
	OperatorIP        string                 `protobuf:"bytes,9,opt,name=operatorIP,proto3" json:"operatorIP,omitempty"`
	OperatorEquipment string                 `protobuf:"bytes,10,opt,name=operator_equipment,json=operatorEquipment,proto3" json:"operator_equipment,omitempty"`
	OperatorCompany   string                 `protobuf:"bytes,11,opt,name=operator_company,json=operatorCompany,proto3" json:"operator_company,omitempty"`
	OperatorProject   string                 `protobuf:"bytes,12,opt,name=operator_project,json=operatorProject,proto3" json:"operator_project,omitempty"`
	LogLevel          string                 `protobuf:"bytes,13,opt,name=log_level,json=logLevel,proto3" json:"log_level,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *LogRequest) Reset() {
	*x = LogRequest{}
	mi := &file_proto_log_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogRequest) ProtoMessage() {}

func (x *LogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_log_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogRequest.ProtoReflect.Descriptor instead.
func (*LogRequest) Descriptor() ([]byte, []int) {
	return file_proto_log_proto_rawDescGZIP(), []int{0}
}

func (x *LogRequest) GetSchema() string {
	if x != nil {
		return x.Schema
	}
	return ""
}

func (x *LogRequest) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *LogRequest) GetOutput() string {
	if x != nil {
		return x.Output
	}
	return ""
}

func (x *LogRequest) GetDetail() string {
	if x != nil {
		return x.Detail
	}
	return ""
}

func (x *LogRequest) GetErrorInfo() string {
	if x != nil {
		return x.ErrorInfo
	}
	return ""
}

func (x *LogRequest) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *LogRequest) GetOperatorID() string {
	if x != nil {
		return x.OperatorID
	}
	return ""
}

func (x *LogRequest) GetOperator() string {
	if x != nil {
		return x.Operator
	}
	return ""
}

func (x *LogRequest) GetOperatorIP() string {
	if x != nil {
		return x.OperatorIP
	}
	return ""
}

func (x *LogRequest) GetOperatorEquipment() string {
	if x != nil {
		return x.OperatorEquipment
	}
	return ""
}

func (x *LogRequest) GetOperatorCompany() string {
	if x != nil {
		return x.OperatorCompany
	}
	return ""
}

func (x *LogRequest) GetOperatorProject() string {
	if x != nil {
		return x.OperatorProject
	}
	return ""
}

func (x *LogRequest) GetLogLevel() string {
	if x != nil {
		return x.LogLevel
	}
	return ""
}

type LogResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LogResponse) Reset() {
	*x = LogResponse{}
	mi := &file_proto_log_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogResponse) ProtoMessage() {}

func (x *LogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_log_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogResponse.ProtoReflect.Descriptor instead.
func (*LogResponse) Descriptor() ([]byte, []int) {
	return file_proto_log_proto_rawDescGZIP(), []int{1}
}

func (x *LogResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_proto_log_proto protoreflect.FileDescriptor

const file_proto_log_proto_rawDesc = "" +
	"\n" +
	"\x0fproto/log.proto\x12\x05proto\"\xa3\x03\n" +
	"\n" +
	"LogRequest\x12\x16\n" +
	"\x06schema\x18\x01 \x01(\tR\x06schema\x12\x16\n" +
	"\x06module\x18\x02 \x01(\tR\x06module\x12\x16\n" +
	"\x06output\x18\x03 \x01(\tR\x06output\x12\x16\n" +
	"\x06detail\x18\x04 \x01(\tR\x06detail\x12\x1d\n" +
	"\n" +
	"error_info\x18\x05 \x01(\tR\terrorInfo\x12\x18\n" +
	"\aservice\x18\x06 \x01(\tR\aservice\x12\x1e\n" +
	"\n" +
	"operatorID\x18\a \x01(\tR\n" +
	"operatorID\x12\x1a\n" +
	"\boperator\x18\b \x01(\tR\boperator\x12\x1e\n" +
	"\n" +
	"operatorIP\x18\t \x01(\tR\n" +
	"operatorIP\x12-\n" +
	"\x12operator_equipment\x18\n" +
	" \x01(\tR\x11operatorEquipment\x12)\n" +
	"\x10operator_company\x18\v \x01(\tR\x0foperatorCompany\x12)\n" +
	"\x10operator_project\x18\f \x01(\tR\x0foperatorProject\x12\x1b\n" +
	"\tlog_level\x18\r \x01(\tR\blogLevel\"'\n" +
	"\vLogResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess2>\n" +
	"\n" +
	"LogService\x120\n" +
	"\aSendLog\x12\x11.proto.LogRequest\x1a\x12.proto.LogResponseB$Z\"github.com/vkeeps/agera-logs/protob\x06proto3"

var (
	file_proto_log_proto_rawDescOnce sync.Once
	file_proto_log_proto_rawDescData []byte
)

func file_proto_log_proto_rawDescGZIP() []byte {
	file_proto_log_proto_rawDescOnce.Do(func() {
		file_proto_log_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_log_proto_rawDesc), len(file_proto_log_proto_rawDesc)))
	})
	return file_proto_log_proto_rawDescData
}

var file_proto_log_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_log_proto_goTypes = []any{
	(*LogRequest)(nil),  // 0: proto.LogRequest
	(*LogResponse)(nil), // 1: proto.LogResponse
}
var file_proto_log_proto_depIdxs = []int32{
	0, // 0: proto.LogService.SendLog:input_type -> proto.LogRequest
	1, // 1: proto.LogService.SendLog:output_type -> proto.LogResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_log_proto_init() }
func file_proto_log_proto_init() {
	if File_proto_log_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_log_proto_rawDesc), len(file_proto_log_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_log_proto_goTypes,
		DependencyIndexes: file_proto_log_proto_depIdxs,
		MessageInfos:      file_proto_log_proto_msgTypes,
	}.Build()
	File_proto_log_proto = out.File
	file_proto_log_proto_goTypes = nil
	file_proto_log_proto_depIdxs = nil
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: version/version.proto

package version

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_version_version_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_version_version_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_version_version_proto_rawDescGZIP(), []int{0}
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version        string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	BuildDate      string `protobuf:"bytes,2,opt,name=build_date,json=buildDate,proto3" json:"build_date,omitempty"`
	GitBranch      string `protobuf:"bytes,3,opt,name=git_branch,json=gitBranch,proto3" json:"git_branch,omitempty"`
	GitCommit      string `protobuf:"bytes,4,opt,name=git_commit,json=gitCommit,proto3" json:"git_commit,omitempty"`
	GitTag         string `protobuf:"bytes,5,opt,name=git_tag,json=gitTag,proto3" json:"git_tag,omitempty"`
	GoVersion      string `protobuf:"bytes,6,opt,name=go_version,json=goVersion,proto3" json:"go_version,omitempty"`
	Compiler       string `protobuf:"bytes,7,opt,name=compiler,proto3" json:"compiler,omitempty"`
	Platform       string `protobuf:"bytes,8,opt,name=platform,proto3" json:"platform,omitempty"`
	KubectlVersion string `protobuf:"bytes,9,opt,name=kubectl_version,json=kubectlVersion,proto3" json:"kubectl_version,omitempty"`
	HelmVersion    string `protobuf:"bytes,10,opt,name=helm_version,json=helmVersion,proto3" json:"helm_version,omitempty"`
	GitRepo        string `protobuf:"bytes,11,opt,name=git_repo,json=gitRepo,proto3" json:"git_repo,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_version_version_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_version_version_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_version_version_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Response) GetBuildDate() string {
	if x != nil {
		return x.BuildDate
	}
	return ""
}

func (x *Response) GetGitBranch() string {
	if x != nil {
		return x.GitBranch
	}
	return ""
}

func (x *Response) GetGitCommit() string {
	if x != nil {
		return x.GitCommit
	}
	return ""
}

func (x *Response) GetGitTag() string {
	if x != nil {
		return x.GitTag
	}
	return ""
}

func (x *Response) GetGoVersion() string {
	if x != nil {
		return x.GoVersion
	}
	return ""
}

func (x *Response) GetCompiler() string {
	if x != nil {
		return x.Compiler
	}
	return ""
}

func (x *Response) GetPlatform() string {
	if x != nil {
		return x.Platform
	}
	return ""
}

func (x *Response) GetKubectlVersion() string {
	if x != nil {
		return x.KubectlVersion
	}
	return ""
}

func (x *Response) GetHelmVersion() string {
	if x != nil {
		return x.HelmVersion
	}
	return ""
}

func (x *Response) GetGitRepo() string {
	if x != nil {
		return x.GitRepo
	}
	return ""
}

var File_version_version_proto protoreflect.FileDescriptor

var file_version_version_proto_rawDesc = []byte{
	0x0a, 0x15, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x09,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xd8, 0x02, 0x0a, 0x08, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x44, 0x61, 0x74, 0x65, 0x12,
	0x1d, 0x0a, 0x0a, 0x67, 0x69, 0x74, 0x5f, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x67, 0x69, 0x74, 0x42, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x1d,
	0x0a, 0x0a, 0x67, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x67, 0x69, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x17, 0x0a,
	0x07, 0x67, 0x69, 0x74, 0x5f, 0x74, 0x61, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x67, 0x69, 0x74, 0x54, 0x61, 0x67, 0x12, 0x1d, 0x0a, 0x0a, 0x67, 0x6f, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x67, 0x6f, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65,
	0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65,
	0x72, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x27, 0x0a,
	0x0f, 0x6b, 0x75, 0x62, 0x65, 0x63, 0x74, 0x6c, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x6b, 0x75, 0x62, 0x65, 0x63, 0x74, 0x6c, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c, 0x68, 0x65, 0x6c, 0x6d, 0x5f, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x68, 0x65,
	0x6c, 0x6d, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x69, 0x74,
	0x5f, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67, 0x69, 0x74,
	0x52, 0x65, 0x70, 0x6f, 0x32, 0x54, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x49, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x2e, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x19, 0x92, 0x41, 0x02, 0x62, 0x00, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x12, 0x0c, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75, 0x63, 0x2d, 0x63, 0x6e, 0x7a,
	0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x34, 0x2f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x3b, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_version_version_proto_rawDescOnce sync.Once
	file_version_version_proto_rawDescData = file_version_version_proto_rawDesc
)

func file_version_version_proto_rawDescGZIP() []byte {
	file_version_version_proto_rawDescOnce.Do(func() {
		file_version_version_proto_rawDescData = protoimpl.X.CompressGZIP(file_version_version_proto_rawDescData)
	})
	return file_version_version_proto_rawDescData
}

var file_version_version_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_version_version_proto_goTypes = []interface{}{
	(*Request)(nil),  // 0: version.Request
	(*Response)(nil), // 1: version.Response
}
var file_version_version_proto_depIdxs = []int32{
	0, // 0: version.Version.Version:input_type -> version.Request
	1, // 1: version.Version.Version:output_type -> version.Response
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_version_version_proto_init() }
func file_version_version_proto_init() {
	if File_version_version_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_version_version_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_version_version_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_version_version_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_version_version_proto_goTypes,
		DependencyIndexes: file_version_version_proto_depIdxs,
		MessageInfos:      file_version_version_proto_msgTypes,
	}.Build()
	File_version_version_proto = out.File
	file_version_version_proto_rawDesc = nil
	file_version_version_proto_goTypes = nil
	file_version_version_proto_depIdxs = nil
}

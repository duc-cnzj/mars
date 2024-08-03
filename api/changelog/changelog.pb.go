// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: changelog/changelog.proto

package changelog

import (
	reflect "reflect"
	sync "sync"

	types "github.com/duc-cnzj/mars/api/v4/types"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ShowRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProjectId   int64 `protobuf:"varint,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	OnlyChanged bool  `protobuf:"varint,2,opt,name=only_changed,json=onlyChanged,proto3" json:"only_changed,omitempty"`
}

func (x *ShowRequest) Reset() {
	*x = ShowRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_changelog_changelog_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRequest) ProtoMessage() {}

func (x *ShowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_changelog_changelog_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowRequest.ProtoReflect.Descriptor instead.
func (*ShowRequest) Descriptor() ([]byte, []int) {
	return file_changelog_changelog_proto_rawDescGZIP(), []int{0}
}

func (x *ShowRequest) GetProjectId() int64 {
	if x != nil {
		return x.ProjectId
	}
	return 0
}

func (x *ShowRequest) GetOnlyChanged() bool {
	if x != nil {
		return x.OnlyChanged
	}
	return false
}

type ShowResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*types.ChangelogModel `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ShowResponse) Reset() {
	*x = ShowResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_changelog_changelog_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowResponse) ProtoMessage() {}

func (x *ShowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_changelog_changelog_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowResponse.ProtoReflect.Descriptor instead.
func (*ShowResponse) Descriptor() ([]byte, []int) {
	return file_changelog_changelog_proto_rawDescGZIP(), []int{1}
}

func (x *ShowResponse) GetItems() []*types.ChangelogModel {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_changelog_changelog_proto protoreflect.FileDescriptor

var file_changelog_changelog_proto_rawDesc = []byte{
	0x0a, 0x19, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x2f, 0x63, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x58, 0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x26, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02, 0x20, 0x00, 0x52, 0x09, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x6f, 0x6e, 0x6c, 0x79, 0x5f,
	0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x6f,
	0x6e, 0x6c, 0x79, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x64, 0x22, 0x3b, 0x0a, 0x0c, 0x53, 0x68,
	0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x05, 0x69, 0x74,
	0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x32, 0x73, 0x0a, 0x09, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x6c, 0x6f, 0x67, 0x12, 0x66, 0x0a, 0x04, 0x53, 0x68, 0x6f, 0x77, 0x12, 0x16, 0x2e, 0x63,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67,
	0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2d, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x27, 0x12, 0x25, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64,
	0x7d, 0x2f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x73, 0x42, 0x35, 0x5a, 0x33,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75, 0x63, 0x2d, 0x63,
	0x6e, 0x7a, 0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x34, 0x2f,
	0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x6c, 0x6f, 0x67, 0x3b, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x6c, 0x6f, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_changelog_changelog_proto_rawDescOnce sync.Once
	file_changelog_changelog_proto_rawDescData = file_changelog_changelog_proto_rawDesc
)

func file_changelog_changelog_proto_rawDescGZIP() []byte {
	file_changelog_changelog_proto_rawDescOnce.Do(func() {
		file_changelog_changelog_proto_rawDescData = protoimpl.X.CompressGZIP(file_changelog_changelog_proto_rawDescData)
	})
	return file_changelog_changelog_proto_rawDescData
}

var file_changelog_changelog_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_changelog_changelog_proto_goTypes = []interface{}{
	(*ShowRequest)(nil),          // 0: changelog.ShowRequest
	(*ShowResponse)(nil),         // 1: changelog.ShowResponse
	(*types.ChangelogModel)(nil), // 2: types.ChangelogModel
}
var file_changelog_changelog_proto_depIdxs = []int32{
	2, // 0: changelog.ShowResponse.items:type_name -> types.ChangelogModel
	0, // 1: changelog.Changelog.Show:input_type -> changelog.ShowRequest
	1, // 2: changelog.Changelog.Show:output_type -> changelog.ShowResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_changelog_changelog_proto_init() }
func file_changelog_changelog_proto_init() {
	if File_changelog_changelog_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_changelog_changelog_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowRequest); i {
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
		file_changelog_changelog_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowResponse); i {
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
			RawDescriptor: file_changelog_changelog_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_changelog_changelog_proto_goTypes,
		DependencyIndexes: file_changelog_changelog_proto_depIdxs,
		MessageInfos:      file_changelog_changelog_proto_msgTypes,
	}.Build()
	File_changelog_changelog_proto = out.File
	file_changelog_changelog_proto_rawDesc = nil
	file_changelog_changelog_proto_goTypes = nil
	file_changelog_changelog_proto_depIdxs = nil
}

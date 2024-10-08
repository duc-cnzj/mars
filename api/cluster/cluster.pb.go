// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v4.25.1
// source: cluster/cluster.proto

package cluster

import (
	reflect "reflect"
	sync "sync"

	websocket "github.com/duc-cnzj/mars/api/v5/websocket"
	_ "github.com/google/gnostic/openapiv3"
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

type Status int32

const (
	Status_StatusUnknown Status = 0
	Status_StatusBad     Status = 1
	Status_StatusNotGood Status = 2
	Status_StatusHealth  Status = 3
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "StatusUnknown",
		1: "StatusBad",
		2: "StatusNotGood",
		3: "StatusHealth",
	}
	Status_value = map[string]int32{
		"StatusUnknown": 0,
		"StatusBad":     1,
		"StatusNotGood": 2,
		"StatusHealth":  3,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_cluster_cluster_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_cluster_cluster_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_cluster_cluster_proto_rawDescGZIP(), []int{0}
}

type InfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *websocket.ClusterInfo `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *InfoResponse) Reset() {
	*x = InfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_cluster_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoResponse) ProtoMessage() {}

func (x *InfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_cluster_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoResponse.ProtoReflect.Descriptor instead.
func (*InfoResponse) Descriptor() ([]byte, []int) {
	return file_cluster_cluster_proto_rawDescGZIP(), []int{0}
}

func (x *InfoResponse) GetItem() *websocket.ClusterInfo {
	if x != nil {
		return x.Item
	}
	return nil
}

type InfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *InfoRequest) Reset() {
	*x = InfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cluster_cluster_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoRequest) ProtoMessage() {}

func (x *InfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cluster_cluster_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoRequest.ProtoReflect.Descriptor instead.
func (*InfoRequest) Descriptor() ([]byte, []int) {
	return file_cluster_cluster_proto_rawDescGZIP(), []int{1}
}

var File_cluster_cluster_proto protoreflect.FileDescriptor

var file_cluster_cluster_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x77, 0x65,
	0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3a, 0x0a, 0x0c, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65,
	0x74, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x69,
	0x74, 0x65, 0x6d, 0x22, 0x0d, 0x0a, 0x0b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2a, 0x4f, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x11, 0x0a, 0x0d,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00, 0x12,
	0x0d, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x61, 0x64, 0x10, 0x01, 0x12, 0x11,
	0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4e, 0x6f, 0x74, 0x47, 0x6f, 0x6f, 0x64, 0x10,
	0x02, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x10, 0x03, 0x32, 0x79, 0x0a, 0x07, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x6e,
	0x0a, 0x0b, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x2e,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x49, 0x6e,
	0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x32, 0xba, 0x47, 0x16, 0x12,
	0x12, 0xe6, 0x9f, 0xa5, 0xe7, 0x9c, 0x8b, 0xe9, 0x9b, 0x86, 0xe7, 0xbe, 0xa4, 0xe4, 0xbf, 0xa1,
	0xe6, 0x81, 0xaf, 0x5a, 0x00, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x13, 0x12, 0x11, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x42, 0x31,
	0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75, 0x63,
	0x2d, 0x63, 0x6e, 0x7a, 0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x35, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x3b, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cluster_cluster_proto_rawDescOnce sync.Once
	file_cluster_cluster_proto_rawDescData = file_cluster_cluster_proto_rawDesc
)

func file_cluster_cluster_proto_rawDescGZIP() []byte {
	file_cluster_cluster_proto_rawDescOnce.Do(func() {
		file_cluster_cluster_proto_rawDescData = protoimpl.X.CompressGZIP(file_cluster_cluster_proto_rawDescData)
	})
	return file_cluster_cluster_proto_rawDescData
}

var file_cluster_cluster_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cluster_cluster_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_cluster_cluster_proto_goTypes = []interface{}{
	(Status)(0),                   // 0: cluster.Status
	(*InfoResponse)(nil),          // 1: cluster.InfoResponse
	(*InfoRequest)(nil),           // 2: cluster.InfoRequest
	(*websocket.ClusterInfo)(nil), // 3: websocket.ClusterInfo
}
var file_cluster_cluster_proto_depIdxs = []int32{
	3, // 0: cluster.InfoResponse.item:type_name -> websocket.ClusterInfo
	2, // 1: cluster.Cluster.ClusterInfo:input_type -> cluster.InfoRequest
	1, // 2: cluster.Cluster.ClusterInfo:output_type -> cluster.InfoResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cluster_cluster_proto_init() }
func file_cluster_cluster_proto_init() {
	if File_cluster_cluster_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cluster_cluster_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfoResponse); i {
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
		file_cluster_cluster_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfoRequest); i {
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
			RawDescriptor: file_cluster_cluster_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cluster_cluster_proto_goTypes,
		DependencyIndexes: file_cluster_cluster_proto_depIdxs,
		EnumInfos:         file_cluster_cluster_proto_enumTypes,
		MessageInfos:      file_cluster_cluster_proto_msgTypes,
	}.Build()
	File_cluster_cluster_proto = out.File
	file_cluster_cluster_proto_rawDesc = nil
	file_cluster_cluster_proto_goTypes = nil
	file_cluster_cluster_proto_depIdxs = nil
}

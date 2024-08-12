// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v4.25.1
// source: event/event.proto

package event

import (
	reflect "reflect"
	sync "sync"

	types "github.com/duc-cnzj/mars/api/v4/types"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page       *int32                `protobuf:"varint,1,opt,name=page,proto3,oneof" json:"page,omitempty"`
	PageSize   *int32                `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3,oneof" json:"page_size,omitempty"`
	ActionType types.EventActionType `protobuf:"varint,3,opt,name=action_type,json=actionType,proto3,enum=types.EventActionType" json:"action_type,omitempty"`
	// 模糊搜索 message 和 username
	Search string `protobuf:"bytes,4,opt,name=search,proto3" json:"search,omitempty"`
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_event_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRequest.ProtoReflect.Descriptor instead.
func (*ListRequest) Descriptor() ([]byte, []int) {
	return file_event_event_proto_rawDescGZIP(), []int{0}
}

func (x *ListRequest) GetPage() int32 {
	if x != nil && x.Page != nil {
		return *x.Page
	}
	return 0
}

func (x *ListRequest) GetPageSize() int32 {
	if x != nil && x.PageSize != nil {
		return *x.PageSize
	}
	return 0
}

func (x *ListRequest) GetActionType() types.EventActionType {
	if x != nil {
		return x.ActionType
	}
	return types.EventActionType(0)
}

func (x *ListRequest) GetSearch() string {
	if x != nil {
		return x.Search
	}
	return ""
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page     int32               `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32               `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	Items    []*types.EventModel `protobuf:"bytes,3,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListResponse) Reset() {
	*x = ListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListResponse) ProtoMessage() {}

func (x *ListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_event_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListResponse.ProtoReflect.Descriptor instead.
func (*ListResponse) Descriptor() ([]byte, []int) {
	return file_event_event_proto_rawDescGZIP(), []int{1}
}

func (x *ListResponse) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *ListResponse) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListResponse) GetItems() []*types.EventModel {
	if x != nil {
		return x.Items
	}
	return nil
}

type ShowRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ShowRequest) Reset() {
	*x = ShowRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_event_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRequest) ProtoMessage() {}

func (x *ShowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_event_event_proto_msgTypes[2]
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
	return file_event_event_proto_rawDescGZIP(), []int{2}
}

func (x *ShowRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type ShowResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *types.EventModel `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *ShowResponse) Reset() {
	*x = ShowResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_event_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowResponse) ProtoMessage() {}

func (x *ShowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_event_event_proto_msgTypes[3]
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
	return file_event_event_proto_rawDescGZIP(), []int{3}
}

func (x *ShowResponse) GetItem() *types.EventModel {
	if x != nil {
		return x.Item
	}
	return nil
}

var File_event_event_proto protoreflect.FileDescriptor

var file_event_event_proto_rawDesc = []byte{
	0x0a, 0x11, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68,
	0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x33, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xb0, 0x01, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x48, 0x00, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20, 0x0a, 0x09,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x48,
	0x01, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01, 0x12, 0x37,
	0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x42,
	0x07, 0x0a, 0x05, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x70, 0x61, 0x67,
	0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x68, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61,
	0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70,
	0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x27, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73,
	0x22, 0x2a, 0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1b, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x0b, 0xe2, 0x41, 0x01,
	0x02, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x35, 0x0a, 0x0c,
	0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x04, 0x69,
	0x74, 0x65, 0x6d, 0x32, 0xdd, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x61, 0x0a,
	0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x12, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x30,
	0xba, 0x47, 0x1a, 0x12, 0x18, 0xe7, 0x94, 0xa8, 0xe6, 0x88, 0xb7, 0xe6, 0x93, 0x8d, 0xe4, 0xbd,
	0x9c, 0xe4, 0xba, 0x8b, 0xe4, 0xbb, 0xb6, 0xe5, 0x88, 0x97, 0xe8, 0xa1, 0xa8, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x0d, 0x12, 0x0b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x12, 0x71, 0x0a, 0x04, 0x53, 0x68, 0x6f, 0x77, 0x12, 0x12, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x40, 0xba, 0x47, 0x25, 0x12, 0x23, 0xe6, 0x9f, 0xa5, 0xe7, 0x9c, 0x8b, 0xe4, 0xba,
	0x8b, 0xe4, 0xbb, 0xb6, 0xe8, 0xaf, 0xa6, 0xe6, 0x83, 0x85, 0x2c, 0x20, 0xe5, 0x8c, 0x85, 0xe5,
	0x90, 0xab, 0x20, 0x6f, 0x6c, 0x64, 0x2c, 0x20, 0x6e, 0x65, 0x77, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x12, 0x12, 0x10, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b,
	0x69, 0x64, 0x7d, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x64, 0x75, 0x63, 0x2d, 0x63, 0x6e, 0x7a, 0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x34, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x3b, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_event_event_proto_rawDescOnce sync.Once
	file_event_event_proto_rawDescData = file_event_event_proto_rawDesc
)

func file_event_event_proto_rawDescGZIP() []byte {
	file_event_event_proto_rawDescOnce.Do(func() {
		file_event_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_event_event_proto_rawDescData)
	})
	return file_event_event_proto_rawDescData
}

var file_event_event_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_event_event_proto_goTypes = []interface{}{
	(*ListRequest)(nil),        // 0: event.ListRequest
	(*ListResponse)(nil),       // 1: event.ListResponse
	(*ShowRequest)(nil),        // 2: event.ShowRequest
	(*ShowResponse)(nil),       // 3: event.ShowResponse
	(types.EventActionType)(0), // 4: types.EventActionType
	(*types.EventModel)(nil),   // 5: types.EventModel
}
var file_event_event_proto_depIdxs = []int32{
	4, // 0: event.ListRequest.action_type:type_name -> types.EventActionType
	5, // 1: event.ListResponse.items:type_name -> types.EventModel
	5, // 2: event.ShowResponse.item:type_name -> types.EventModel
	0, // 3: event.Event.List:input_type -> event.ListRequest
	2, // 4: event.Event.Show:input_type -> event.ShowRequest
	1, // 5: event.Event.List:output_type -> event.ListResponse
	3, // 6: event.Event.Show:output_type -> event.ShowResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_event_event_proto_init() }
func file_event_event_proto_init() {
	if File_event_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_event_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRequest); i {
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
		file_event_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListResponse); i {
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
		file_event_event_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_event_event_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
	file_event_event_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_event_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_event_event_proto_goTypes,
		DependencyIndexes: file_event_event_proto_depIdxs,
		MessageInfos:      file_event_event_proto_msgTypes,
	}.Build()
	File_event_event_proto = out.File
	file_event_event_proto_rawDesc = nil
	file_event_event_proto_goTypes = nil
	file_event_event_proto_depIdxs = nil
}
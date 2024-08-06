// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: file/file.proto

package file

import (
	types "github.com/duc-cnzj/mars/api/v4/types"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteRequest) Reset() {
	*x = DeleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteRequest) ProtoMessage() {}

func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteRequest.ProtoReflect.Descriptor instead.
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{0}
}

func (x *DeleteRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteResponse) Reset() {
	*x = DeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteResponse) ProtoMessage() {}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteResponse.ProtoReflect.Descriptor instead.
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{1}
}

type DiskInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DiskInfoRequest) Reset() {
	*x = DiskInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DiskInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DiskInfoRequest) ProtoMessage() {}

func (x *DiskInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DiskInfoRequest.ProtoReflect.Descriptor instead.
func (*DiskInfoRequest) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{2}
}

type DiskInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Usage         int64  `protobuf:"varint,1,opt,name=usage,proto3" json:"usage,omitempty"`
	HumanizeUsage string `protobuf:"bytes,2,opt,name=humanize_usage,json=humanizeUsage,proto3" json:"humanize_usage,omitempty"`
}

func (x *DiskInfoResponse) Reset() {
	*x = DiskInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DiskInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DiskInfoResponse) ProtoMessage() {}

func (x *DiskInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DiskInfoResponse.ProtoReflect.Descriptor instead.
func (*DiskInfoResponse) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{3}
}

func (x *DiskInfoResponse) GetUsage() int64 {
	if x != nil {
		return x.Usage
	}
	return 0
}

func (x *DiskInfoResponse) GetHumanizeUsage() string {
	if x != nil {
		return x.HumanizeUsage
	}
	return ""
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page           *int32 `protobuf:"varint,1,opt,name=page,proto3,oneof" json:"page,omitempty"`
	PageSize       *int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3,oneof" json:"page_size,omitempty"`
	WithoutDeleted bool   `protobuf:"varint,3,opt,name=without_deleted,json=withoutDeleted,proto3" json:"without_deleted,omitempty"`
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[4]
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
	return file_file_file_proto_rawDescGZIP(), []int{4}
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

func (x *ListRequest) GetWithoutDeleted() bool {
	if x != nil {
		return x.WithoutDeleted
	}
	return false
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page     int32              `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32              `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	Items    []*types.FileModel `protobuf:"bytes,3,rep,name=items,proto3" json:"items,omitempty"`
	Count    int32              `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *ListResponse) Reset() {
	*x = ListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListResponse) ProtoMessage() {}

func (x *ListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[5]
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
	return file_file_file_proto_rawDescGZIP(), []int{5}
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

func (x *ListResponse) GetItems() []*types.FileModel {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *ListResponse) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type MaxUploadSizeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MaxUploadSizeRequest) Reset() {
	*x = MaxUploadSizeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MaxUploadSizeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MaxUploadSizeRequest) ProtoMessage() {}

func (x *MaxUploadSizeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MaxUploadSizeRequest.ProtoReflect.Descriptor instead.
func (*MaxUploadSizeRequest) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{6}
}

type MaxUploadSizeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HumanizeSize string `protobuf:"bytes,1,opt,name=humanize_size,json=humanizeSize,proto3" json:"humanize_size,omitempty"`
	Bytes        uint32 `protobuf:"varint,2,opt,name=bytes,proto3" json:"bytes,omitempty"`
}

func (x *MaxUploadSizeResponse) Reset() {
	*x = MaxUploadSizeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MaxUploadSizeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MaxUploadSizeResponse) ProtoMessage() {}

func (x *MaxUploadSizeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MaxUploadSizeResponse.ProtoReflect.Descriptor instead.
func (*MaxUploadSizeResponse) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{7}
}

func (x *MaxUploadSizeResponse) GetHumanizeSize() string {
	if x != nil {
		return x.HumanizeSize
	}
	return ""
}

func (x *MaxUploadSizeResponse) GetBytes() uint32 {
	if x != nil {
		return x.Bytes
	}
	return 0
}

type ShowRecordsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ShowRecordsRequest) Reset() {
	*x = ShowRecordsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRecordsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRecordsRequest) ProtoMessage() {}

func (x *ShowRecordsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowRecordsRequest.ProtoReflect.Descriptor instead.
func (*ShowRecordsRequest) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{8}
}

func (x *ShowRecordsRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type ShowRecordsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []string `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ShowRecordsResponse) Reset() {
	*x = ShowRecordsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_file_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRecordsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRecordsResponse) ProtoMessage() {}

func (x *ShowRecordsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_file_file_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowRecordsResponse.ProtoReflect.Descriptor instead.
func (*ShowRecordsResponse) Descriptor() ([]byte, []int) {
	return file_file_file_proto_rawDescGZIP(), []int{9}
}

func (x *ShowRecordsResponse) GetItems() []string {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_file_file_proto protoreflect.FileDescriptor

var file_file_file_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x69, 0x6c, 0x65, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f,
	0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x2b, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42,
	0x0a, 0xe0, 0x41, 0x02, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x28, 0x01, 0x52, 0x02, 0x69, 0x64, 0x22,
	0x10, 0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x11, 0x0a, 0x0f, 0x44, 0x69, 0x73, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x4f, 0x0a, 0x10, 0x44, 0x69, 0x73, 0x6b, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x75, 0x73, 0x61, 0x67, 0x65, 0x12, 0x25,
	0x0a, 0x0e, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x5f, 0x75, 0x73, 0x61, 0x67, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x69, 0x7a, 0x65,
	0x55, 0x73, 0x61, 0x67, 0x65, 0x22, 0x88, 0x01, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20,
	0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x48, 0x01, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01,
	0x12, 0x27, 0x0a, 0x0f, 0x77, 0x69, 0x74, 0x68, 0x6f, 0x75, 0x74, 0x5f, 0x64, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x77, 0x69, 0x74, 0x68, 0x6f,
	0x75, 0x74, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x70, 0x61,
	0x67, 0x65, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65,
	0x22, 0x7d, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x70, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x26, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4d, 0x6f, 0x64,
	0x65, 0x6c, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22,
	0x16, 0x0a, 0x14, 0x4d, 0x61, 0x78, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x52, 0x0a, 0x15, 0x4d, 0x61, 0x78, 0x55, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x23, 0x0a, 0x0d, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x68, 0x75, 0x6d, 0x61, 0x6e, 0x69, 0x7a,
	0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x62, 0x79, 0x74, 0x65, 0x73, 0x22, 0x30, 0x0a, 0x12, 0x53,
	0x68, 0x6f, 0x77, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1a, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xe0,
	0x41, 0x02, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x2b, 0x0a,
	0x13, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x32, 0xb2, 0x04, 0x0a, 0x04, 0x46,
	0x69, 0x6c, 0x65, 0x12, 0x52, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x11, 0x2e, 0x66, 0x69,
	0x6c, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12,
	0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x23, 0x92, 0x41, 0x0e, 0x12, 0x0c, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe5,
	0x88, 0x97, 0xe8, 0xa1, 0xa8, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x73, 0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x12, 0x18, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x53, 0x68,
	0x6f, 0x77, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x19, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x63, 0x6f,
	0x72, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2f, 0x92, 0x41, 0x0e,
	0x12, 0x0c, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe8, 0xaf, 0xa6, 0xe6, 0x83, 0x85, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x18, 0x12, 0x16, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x5d, 0x0a, 0x06,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x66, 0x69,
	0x6c, 0x65, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x28, 0x92, 0x41, 0x0e, 0x12, 0x0c, 0xe5, 0x88, 0xa0, 0xe9, 0x99, 0xa4, 0xe6, 0x96,
	0x87, 0xe4, 0xbb, 0xb6, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x2a, 0x0f, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x74, 0x0a, 0x08, 0x44,
	0x69, 0x73, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x15, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x44,
	0x69, 0x73, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16,
	0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x44, 0x69, 0x73, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x39, 0x92, 0x41, 0x1a, 0x12, 0x18, 0xe6, 0x9f, 0xa5,
	0xe7, 0x9c, 0x8b, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe7, 0x9b, 0xae, 0xe5, 0xbd, 0x95, 0xe5,
	0xa4, 0xa7, 0xe5, 0xb0, 0x8f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x12, 0x14, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x69, 0x6e, 0x66,
	0x6f, 0x12, 0x8b, 0x01, 0x0a, 0x0d, 0x4d, 0x61, 0x78, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53,
	0x69, 0x7a, 0x65, 0x12, 0x1a, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x4d, 0x61, 0x78, 0x55, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1b, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x4d, 0x61, 0x78, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x53, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x41, 0x92, 0x41,
	0x1c, 0x12, 0x18, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0xe6, 0x9c, 0x80, 0xe5, 0xa4, 0xa7, 0xe4,
	0xb8, 0x8a, 0xe4, 0xbc, 0xa0, 0xe5, 0xa4, 0xa7, 0xe5, 0xb0, 0x8f, 0x62, 0x00, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1c, 0x12, 0x1a, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f,
	0x6d, 0x61, 0x78, 0x5f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x42,
	0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75,
	0x63, 0x2d, 0x63, 0x6e, 0x7a, 0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x34, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x3b, 0x66, 0x69, 0x6c, 0x65, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_file_file_proto_rawDescOnce sync.Once
	file_file_file_proto_rawDescData = file_file_file_proto_rawDesc
)

func file_file_file_proto_rawDescGZIP() []byte {
	file_file_file_proto_rawDescOnce.Do(func() {
		file_file_file_proto_rawDescData = protoimpl.X.CompressGZIP(file_file_file_proto_rawDescData)
	})
	return file_file_file_proto_rawDescData
}

var file_file_file_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_file_file_proto_goTypes = []interface{}{
	(*DeleteRequest)(nil),         // 0: file.DeleteRequest
	(*DeleteResponse)(nil),        // 1: file.DeleteResponse
	(*DiskInfoRequest)(nil),       // 2: file.DiskInfoRequest
	(*DiskInfoResponse)(nil),      // 3: file.DiskInfoResponse
	(*ListRequest)(nil),           // 4: file.ListRequest
	(*ListResponse)(nil),          // 5: file.ListResponse
	(*MaxUploadSizeRequest)(nil),  // 6: file.MaxUploadSizeRequest
	(*MaxUploadSizeResponse)(nil), // 7: file.MaxUploadSizeResponse
	(*ShowRecordsRequest)(nil),    // 8: file.ShowRecordsRequest
	(*ShowRecordsResponse)(nil),   // 9: file.ShowRecordsResponse
	(*types.FileModel)(nil),       // 10: types.FileModel
}
var file_file_file_proto_depIdxs = []int32{
	10, // 0: file.ListResponse.items:type_name -> types.FileModel
	4,  // 1: file.File.List:input_type -> file.ListRequest
	8,  // 2: file.File.ShowRecords:input_type -> file.ShowRecordsRequest
	0,  // 3: file.File.Delete:input_type -> file.DeleteRequest
	2,  // 4: file.File.DiskInfo:input_type -> file.DiskInfoRequest
	6,  // 5: file.File.MaxUploadSize:input_type -> file.MaxUploadSizeRequest
	5,  // 6: file.File.List:output_type -> file.ListResponse
	9,  // 7: file.File.ShowRecords:output_type -> file.ShowRecordsResponse
	1,  // 8: file.File.Delete:output_type -> file.DeleteResponse
	3,  // 9: file.File.DiskInfo:output_type -> file.DiskInfoResponse
	7,  // 10: file.File.MaxUploadSize:output_type -> file.MaxUploadSizeResponse
	6,  // [6:11] is the sub-list for method output_type
	1,  // [1:6] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_file_file_proto_init() }
func file_file_file_proto_init() {
	if File_file_file_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_file_file_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteRequest); i {
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
		file_file_file_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteResponse); i {
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
		file_file_file_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DiskInfoRequest); i {
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
		file_file_file_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DiskInfoResponse); i {
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
		file_file_file_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
		file_file_file_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
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
		file_file_file_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MaxUploadSizeRequest); i {
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
		file_file_file_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MaxUploadSizeResponse); i {
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
		file_file_file_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowRecordsRequest); i {
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
		file_file_file_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowRecordsResponse); i {
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
	file_file_file_proto_msgTypes[4].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_file_file_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_file_file_proto_goTypes,
		DependencyIndexes: file_file_file_proto_depIdxs,
		MessageInfos:      file_file_file_proto_msgTypes,
	}.Build()
	File_file_file_proto = out.File
	file_file_file_proto_rawDesc = nil
	file_file_file_proto_goTypes = nil
	file_file_file_proto_depIdxs = nil
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: repo/repo.proto

package repo

import (
	reflect "reflect"
	sync "sync"

	mars "github.com/duc-cnzj/mars/api/v4/mars"
	types "github.com/duc-cnzj/mars/api/v4/types"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
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

	Page     *int32 `protobuf:"varint,1,opt,name=page,proto3,oneof" json:"page,omitempty"`
	PageSize *int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3,oneof" json:"page_size,omitempty"`
	Enabled  *bool  `protobuf:"varint,3,opt,name=enabled,proto3,oneof" json:"enabled,omitempty"`
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[0]
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
	return file_repo_repo_proto_rawDescGZIP(), []int{0}
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

func (x *ListRequest) GetEnabled() bool {
	if x != nil && x.Enabled != nil {
		return *x.Enabled
	}
	return false
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page     int32              `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32              `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	Count    int32              `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Items    []*types.RepoModel `protobuf:"bytes,4,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListResponse) Reset() {
	*x = ListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListResponse) ProtoMessage() {}

func (x *ListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[1]
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
	return file_repo_repo_proto_rawDescGZIP(), []int{1}
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

func (x *ListResponse) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *ListResponse) GetItems() []*types.RepoModel {
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
		mi := &file_repo_repo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRequest) ProtoMessage() {}

func (x *ShowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[2]
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
	return file_repo_repo_proto_rawDescGZIP(), []int{2}
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

	Item *types.RepoModel `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *ShowResponse) Reset() {
	*x = ShowResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowResponse) ProtoMessage() {}

func (x *ShowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[3]
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
	return file_repo_repo_proto_rawDescGZIP(), []int{3}
}

func (x *ShowResponse) GetItem() *types.RepoModel {
	if x != nil {
		return x.Item
	}
	return nil
}

type ToggleEnabledRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Enabled bool  `protobuf:"varint,2,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *ToggleEnabledRequest) Reset() {
	*x = ToggleEnabledRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToggleEnabledRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleEnabledRequest) ProtoMessage() {}

func (x *ToggleEnabledRequest) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleEnabledRequest.ProtoReflect.Descriptor instead.
func (*ToggleEnabledRequest) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{4}
}

func (x *ToggleEnabledRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ToggleEnabledRequest) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

type ToggleEnabledResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *types.RepoModel `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *ToggleEnabledResponse) Reset() {
	*x = ToggleEnabledResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToggleEnabledResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleEnabledResponse) ProtoMessage() {}

func (x *ToggleEnabledResponse) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleEnabledResponse.ProtoReflect.Descriptor instead.
func (*ToggleEnabledResponse) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{5}
}

func (x *ToggleEnabledResponse) GetItem() *types.RepoModel {
	if x != nil {
		return x.Item
	}
	return nil
}

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	GitProjectId *int32       `protobuf:"varint,2,opt,name=git_project_id,json=gitProjectId,proto3,oneof" json:"git_project_id,omitempty"`
	MarsConfig   *mars.Config `protobuf:"bytes,3,opt,name=mars_config,json=marsConfig,proto3" json:"mars_config,omitempty"`
	NeedGitRepo  bool         `protobuf:"varint,4,opt,name=need_git_repo,json=needGitRepo,proto3" json:"need_git_repo,omitempty"`
}

func (x *CreateRequest) Reset() {
	*x = CreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateRequest) ProtoMessage() {}

func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateRequest.ProtoReflect.Descriptor instead.
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{6}
}

func (x *CreateRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateRequest) GetGitProjectId() int32 {
	if x != nil && x.GitProjectId != nil {
		return *x.GitProjectId
	}
	return 0
}

func (x *CreateRequest) GetMarsConfig() *mars.Config {
	if x != nil {
		return x.MarsConfig
	}
	return nil
}

func (x *CreateRequest) GetNeedGitRepo() bool {
	if x != nil {
		return x.NeedGitRepo
	}
	return false
}

type CreateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *types.RepoModel `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *CreateResponse) Reset() {
	*x = CreateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateResponse) ProtoMessage() {}

func (x *CreateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateResponse.ProtoReflect.Descriptor instead.
func (*CreateResponse) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{7}
}

func (x *CreateResponse) GetItem() *types.RepoModel {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           int32        `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name         string       `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	GitProjectId *int32       `protobuf:"varint,3,opt,name=git_project_id,json=gitProjectId,proto3,oneof" json:"git_project_id,omitempty"`
	MarsConfig   *mars.Config `protobuf:"bytes,4,opt,name=mars_config,json=marsConfig,proto3" json:"mars_config,omitempty"`
	NeedGitRepo  bool         `protobuf:"varint,5,opt,name=need_git_repo,json=needGitRepo,proto3" json:"need_git_repo,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{8}
}

func (x *UpdateRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *UpdateRequest) GetGitProjectId() int32 {
	if x != nil && x.GitProjectId != nil {
		return *x.GitProjectId
	}
	return 0
}

func (x *UpdateRequest) GetMarsConfig() *mars.Config {
	if x != nil {
		return x.MarsConfig
	}
	return nil
}

func (x *UpdateRequest) GetNeedGitRepo() bool {
	if x != nil {
		return x.NeedGitRepo
	}
	return false
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *types.RepoModel `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_repo_repo_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_repo_repo_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_repo_repo_proto_rawDescGZIP(), []int{9}
}

func (x *UpdateResponse) GetItem() *types.RepoModel {
	if x != nil {
		return x.Item
	}
	return nil
}

var File_repo_repo_proto protoreflect.FileDescriptor

var file_repo_repo_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d,
	0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x01, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x48, 0x00, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20, 0x0a, 0x09,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x48,
	0x01, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01, 0x12, 0x1d,
	0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x48,
	0x02, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a,
	0x05, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f,
	0x73, 0x69, 0x7a, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64,
	0x22, 0x7d, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x70, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x26, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52,
	0x65, 0x70, 0x6f, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22,
	0x29, 0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xe0, 0x41, 0x02, 0xfa,
	0x42, 0x04, 0x1a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x34, 0x0a, 0x0c, 0x53, 0x68,
	0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x52, 0x65, 0x70, 0x6f, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d,
	0x22, 0x51, 0x0a, 0x14, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xe0, 0x41, 0x02, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x20, 0x00,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62,
	0x6c, 0x65, 0x64, 0x22, 0x3d, 0x0a, 0x15, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x45, 0x6e, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x04, 0x69, 0x74,
	0x65, 0x6d, 0x22, 0xd2, 0x01, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x1c, 0xe0, 0x41, 0x02, 0xfa, 0x42, 0x16, 0x72, 0x14, 0x20, 0x01, 0x32, 0x10,
	0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a, 0x30, 0x2d, 0x39, 0x5f, 0x2d, 0x5d, 0x2b, 0x24,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x29, 0x0a, 0x0e, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00,
	0x52, 0x0c, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x88, 0x01,
	0x01, 0x12, 0x2d, 0x0a, 0x0b, 0x6d, 0x61, 0x72, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x52, 0x0a, 0x6d, 0x61, 0x72, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x22, 0x0a, 0x0d, 0x6e, 0x65, 0x65, 0x64, 0x5f, 0x67, 0x69, 0x74, 0x5f, 0x72, 0x65, 0x70,
	0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x6e, 0x65, 0x65, 0x64, 0x47, 0x69, 0x74,
	0x52, 0x65, 0x70, 0x6f, 0x42, 0x11, 0x0a, 0x0f, 0x5f, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x22, 0x36, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x69, 0x74, 0x65,
	0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x52, 0x65, 0x70, 0x6f, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22,
	0xee, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1a, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xe0,
	0x41, 0x02, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x12, 0x30, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1c, 0xe0, 0x41, 0x02,
	0xfa, 0x42, 0x16, 0x72, 0x14, 0x20, 0x01, 0x32, 0x10, 0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d,
	0x5a, 0x30, 0x2d, 0x39, 0x5f, 0x2d, 0x5d, 0x2b, 0x24, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x29, 0x0a, 0x0e, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x0c, 0x67, 0x69, 0x74, 0x50, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x88, 0x01, 0x01, 0x12, 0x2d, 0x0a, 0x0b, 0x6d, 0x61,
	0x72, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0a, 0x6d,
	0x61, 0x72, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x22, 0x0a, 0x0d, 0x6e, 0x65, 0x65,
	0x64, 0x5f, 0x67, 0x69, 0x74, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0b, 0x6e, 0x65, 0x65, 0x64, 0x47, 0x69, 0x74, 0x52, 0x65, 0x70, 0x6f, 0x42, 0x11, 0x0a,
	0x0f, 0x5f, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64,
	0x22, 0x36, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x4d, 0x6f, 0x64,
	0x65, 0x6c, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x32, 0xff, 0x03, 0x0a, 0x04, 0x52, 0x65, 0x70,
	0x6f, 0x12, 0x58, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x11, 0x2e, 0x72, 0x65, 0x70, 0x6f,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x72,
	0x65, 0x70, 0x6f, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x29, 0x92, 0x41, 0x14, 0x12, 0x12, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0x20, 0x72, 0x65,
	0x70, 0x6f, 0x20, 0xe5, 0x88, 0x97, 0xe8, 0xa1, 0xa8, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12,
	0x0a, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x12, 0x5a, 0x0a, 0x06, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x72, 0x65, 0x70,
	0x6f, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x25, 0x92, 0x41, 0x0d, 0x12, 0x0b, 0xe5, 0x88, 0x9b, 0xe5, 0xbb, 0xba, 0x20, 0x72, 0x65,
	0x70, 0x6f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x3a, 0x01, 0x2a, 0x22, 0x0a, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x12, 0x5d, 0x0a, 0x04, 0x53, 0x68, 0x6f, 0x77, 0x12,
	0x11, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x12, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2e, 0x92, 0x41, 0x14, 0x12, 0x12, 0xe8, 0x8e, 0xb7,
	0xe5, 0x8f, 0x96, 0x20, 0x72, 0x65, 0x70, 0x6f, 0x20, 0xe8, 0xaf, 0xa6, 0xe6, 0x83, 0x85, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x11, 0x12, 0x0f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x70, 0x6f,
	0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x5f, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x12, 0x13, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2a, 0x92, 0x41, 0x0d,
	0x12, 0x0b, 0xe6, 0x9b, 0xb4, 0xe6, 0x96, 0xb0, 0x20, 0x72, 0x65, 0x70, 0x6f, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x14, 0x3a, 0x01, 0x2a, 0x1a, 0x0f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x70,
	0x6f, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x80, 0x01, 0x0a, 0x0d, 0x54, 0x6f, 0x67, 0x67,
	0x6c, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x1a, 0x2e, 0x72, 0x65, 0x70, 0x6f,
	0x2e, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x54, 0x6f, 0x67,
	0x67, 0x6c, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x36, 0x92, 0x41, 0x0f, 0x12, 0x0d, 0xe5, 0xbc, 0x80, 0xe5, 0x90, 0xaf, 0x2f,
	0xe5, 0x85, 0xb3, 0xe9, 0x97, 0xad, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1e, 0x3a, 0x01, 0x2a, 0x22,
	0x19, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x2f, 0x74, 0x6f, 0x67, 0x67,
	0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75, 0x63, 0x2d, 0x63, 0x6e, 0x7a,
	0x6a, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x34, 0x2f, 0x72, 0x65,
	0x70, 0x6f, 0x3b, 0x72, 0x65, 0x70, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_repo_repo_proto_rawDescOnce sync.Once
	file_repo_repo_proto_rawDescData = file_repo_repo_proto_rawDesc
)

func file_repo_repo_proto_rawDescGZIP() []byte {
	file_repo_repo_proto_rawDescOnce.Do(func() {
		file_repo_repo_proto_rawDescData = protoimpl.X.CompressGZIP(file_repo_repo_proto_rawDescData)
	})
	return file_repo_repo_proto_rawDescData
}

var file_repo_repo_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_repo_repo_proto_goTypes = []interface{}{
	(*ListRequest)(nil),           // 0: repo.ListRequest
	(*ListResponse)(nil),          // 1: repo.ListResponse
	(*ShowRequest)(nil),           // 2: repo.ShowRequest
	(*ShowResponse)(nil),          // 3: repo.ShowResponse
	(*ToggleEnabledRequest)(nil),  // 4: repo.ToggleEnabledRequest
	(*ToggleEnabledResponse)(nil), // 5: repo.ToggleEnabledResponse
	(*CreateRequest)(nil),         // 6: repo.CreateRequest
	(*CreateResponse)(nil),        // 7: repo.CreateResponse
	(*UpdateRequest)(nil),         // 8: repo.UpdateRequest
	(*UpdateResponse)(nil),        // 9: repo.UpdateResponse
	(*types.RepoModel)(nil),       // 10: types.RepoModel
	(*mars.Config)(nil),           // 11: mars.Config
}
var file_repo_repo_proto_depIdxs = []int32{
	10, // 0: repo.ListResponse.items:type_name -> types.RepoModel
	10, // 1: repo.ShowResponse.item:type_name -> types.RepoModel
	10, // 2: repo.ToggleEnabledResponse.item:type_name -> types.RepoModel
	11, // 3: repo.CreateRequest.mars_config:type_name -> mars.Config
	10, // 4: repo.CreateResponse.item:type_name -> types.RepoModel
	11, // 5: repo.UpdateRequest.mars_config:type_name -> mars.Config
	10, // 6: repo.UpdateResponse.item:type_name -> types.RepoModel
	0,  // 7: repo.Repo.List:input_type -> repo.ListRequest
	6,  // 8: repo.Repo.Create:input_type -> repo.CreateRequest
	2,  // 9: repo.Repo.Show:input_type -> repo.ShowRequest
	8,  // 10: repo.Repo.Update:input_type -> repo.UpdateRequest
	4,  // 11: repo.Repo.ToggleEnabled:input_type -> repo.ToggleEnabledRequest
	1,  // 12: repo.Repo.List:output_type -> repo.ListResponse
	7,  // 13: repo.Repo.Create:output_type -> repo.CreateResponse
	3,  // 14: repo.Repo.Show:output_type -> repo.ShowResponse
	9,  // 15: repo.Repo.Update:output_type -> repo.UpdateResponse
	5,  // 16: repo.Repo.ToggleEnabled:output_type -> repo.ToggleEnabledResponse
	12, // [12:17] is the sub-list for method output_type
	7,  // [7:12] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_repo_repo_proto_init() }
func file_repo_repo_proto_init() {
	if File_repo_repo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_repo_repo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_repo_repo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_repo_repo_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_repo_repo_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_repo_repo_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToggleEnabledRequest); i {
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
		file_repo_repo_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToggleEnabledResponse); i {
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
		file_repo_repo_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateRequest); i {
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
		file_repo_repo_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateResponse); i {
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
		file_repo_repo_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
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
		file_repo_repo_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
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
	file_repo_repo_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_repo_repo_proto_msgTypes[6].OneofWrappers = []interface{}{}
	file_repo_repo_proto_msgTypes[8].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_repo_repo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_repo_repo_proto_goTypes,
		DependencyIndexes: file_repo_repo_proto_depIdxs,
		MessageInfos:      file_repo_repo_proto_msgTypes,
	}.Build()
	File_repo_repo_proto = out.File
	file_repo_repo_proto_rawDesc = nil
	file_repo_repo_proto_goTypes = nil
	file_repo_repo_proto_depIdxs = nil
}

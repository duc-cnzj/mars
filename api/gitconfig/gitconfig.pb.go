// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: gitconfig/gitconfig.proto

package gitconfig

import (
	reflect "reflect"
	sync "sync"

	mars "github.com/duc-cnzj/mars/api/v4/mars"
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

type FileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId string `protobuf:"bytes,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
	Branch       string `protobuf:"bytes,2,opt,name=branch,proto3" json:"branch,omitempty"`
}

func (x *FileRequest) Reset() {
	*x = FileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileRequest) ProtoMessage() {}

func (x *FileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileRequest.ProtoReflect.Descriptor instead.
func (*FileRequest) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{0}
}

func (x *FileRequest) GetGitProjectId() string {
	if x != nil {
		return x.GitProjectId
	}
	return ""
}

func (x *FileRequest) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

type FileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data     string          `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Type     string          `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Elements []*mars.Element `protobuf:"bytes,3,rep,name=elements,proto3" json:"elements,omitempty"`
}

func (x *FileResponse) Reset() {
	*x = FileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileResponse) ProtoMessage() {}

func (x *FileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileResponse.ProtoReflect.Descriptor instead.
func (*FileResponse) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{1}
}

func (x *FileResponse) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

func (x *FileResponse) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FileResponse) GetElements() []*mars.Element {
	if x != nil {
		return x.Elements
	}
	return nil
}

type ShowRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId int64  `protobuf:"varint,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
	Branch       string `protobuf:"bytes,2,opt,name=branch,proto3" json:"branch,omitempty"`
}

func (x *ShowRequest) Reset() {
	*x = ShowRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRequest) ProtoMessage() {}

func (x *ShowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[2]
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
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{2}
}

func (x *ShowRequest) GetGitProjectId() int64 {
	if x != nil {
		return x.GitProjectId
	}
	return 0
}

func (x *ShowRequest) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

type ShowResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Branch string       `protobuf:"bytes,1,opt,name=branch,proto3" json:"branch,omitempty"`
	Config *mars.Config `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ShowResponse) Reset() {
	*x = ShowResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowResponse) ProtoMessage() {}

func (x *ShowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[3]
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
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{3}
}

func (x *ShowResponse) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

func (x *ShowResponse) GetConfig() *mars.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type GlobalConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId int64 `protobuf:"varint,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
}

func (x *GlobalConfigRequest) Reset() {
	*x = GlobalConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GlobalConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GlobalConfigRequest) ProtoMessage() {}

func (x *GlobalConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GlobalConfigRequest.ProtoReflect.Descriptor instead.
func (*GlobalConfigRequest) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{4}
}

func (x *GlobalConfigRequest) GetGitProjectId() int64 {
	if x != nil {
		return x.GitProjectId
	}
	return 0
}

type GlobalConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Enabled bool         `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	Config  *mars.Config `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *GlobalConfigResponse) Reset() {
	*x = GlobalConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GlobalConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GlobalConfigResponse) ProtoMessage() {}

func (x *GlobalConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GlobalConfigResponse.ProtoReflect.Descriptor instead.
func (*GlobalConfigResponse) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{5}
}

func (x *GlobalConfigResponse) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *GlobalConfigResponse) GetConfig() *mars.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId int64        `protobuf:"varint,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
	Config       *mars.Config `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[6]
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
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateRequest) GetGitProjectId() int64 {
	if x != nil {
		return x.GitProjectId
	}
	return 0
}

func (x *UpdateRequest) GetConfig() *mars.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Config *mars.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[7]
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
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateResponse) GetConfig() *mars.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type ToggleGlobalStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId int64 `protobuf:"varint,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
	Enabled      bool  `protobuf:"varint,2,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

func (x *ToggleGlobalStatusRequest) Reset() {
	*x = ToggleGlobalStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToggleGlobalStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleGlobalStatusRequest) ProtoMessage() {}

func (x *ToggleGlobalStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleGlobalStatusRequest.ProtoReflect.Descriptor instead.
func (*ToggleGlobalStatusRequest) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{8}
}

func (x *ToggleGlobalStatusRequest) GetGitProjectId() int64 {
	if x != nil {
		return x.GitProjectId
	}
	return 0
}

func (x *ToggleGlobalStatusRequest) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

type DefaultChartValuesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GitProjectId int64  `protobuf:"varint,1,opt,name=git_project_id,json=gitProjectId,proto3" json:"git_project_id,omitempty"`
	Branch       string `protobuf:"bytes,2,opt,name=branch,proto3" json:"branch,omitempty"`
}

func (x *DefaultChartValuesRequest) Reset() {
	*x = DefaultChartValuesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DefaultChartValuesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DefaultChartValuesRequest) ProtoMessage() {}

func (x *DefaultChartValuesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DefaultChartValuesRequest.ProtoReflect.Descriptor instead.
func (*DefaultChartValuesRequest) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{9}
}

func (x *DefaultChartValuesRequest) GetGitProjectId() int64 {
	if x != nil {
		return x.GitProjectId
	}
	return 0
}

func (x *DefaultChartValuesRequest) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

type DefaultChartValuesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *DefaultChartValuesResponse) Reset() {
	*x = DefaultChartValuesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DefaultChartValuesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DefaultChartValuesResponse) ProtoMessage() {}

func (x *DefaultChartValuesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DefaultChartValuesResponse.ProtoReflect.Descriptor instead.
func (*DefaultChartValuesResponse) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{10}
}

func (x *DefaultChartValuesResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type ToggleGlobalStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ToggleGlobalStatusResponse) Reset() {
	*x = ToggleGlobalStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitconfig_gitconfig_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ToggleGlobalStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ToggleGlobalStatusResponse) ProtoMessage() {}

func (x *ToggleGlobalStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitconfig_gitconfig_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ToggleGlobalStatusResponse.ProtoReflect.Descriptor instead.
func (*ToggleGlobalStatusResponse) Descriptor() ([]byte, []int) {
	return file_gitconfig_gitconfig_proto_rawDescGZIP(), []int{11}
}

var File_gitconfig_gitconfig_proto protoreflect.FileDescriptor

var file_gitconfig_gitconfig_proto_rawDesc = []byte{
	0x0a, 0x19, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x67, 0x69, 0x74, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x67, 0x69, 0x74,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x6d,
	0x61, 0x72, 0x73, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c,
	0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5d, 0x0a, 0x0b,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x0e, 0x67,
	0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x20, 0x01, 0x52, 0x0c, 0x67, 0x69,
	0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x06, 0x62, 0x72,
	0x61, 0x6e, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72,
	0x02, 0x20, 0x01, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x22, 0x61, 0x0a, 0x0c, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x29, 0x0a, 0x08, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x45, 0x6c, 0x65,
	0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x54,
	0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a,
	0x0e, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02, 0x20, 0x00, 0x52, 0x0c,
	0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06,
	0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x72,
	0x61, 0x6e, 0x63, 0x68, 0x22, 0x4c, 0x0a, 0x0c, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x24, 0x0a, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6d,
	0x61, 0x72, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x22, 0x44, 0x0a, 0x13, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x0e, 0x67, 0x69, 0x74,
	0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02, 0x20, 0x00, 0x52, 0x0c, 0x67, 0x69, 0x74, 0x50,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x22, 0x56, 0x0a, 0x14, 0x47, 0x6c, 0x6f, 0x62,
	0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x24, 0x0a, 0x06, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6d, 0x61, 0x72,
	0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x22, 0x64, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x2d, 0x0a, 0x0e, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02,
	0x20, 0x00, 0x52, 0x0c, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64,
	0x12, 0x24, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x6d, 0x61, 0x72, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x36, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6d, 0x61, 0x72, 0x73, 0x2e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x64,
	0x0a, 0x19, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x0e, 0x67,
	0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02, 0x20, 0x00, 0x52, 0x0c, 0x67, 0x69,
	0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e,
	0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x6e, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x22, 0x62, 0x0a, 0x19, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x43,
	0x68, 0x61, 0x72, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x2d, 0x0a, 0x0e, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x22, 0x02,
	0x20, 0x00, 0x52, 0x0c, 0x67, 0x69, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x22, 0x32, 0x0a, 0x1a, 0x44, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x43, 0x68, 0x61, 0x72, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x1c, 0x0a, 0x1a,
	0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xeb, 0x06, 0x0a, 0x09, 0x47,
	0x69, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x86, 0x01, 0x0a, 0x04, 0x53, 0x68, 0x6f,
	0x77, 0x12, 0x16, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x53, 0x68,
	0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x67, 0x69, 0x74, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x4d, 0xba, 0x47, 0x14, 0x12, 0x12, 0xe6, 0x9f, 0xa5, 0xe7, 0x9c, 0x8b, 0xe9,
	0xa1, 0xb9, 0xe7, 0x9b, 0xae, 0xe9, 0x85, 0x8d, 0xe7, 0xbd, 0xae, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x30, 0x12, 0x2e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x69, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0xae, 0x01, 0x0a, 0x0c, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x1e, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47,
	0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47,
	0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x5d, 0xba, 0x47, 0x22, 0x12, 0x20, 0xe6, 0x9f, 0xa5, 0xe7, 0x9c, 0x8b,
	0xe9, 0xa1, 0xb9, 0xe7, 0x9b, 0xae, 0x20, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x20, 0xe9, 0x85, 0x8d, 0xe7, 0xbd, 0xae, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x32,
	0x12, 0x30, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x69, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x73, 0x2f, 0x7b, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x5f, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0xbc, 0x01, 0x0a, 0x12, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x47, 0x6c, 0x6f,
	0x62, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x24, 0x2e, 0x67, 0x69, 0x74, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x54, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x47, 0x6c, 0x6f, 0x62,
	0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x25, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x54, 0x6f, 0x67, 0x67,
	0x6c, 0x65, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x59, 0xba, 0x47, 0x1b, 0x12, 0x19, 0xe5, 0xbc, 0x80,
	0xe5, 0x90, 0xaf, 0x2f, 0xe5, 0x85, 0xb3, 0xe9, 0x97, 0xad, 0xe5, 0x85, 0xa8, 0xe5, 0xb1, 0x80,
	0xe9, 0x85, 0x8d, 0xe7, 0xbd, 0xae, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x35, 0x3a, 0x01, 0x2a, 0x22,
	0x30, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x69, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x2f, 0x7b, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f,
	0x69, 0x64, 0x7d, 0x2f, 0x74, 0x6f, 0x67, 0x67, 0x6c, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x8f, 0x01, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x18, 0x2e, 0x67,
	0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x50, 0xba, 0x47, 0x14, 0x12, 0x12, 0xe6, 0x9b, 0xb4, 0xe6, 0x96, 0xb0, 0xe5, 0x85,
	0xa8, 0xe5, 0xb1, 0x80, 0xe9, 0x85, 0x8d, 0xe7, 0xbd, 0xae, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x33,
	0x3a, 0x01, 0x2a, 0x1a, 0x2e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x69, 0x74, 0x2f, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x6d, 0x61, 0x72, 0x73, 0x5f, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0xd2, 0x01, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x44, 0x65, 0x66, 0x61, 0x75,
	0x6c, 0x74, 0x43, 0x68, 0x61, 0x72, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12, 0x24, 0x2e,
	0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x43, 0x68, 0x61, 0x72, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e,
	0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x43, 0x68, 0x61, 0x72, 0x74, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x6c, 0xba, 0x47, 0x30, 0x12,
	0x2e, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0xe9, 0xa1, 0xb9, 0xe7, 0x9b, 0xae, 0x20, 0x68, 0x65,
	0x6c, 0x6d, 0x20, 0x63, 0x68, 0x61, 0x72, 0x74, 0x73, 0x20, 0xe7, 0x9a, 0x84, 0xe9, 0xbb, 0x98,
	0xe8, 0xae, 0xa4, 0x20, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x2e, 0x79, 0x61, 0x6d, 0x6c, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x33, 0x12, 0x31, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x69, 0x74, 0x2f,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x67, 0x69, 0x74, 0x5f, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x75, 0x63, 0x2d, 0x63, 0x6e, 0x7a, 0x6a, 0x2f,
	0x6d, 0x61, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x34, 0x2f, 0x67, 0x69, 0x74, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x3b, 0x67, 0x69, 0x74, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gitconfig_gitconfig_proto_rawDescOnce sync.Once
	file_gitconfig_gitconfig_proto_rawDescData = file_gitconfig_gitconfig_proto_rawDesc
)

func file_gitconfig_gitconfig_proto_rawDescGZIP() []byte {
	file_gitconfig_gitconfig_proto_rawDescOnce.Do(func() {
		file_gitconfig_gitconfig_proto_rawDescData = protoimpl.X.CompressGZIP(file_gitconfig_gitconfig_proto_rawDescData)
	})
	return file_gitconfig_gitconfig_proto_rawDescData
}

var file_gitconfig_gitconfig_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_gitconfig_gitconfig_proto_goTypes = []interface{}{
	(*FileRequest)(nil),                // 0: gitconfig.FileRequest
	(*FileResponse)(nil),               // 1: gitconfig.FileResponse
	(*ShowRequest)(nil),                // 2: gitconfig.ShowRequest
	(*ShowResponse)(nil),               // 3: gitconfig.ShowResponse
	(*GlobalConfigRequest)(nil),        // 4: gitconfig.GlobalConfigRequest
	(*GlobalConfigResponse)(nil),       // 5: gitconfig.GlobalConfigResponse
	(*UpdateRequest)(nil),              // 6: gitconfig.UpdateRequest
	(*UpdateResponse)(nil),             // 7: gitconfig.UpdateResponse
	(*ToggleGlobalStatusRequest)(nil),  // 8: gitconfig.ToggleGlobalStatusRequest
	(*DefaultChartValuesRequest)(nil),  // 9: gitconfig.DefaultChartValuesRequest
	(*DefaultChartValuesResponse)(nil), // 10: gitconfig.DefaultChartValuesResponse
	(*ToggleGlobalStatusResponse)(nil), // 11: gitconfig.ToggleGlobalStatusResponse
	(*mars.Element)(nil),               // 12: mars.Element
	(*mars.Config)(nil),                // 13: mars.Config
}
var file_gitconfig_gitconfig_proto_depIdxs = []int32{
	12, // 0: gitconfig.FileResponse.elements:type_name -> mars.Element
	13, // 1: gitconfig.ShowResponse.config:type_name -> mars.Config
	13, // 2: gitconfig.GlobalConfigResponse.config:type_name -> mars.Config
	13, // 3: gitconfig.UpdateRequest.config:type_name -> mars.Config
	13, // 4: gitconfig.UpdateResponse.config:type_name -> mars.Config
	2,  // 5: gitconfig.GitConfig.Show:input_type -> gitconfig.ShowRequest
	4,  // 6: gitconfig.GitConfig.GlobalConfig:input_type -> gitconfig.GlobalConfigRequest
	8,  // 7: gitconfig.GitConfig.ToggleGlobalStatus:input_type -> gitconfig.ToggleGlobalStatusRequest
	6,  // 8: gitconfig.GitConfig.Update:input_type -> gitconfig.UpdateRequest
	9,  // 9: gitconfig.GitConfig.GetDefaultChartValues:input_type -> gitconfig.DefaultChartValuesRequest
	3,  // 10: gitconfig.GitConfig.Show:output_type -> gitconfig.ShowResponse
	5,  // 11: gitconfig.GitConfig.GlobalConfig:output_type -> gitconfig.GlobalConfigResponse
	11, // 12: gitconfig.GitConfig.ToggleGlobalStatus:output_type -> gitconfig.ToggleGlobalStatusResponse
	7,  // 13: gitconfig.GitConfig.Update:output_type -> gitconfig.UpdateResponse
	10, // 14: gitconfig.GitConfig.GetDefaultChartValues:output_type -> gitconfig.DefaultChartValuesResponse
	10, // [10:15] is the sub-list for method output_type
	5,  // [5:10] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_gitconfig_gitconfig_proto_init() }
func file_gitconfig_gitconfig_proto_init() {
	if File_gitconfig_gitconfig_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gitconfig_gitconfig_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileRequest); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileResponse); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_gitconfig_gitconfig_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_gitconfig_gitconfig_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GlobalConfigRequest); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GlobalConfigResponse); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
		file_gitconfig_gitconfig_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
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
		file_gitconfig_gitconfig_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToggleGlobalStatusRequest); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DefaultChartValuesRequest); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DefaultChartValuesResponse); i {
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
		file_gitconfig_gitconfig_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ToggleGlobalStatusResponse); i {
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
			RawDescriptor: file_gitconfig_gitconfig_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_gitconfig_gitconfig_proto_goTypes,
		DependencyIndexes: file_gitconfig_gitconfig_proto_depIdxs,
		MessageInfos:      file_gitconfig_gitconfig_proto_msgTypes,
	}.Build()
	File_gitconfig_gitconfig_proto = out.File
	file_gitconfig_gitconfig_proto_rawDesc = nil
	file_gitconfig_gitconfig_proto_goTypes = nil
	file_gitconfig_gitconfig_proto_depIdxs = nil
}

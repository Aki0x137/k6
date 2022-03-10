// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: distributed.proto

package distributed

import (
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

type RegisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_distributed_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_distributed_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_distributed_proto_rawDescGZIP(), []int{0}
}

type RegisterResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceID uint32 `protobuf:"varint,1,opt,name=instanceID,proto3" json:"instanceID,omitempty"`
	Archive    []byte `protobuf:"bytes,2,opt,name=archive,proto3" json:"archive,omitempty"` // TODO: send this with a `stream` of bytes chunks
	Options    []byte `protobuf:"bytes,3,opt,name=options,proto3" json:"options,omitempty"`
}

func (x *RegisterResponse) Reset() {
	*x = RegisterResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_distributed_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterResponse) ProtoMessage() {}

func (x *RegisterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_distributed_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterResponse.ProtoReflect.Descriptor instead.
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return file_distributed_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterResponse) GetInstanceID() uint32 {
	if x != nil {
		return x.InstanceID
	}
	return 0
}

func (x *RegisterResponse) GetArchive() []byte {
	if x != nil {
		return x.Archive
	}
	return nil
}

func (x *RegisterResponse) GetOptions() []byte {
	if x != nil {
		return x.Options
	}
	return nil
}

type AgentMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TODO: actually use random session IDs to prevent spoofing
	//
	// Types that are assignable to Message:
	//	*AgentMessage_InitInstanceID
	//	*AgentMessage_SignalAndWaitOnID
	//	*AgentMessage_GetOrCreateDataWithID
	//	*AgentMessage_CreatedData
	Message isAgentMessage_Message `protobuf_oneof:"Message"`
}

func (x *AgentMessage) Reset() {
	*x = AgentMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_distributed_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AgentMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgentMessage) ProtoMessage() {}

func (x *AgentMessage) ProtoReflect() protoreflect.Message {
	mi := &file_distributed_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AgentMessage.ProtoReflect.Descriptor instead.
func (*AgentMessage) Descriptor() ([]byte, []int) {
	return file_distributed_proto_rawDescGZIP(), []int{2}
}

func (m *AgentMessage) GetMessage() isAgentMessage_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (x *AgentMessage) GetInitInstanceID() uint32 {
	if x, ok := x.GetMessage().(*AgentMessage_InitInstanceID); ok {
		return x.InitInstanceID
	}
	return 0
}

func (x *AgentMessage) GetSignalAndWaitOnID() string {
	if x, ok := x.GetMessage().(*AgentMessage_SignalAndWaitOnID); ok {
		return x.SignalAndWaitOnID
	}
	return ""
}

func (x *AgentMessage) GetGetOrCreateDataWithID() string {
	if x, ok := x.GetMessage().(*AgentMessage_GetOrCreateDataWithID); ok {
		return x.GetOrCreateDataWithID
	}
	return ""
}

func (x *AgentMessage) GetCreatedData() *DataPacket {
	if x, ok := x.GetMessage().(*AgentMessage_CreatedData); ok {
		return x.CreatedData
	}
	return nil
}

type isAgentMessage_Message interface {
	isAgentMessage_Message()
}

type AgentMessage_InitInstanceID struct {
	InitInstanceID uint32 `protobuf:"varint,1,opt,name=initInstanceID,proto3,oneof"`
}

type AgentMessage_SignalAndWaitOnID struct {
	SignalAndWaitOnID string `protobuf:"bytes,2,opt,name=signalAndWaitOnID,proto3,oneof"`
}

type AgentMessage_GetOrCreateDataWithID struct {
	GetOrCreateDataWithID string `protobuf:"bytes,3,opt,name=getOrCreateDataWithID,proto3,oneof"`
}

type AgentMessage_CreatedData struct {
	CreatedData *DataPacket `protobuf:"bytes,4,opt,name=createdData,proto3,oneof"`
}

func (*AgentMessage_InitInstanceID) isAgentMessage_Message() {}

func (*AgentMessage_SignalAndWaitOnID) isAgentMessage_Message() {}

func (*AgentMessage_GetOrCreateDataWithID) isAgentMessage_Message() {}

func (*AgentMessage_CreatedData) isAgentMessage_Message() {}

type ControllerMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceID uint32 `protobuf:"varint,1,opt,name=instanceID,proto3" json:"instanceID,omitempty"`
	// Types that are assignable to Message:
	//	*ControllerMessage_DoneWaitWithID
	//	*ControllerMessage_CreateDataWithID
	//	*ControllerMessage_DataWithID
	Message isControllerMessage_Message `protobuf_oneof:"Message"`
}

func (x *ControllerMessage) Reset() {
	*x = ControllerMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_distributed_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ControllerMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ControllerMessage) ProtoMessage() {}

func (x *ControllerMessage) ProtoReflect() protoreflect.Message {
	mi := &file_distributed_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ControllerMessage.ProtoReflect.Descriptor instead.
func (*ControllerMessage) Descriptor() ([]byte, []int) {
	return file_distributed_proto_rawDescGZIP(), []int{3}
}

func (x *ControllerMessage) GetInstanceID() uint32 {
	if x != nil {
		return x.InstanceID
	}
	return 0
}

func (m *ControllerMessage) GetMessage() isControllerMessage_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (x *ControllerMessage) GetDoneWaitWithID() string {
	if x, ok := x.GetMessage().(*ControllerMessage_DoneWaitWithID); ok {
		return x.DoneWaitWithID
	}
	return ""
}

func (x *ControllerMessage) GetCreateDataWithID() string {
	if x, ok := x.GetMessage().(*ControllerMessage_CreateDataWithID); ok {
		return x.CreateDataWithID
	}
	return ""
}

func (x *ControllerMessage) GetDataWithID() *DataPacket {
	if x, ok := x.GetMessage().(*ControllerMessage_DataWithID); ok {
		return x.DataWithID
	}
	return nil
}

type isControllerMessage_Message interface {
	isControllerMessage_Message()
}

type ControllerMessage_DoneWaitWithID struct {
	DoneWaitWithID string `protobuf:"bytes,2,opt,name=doneWaitWithID,proto3,oneof"`
}

type ControllerMessage_CreateDataWithID struct {
	CreateDataWithID string `protobuf:"bytes,3,opt,name=createDataWithID,proto3,oneof"`
}

type ControllerMessage_DataWithID struct {
	DataWithID *DataPacket `protobuf:"bytes,4,opt,name=dataWithID,proto3,oneof"`
}

func (*ControllerMessage_DoneWaitWithID) isControllerMessage_Message() {}

func (*ControllerMessage_CreateDataWithID) isControllerMessage_Message() {}

func (*ControllerMessage_DataWithID) isControllerMessage_Message() {}

type DataPacket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data  []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Error string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *DataPacket) Reset() {
	*x = DataPacket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_distributed_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataPacket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataPacket) ProtoMessage() {}

func (x *DataPacket) ProtoReflect() protoreflect.Message {
	mi := &file_distributed_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataPacket.ProtoReflect.Descriptor instead.
func (*DataPacket) Descriptor() ([]byte, []int) {
	return file_distributed_proto_rawDescGZIP(), []int{4}
}

func (x *DataPacket) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DataPacket) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *DataPacket) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_distributed_proto protoreflect.FileDescriptor

var file_distributed_proto_rawDesc = []byte{
	0x0a, 0x11, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64,
	0x22, 0x11, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x66, 0x0a, 0x10, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x72, 0x63, 0x68, 0x69,
	0x76, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x61, 0x72, 0x63, 0x68, 0x69, 0x76,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0xe8, 0x01, 0x0a, 0x0c,
	0x41, 0x67, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x28, 0x0a, 0x0e,
	0x69, 0x6e, 0x69, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x48, 0x00, 0x52, 0x0e, 0x69, 0x6e, 0x69, 0x74, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x12, 0x2e, 0x0a, 0x11, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x41, 0x6e, 0x64, 0x57, 0x61, 0x69, 0x74, 0x4f, 0x6e, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x11, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x41, 0x6e, 0x64, 0x57, 0x61,
	0x69, 0x74, 0x4f, 0x6e, 0x49, 0x44, 0x12, 0x36, 0x0a, 0x15, 0x67, 0x65, 0x74, 0x4f, 0x72, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x15, 0x67, 0x65, 0x74, 0x4f, 0x72, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x12, 0x3b,
	0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x64, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x48, 0x00, 0x52, 0x0b,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x42, 0x09, 0x0a, 0x07, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xd1, 0x01, 0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x44, 0x12, 0x28, 0x0a, 0x0e,
	0x64, 0x6f, 0x6e, 0x65, 0x57, 0x61, 0x69, 0x74, 0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0e, 0x64, 0x6f, 0x6e, 0x65, 0x57, 0x61, 0x69, 0x74,
	0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x12, 0x2c, 0x0a, 0x10, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x10, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x57, 0x69,
	0x74, 0x68, 0x49, 0x44, 0x12, 0x39, 0x0a, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x57, 0x69, 0x74, 0x68,
	0x49, 0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72,
	0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x50, 0x61, 0x63, 0x6b, 0x65,
	0x74, 0x48, 0x00, 0x52, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x57, 0x69, 0x74, 0x68, 0x49, 0x44, 0x42,
	0x09, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x46, 0x0a, 0x0a, 0x44, 0x61,
	0x74, 0x61, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x14, 0x0a, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x32, 0xb2, 0x01, 0x0a, 0x0f, 0x44, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x65, 0x64, 0x54, 0x65, 0x73, 0x74, 0x12, 0x49, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x12, 0x1c, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64,
	0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1d, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x2e, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x54, 0x0a, 0x11, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x41, 0x6e, 0x64, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x19, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62,
	0x75, 0x74, 0x65, 0x64, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x1a, 0x1e, 0x2e, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x6f, 0x2e, 0x6b, 0x36,
	0x2e, 0x69, 0x6f, 0x2f, 0x6b, 0x36, 0x2f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x64, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_distributed_proto_rawDescOnce sync.Once
	file_distributed_proto_rawDescData = file_distributed_proto_rawDesc
)

func file_distributed_proto_rawDescGZIP() []byte {
	file_distributed_proto_rawDescOnce.Do(func() {
		file_distributed_proto_rawDescData = protoimpl.X.CompressGZIP(file_distributed_proto_rawDescData)
	})
	return file_distributed_proto_rawDescData
}

var file_distributed_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_distributed_proto_goTypes = []interface{}{
	(*RegisterRequest)(nil),   // 0: distributed.RegisterRequest
	(*RegisterResponse)(nil),  // 1: distributed.RegisterResponse
	(*AgentMessage)(nil),      // 2: distributed.AgentMessage
	(*ControllerMessage)(nil), // 3: distributed.ControllerMessage
	(*DataPacket)(nil),        // 4: distributed.DataPacket
}
var file_distributed_proto_depIdxs = []int32{
	4, // 0: distributed.AgentMessage.createdData:type_name -> distributed.DataPacket
	4, // 1: distributed.ControllerMessage.dataWithID:type_name -> distributed.DataPacket
	0, // 2: distributed.DistributedTest.Register:input_type -> distributed.RegisterRequest
	2, // 3: distributed.DistributedTest.CommandAndControl:input_type -> distributed.AgentMessage
	1, // 4: distributed.DistributedTest.Register:output_type -> distributed.RegisterResponse
	3, // 5: distributed.DistributedTest.CommandAndControl:output_type -> distributed.ControllerMessage
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_distributed_proto_init() }
func file_distributed_proto_init() {
	if File_distributed_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_distributed_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterRequest); i {
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
		file_distributed_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterResponse); i {
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
		file_distributed_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AgentMessage); i {
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
		file_distributed_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ControllerMessage); i {
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
		file_distributed_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataPacket); i {
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
	file_distributed_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*AgentMessage_InitInstanceID)(nil),
		(*AgentMessage_SignalAndWaitOnID)(nil),
		(*AgentMessage_GetOrCreateDataWithID)(nil),
		(*AgentMessage_CreatedData)(nil),
	}
	file_distributed_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*ControllerMessage_DoneWaitWithID)(nil),
		(*ControllerMessage_CreateDataWithID)(nil),
		(*ControllerMessage_DataWithID)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_distributed_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_distributed_proto_goTypes,
		DependencyIndexes: file_distributed_proto_depIdxs,
		MessageInfos:      file_distributed_proto_msgTypes,
	}.Build()
	File_distributed_proto = out.File
	file_distributed_proto_rawDesc = nil
	file_distributed_proto_goTypes = nil
	file_distributed_proto_depIdxs = nil
}

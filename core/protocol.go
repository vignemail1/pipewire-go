// Package core - protocol.go
// PipeWire protocol message types and constants

package core

// PipeWireProtocolVersion is the current protocol version
const PipeWireProtocolVersion = 3

// ObjectType represents the type of PipeWire object
type ObjectType string

const (
	ObjectTypeCore     ObjectType = "Core"
	ObjectTypeRegistry ObjectType = "Registry"
	ObjectTypeNode     ObjectType = "Node"
	ObjectTypePort     ObjectType = "Port"
	ObjectTypeLink     ObjectType = "Link"
	ObjectTypeClient   ObjectType = "Client"
	ObjectTypeModule   ObjectType = "Module"
	ObjectTypeFactory  ObjectType = "Factory"
)

// MethodID represents method IDs for protocol messages
type MethodID uint32

const (
	// Core methods
	CoreMethodAddListener    MethodID = 0
	CoreMethodRemoveListener MethodID = 1
	CoreMethodHelloV0        MethodID = 2
	CoreMethodSync           MethodID = 3
	CoreMethodPing           MethodID = 4
	CoreMethodPong           MethodID = 5
	CoreMethodError          MethodID = 6
	CoreMethodRemoveObject   MethodID = 7
	CoreMethodUpdateTypes    MethodID = 8
	CoreMethodUpdateProps    MethodID = 9
	CoreMethodBound          MethodID = 10
	CoreMethodAddMem         MethodID = 11
	CoreMethodRemoveMem      MethodID = 12
	CoreMethodBoundMeta      MethodID = 13
	CoreMethodAddPool        MethodID = 14
	CoreMethodRemovePool     MethodID = 15
	CoreMethodInitialized    MethodID = 16

	// Registry methods
	RegistryMethodAddListener MethodID = 0
	RegistryMethodRemoveListener MethodID = 1
	RegistryMethodGlobal      MethodID = 2
	RegistryMethodGlobalRemove MethodID = 3
	RegistryMethodBind        MethodID = 4

	// Node methods
	NodeMethodAddListener    MethodID = 0
	NodeMethodRemoveListener MethodID = 1
	NodeMethodSetParam       MethodID = 2
	NodeMethodSetIOConf      MethodID = 3
)

// EventType represents event types from server
type EventType uint32

const (
	EventTypeHello         EventType = 0
	EventTypeAddMem        EventType = 1
	EventTypeRemoveMem     EventType = 2
	EventTypeSync          EventType = 3
	EventTypePing          EventType = 4
	EventTypePong          EventType = 5
	EventTypeError         EventType = 6
	EventTypeRemoveObject  EventType = 7
	EventTypeUpdateTypes   EventType = 8
	EventTypeUpdateProps   EventType = 9
	EventTypeBound         EventType = 10
	EventTypeAddPool       EventType = 11
	EventTypeRemovePool    EventType = 12
	EventTypeBoundMeta     EventType = 13
	EventTypeInitialized   EventType = 14
)

// RegistryEventType represents registry-specific events
type RegistryEventType uint32

const (
	RegistryEventTypeGlobal       RegistryEventType = 0
	RegistryEventTypeGlobalRemove RegistryEventType = 1
)

// NodeEventType represents node-specific events
type NodeEventType uint32

const (
	NodeEventTypeInfo       NodeEventType = 0
	NodeEventTypeParam      NodeEventType = 1
	NodeEventTypeStateChanged NodeEventType = 2
	NodeEventTypeIOConf     NodeEventType = 3
)

// PortEventType represents port-specific events
type PortEventType uint32

const (
	PortEventTypeInfo       PortEventType = 0
	PortEventTypeParam      PortEventType = 1
	PortEventTypeStateChanged PortEventType = 2
)

// LinkEventType represents link-specific events
type LinkEventType uint32

const (
	LinkEventTypeInfo       LinkEventType = 0
	LinkEventTypeStateChanged LinkEventType = 1
)

// Message represents a PipeWire protocol message
type Message struct {
	ObjectID   uint32         // Target object ID
	Opcode     uint32         // Method/event ID
	Sequence   uint32         // Sequence number for sync
	Args       []interface{}  // Message arguments
	PODData    []byte         // Raw POD data
}

// String returns a human-readable message representation
func (m *Message) String() string {
	return "Message{ObjectID:" + string(rune(m.ObjectID)) + "}"
}

// RequestContext holds information about a request waiting for response
type RequestContext struct {
	Sequence  uint32
	Result    chan interface{}
	Error     chan error
	Timeout   chan struct{}
}

// ProtocolState represents the state of the protocol negotiation
type ProtocolState uint32

const (
	ProtocolStateInitial    ProtocolState = 0
	ProtocolStateHelloSent  ProtocolState = 1
	ProtocolStateHelloRecv  ProtocolState = 2
	ProtocolStateReady      ProtocolState = 3
	ProtocolStateClosed     ProtocolState = 4
)

// ProtocolFeature represents optional protocol features
type ProtocolFeature uint32

const (
	ProtocolFeatureBindObjectID ProtocolFeature = 1 << iota
	ProtocolFeatureMonitor
	ProtocolFeatureFairqueue
	ProtocolFeatureRtKits
)

// MemType represents different memory types
type MemType uint32

const (
	MemTypeShared  MemType = 0 // Shared memory
	MemTypeDMABuf  MemType = 1 // DMA buffer
	MemTypeMMAP    MemType = 2 // MMAP
	MemTypeDMAHeap MemType = 3 // DMA heap
)

// Permission represents access permissions
type Permission uint32

const (
	PermissionRead  Permission = 1 << 0
	PermissionWrite Permission = 1 << 1
	PermissionX     Permission = 1 << 2
)

// ParamDirection represents parameter direction
type ParamDirection uint32

const (
	ParamDirectionInput  ParamDirection = 0
	ParamDirectionOutput ParamDirection = 1
	ParamDirectionBoth   ParamDirection = 2
)

// ParameterType represents different parameter types in POD format
type ParameterType uint32

const (
	ParamTypeNone       ParameterType = 0
	ParamTypePropsType  ParameterType = 1
	ParamTypePropsMask  ParameterType = 2
	ParamTypeFormat     ParameterType = 3
	ParamTypeBuffers    ParameterType = 4
	ParamTypeMeta       ParameterType = 5
	ParamTypeIO         ParameterType = 6
	ParamTypeEnumFormat ParameterType = 7
	ParamTypeEnumProfile ParameterType = 8
	ParamTypeEnumRoute  ParameterType = 9
	ParamTypeLatency    ParameterType = 10
	ParamTypeProcessLatency ParameterType = 11
)

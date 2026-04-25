package guard

type Capability uint32

const (
	CapReadFiles Capability = 1 << iota
	CapWriteFiles
	CapExecProcess
	CapNetworkOutbound
	CapNetworkInbound
	CapReadAuditTrail
	CapWriteKnowledgeMesh
	CapManageNodes
	CapFleetAdmin
)

type Role uint32

const (
	RoleObserver Role = Role(CapReadFiles | CapReadAuditTrail)
	RoleAgent    Role = Role(uint32(RoleObserver) | uint32(CapWriteFiles|CapWriteKnowledgeMesh|CapNetworkOutbound))
	RoleAdmin    Role = 0xFFFFFFFF
)

func (c Capability) String() string {
	switch c {
	case CapReadFiles: return "ReadFiles"
	case CapWriteFiles: return "WriteFiles"
	case CapExecProcess: return "ExecProcess"
	case CapNetworkOutbound: return "NetworkOutbound"
	case CapNetworkInbound: return "NetworkInbound"
	case CapReadAuditTrail: return "ReadAuditTrail"
	case CapWriteKnowledgeMesh: return "WriteKnowledgeMesh"
	case CapManageNodes: return "ManageNodes"
	case CapFleetAdmin: return "FleetAdmin"
	default: return "Unknown"
	}
}

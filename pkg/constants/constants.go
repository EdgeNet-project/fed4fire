package constants

// Default value for new deployments.
const (
	DefaultCpuRequest    = "0.01"
	DefaultMemoryRequest = "16Mi"
)

const (
	EdgeNetLabelCountryISO = "edge-net.io/country-iso"
	EdgeNetLabelLatitude   = "edge-net.io/lat"
	EdgeNetLabelLongitude  = "edge-net.io/lon"
)

const (
	HttpHeaderCertificate = "X-Fed4Fire-Certificate"
	HttpHeaderUser        = "X-Fed4Fire-User"
)

// Error messages specific to this AM.
const (
	ErrorBadAction        = "Unsupported action"
	ErrorBadCredentials   = "Invalid credentials"
	ErrorBadTime          = "Failed to parse time"
	ErrorBadIdentifier    = "Failed to parse identifier"
	ErrorBuildResources   = "Failed to build resources"
	ErrorCreateResource   = "Failed to create resource"
	ErrorDeleteResource   = "Failed to delete resource"
	ErrorGetResource      = "Failed to get resource"
	ErrorListResources    = "Failed to list resources"
	ErrorUpdateResource   = "Failed to update resource"
	ErrorSerializeRspec   = "Failed to serialize rspec"
	ErrorDeserializeRspec = "Failed to deserialize rspec"
)

// Names for Kubernetes objects labels and annotations.
const (
	Fed4FireClientId   = "fed4fire.eu/client-id"
	Fed4FireExpires    = "fed4fire.eu/expires"
	Fed4FireSlice      = "fed4fire.eu/slice"
	Fed4FireSliceHash  = "fed4fire.eu/slice-hash"
	Fed4FireSliver     = "fed4fire.eu/sliver"
	Fed4FireSliverName = "fed4fire.eu/sliver-name"
	Fed4FireUser       = "fed4fire.eu/user"
)

const (
	GeniActionStart = "geni_start"
)

// https://groups.geni.net/geni/attachment/wiki/GAPI_AM_API_V3/CommonConcepts/geni-error-codes.xml
const (
	// Success
	GeniCodeSuccess = 0
	// Bad Arguments: malformed
	GeniCodeBadargs = 1
	// Error (other)
	GeniCodeError = 2
	// Operation Forbidden: eg supplied credentials do not provide sufficient privileges (on the given slice)
	GeniCodeForbidden = 3
	// Bad Version (eg of RSpec)
	GeniCodeBadversion = 4
	// Server Error
	GeniCodeServerror = 5
	// Too Big (eg request RSpec)
	GeniCodeToobig = 6
	// Operation Refused
	GeniCodeRefused = 7
	// Operation Timed Out
	GeniCodeTimedout = 8
	// Database Error
	GeniCodeDberror = 9
	// RPC Error
	GeniCodeRpcerror = 10
	// Unavailable (eg server in lockdown)
	GeniCodeUnavailable = 11
	// Search Failed (eg for slice)
	GeniCodeSearchfailed = 12
	// Operation Unsupported
	GeniCodeUnsupported = 13
	// Busy (resource, slice, or server); try again later
	GeniCodeBusy = 14
	// Expired (eg slice)
	GeniCodeExpired = 15
	// In Progress
	GeniCodeInprogress = 16
	// Already Exists (eg slice)
	GeniCodeAlreadyexists = 17
)

// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3/CommonConcepts#SliverAllocationStates
const (
	// The sliver does not exist. This is the small black circle in typical state diagrams.
	GeniStateUnallocated = "geni_unallocated"
	// The sliver exists, defines particular resources, and is in a slice.
	// The aggregate has not (if possible) done any time consuming or expensive work to instantiate the resources,
	// provision them, or make it difficult to revert the slice to the state prior to allocating this sliver.
	// This state is what the aggregate is offering the experimenter.
	GeniStateAllocated = "geni_allocated"
	// The aggregate has started instantiating resources, and otherwise making changes to resources
	// and the slice to make the resources available to the experimenter.
	// At this point, operational states are valid to specify further when
	// the resources are available for experimenter use.
	GeniStateProvisioned = "geni_provisioned"
)

// geni_operational_state
const (
	// Required for aggregates to support. A wait state.
	// The sliver is still being allocated and provisioned, and other operational states are not yet valid.
	GeniStatePendingAllocation = "geni_pending_allocation"
	// A final state. The resource is not usable / accessible by the experimenter,
	// and requires explicit experimenter action before it is usable/accessible by the experimenter.
	GeniStateNotReady = "geni_notready"
	// A wait state. The resource is in process of changing to geni_ready,
	// and on success will do so without additional experimenter action.
	// For example, the resource may be powering on.
	GeniStateConfiguring = "geni_configuring"
	// A wait state. The resource is in process of changing to geni_notready,
	// and on success will do so without additional experimenter action.
	// For example, the resource may be powering off.
	GeniStateStopping = "geni_stopping"
	// A final state. The resource is usable/accessible by the experimenter, and ready for slice operations.
	GeniStateReady = "geni_ready"
	// A wait state. The resource is performing some operational action,
	// but remains accessible/usable by the experimenter.
	// Upon completion of the action, the resource will return to geni_ready.
	GeniStateReadyBusy = "geni_ready_busy"
	// A final state. Some operational action failed, rendering the resource unusable.
	// An administrator action, undefined by this API,
	// may be required to return the resource to another operational state.
	GeniStateFailed = "geni_failed"
)

const (
	// Performing multiple Allocates without a delete is an error condition because the aggregate
	// only supports a single sliver per slice or does not allow incrementally adding new slivers.
	GeniAllocateSingle = "geni_single"
	// Additional calls to Allocate must be disjoint from slivers allocated with previous calls
	// (no references or dependencies on existing slivers).
	// The topologies must be disjoint in that there can be no connection or other reference
	// from one topology to the other.
	GeniAllocateDisjoint = "geni_disjoint"
	// Multiple slivers can exist and be incrementally added, including those which connect or overlap in some way.
	GeniAllocateMany = "geny_many"
)

const (
	// https://groups.geni.net/geni/wiki/TIEDABACCredential
	GeniCredentialTypeAbac = "geni_abac"
	// https://groups.geni.net/geni/wiki/GeniApiCredentials
	GeniCredentialTypeSfa = "geni_sfa"
)

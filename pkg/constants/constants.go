package constants

// Default value for new deployments.
const (
	DefaultCpuRequest    = "0.01"
	DefaultMemoryRequest = "16Mi"
)

const (
	HttpHeaderCertificate = "X-Fed4Fire-Certificate"
	HttpHeaderUser        = "X-Fed4Fire-User"
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

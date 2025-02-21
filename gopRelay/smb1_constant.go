package gopRelay

/*
 *
 * Retrieved from https://github.com/miketeo/pysmb/blob/dev-1.2.x/python3/smb/smb_constants.py
 *
 */

// Values for Command field in SMB message header
const SMB_COM_CREATE_DIRECTORY = 0x00
const SMB_COM_DELETE_DIRECTORY = 0x01
const SMB_COM_CLOSE = 0x04
const SMB_COM_DELETE = 0x06
const SMB_COM_RENAME = 0x07
const SMB_COM_TRANSACTION = 0x25
const SMB_COM_ECHO = 0x2B
const SMB_COM_OPEN_ANDX = 0x2D
const SMB_COM_READ_ANDX = 0x2E
const SMB_COM_WRITE_ANDX = 0x2F
const SMB_COM_TRANSACTION2 = 0x32
const SMB_COM_NEGOTIATE = 0x72
const SMB_COM_SESSION_SETUP_ANDX = 0x73
const SMB_COM_TREE_CONNECT_ANDX = 0x75
const SMB_COM_NT_TRANSACT = 0xA0
const SMB_COM_NT_CREATE_ANDX = 0xA2

var SMB_COMMAND_NAMES = map[byte]string{
	0x00: "SMB_COM_CREATE_DIRECTORY",
	0x01: "SMB_COM_DELETE_DIRECTORY",
	0x04: "SMB_COM_CLOSE",
	0x06: "SMB_COM_DELETE",
	0x25: "SMB_COM_TRANSACTION",
	0x2B: "SMB_COM_ECHO",
	0x2D: "SMB_COM_OPEN_ANDX",
	0x2E: "SMB_COM_READ_ANDX",
	0x2F: "SMB_COM_WRITE_ANDX",
	0x32: "SMB_COM_TRANSACTION2",
	0x72: "SMB_COM_NEGOTIATE",
	0x73: "SMB_COM_SESSION_SETUP_ANDX",
	0x75: "SMB_COM_TREE_CONNECT_ANDX",
	0xA0: "SMB_COM_NT_TRANSACT",
	0xA2: "SMB_COM_NT_CREATE_ANDX",
}

// Bitmask for Flags field in SMB message header
const SMB_FLAGS_LOCK_AND_READ_OK = 0x01    // LANMAN1.0
const SMB_FLAGS_BUF_AVAIL = 0x02           // LANMAN1.0, Obsolete
const SMB_FLAGS_CASE_INSENSITIVE = 0x08    // LANMAN1.0, Obsolete
const SMB_FLAGS_CANONICALIZED_PATHS = 0x10 // LANMAN1.0, Obsolete
const SMB_FLAGS_OPLOCK = 0x20              // LANMAN1.0, Obsolete
const SMB_FLAGS_OPBATCH = 0x40             // LANMAN1.0, Obsolete
const SMB_FLAGS_REPLY = 0x80               // LANMAN1.0

// Bitmask for Flags2 field in SMB message header
const SMB_FLAGS2_LONG_NAMES = 0x0001             // LANMAN2.0
const SMB_FLAGS2_EAS = 0x0002                    // LANMAN1.2
const SMB_FLAGS2_SMB_SECURITY_SIGNATURE = 0x0004 // NT LANMAN
const SMB_FLAGS2_IS_LONG_NAME = 0x0040           // NT LANMAN
const SMB_FLAGS2_DFS = 0x1000                    // NT LANMAN
const SMB_FLAGS2_REPARSE_PATH = 0x0400           //
const SMB_FLAGS2_EXTENDED_SECURITY = 0x0800      //
const SMB_FLAGS2_PAGING_IO = 0x2000              // NT LANMAN
const SMB_FLAGS2_NT_STATUS = 0x4000              // NT LANMAN
const SMB_FLAGS2_UNICODE = 0x8000                // NT LANMAN

// Bitmask for Capabilities field in SMB_COM_SESSION_SETUP_ANDX response
// [MS-SMB]: 2.2.4.5.2.1 (Capabilities field)
const CAP_RAW_MODE = 0x01
const CAP_MPX_MODE = 0x02
const CAP_UNICODE = 0x04
const CAP_LARGE_FILES = 0x08
const CAP_NT_SMBS = 0x10
const CAP_RPC_REMOTE_APIS = 0x20
const CAP_STATUS32 = 0x40
const CAP_LEVEL_II_OPLOCKS = 0x80
const CAP_LOCK_AND_READ = 0x0100
const CAP_NT_FIND = 0x0200
const CAP_DFS = 0x1000
const CAP_INFOLEVEL_PASSTHRU = 0x2000
const CAP_LARGE_READX = 0x4000
const CAP_LARGE_WRITEX = 0x8000
const CAP_LWIO = 0x010000
const CAP_UNIX = 0x800000
const CAP_COMPRESSED = 0x02000000
const CAP_DYNAMIC_REAUTH = 0x20000000
const CAP_PERSISTENT_HANDLES = 0x40000000
const CAP_EXTENDED_SECURITY = 0x80000000

// Value for Action field in SMB_COM_SESSION_SETUP_ANDX response
const SMB_SETUP_GUEST = 0x0001
const SMB_SETUP_USE_LANMAN_KEY = 0x0002

// Bitmask for SecurityMode field in SMB_COM_NEGOTIATE response
const NEGOTIATE_USER_SECURITY = 0x01
const NEGOTIATE_ENCRYPT_PASSWORDS = 0x02
const NEGOTIATE_SECURITY_SIGNATURES_ENABLE = 0x04
const NEGOTIATE_SECURITY_SIGNATURES_REQUIRE = 0x08

// Available constants for Service field in SMB_COM_TREE_CONNECT_ANDX request
// [MS-CIFS]: 2.2.4.55.1 (Service field)
const SERVICE_PRINTER = "LPT1:"
const SERVICE_NAMED_PIPE = "IPC"
const SERVICE_COMM = "COMM"
const SERVICE_ANY = "?????"

// Bitmask for Flags field in SMB_COM_NT_CREATE_ANDX request
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB]: 2.2.4.9.1
const NT_CREATE_REQUEST_OPLOCK = 0x02
const NT_CREATE_REQUEST_OPBATCH = 0x04
const NT_CREATE_OPEN_TARGET_DIR = 0x08
const NT_CREATE_REQUEST_EXTENDED_RESPONSE = 0x10 // Defined in [MS-SMB]: 2.2.4.9.1

// Bitmask for DesiredAccess field in SMB_COM_NT_CREATE_ANDX request
// and SMB2CreateRequest class
// Also used for MaximalAccess field in SMB2TreeConnectResponse class
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB2]: 2.2.13.1.1
const FILE_READ_DATA = 0x01
const FILE_WRITE_DATA = 0x02
const FILE_APPEND_DATA = 0x04
const FILE_READ_EA = 0x08
const FILE_WRITE_EA = 0x10
const FILE_EXECUTE = 0x20
const FILE_DELETE_CHILD = 0x40
const FILE_READ_ATTRIBUTES = 0x80
const FILE_WRITE_ATTRIBUTES = 0x0100
const DELETE = 0x010000
const READ_CONTROL = 0x020000
const WRITE_DAC = 0x040000
const WRITE_OWNER = 0x080000
const SYNCHRONIZE = 0x100000
const ACCESS_SYSTEM_SECURITY = 0x01000000
const MAXIMUM_ALLOWED = 0x02000000
const GENERIC_ALL = 0x10000000
const GENERIC_EXECUTE = 0x20000000
const GENERIC_WRITE = 0x40000000
const GENERIC_READ = 0x80000000

// SMB_EXT_FILE_ATTR bitmask ([MS-CIFS]: 2.2.1.2.3)
// Includes extensions defined in [MS-SMB] 2.2.1.2.1
// Bitmask for FileAttributes field in SMB_COM_NT_CREATE_ANDX request ([MS-CIFS]: 2.2.4.64.1)
// Also used for FileAttributes field in SMB2CreateRequest class ([MS-SMB2]: 2.2.13)
const ATTR_READONLY = 0x01
const ATTR_HIDDEN = 0x02
const ATTR_SYSTEM = 0x04
const ATTR_DIRECTORY = 0x10
const ATTR_ARCHIVE = 0x20
const ATTR_NORMAL = 0x80
const ATTR_TEMPORARY = 0x0100
const ATTR_SPARSE = 0x0200
const ATTR_REPARSE_POINT = 0x0400
const ATTR_COMPRESSED = 0x0800
const ATTR_OFFLINE = 0x1000
const ATTR_NOT_CONTENT_INDEXED = 0x2000
const ATTR_ENCRYPTED = 0x4000
const POSIX_SEMANTICS = 0x01000000
const BACKUP_SEMANTICS = 0x02000000
const DELETE_ON_CLOSE = 0x04000000
const SEQUENTIAL_SCAN = 0x08000000
const RANDOM_ACCESS = 0x10000000
const NO_BUFFERING = 0x20000000
const WRITE_THROUGH = 0x80000000

// Bitmask for ShareAccess field in SMB_COM_NT_CREATE_ANDX request
// and SMB2CreateRequest class
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB2]: 2.2.13
const FILE_SHARE_NONE = 0x00
const FILE_SHARE_READ = 0x01
const FILE_SHARE_WRITE = 0x02
const FILE_SHARE_DELETE = 0x04

// Values for CreateDisposition field in SMB_COM_NT_CREATE_ANDX request
// and SMB2CreateRequest class
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB2]: 2.2.13
const FILE_SUPERSEDE = 0x00
const FILE_OPEN = 0x01
const FILE_CREATE = 0x02
const FILE_OPEN_IF = 0x03
const FILE_OVERWRITE = 0x04
const FILE_OVERWRITE_IF = 0x05

// Bitmask for CreateOptions field in SMB_COM_NT_CREATE_ANDX request
// and SMB2CreateRequest class
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB2]: 2.2.13
const FILE_DIRECTORY_FILE = 0x01
const FILE_WRITE_THROUGH = 0x02
const FILE_SEQUENTIAL_ONLY = 0x04
const FILE_NO_INTERMEDIATE_BUFFERING = 0x08
const FILE_SYNCHRONOUS_IO_ALERT = 0x10
const FILE_SYNCHRONOUS_IO_NONALERT = 0x20
const FILE_NON_DIRECTORY_FILE = 0x40
const FILE_CREATE_TREE_CONNECTION = 0x80
const FILE_COMPLETE_IF_OPLOCKED = 0x0100
const FILE_NO_EA_KNOWLEDGE = 0x0200
const FILE_OPEN_FOR_RECOVERY = 0x0400
const FILE_RANDOM_ACCESS = 0x0800
const FILE_DELETE_ON_CLOSE = 0x1000
const FILE_OPEN_BY_FILE_ID = 0x2000
const FILE_OPEN_FOR_BACKUP_INTENT = 0x4000
const FILE_NO_COMPRESSION = 0x8000
const FILE_RESERVE_OPFILTER = 0x100000
const FILE_OPEN_NO_RECALL = 0x400000
const FILE_OPEN_FOR_FREE_SPACE_QUERY = 0x800000

// Values for ImpersonationLevel field in SMB_COM_NT_CREATE_ANDX request
// and SMB2CreateRequest class
// For interpretations about these values, refer to [MS-WSO] and [MSDN-IMPERS]
// [MS-CIFS]: 2.2.4.64.1
// [MS-SMB]: 2.2.4.9.1
// [MS-SMB2]: 2.2.13
const SEC_ANONYMOUS = 0x00
const SEC_IDENTIFY = 0x01
const SEC_IMPERSONATE = 0x02
const SEC_DELEGATION = 0x03 // Defined in [MS-SMB]: 2.2.4.9.1

// Values for SecurityFlags field in SMB_COM_NT_CREATE_ANDX request
// [MS-CIFS]: 2.2.4.64.1
const SMB_SECURITY_CONTEXT_TRACKING = 0x01
const SMB_SECURITY_EFFECTIVE_ONLY = 0x02

// Bitmask for Flags field in SMB_COM_TRANSACTION2 request
// [MS-CIFS]: 2.2.4.46.1
const DISCONNECT_TID = 0x01
const NO_RESPONSE = 0x02

// Bitmask for basic file attributes
// [MS-CIFS]: 2.2.1.2.4
const SMB_FILE_ATTRIBUTE_NORMAL = 0x00
const SMB_FILE_ATTRIBUTE_READONLY = 0x01
const SMB_FILE_ATTRIBUTE_HIDDEN = 0x02
const SMB_FILE_ATTRIBUTE_SYSTEM = 0x04
const SMB_FILE_ATTRIBUTE_VOLUME = 0x08 // Unsupported for listPath() operations
const SMB_FILE_ATTRIBUTE_DIRECTORY = 0x10
const SMB_FILE_ATTRIBUTE_ARCHIVE = 0x20

// SMB_FILE_ATTRIBUTE_INCL_NORMAL is a special placeholder to include normal files
// with other search attributes for listPath() operations. It is not defined in the MS-CIFS specs.
const SMB_FILE_ATTRIBUTE_INCL_NORMAL = 0x10000

// Do not use the following values for listPath() operations as they are not supported for SMB2
const SMB_SEARCH_ATTRIBUTE_READONLY = 0x0100
const SMB_SEARCH_ATTRIBUTE_HIDDEN = 0x0200
const SMB_SEARCH_ATTRIBUTE_SYSTEM = 0x0400
const SMB_SEARCH_ATTRIBUTE_DIRECTORY = 0x1000
const SMB_SEARCH_ATTRIBUTE_ARCHIVE = 0x2000

// Bitmask for OptionalSupport field in SMB_COM_TREE_CONNECT_ANDX response
const SMB_TREE_CONNECTX_SUPPORT_SEARCH = 0x0001
const SMB_TREE_CONNECTX_SUPPORT_DFS = 0x0002

// Bitmask for security information fields, specified as
// AdditionalInformation in SMB2
// [MS-SMB]: 2.2.7.4
// [MS-SMB2]: 2.2.37
const OWNER_SECURITY_INFORMATION = 0x00000001
const GROUP_SECURITY_INFORMATION = 0x00000002
const DACL_SECURITY_INFORMATION = 0x00000004
const SACL_SECURITY_INFORMATION = 0x00000008
const LABEL_SECURITY_INFORMATION = 0x00000010
const ATTRIBUTE_SECURITY_INFORMATION = 0x00000020
const SCOPE_SECURITY_INFORMATION = 0x00000040
const BACKUP_SECURITY_INFORMATION = 0x00010000

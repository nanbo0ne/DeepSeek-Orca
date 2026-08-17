// Package product contains the canonical public identity for O.R.C.A.
package product

const (
	Name              = "O.R.C.A"
	WindowsName       = "O.R.C.A for Windows"
	FullName          = "Open Reasoning & Computing Agent"
	AssistantName     = "Orca"
	Repository        = "https://github.com/nanbo0ne/O.R.C.A-for-Windows"
	RepositorySlug    = "nanbo0ne/O.R.C.A-for-Windows"
	ConfigDirName     = "orca"
	ProjectConfigName = "orca.toml"
	ProjectStateDir   = ".orca"
	InstructionName   = "ORCA.md"
	EnvPrefix         = "ORCA_"
	SingleInstanceID  = "com.orca.windows"
)

// Legacy identifiers remain read-only migration aliases. They must never be
// presented as the current product identity.
const (
	LegacyConfigDirName     = "deepseek-orca"
	LegacyProjectConfigName = "deepseek-orca.toml"
	LegacyProjectStateDir   = ".deepseek-orca"
	LegacyInstructionName   = "DEEPSEEK_ORCA.md"
	LegacyEnvPrefix         = "DEEPSEEK_ORCA_"
	LegacySingleInstanceID  = "com.deepseek-orca.desktop"
)

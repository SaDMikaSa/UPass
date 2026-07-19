package common

var VaultMagic = []byte{0x55, 0x50, 0x41, 0x53}

const VaultVersion = 1

const (
	SaltSize         = 16
	Argon2Time       = 2
	Argon2Memory     = 64 * 1024
	MinStrengthScore = 3
)
const (
	MagicSize       = 4
	VersionSize     = 1
	ArgonTimeSize   = 1
	ArgonMemorySize = 4
	EmpLenSize      = 2
	BaseHeaderSize  = MagicSize + VersionSize + ArgonTimeSize + ArgonMemorySize + EmpLenSize // 12
)

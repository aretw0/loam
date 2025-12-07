package loam

// Option is a functional option for configuring a Vault.
type Option func(*Vault) error

// WithAutoInit enables automatic initialization of the vault directory and git repository
// if they do not exist.
func WithAutoInit(auto bool) Option {
	return func(v *Vault) error {
		// We store this config in the Vault struct (needs update)
		// For now we can handle logic in NewVault, but we need a place to store "isAutoInit" if needed later?
		// Actually, NewVault applies options *before* final setup.
		// Let's assume we add fields to Vault struct or handle it during apply.
		// Since 'Vault' struct is what we return, we might need an intermediate 'config' struct
		// OR we add these flags to Vault.
		v.autoInit = auto
		return nil
	}
}

// WithGitless forces the vault to operate without Git interactions.
// If false, it tries to auto-detect if git is available.
func WithGitless(gitless bool) Option {
	return func(v *Vault) error {
		v.isGitless = gitless
		return nil
	}
}

// WithTempDir forces the vault to use a temporary directory for safety.
// This is useful for tests and examples.
func WithTempDir() Option {
	return func(v *Vault) error {
		v.forceTemp = true
		return nil
	}
}

// WithMustExist enforces that the vault directory must already exist.
// It overrides AutoInit behavior.
func WithMustExist() Option {
	return func(v *Vault) error {
		v.mustExist = true
		return nil
	}
}

// Internal validation of options could go here if needed.

package providerindexstorage

type ProviderListNotFoundError struct {
	BaseError
}

func (p *ProviderListNotFoundError) Error() string {
	if p.Cause != nil {
		return "Provider list not found in storage (" + p.Cause.Error() + ")."
	}
	return "Provider list not found in storage."
}
